// Package execution defines the host-owned invocation boundary. Validation here
// checks shape and correlation, not authorization or artifact registration.
package execution

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"regexp"
	"strings"
	"time"
)

var (
	ErrInvalidRequest   = errors.New("invalid execution request")
	ErrInvalidResult    = errors.New("invalid execution result")
	ErrArtifactMismatch = errors.New("execution artifact digest mismatch")
	ErrUnavailable      = errors.New("execution capacity unavailable")
)

// Executor does not own worker lifecycle or durable application state.
type Executor interface {
	Execute(context.Context, Request) (Result, error)
}

type Artifact struct {
	ID, Version, Digest, Entrypoint string
}

// Verify checks the bytes of a complete immutable artifact, not its individual
// source files. The registry must define and pin the artifact serialization.
func (a Artifact) Verify(bundle []byte) error {
	if a.Digest != ArtifactDigest(bundle) {
		return ErrArtifactMismatch
	}
	return nil
}

// ArtifactDigest identifies canonical artifact bytes.
func ArtifactDigest(bundle []byte) string {
	sum := sha256.Sum256(bundle)
	return "sha256:" + hex.EncodeToString(sum[:])
}

// Runtime identifies a capability and isolation profile, not a claim that the
// runtime is implemented or available.
type Runtime struct {
	Kind, Profile string
}

type Limits struct {
	CPUTime                              time.Duration
	MemoryBytes, InputBytes, OutputBytes int64
}

// Authorization is constructed by Brain after admission, not by the caller or
// worker. It describes one admitted operation and is not a reusable credential.
type Authorization struct {
	ProjectID, OriginalCaller, Caller, Capability, Resource string
}

type Request struct {
	ID, TraceID, ParentID, ProjectID string
	Artifact                         Artifact
	Runtime                          Runtime
	Input                            json.RawMessage
	Context                          json.RawMessage
	Deadline                         time.Time
	Limits                           Limits
	Auth                             Authorization
	Idempotent                       bool
	IdempotencyKey                   string
}

// Version is an immutable registry identifier, never a selector. The registry
// must additionally resolve it and compare its digest before dispatch.
var exactVersion = regexp.MustCompile(`^[0-9]+(\.[0-9]+){0,2}(-[0-9A-Za-z.-]+)?(\+[0-9A-Za-z.-]+)?$`)
var digest = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)

func present(value string) bool {
	return strings.TrimSpace(value) != "" && len(value) <= 1024
}

func (r Request) Validate(now time.Time) error {
	if !present(r.ID) || !present(r.TraceID) || !present(r.ProjectID) || r.ParentID == r.ID || (r.ParentID != "" && !present(r.ParentID)) ||
		r.Auth.ProjectID != r.ProjectID || !present(r.Auth.OriginalCaller) || !present(r.Auth.Caller) || !present(r.Auth.Capability) ||
		!present(r.Artifact.ID) || len(r.Artifact.Version) > 128 || !exactVersion.MatchString(r.Artifact.Version) || !digest.MatchString(r.Artifact.Digest) || !present(r.Artifact.Entrypoint) ||
		!present(r.Runtime.Kind) || !present(r.Runtime.Profile) || r.Deadline.IsZero() || r.Limits.CPUTime <= 0 || r.Limits.MemoryBytes <= 0 || r.Limits.InputBytes <= 0 || r.Limits.OutputBytes <= 0 ||
		!json.Valid(r.Input) || int64(len(r.Input)) > r.Limits.InputBytes || (len(r.Context) != 0 && !json.Valid(r.Context)) || int64(len(r.Context)) > r.Limits.InputBytes-int64(len(r.Input)) ||
		(r.Idempotent && !present(r.IdempotencyKey)) || (r.IdempotencyKey != "" && !present(r.IdempotencyKey)) {
		return ErrInvalidRequest
	}
	if !now.Before(r.Deadline) {
		return context.DeadlineExceeded
	}
	return nil
}

type Status string

const (
	Succeeded Status = "succeeded"
	Failed    Status = "failed"
	Cancelled Status = "cancelled"
)

type Failure struct {
	Code, Message string
}

type Usage struct {
	CPUTime         time.Duration
	PeakMemoryBytes int64
}

type Result struct {
	ID, WorkerID string
	Status       Status
	Output       json.RawMessage
	Error        *Failure
	Duration     time.Duration
	Usage        Usage
	// Effects are untrusted intents. Brain must authorize and validate each
	// intent independently before committing it.
	Effects []json.RawMessage
}

func (v Result) Validate(r Request, assignedWorker string) error {
	if !present(assignedWorker) || v.ID != r.ID || v.WorkerID != assignedWorker || v.Duration < 0 || v.Usage.CPUTime < 0 || v.Usage.PeakMemoryBytes < 0 || r.Limits.OutputBytes <= 0 {
		return ErrInvalidResult
	}
	if v.Status == Succeeded {
		if v.Error != nil || !json.Valid(v.Output) {
			return ErrInvalidResult
		}
	} else if v.Status == Failed || v.Status == Cancelled {
		if v.Error == nil || !present(v.Error.Code) || len(v.Effects) != 0 || len(v.Output) != 0 {
			return ErrInvalidResult
		}
	} else {
		return ErrInvalidResult
	}
	remaining := r.Limits.OutputBytes - int64(len(v.Output))
	if v.Error != nil {
		remaining -= int64(len(v.Error.Code)) + int64(len(v.Error.Message))
	}
	if remaining < 0 {
		return ErrInvalidResult
	}
	for _, effect := range v.Effects {
		if !json.Valid(effect) || int64(len(effect)) > remaining {
			return ErrInvalidResult
		}
		remaining -= int64(len(effect))
	}
	return nil
}
