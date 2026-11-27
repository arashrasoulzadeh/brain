package synapse

import (
	"encoding/json"
	"math"
	"time"

	"brain/internal/execution"
	pb "brain/proto/synapse/v1"
)

func leaseWire(l Lease) *pb.Lease {
	return &pb.Lease{SessionId: l.SessionID, Attempt: l.Attempt, ExecutionId: l.ExecutionID}
}

func readLease(l *pb.Lease) (Lease, error) {
	if l == nil || !validID(l.SessionId) || !validID(l.ExecutionId) || l.Attempt == 0 {
		return Lease{}, ErrProtocol
	}
	return Lease{SessionID: l.SessionId, Attempt: l.Attempt, ExecutionID: l.ExecutionId}, nil
}

func runtimeWire(r execution.Runtime) *pb.Runtime {
	return &pb.Runtime{Kind: r.Kind, IsolationProfile: r.Profile}
}

func readRuntime(r *pb.Runtime) (execution.Runtime, error) {
	if r == nil || !validID(r.Kind) || !validID(r.IsolationProfile) {
		return execution.Runtime{}, ErrProtocol
	}
	return execution.Runtime{Kind: r.Kind, Profile: r.IsolationProfile}, nil
}

func assignmentWire(a Assignment) *pb.Assignment {
	r := a.Request
	return &pb.Assignment{
		Lease: leaseWire(a.Lease), TraceId: r.TraceID, ParentExecutionId: r.ParentID, ProjectId: r.ProjectID,
		Artifact: &pb.Artifact{ProjectScope: r.Artifact.ProjectID, Id: r.Artifact.ID, Version: r.Artifact.Version, Sha256Digest: r.Artifact.Digest, Entrypoint: r.Artifact.Entrypoint},
		Runtime: runtimeWire(r.Runtime), InputJson: r.Input, ContextJson: r.Context, DeadlineUnixNanos: r.Deadline.UnixNano(),
		Limits: &pb.Limits{CpuNanos: uint64(r.Limits.CPUTime), MemoryBytes: uint64(r.Limits.MemoryBytes), InputBytes: uint64(r.Limits.InputBytes), OutputBytes: uint64(r.Limits.OutputBytes)},
		Authorization: &pb.Authorization{OriginalCaller: r.Auth.OriginalCaller, Caller: r.Auth.Caller, Capability: r.Auth.Capability, Resource: r.Auth.Resource},
		Idempotent: r.Idempotent, IdempotencyKey: r.IdempotencyKey,
	}
}

func readAssignment(a *pb.Assignment) (Assignment, error) {
	if a == nil || a.Artifact == nil || a.Limits == nil || a.Authorization == nil ||
		a.Limits.CpuNanos > math.MaxInt64 || a.Limits.MemoryBytes > math.MaxInt64 || a.Limits.InputBytes > math.MaxInt64 || a.Limits.OutputBytes > math.MaxInt64 {
		return Assignment{}, ErrProtocol
	}
	lease, err := readLease(a.Lease)
	if err != nil {
		return Assignment{}, err
	}
	runtime, err := readRuntime(a.Runtime)
	if err != nil {
		return Assignment{}, err
	}
	r := execution.Request{
		ID: lease.ExecutionID, TraceID: a.TraceId, ParentID: a.ParentExecutionId, ProjectID: a.ProjectId,
		Artifact: execution.Artifact{ProjectID: a.Artifact.ProjectScope, ID: a.Artifact.Id, Version: a.Artifact.Version, Digest: a.Artifact.Sha256Digest, Entrypoint: a.Artifact.Entrypoint},
		Runtime: runtime, Input: a.InputJson, Context: a.ContextJson, Deadline: time.Unix(0, a.DeadlineUnixNanos),
		Limits: execution.Limits{CPUTime: time.Duration(a.Limits.CpuNanos), MemoryBytes: int64(a.Limits.MemoryBytes), InputBytes: int64(a.Limits.InputBytes), OutputBytes: int64(a.Limits.OutputBytes)},
		Auth: execution.Authorization{ProjectID: a.ProjectId, OriginalCaller: a.Authorization.OriginalCaller, Caller: a.Authorization.Caller, Capability: a.Authorization.Capability, Resource: a.Authorization.Resource},
		Idempotent: a.Idempotent, IdempotencyKey: a.IdempotencyKey,
	}
	return Assignment{Lease: lease, Request: r}, nil
}

func resultWire(l Lease, r execution.Result) *pb.Result {
	out := &pb.Result{Lease: leaseWire(l), OutputJson: r.Output, DurationNanos: uint64(r.Duration), CpuNanos: uint64(r.Usage.CPUTime), PeakMemoryBytes: uint64(r.Usage.PeakMemoryBytes)}
	switch r.Status {
	case execution.Succeeded:
		out.Status = pb.Result_SUCCEEDED
	case execution.Failed:
		out.Status = pb.Result_FAILED
	case execution.Cancelled:
		out.Status = pb.Result_CANCELLED
	}
	if r.Error != nil {
		out.Error = &pb.Failure{Code: r.Error.Code, Message: r.Error.Message}
	}
	for _, effect := range r.Effects {
		out.EffectIntentsJson = append(out.EffectIntentsJson, effect)
	}
	return out
}

func readResult(r *pb.Result, workerID string) (Lease, execution.Result, error) {
	if r == nil || r.DurationNanos > math.MaxInt64 || r.CpuNanos > math.MaxInt64 || r.PeakMemoryBytes > math.MaxInt64 {
		return Lease{}, execution.Result{}, ErrProtocol
	}
	lease, err := readLease(r.Lease)
	if err != nil {
		return Lease{}, execution.Result{}, err
	}
	out := execution.Result{ID: lease.ExecutionID, WorkerID: workerID, Output: r.OutputJson, Duration: time.Duration(r.DurationNanos), Usage: execution.Usage{CPUTime: time.Duration(r.CpuNanos), PeakMemoryBytes: int64(r.PeakMemoryBytes)}}
	switch r.Status {
	case pb.Result_SUCCEEDED:
		out.Status = execution.Succeeded
	case pb.Result_FAILED:
		out.Status = execution.Failed
	case pb.Result_CANCELLED:
		out.Status = execution.Cancelled
	default:
		return Lease{}, execution.Result{}, ErrProtocol
	}
	if r.Error != nil {
		out.Error = &execution.Failure{Code: r.Error.Code, Message: r.Error.Message}
	}
	for _, effect := range r.EffectIntentsJson {
		out.Effects = append(out.Effects, json.RawMessage(effect))
	}
	return lease, out, nil
}

func chunkWire(c Chunk) *pb.ArtifactChunk {
	return &pb.ArtifactChunk{Lease: leaseWire(c.Lease), Offset: uint64(c.Offset), TotalBytes: uint64(c.Total), Data: c.Data, FinalChunk: c.Final}
}

func readChunk(c *pb.ArtifactChunk) (Chunk, error) {
	if c == nil || c.Offset > math.MaxInt64 || c.TotalBytes > math.MaxInt64 {
		return Chunk{}, ErrProtocol
	}
	l, err := readLease(c.Lease)
	if err != nil {
		return Chunk{}, err
	}
	return Chunk{Lease: l, Offset: int64(c.Offset), Total: int64(c.TotalBytes), Data: c.Data, Final: c.FinalChunk}, nil
}
