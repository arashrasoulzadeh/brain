package cortex_test

import (
	"context"
	"errors"
	"testing"

	"brain/internal/execution"
)

func TestNestedPlacementRejectsOneSlotWithoutQueueing(t *testing.T) {
	f := setup(t, 4, 4, 4, 1)
	register(t, f, "only", 1, executorFunc(func(ctx context.Context, parent execution.Request) (execution.Result, error) {
		child := f.request("child", parent.ProjectID)
		child.ParentID, child.TraceID, child.Deadline = parent.ID, parent.TraceID, parent.Deadline
		if _, err := f.scheduler.Execute(ctx, child); !errors.Is(err, execution.ErrInvalidRequest) {
			t.Errorf("unbound parent accepted: %v", err)
		}
		ctx = execution.WithParent(ctx, parent)
		if _, err := f.scheduler.Execute(ctx, child); !errors.Is(err, execution.ErrUnavailable) {
			t.Errorf("saturated child did not fail closed: %v", err)
		}
		if f.scheduler.Snapshot().Queued != 0 {
			t.Error("child queued behind its parent")
		}
		child.ProjectID, child.Auth.ProjectID = "foreign", "foreign"
		if _, err := f.scheduler.Execute(ctx, child); !errors.Is(err, execution.ErrInvalidRequest) {
			t.Errorf("foreign child admitted: %v", err)
		}
		return result(parent, "only"), nil
	}), runtime)
	if _, err := f.scheduler.Execute(context.Background(), f.request("root", "alpha")); err != nil {
		t.Fatal(err)
	}
}

func TestNestedPlacementPreservesTraceAndRejectsGrandchild(t *testing.T) {
	f := setup(t, 4, 4, 4, 2)
	register(t, f, "a", 1, executorFunc(func(ctx context.Context, parent execution.Request) (execution.Result, error) {
		child := f.request("child", parent.ProjectID)
		child.ParentID, child.TraceID, child.Deadline = parent.ID, parent.TraceID, parent.Deadline
		if _, err := f.scheduler.Execute(execution.WithParent(ctx, parent), child); err != nil {
			return execution.Result{}, err
		}
		return result(parent, "a"), nil
	}), runtime)
	register(t, f, "b", 1, executorFunc(func(ctx context.Context, child execution.Request) (execution.Result, error) {
		if child.ParentID != "root" || child.TraceID != "root" {
			t.Error("child trace not preserved")
		}
		grandchild := f.request("grandchild", child.ProjectID)
		grandchild.ParentID, grandchild.TraceID, grandchild.Deadline = child.ID, child.TraceID, child.Deadline
		if _, err := f.scheduler.Execute(execution.WithParent(ctx, child), grandchild); !errors.Is(err, execution.ErrInvalidRequest) {
			t.Errorf("grandchild admitted: %v", err)
		}
		return result(child, "b"), nil
	}), runtime)
	if _, err := f.scheduler.Execute(context.Background(), f.request("root", "alpha")); err != nil {
		t.Fatal(err)
	}
	trace := f.scheduler.Snapshot().Recent
	if len(trace) != 2 || trace[1].ParentID != "root" || trace[1].TraceID != "root" {
		t.Fatalf("missing child trace: %+v", trace)
	}
}
