package plugin_test

import (
	"errors"
	"testing"

	"brain/internal/plugin"
)

func TestRegistryInvokeRequiresLifecycleAndBoundWorker(t *testing.T) {
	registry := plugin.NewRegistry(1)
	manifest := plugin.Manifest{ID: "sample", Title: "Sample", Description: "sample", Version: "1.0.0", HostAPI: 1, Capabilities: []string{"sample.read"}, Routes: []plugin.Route{{Operation: "sample.read", Function: "read", Capability: "sample.read"}}}
	if err := registry.Register(manifest); err != nil {
		t.Fatal(err)
	}
	if err := registry.Install("project", "sample", "1.0.0"); err != nil {
		t.Fatal(err)
	}
	if err := registry.SetReadiness("sample", "1.0.0", true); err != nil {
		t.Fatal(err)
	}
	if err := registry.Grant("project", "sample", "1.0.0", "sample.read"); err != nil {
		t.Fatal(err)
	}
	if err := registry.Enable("project", "sample", "1.0.0"); err != nil {
		t.Fatal(err)
	}
	request := plugin.WorkerRequest{Protocol: 1, ProjectID: "project", PluginID: "sample", Version: "1.0.0", Operation: "sample.read", Payload: []byte(`{"ok":true}`)}
	if _, err := registry.Invoke(request); !errors.Is(err, plugin.ErrWorkerNotReady) {
		t.Fatalf("unbound invoke error = %v", err)
	}
	worker := &contractWorker{ready: true}
	if err := registry.BindWorker("project", "sample", "1.0.0", worker); err != nil {
		t.Fatal(err)
	}
	response, err := registry.Invoke(request)
	if err != nil || !response.Success {
		t.Fatalf("Invoke() = %#v, %v", response, err)
	}
}
