package plugin_test

import (
	"context"
	"errors"
	"testing"

	"brain/internal/plugin"
	"brain/internal/project"
)

func TestWasmRunnerCreatesFreshRuntimePerInvocation(t *testing.T) {
	runner := &plugin.WasmRunner{}
	for range 2 {
		if err := runner.Run(context.Background(), []byte("not wasm")); !errors.Is(err, plugin.ErrWasmModuleInvalid) {
			t.Fatalf("Run error = %v", err)
		}
	}
	if runner.Invocations() != 2 {
		t.Fatalf("invocations = %d, want 2", runner.Invocations())
	}
}

func TestWasmRunnerRequiresHostProjectContext(t *testing.T) {
	runner := &plugin.WasmRunner{}
	ctx := project.WithProject(context.Background(), project.Project{ID: "dev"})
	if err := runner.RunForProject(ctx, "prod", []byte("not wasm")); !errors.Is(err, plugin.ErrWasmProjectMismatch) {
		t.Fatalf("cross-project WASM execution accepted: %v", err)
	}
	if err := runner.RunForProject(context.Background(), "dev", []byte("not wasm")); !errors.Is(err, plugin.ErrWasmProjectContextRequired) {
		t.Fatalf("unbound WASM execution accepted: %v", err)
	}
}
