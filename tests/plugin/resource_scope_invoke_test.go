package plugin_test

import (
	"errors"
	"testing"

	"brain/internal/plugin"
)

func TestRegistryEnforcesDeclaredResourceBeforeWorker(t *testing.T) {
	r := plugin.NewRegistry(1)
	m := plugin.Manifest{ID: "sample", Version: "1", HostAPI: 1, Capabilities: []string{"read"}, Routes: []plugin.Route{{Operation: "read", Function: "read", Capability: "read", ResourceField: "id"}}}
	if err := r.Register(m); err != nil {
		t.Fatal(err)
	}
	if err := r.Install("alpha", m.ID, m.Version); err != nil {
		t.Fatal(err)
	}
	if err := r.SetReadiness(m.ID, m.Version, true); err != nil {
		t.Fatal(err)
	}
	if err := r.GrantResource("alpha", m.ID, m.Version, "read", "allowed"); err != nil {
		t.Fatal(err)
	}
	if err := r.Enable("alpha", m.ID, m.Version); err != nil {
		t.Fatal(err)
	}
	worker := &countingWorker{}
	if err := r.BindWorker("alpha", m.ID, m.Version, worker); err != nil {
		t.Fatal(err)
	}
	request := plugin.WorkerRequest{Protocol: 1, ProjectID: "alpha", PluginID: m.ID, Version: m.Version, Operation: "read", Payload: []byte(`{"id":"other"}`)}
	if _, err := r.Invoke(request); !errors.Is(err, plugin.ErrGrantDenied) || worker.calls != 0 {
		t.Fatalf("out-of-scope request reached worker: %v, calls=%d", err, worker.calls)
	}
	request.Payload = []byte(`{"id":"allowed"}`)
	if result, err := r.Invoke(request); err != nil || !result.Success || worker.calls != 1 {
		t.Fatalf("allowed request failed: %+v %v", result, err)
	}
}
