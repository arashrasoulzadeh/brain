package plugin_test

import (
	"context"
	"errors"
	"testing"

	"brain/internal/plugin"
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
