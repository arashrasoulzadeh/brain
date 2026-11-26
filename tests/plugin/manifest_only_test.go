package plugin_test

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"brain/internal/plugin"
)

func TestManifestOnlyPluginRunsWithoutTypeScript(t *testing.T) {
	node, err := exec.LookPath("node")
	if err != nil {
		t.Skip("Node is required for the worker contract")
	}
	root := t.TempDir()
	manifestPath := filepath.Join(root, "plugin.yaml")
	yaml := `id: dynamic
runtime: declarative
version: "1"
host_api: 1
capabilities:
  - dynamic.write
  - dynamic.read
  - dynamic.approve
routes:
  - operation: dynamic.write
    function: write
    capability: dynamic.write
    helper: sdk.resource
    storage:
      action: state.create
      key: id
      target: record
      values:
        id: $input.id
        status: pending
    response:
      input:
        - name: record
          type: object
          required: true
          error: record required
      result: $input.record
  - operation: dynamic.read
    function: read
    capability: dynamic.read
    helper: sdk.resource
    storage:
      action: state.read
      key: id
      target: record
    response:
      input:
        - name: record
          type: object
          required: true
          error: record not found
      result: $input.record
  - operation: dynamic.approve
    function: approve
    capability: dynamic.approve
    helper: sdk.resource
    storage:
      action: state.patch
      key: id
      target: record
      guard:
        field: status
        allowed:
          - pending
        error: invalid transition
      values:
        status: approved
    response:
      input:
        - name: record
          type: object
          required: true
          error: record not found
      result: $input.record
`
	if err := os.WriteFile(manifestPath, []byte(yaml), 0o600); err != nil {
		t.Fatal(err)
	}
	manifest, err := plugin.LoadManifest(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	if manifest.EntryPoint() != "" || len(manifest.SourceFiles) != 0 {
		t.Fatal("manifest-only plugin unexpectedly needs code")
	}
	script, err := filepath.Abs("../../scripts/typescript-worker.mjs")
	if err != nil {
		t.Fatal(err)
	}
	worker, err := plugin.NewManifestTypeScriptWorker(node, script, filepath.Join(root, "descriptor.json"), root, "project", manifest, plugin.IsolationPolicy{})
	if err != nil {
		t.Fatal(err)
	}
	if err := worker.Start(); err != nil {
		t.Fatal(err)
	}
	defer worker.Stop()
	registry := plugin.NewRegistry(1)
	if err := registry.Register(manifest); err != nil {
		t.Fatal(err)
	}
	if err := registry.Install("project", manifest.ID, manifest.Version); err != nil {
		t.Fatal(err)
	}
	if err := registry.SetReadiness(manifest.ID, manifest.Version, true); err != nil {
		t.Fatal(err)
	}
	for _, capability := range manifest.Capabilities {
		if err := registry.Grant("project", manifest.ID, manifest.Version, capability); err != nil {
			t.Fatal(err)
		}
	}
	if err := registry.Enable("project", manifest.ID, manifest.Version); err != nil {
		t.Fatal(err)
	}
	if err := registry.BindWorker("project", manifest.ID, manifest.Version, worker); err != nil {
		t.Fatal(err)
	}
	request := plugin.WorkerRequest{Protocol: 1, ProjectID: "project", PluginID: manifest.ID, Version: manifest.Version, Operation: "dynamic.write", Payload: []byte(`{"id":"one","record":{"status":"forged"}}`)}
	if _, err := registry.Invoke(request); !errors.Is(err, plugin.ErrManifestStorageUnavailable) {
		t.Fatalf("storage unavailable: %v", err)
	}
	storage, err := plugin.OpenSQLiteStorage(filepath.Join(root, "data.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()
	registry.BindSQLiteStorage(storage)
	for _, operation := range []string{"dynamic.write", "dynamic.read"} {
		request.Operation = operation
		response, err := registry.Invoke(request)
		if err != nil || !response.Success {
			t.Fatalf("invoke %s: %#v %v", operation, response, err)
		}
		var payload map[string]any
		if err := json.Unmarshal(response.Payload, &payload); err != nil || payload["status"] != "pending" || payload["id"] != "one" {
			t.Fatalf("response: %s %v", response.Payload, err)
		}
	}
	request.Operation = "dynamic.approve"
	if response, err := registry.Invoke(request); err != nil || !response.Success {
		t.Fatalf("approve: %#v %v", response, err)
	}
	if _, err := registry.Invoke(request); err == nil || err.Error() != "invalid transition" {
		t.Fatalf("invalid transition accepted: %v", err)
	}
	request.Operation = "dynamic.read"
	request.Payload = []byte(`{"id":"missing","record":{"status":"forged"}}`)
	if response, err := registry.Invoke(request); err != nil || response.Success {
		t.Fatalf("forged read: %#v %v", response, err)
	}
}
