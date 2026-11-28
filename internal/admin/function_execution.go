package admin

import (
	"context"
	"encoding/json"
	"time"

	"brain/internal/execution"
)

type functionExecution struct {
	catalog *execution.Catalog
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
	bundle, err := execution.BuildBundle([]execution.BundleFile{{Path: "source.mjs", Data:[]byte(source)}}, execution.MaxBundleBytes)
	if err != nil {
		return nil, err
	}
	key := execution.ArtifactKey{ProjectID: project, ID: id, Version: "1"}
	release, err := f.catalog.RegisterTransient(execution.Registration{Key: key, Runtime: f.runtime, Entrypoints:[]string{"functions.call"}, Bundle: bundle})
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
		deadline = time.Now().Add(15*time.Second)
	}
	req := execution.Request{ID: id, TraceID: id, ProjectID: project, Artifact: artifact, Runtime: runtime, Input: input, Deadline: deadline, Limits: execution.Limits{CPUTime: 30 * time.Second, MemoryBytes: 2 << 30, InputBytes: 1 << 20, OutputBytes: 1 << 20}, Auth: execution.Authorization{ProjectID: project, OriginalCaller: principal, Caller: principal, Capability: "functions.call", Resource: name}}
	result, err := f.executor.Execute(ctx, req)
	if err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := result.Validate(req, result.WorkerID); err != nil || result.Status != execution.Succeeded || len(result.Effects) != 0 {
		return nil, execution.ErrInvalidResult
	}
	return result.Output, nil
}
