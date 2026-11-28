package neuron_test

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"brain/internal/execution"
	"brain/internal/neuron"
	"brain/internal/plugin"
)

func TestPluginRuntimePreservesPreparedContextAndEffects(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Linux trusted launcher fixture; real sandbox parity requires separate platform CI")
	}
	root := t.TempDir()
	for _, path := range []string{"scripts", "plugins/sdk-typescript"} {
		if err := os.MkdirAll(filepath.Join(root, path), 0o700); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(root, "package.json"), []byte(`{"type":"module"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	launcher := filepath.Join(root, "scripts", "run-sandboxed-typescript-worker.sh")
	logPath := filepath.Join(root, "invocation.json")
	script := "#!/bin/sh\n[ -L \"$WORKER_BUNDLE_PARENT/sdk-typescript\" ] || exit 1\n[ -L \"$WORKER_BUNDLE_PARENT/package.json\" ] || exit 1\nprintf '%s\\n' '{\"ready\":true}'\nIFS= read -r line || exit 1\nprintf '%s' \"$line\" > '" + logPath + "'\nprintf '%s\\n' '{\"jsonrpc\":\"2.0\",\"id\":\"call\",\"result\":{\"protocol\":1,\"success\":true,\"payload\":{\"ok\":true},\"observations\":[{\"event\":\"completed\",\"payload\":{\"ok\":true}}]}}'\nwhile IFS= read -r line; do :; done\n"
	if err := os.WriteFile(launcher, []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	runner, err := neuron.NewPluginTypeScriptRuntime(root, false)
	if err != nil {
		t.Fatal(err)
	}
	manifest := "id: sample\nversion: 1.0.0\nhost_api: 1\nentrypoint: src/index.mts\nsource_files:\n  - src/index.mts\ncapabilities:\n  - call\nconfig_schema:\n  label: string\nevents:\n  - name: completed\n    fields: {ok: boolean}\nroutes:\n  - operation: call\n    function: call\n    capability: call\n    helper: sdk.context\n"
	files := standaloneFiles(manifest, "export function call() { return {}; }")
	request := standaloneRequest(t, runner, files)
	request.Auth.Caller = "plugin:sample@1.0.0"
	request.Context, err = json.Marshal(plugin.InvocationContext{
		InvocationID: request.ID, ProjectVerified: true, Config: map[string]any{"label": "prepared"}, Available: []string{},
		Auth: plugin.InvocationAuth{
			Actor: plugin.InvocationActor{Kind: "user", ID: "42", Authenticated: true, Level: "authenticated"},
			Caller: plugin.InvocationActor{Kind: "plugin", ID: "sample@1.0.0", Level: "system"},
			Permissions: []plugin.InvocationPermission{{Capability: "call", Operation: "call"}},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	result, err := runner.Run(context.Background(), request, files)
	if err != nil || result.Status != execution.Succeeded || string(result.Output) != `{"ok":true}` || len(result.Effects) != 1 {
		t.Fatal(result, err)
	}
	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}
	var recorded struct {
		Params struct {
			ProjectID   string                   `json:"projectId"`
			HostContext plugin.InvocationContext `json:"hostContext"`
		}
	}
	if err := json.Unmarshal(data, &recorded); err != nil || recorded.Params.ProjectID != "alpha" || recorded.Params.HostContext.Config["label"] != "prepared" || recorded.Params.HostContext.Auth.Actor.ID != "42" {
		t.Fatal(string(data), err)
	}
	request.Auth.Caller = "host"
	if _, err := runner.Run(context.Background(), request, files); !errors.Is(err, plugin.ErrInvalidWorkerRequest) {
		t.Fatal("mismatched caller accepted", err)
	}
}
