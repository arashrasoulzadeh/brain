package plugin_test

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"brain/internal/plugin"
)

func TestManifestStorageIsScopedByProjectPluginAndVersion(t *testing.T) {
	storage, err := plugin.OpenSQLiteStorage(filepath.Join(t.TempDir(), "data.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()
	registry := plugin.NewRegistry(1)
	registry.BindSQLiteStorage(storage)
	worker := &storageCommitWorker{handle: func(request plugin.WorkerRequest) plugin.WorkerResponse {
		return plugin.WorkerResponse{Protocol: 1, Success: true, Payload: request.Payload}
	}}
	for _, id := range []string{"custom", "other"} {
		for _, version := range []string{"1", "2"} {
			manifest := plugin.Manifest{
				ID: id, Version: version, HostAPI: 1, Capabilities: []string{"write", "read"},
				Routes: []plugin.Route{
					{Operation: "write", Function: "write", Capability: "write", Storage: &plugin.StorageSpec{Action: "state.create", Key: "id", Target: "record", Values: map[string]any{"id": "$input.id", "status": version}}},
					{Operation: "read", Function: "read", Capability: "read", Storage: &plugin.StorageSpec{Action: "state.read", Key: "id", Target: "record"}},
				},
			}
			if err := registry.Register(manifest); err != nil {
				t.Fatal(err)
			}
			if err := registry.SetReadiness(id, version, true); err != nil {
				t.Fatal(err)
			}
			for _, project := range []string{"project-a", "project-b"} {
				if err := registry.Install(project, id, version); err != nil {
					t.Fatal(err)
				}
				for _, capability := range manifest.Capabilities {
					if err := registry.Grant(project, id, version, capability); err != nil {
						t.Fatal(err)
					}
				}
				if err := registry.Enable(project, id, version); err != nil {
					t.Fatal(err)
				}
				if err := registry.BindWorker(project, id, version, worker); err != nil {
					t.Fatal(err)
				}
			}
		}
	}
	invoke := func(project, id, version, op string) map[string]any {
		t.Helper()
		response, err := registry.Invoke(plugin.WorkerRequest{
			Protocol: 1, ProjectID: project, PluginID: id, Version: version, Operation: op,
			Payload: []byte(`{"id":" item ","record":{"status":"spoof"}}`),
		})
		if err != nil || !response.Success {
			t.Fatalf("invoke: %#v %v", response, err)
		}
		var payload map[string]any
		if err := json.Unmarshal(response.Payload, &payload); err != nil {
			t.Fatal(err)
		}
		return payload
	}
	invoke("project-a", "custom", "1", "write")
	for _, scope := range [][2]string{{"project-b", "custom"}, {"project-a", "other"}} {
		if result := invoke(scope[0], scope[1], "1", "read"); result["record"] != nil {
			t.Fatalf("cross-scope data or caller-supplied record exposed: %v", result)
		}
	}
	invoke("project-a", "custom", "2", "write")
	result := invoke("project-a", "custom", "1", "read")
	record, ok := result["record"].(map[string]any)
	if !ok || record["status"] != "2" || record["id"] != "item" {
		t.Fatalf("wrong registered version or key: %v", result)
	}
}

func TestStorageManifestRejectsInvalidOperations(t *testing.T) {
	for _, spec := range []*plugin.StorageSpec{
		{Action: "unknown", Key: "id"},
		{Action: "state.read", Key: "id", FixedKey: "x", Target: "record"},
		{Action: "state.read", Key: "id", Target: "id"},
		{Action: "state.create", Key: "id", Target: "record", Values: map[string]any{"value": "$eval.x"}},
		{Action: "state.create", Key: "id", Target: "record", Values: map[string]any{"__proto__": true}},
	} {
		registry := plugin.NewRegistry(1)
		manifest := plugin.Manifest{ID: "sample", Version: "1", HostAPI: 1, Capabilities: []string{"write"}, Routes: []plugin.Route{{Operation: "write", Function: "write", Capability: "write", Storage: spec}}}
		if err := registry.Register(manifest); err == nil {
			t.Fatalf("invalid storage accepted: %#v", spec)
		}
	}
}
