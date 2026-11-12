package project_test

import (
	"context"
	"testing"

	"brain/internal/project"
)

func TestCallerContextIsBoundToTransactionContext(t *testing.T) {
	tx := project.WithCaller(context.Background(), "dev", "user-1")
	caller, ok := project.CallerFromContext(tx)
	if !ok || caller.ProjectID != "dev" || caller.ActorID != "user-1" {
		t.Fatalf("caller = %#v, %v", caller, ok)
	}
	if _, ok := project.CallerFromContext(context.Background()); ok {
		t.Fatal("caller leaked outside transaction context")
	}
}
