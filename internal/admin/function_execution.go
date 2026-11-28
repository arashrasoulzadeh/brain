package admin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"brain/internal/cortex"
	"brain/internal/execution"
	"brain/internal/synapse"
)

var (
	ErrFunctionExecutionFailed      = errors.New("function execution failed")
	ErrFunctionExecutionUnavailable = errors.New("function execution capacity unavailable")
	ErrFunctionExecutionTimeout     = errors.New("function execution deadline exceeded")
	ErrFunctionResultInvalid        = errors.New("function must return a JSON object, array, or null")
)

// scheduledFunctionError keeps transport internals and arbitrary worker error
// text out of the public response and operator logs.
func scheduledFunctionError(err error) error {
	switch {
	case errors.Is(err, context.DeadlineExceeded):
		return ErrFunctionExecutionTimeout
	case errors.Is(err, execution.ErrUnavailable), errors.Is(err, execution.ErrCatalogLimit),
		errors.Is(err, cortex.ErrProjectQuota), errors.Is(err, cortex.ErrQueueFull), errors.Is(err, cortex.ErrSessionLost),
		errors.Is(err, synapse.ErrClosed), errors.Is(err, synapse.ErrBackpressure):
		return ErrFunctionExecutionUnavailable
	case errors.Is(err, execution.ErrInvalidResult):
		return fmt.Errorf("%w: invalid Neuron result", ErrFunctionExecutionFailed)
	case errors.Is(err, execution.ErrInvalidRequest), errors.Is(err, execution.ErrArtifactNotFound), errors.Is(err, execution.ErrArtifactMismatch):
		return fmt.Errorf("%w: execution artifact or request rejected", ErrFunctionExecutionFailed)
	default:
		return ErrFunctionExecutionFailed
	}
}

type functionExecution struct {
	catalog  *execution.Catalog
	executor execution.Executor
	runtime  execution.Runtime
}

// BindFunctionExecution is called during startup, before accepting requests.
func (s *LifecycleService) BindFunctionExecution(catalog *execution.Catalog, executor execution.Executor, runtime execution.Runtime) error {
	if catalog == nil || executor == nil || runtime.Kind != execution.UserFunctionRuntimeKind || runtime.Profile == "" || s.functionExecution != nil {
		return execution.ErrInvalidRequest
	}
	s.functionExecution = &functionExecution{catalog: catalog, executor: executor, runtime: runtime}
	return nil
}

func (f *functionExecution) run(ctx context.Context, project, name, source, id, principal string, input json.RawMessage) ([]byte, error) {
	bundle, err := execution.BuildBundle([]execution.BundleFile{{Path: "source.mjs", Data: []byte(source)}}, execution.MaxBundleBytes)
	if err != nil {
		return nil, err
	}
	key := execution.ArtifactKey{ProjectID: project, ID: id, Version: "1"}
	release, err := f.catalog.RegisterTransient(execution.Registration{Key: key, Runtime: f.runtime, Entrypoints: []string{"functions.call"}, Bundle: bundle})
	if err != nil {
		return nil, err
	}
	defer release()
	artifact, runtime, err := f.catalog.Resolve(key, "functions.call")
	if err != nil {
		return nil, err
	}
	deadline, ok := ctx.Deadline()
	if !ok {
		deadline = time.Now().Add(15 * time.Second)
	}
	req := execution.Request{ID: id, TraceID: id, ProjectID: project, Artifact: artifact, Runtime: runtime, Input: input, Deadline: deadline, Limits: execution.Limits{CPUTime: 30 * time.Second, MemoryBytes: 2 << 30, InputBytes: 1 << 20, OutputBytes: 1 << 20}, Auth: execution.Authorization{ProjectID: project, OriginalCaller: principal, Caller: principal, Capability: "functions.call", Resource: name}}
	result, err := f.executor.Execute(ctx, req)
	if err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := result.Validate(req, result.WorkerID); err != nil || len(result.Effects) != 0 {
		return nil, execution.ErrInvalidResult
	}
	if result.Status != execution.Succeeded {
		return nil, ErrFunctionExecutionFailed
	}
	return result.Output, nil
}
