package neuron

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"brain/internal/execution"
	"brain/internal/plugin"
)

// NewPluginTypeScriptRuntime is the full prepared-plugin compatibility profile.
// The standalone profile remains separate and retains its stricter admission.
func NewPluginTypeScriptRuntime(root string, localMacOS bool) (*TypeScriptRuntime, error) {
	runner, err := NewTypeScriptRuntime(root, localMacOS)
	if err != nil {
		return nil, err
	}
	runner.capability.Kind = plugin.ScheduledTypeScriptKind
	return runner, nil
}

func (r *TypeScriptRuntime) probePlugin(ctx context.Context) error {
	files := []execution.BundleFile{
		{Path: "plugin.yaml", Data: []byte("id: neuron-plugin-probe\nversion: 1.0.0\nhost_api: 1\nentrypoint: index.ts\nsource_files:\n  - index.ts\nimports:\n  - ../sdk-typescript/src/index.js\ncapabilities:\n  - health\nevents:\n  - name: completed\n    fields: {ok: boolean}\nroutes:\n  - operation: health\n    function: health\n    helper: sdk.context\n    capability: health\n")},
		{Path: "index.ts", Data: []byte("import { pluginRequest } from '../sdk-typescript/src/index.js';\nexport function health(context: any) { const memory = new WebAssembly.Memory({ initial: 1, maximum: 2 }); context.plugins.observe.fire('completed', { ok: true }); return { accepted: typeof pluginRequest === 'function' && memory.buffer.byteLength === 65536 && context.project.id === 'neuron-probe' && context.auth.can('health') }; }\n")},
	}
	bundle, err := execution.BuildBundle(files, 4096)
	if err != nil {
		return err
	}
	host := plugin.InvocationContext{
		InvocationID: "probe", ProjectVerified: true, Config: map[string]any{}, Available: []string{},
		Auth: plugin.InvocationAuth{
			Actor:       plugin.InvocationActor{Kind: "host", Level: "system"},
			Caller:      plugin.InvocationActor{Kind: "host", Level: "system"},
			Permissions: []plugin.InvocationPermission{{Capability: "health", Operation: "health"}},
		},
	}
	encoded, err := json.Marshal(host)
	if err != nil {
		return err
	}
	request := execution.Request{
		ID: "probe", TraceID: "probe", ProjectID: "neuron-probe",
		Artifact: execution.Artifact{ID: "neuron-plugin-probe", Version: "1.0.0", Digest: execution.ArtifactDigest(bundle), Entrypoint: "health"},
		Runtime:  r.capability, Input: json.RawMessage("null"), Context: encoded, Deadline: time.Now().Add(20 * time.Second),
		Limits: execution.Limits{CPUTime: TypeScriptCPU, MemoryBytes: TypeScriptMemory, InputBytes: 4096, OutputBytes: 4096},
		Auth:   execution.Authorization{ProjectID: "neuron-probe", OriginalCaller: "host", Caller: "host", Capability: "health"},
	}
	result, err := r.Run(ctx, request, files)
	if err != nil {
		return err
	}
	if result.Status != execution.Succeeded || string(result.Output) != "{\"accepted\":true}" || len(result.Effects) != 1 {
		return errors.New("isolated plugin Neuron startup probe failed")
	}
	var intent plugin.ObservationIntent
	if json.Unmarshal(result.Effects[0], &intent) != nil || intent.Event != "completed" || string(intent.Payload) != "{\"ok\":true}" {
		return errors.New("isolated plugin Neuron observation probe failed")
	}
	return nil
}
