package execution

import (
	"bytes"
	"context"
	"time"
)

// CheckRequest verifies structural admission and the exact registered artifact
// and runtime. Caller authorization, quota and lifecycle checks remain Brain's
// responsibility and must precede dispatch and be rechecked before commit.
func (c *Catalog) CheckRequest(request Request, now time.Time) error {
	if err := request.Validate(now); err != nil {
		return err
	}
	key := ArtifactKey{ProjectID: request.Artifact.ProjectID, ID: request.Artifact.ID, Version: request.Artifact.Version}
	artifact, runtime, err := c.Resolve(key, request.Artifact.Entrypoint)
	if err != nil {
		return err
	}
	if artifact != request.Artifact {
		return ErrArtifactMismatch
	}
	if runtime != request.Runtime {
		return ErrInvalidRequest
	}
	return nil
}

// ExecuteAssigned checks a host-selected executor and assignment. It never
// schedules, retries or commits effects. The executor must honor cancellation
// and release its underlying runtime resources; this boundary rejects late
// results even if an adapter fails to do so.
func ExecuteAssigned(ctx context.Context, catalog *Catalog, workerID string, executor Executor, request Request) (Result, error) {
	if catalog == nil || executor == nil || !present(workerID) {
		return Result{}, ErrUnavailable
	}
	if err := catalog.CheckRequest(request, time.Now()); err != nil {
		return Result{}, err
	}
	ctx, cancel := context.WithDeadline(ctx, request.Deadline)
	defer cancel()
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}
	request.Deadline, _ = ctx.Deadline()
	request.Input = bytes.Clone(request.Input)
	request.Context = bytes.Clone(request.Context)
	result, err := executor.Execute(ctx, request)
	if cancellation := ctx.Err(); cancellation != nil {
		return Result{}, cancellation
	}
	if err != nil {
		return Result{}, err
	}
	if err := result.Validate(request, workerID); err != nil {
		return Result{}, err
	}
	return result, nil
}
