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

type executorFunc func(context.Context, execution.Request) (execution.Result, error)

func (f executorFunc) Execute(ctx context.Context, request execution.Request) (execution.Result, error) {
	return f(ctx, request)
}

func assignedRequest(t *testing.T) (*execution.Catalog, execution.Request) {
	t.Helper()
	catalog, err := execution.NewCatalog(2, 4096)
	if err != nil {
		t.Fatal(err)
	}
	reg := registration(t, "project-a", "1.0.0")
	if err := catalog.Register(reg); err != nil {
		t.Fatal(err)
	}
	artifact, runtime, err := catalog.Resolve(reg.Key, "files.metadata")
	if err != nil {
		t.Fatal(err)
	}
	req := validRequest(time.Now())
	req.Artifact, req.Runtime = artifact, runtime
	req.Deadline = time.Now().Add(time.Minute)
	return catalog, req
}

func success(req execution.Request) execution.Result {
	return execution.Result{ID: req.ID, WorkerID: "neuron-1", Status: execution.Succeeded, Output: json.RawMessage(`null`)}
}

func TestAssignedExecutionRequiresPinnedArtifact(t *testing.T) {
	catalog, req := assignedRequest(t)
	if err := catalog.CheckRequest(req, time.Now()); err != nil {
		t.Fatal(err)
	}
	cases := map[string]struct {
		mutate func(*execution.Request)
		want   error
	}{
		"digest":     {func(r *execution.Request) { r.Artifact.Digest = "sha256:" + strings.Repeat("a", 64) }, execution.ErrArtifactMismatch},
		"runtime":    {func(r *execution.Request) { r.Runtime.Profile = "trusted" }, execution.ErrInvalidRequest},
		"version":    {func(r *execution.Request) { r.Artifact.Version = "2.0.0" }, execution.ErrArtifactNotFound},
		"entrypoint": {func(r *execution.Request) { r.Artifact.Entrypoint = "unknown" }, execution.ErrArtifactNotFound},
		"project":    {func(r *execution.Request) { r.Artifact.ProjectID = "project-b" }, execution.ErrInvalidRequest},
		"deadline":   {func(r *execution.Request) { r.Deadline = time.Now().Add(-time.Second) }, context.DeadlineExceeded},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			bad := req
			tc.mutate(&bad)
			never := executorFunc(func(context.Context, execution.Request) (execution.Result, error) {
				t.Fatal("invalid request reached executor")
				return execution.Result{}, nil
			})
			if _, err := execution.ExecuteAssigned(context.Background(), catalog, "neuron-1", never, bad); !errors.Is(err, tc.want) {
				t.Fatalf("got %v, want %v", err, tc.want)
			}
		})
	}
}

func TestAssignedExecutionDeadlineAndResultBoundary(t *testing.T) {
	catalog, req := assignedRequest(t)
	req.Context = json.RawMessage(`{"safe":true}`)
	parent, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	deadline, _ := parent.Deadline()
	run := executorFunc(func(ctx context.Context, got execution.Request) (execution.Result, error) {
		if d, ok := ctx.Deadline(); !ok || !d.Equal(deadline) || !got.Deadline.Equal(deadline) {
			t.Fatal("parent deadline was not propagated")
		}
		got.Input[0] = 'X'
		got.Context[0] = 'X'
		return success(got), nil
	})
	result, err := execution.ExecuteAssigned(parent, catalog, "neuron-1", run, req)
	if err != nil || result.Status != execution.Succeeded || req.Input[0] != '{' || req.Context[0] != '{' {
		t.Fatalf("bad execution or shared mutable input: %v %v", result, err)
	}
	for _, worker := range []string{"", "neuron-2"} {
		want := execution.ErrInvalidResult
		if worker == "" {
			want = execution.ErrUnavailable
		}
		if _, err := execution.ExecuteAssigned(parent, catalog, worker, run, req); !errors.Is(err, want) {
			t.Fatalf("wrong worker accepted: %v", err)
		}
	}
	failure := errors.New("transport disconnected")
	broken := executorFunc(func(context.Context, execution.Request) (execution.Result, error) { return execution.Result{}, failure })
	if _, err := execution.ExecuteAssigned(parent, catalog, "neuron-1", broken, req); !errors.Is(err, failure) {
		t.Fatal("transport error lost")
	}
	if _, err := execution.ExecuteAssigned(parent, catalog, "neuron-1", nil, req); !errors.Is(err, execution.ErrUnavailable) {
		t.Fatal("missing executor did not fail closed")
	}
	if _, err := execution.ExecuteAssigned(parent, nil, "neuron-1", run, req); !errors.Is(err, execution.ErrUnavailable) {
		t.Fatal("missing catalog did not fail closed")
	}
}

func TestAssignedExecutionRejectsLateResult(t *testing.T) {
	catalog, req := assignedRequest(t)
	ctx, cancel := context.WithCancel(context.Background())
	late := executorFunc(func(_ context.Context, got execution.Request) (execution.Result, error) {
		cancel()
		return success(got), nil
	})
	if result, err := execution.ExecuteAssigned(ctx, catalog, "neuron-1", late, req); !errors.Is(err, context.Canceled) || result.ID != "" {
		t.Fatal("cancelled execution exposed result")
	}
	never := executorFunc(func(context.Context, execution.Request) (execution.Result, error) {
		t.Fatal("cancelled request was dispatched")
		return execution.Result{}, nil
	})
	if _, err := execution.ExecuteAssigned(ctx, catalog, "neuron-1", never, req); !errors.Is(err, context.Canceled) {
		t.Fatal("cancelled parent not rejected")
	}
}
