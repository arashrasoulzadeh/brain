package execution_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"brain/internal/execution"
)

func validRequest(now time.Time) execution.Request {
	return execution.Request{
		ID: "invocation-1", TraceID: "trace-1", ProjectID: "project-a",
		Artifact: execution.Artifact{ID: "files", Version: "1.0.0", Digest: "sha256:" + strings.Repeat("a", 64), Entrypoint: "metadata"},
		Runtime:  execution.Runtime{Kind: "typescript", Profile: "sandboxed"},
		Input:    json.RawMessage(`{"key":"a"}`), Deadline: now.Add(time.Second),
		Limits: execution.Limits{CPUTime: time.Second, MemoryBytes: 1024, InputBytes: 128, OutputBytes: 128},
		Auth:   execution.Authorization{ProjectID: "project-a", OriginalCaller: "user:1", Caller: "user:1", Capability: "files.metadata", Resource: "a"},
	}
}

func TestRequestValidation(t *testing.T) {
	now := time.Unix(100, 0)
	if err := validRequest(now).Validate(now); err != nil {
		t.Fatal(err)
	}
	cases := map[string]func(*execution.Request){
		"missing ID":          func(r *execution.Request) { r.ID = "" },
		"missing trace":       func(r *execution.Request) { r.TraceID = "" },
		"self parent":         func(r *execution.Request) { r.ParentID = r.ID },
		"wrong project":       func(r *execution.Request) { r.Auth.ProjectID = "project-b" },
		"missing caller":      func(r *execution.Request) { r.Auth.OriginalCaller = "" },
		"missing capability":  func(r *execution.Request) { r.Auth.Capability = "" },
		"floating version":    func(r *execution.Request) { r.Artifact.Version = "latest" },
		"wildcard version":    func(r *execution.Request) { r.Artifact.Version = "1.x" },
		"range version":       func(r *execution.Request) { r.Artifact.Version = "^1.0.0" },
		"bad digest":          func(r *execution.Request) { r.Artifact.Digest = "sha256:abc" },
		"missing entrypoint":  func(r *execution.Request) { r.Artifact.Entrypoint = "" },
		"missing profile":     func(r *execution.Request) { r.Runtime.Profile = "" },
		"invalid JSON":        func(r *execution.Request) { r.Input = []byte("{") },
		"oversized input":     func(r *execution.Request) { r.Limits.InputBytes = 1 },
		"zero output limit":   func(r *execution.Request) { r.Limits.OutputBytes = 0 },
		"zero CPU limit":      func(r *execution.Request) { r.Limits.CPUTime = 0 },
		"zero memory limit":   func(r *execution.Request) { r.Limits.MemoryBytes = 0 },
		"self reported retry": func(r *execution.Request) { r.Idempotent = true },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			r := validRequest(now)
			mutate(&r)
			if !errors.Is(r.Validate(now), execution.ErrInvalidRequest) {
				t.Fatal("expected invalid request")
			}
		})
	}
	r := validRequest(now)
	r.Deadline = now
	if !errors.Is(r.Validate(now), context.DeadlineExceeded) {
		t.Fatal("expected expired deadline")
	}
}

func TestResultValidation(t *testing.T) {
	r := validRequest(time.Now())
	valid := execution.Result{ID: r.ID, WorkerID: "neuron-1", Status: execution.Succeeded, Output: json.RawMessage(`{"ok":true}`)}
	if err := valid.Validate(r, "neuron-1"); err != nil {
		t.Fatal(err)
	}
	cases := map[string]func(*execution.Result){
		"wrong invocation":      func(v *execution.Result) { v.ID = "other" },
		"wrong worker":          func(v *execution.Result) { v.WorkerID = "other" },
		"unknown status":        func(v *execution.Result) { v.Status = "unknown" },
		"success with error":    func(v *execution.Result) { v.Error = &execution.Failure{Code: "failed"} },
		"failure without error": func(v *execution.Result) { v.Status = execution.Failed },
		"invalid output":        func(v *execution.Result) { v.Output = []byte("{") },
		"oversized output":      func(v *execution.Result) { v.Output = []byte(`"` + strings.Repeat("a", 128) + `"`) },
		"negative duration":     func(v *execution.Result) { v.Duration = -time.Second },
		"oversized combined output": func(v *execution.Result) {
			v.Effects = []json.RawMessage{json.RawMessage(`"` + strings.Repeat("a", 120) + `"`)}
		},
		"invalid effect": func(v *execution.Result) { v.Effects = []json.RawMessage{json.RawMessage(`{`)} },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			v := valid
			mutate(&v)
			if !errors.Is(v.Validate(r, "neuron-1"), execution.ErrInvalidResult) {
				t.Fatal("expected invalid result")
			}
		})
	}
	failed := execution.Result{ID: r.ID, WorkerID: "neuron-1", Status: execution.Failed, Error: &execution.Failure{Code: "runtime_error"}, Effects: []json.RawMessage{json.RawMessage(`{}`)}}
	if !errors.Is(failed.Validate(r, "neuron-1"), execution.ErrInvalidResult) {
		t.Fatal("failed invocation must not propose effects")
	}
}

func TestArtifactVerification(t *testing.T) {
	a := execution.Artifact{Digest: "sha256:ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"}
	if err := a.Verify([]byte("abc")); err != nil {
		t.Fatal(err)
	}
	if !errors.Is(a.Verify([]byte("abd")), execution.ErrArtifactMismatch) {
		t.Fatal("tampered artifact accepted")
	}
}

func TestValidEffectsAndFailureBounds(t *testing.T) {
	r := validRequest(time.Now())
	v := execution.Result{ID: r.ID, WorkerID: "neuron-1", Status: execution.Succeeded, Output: json.RawMessage(`null`), Effects: []json.RawMessage{json.RawMessage(`{}`)}}
	if err := v.Validate(r, "neuron-1"); err != nil {
		t.Fatal(err)
	}
	v = execution.Result{ID: r.ID, WorkerID: "neuron-1", Status: execution.Failed, Error: &execution.Failure{Code: "runtime_error"}}
	if err := v.Validate(r, "neuron-1"); err != nil {
		t.Fatal(err)
	}
	v.Error.Message = strings.Repeat("a", 128)
	if !errors.Is(v.Validate(r, "neuron-1"), execution.ErrInvalidResult) {
		t.Fatal("oversized error accepted")
	}
}
