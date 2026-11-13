package plugin

import (
	"context"
	"errors"

	"github.com/tetratelabs/wazero"
)

var ErrWasmModuleInvalid = errors.New("invalid wasm module")

type WasmRunner struct{ invocations int }

func (r *WasmRunner) Run(ctx context.Context, wasm []byte) error {
	r.invocations++
	runtime := wazero.NewRuntime(ctx)
	defer runtime.Close(ctx)
	compiled, err := runtime.CompileModule(ctx, wasm)
	if err != nil {
		return ErrWasmModuleInvalid
	}
	defer compiled.Close(ctx)
	_, err = runtime.InstantiateModule(ctx, compiled, wazero.NewModuleConfig())
	return err
}

func (r *WasmRunner) Invocations() int { return r.invocations }
