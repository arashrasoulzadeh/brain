package admin_test

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"testing"

	"brain/internal/admin"
	"brain/internal/execution"
	"brain/internal/operator"
)

type scheduledFunctionExecutor func(context.Context, execution.Request) (execution.Result, error)

func (f scheduledFunctionExecutor) Execute(ctx context.Context, req execution.Request) (execution.Result, error) {
	return f(ctx, req)
}

func TestPublishedFunctionDispatchUsesExactScopedBundleWithoutFallback(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(t.TempDir()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(cwd); err != nil {
			t.Error(err)
		}
	})
	service := admin.NewLifecycleService(nil, nil)
	actor := operator.Actor{Kind: operator.ActorOperator, ID: "operator"}
	source := "export const config = { method: 'POST' }; export default () => { throw new Error('not local') }"
	if _, err := service.CreateFunctionDraft(actor, "project", "sample", source); err != nil {
		t.Fatal(err)
	}
	if _, err := service.PublishFunction(actor, "project", "sample", "POST"); err != nil {
		t.Fatal(err)
	}
	catalog, err := execution.NewCatalog(1, 1<<20)
	if err != nil {
		t.Fatal(err)
	}
	runtime := execution.Runtime{Kind: execution.UserFunctionRuntimeKind, Profile: "test"}
	var seen execution.Artifact
	unavailable := true
	executor := scheduledFunctionExecutor(func(ctx context.Context, req execution.Request) (execution.Result, error) {
		seen = req.Artifact
		if req.ProjectID != "project" || req.Artifact.ProjectID != "project" || req.Auth.OriginalCaller != "public" || req.Auth.Caller != "public" || req.Auth.Resource != "sample" || req.Runtime != runtime {
			t.Fatalf("wrong execution identity: %+v", req)
		}
		bundle, err := catalog.Fetch(req.Artifact)
		if err != nil {
			t.Fatal(err)
		}
		files, err := execution.OpenBundle(req.Artifact, bundle, 1<<20)
		if err != nil || len(files) != 1 || string(files[0].Data) != source {
			t.Fatalf("wrong artifact: %+v, %v", files, err)
		}
		if unavailable {
			return execution.Result{}, execution.ErrUnavailable
		}
		return execution.Result{ID: req.ID, WorkerID: "test", Status: execution.Succeeded, Output: json.RawMessage(`{"result":{"remote":true}}`)}, nil
	})
	if err := service.BindFunctionExecution(catalog, executor, runtime); err != nil {
		t.Fatal(err)
	}
	if _, err := service.PublicInvokeFunction("project", "sample", "POST", nil); !errors.Is(err, admin.ErrFunctionExecutionUnavailable) {
		t.Fatalf("unavailable capacity not classified: %v", err)
	}
	if _, err := catalog.Fetch(seen); !errors.Is(err, execution.ErrArtifactNotFound) {
		t.Fatalf("failed invocation leaked bundle: %v", err)
	}
	unavailable = false
	result, err := service.PublicInvokeFunction("project", "sample", "POST", nil)
	if err != nil || result["result"].(map[string]any)["remote"] != true {
		t.Fatalf("remote result: %+v %v", result, err)
	}
	if _, err := catalog.Fetch(seen); !errors.Is(err, execution.ErrArtifactNotFound) {
		t.Fatalf("successful invocation leaked bundle: %v", err)
	}
}

func TestOperatorDraftUsesScheduledArtifactAndHostHandler(t *testing.T) {
	service := admin.NewLifecycleService(nil, nil)
	catalog, err := execution.NewCatalog(1, 1<<20)
	if err != nil {
		t.Fatal(err)
	}
	runtime := execution.Runtime{Kind: execution.UserFunctionRuntimeKind, Profile: "test"}
	var seen execution.Artifact
	executor := scheduledFunctionExecutor(func(ctx context.Context, req execution.Request) (execution.Result, error) {
		seen = req.Artifact
		if req.Artifact.Entrypoint != "functions.invoke" || req.Auth.Capability != "functions.invoke" || req.Auth.Caller != "operator:root" || req.Auth.OriginalCaller != "operator:root" {
			t.Fatalf("wrong draft identity: %+v", req)
		}
		if execution.HostHandler(ctx) == nil {
			t.Fatal("operator draft host handler missing")
		}
		return execution.Result{ID: req.ID, WorkerID: "test", Status: execution.Succeeded, Output: json.RawMessage(`{"result":{"draft":true}}`)}, nil
	})
	if err := service.BindFunctionExecution(catalog, executor, runtime); err != nil {
		t.Fatal(err)
	}
	actor := operator.Actor{Kind: operator.ActorOperator, ID: "root"}
	source := "export const config = { method: 'POST' }; export default () => { throw new Error('no local fallback') }"
	out, err := service.RunFunctionDraft(actor, "project", "sample", source, nil)
	if err != nil || out["result"].(map[string]any)["draft"] != true {
		t.Fatalf("draft result: %+v %v", out, err)
	}
	if _, err := catalog.Fetch(seen); !errors.Is(err, execution.ErrArtifactNotFound) {
		t.Fatal("draft artifact retained", err)
	}
}
