package plugin

import (
	"context"
	"errors"
	"testing"

	"brain/internal/project"
)

func TestCoverageWasmExecution(t *testing.T) {
	runner := &WasmRunner{}
	// A valid empty WebAssembly module (magic and version).
	module := []byte{0, 97, 115, 109, 1, 0, 0, 0}
	ctx := project.WithProject(context.Background(), project.Project{ID: "coverage", Active: true})
	if err := runner.RunForProject(ctx, "other", module); !errors.Is(err, ErrWasmProjectMismatch) {
		t.Fatal(err)
	}
	if runner.Invocations() != 0 {
		t.Fatal("mismatched project executed WASM")
	}
	if err := runner.RunForProject(ctx, "coverage", module); err != nil {
		t.Fatal(err)
	}
	if runner.Invocations() != 1 {
		t.Fatal("successful invocation was not counted")
	}
	if err := runner.Run(ctx, []byte("invalid")); !errors.Is(err, ErrWasmModuleInvalid) {
		t.Fatal(err)
	}
}
