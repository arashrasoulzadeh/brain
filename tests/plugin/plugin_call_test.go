package plugin_test

import (
	"errors"
	"testing"

	"brain/internal/plugin"
)

func TestDeclaredPurePluginCallRequiresBothProjectGrants(t *testing.T) {
	r := plugin.NewRegistry(1)
	target := plugin.Manifest{ID: "target", Version: "1", HostAPI: 1, Declarative: true, Capabilities: []string{"read"}, Routes: []plugin.Route{{Operation: "read", Function: "read", Capability: "read", Helper: "sdk.response", Response: &plugin.ResponseSpec{Output: map[string]any{"ok": true}}}}}
	source := plugin.Manifest{ID: "source", Version: "1", HostAPI: 1, Declarative: true, OptionalDependencies: []string{"target@1"}, Capabilities: []string{"call.target.read"}, Routes: []plugin.Route{{Operation: "read", Function: "read", Capability: "call.target.read", Helper: "host.call", Call: &plugin.PluginCall{Plugin: "target", Version: "1", Operation: "read"}}}}
	for _, m := range []plugin.Manifest{source, target} {
		if err := r.Register(m); err != nil {
			t.Fatal(err)
		}
		if err := r.SetReadiness(m.ID, m.Version, true); err != nil {
			t.Fatal(err)
		}
		if err := r.Install("alpha", m.ID, m.Version); err != nil {
			t.Fatal(err)
		}
		if err := r.Grant("alpha", m.ID, m.Version, m.Capabilities[0]); err != nil {
			t.Fatal(err)
		}
		if err := r.Enable("alpha", m.ID, m.Version); err != nil {
			t.Fatal(err)
		}
		if err := r.BindWorker("alpha", m.ID, m.Version, &contractWorker{ready: true}); err != nil {
			t.Fatal(err)
		}
	}
	request := plugin.WorkerRequest{Protocol: 1, ProjectID: "alpha", PluginID: "source", Version: "1", Operation: "read", Payload: []byte("{}")}
	if err := r.CheckAdminReadAsOperator(request, "root"); err != nil {
		t.Fatalf("delegated read check failed: %v", err)
	}
	for _, invalid := range []struct {
		request plugin.WorkerRequest
		actor   string
		want    error
	}{
		{request, "", plugin.ErrActorDenied},
		{plugin.WorkerRequest{Protocol: 2}, "root", plugin.ErrInvalidWorkerRequest},
		{plugin.WorkerRequest{Protocol: 1, ProjectID: "alpha", PluginID: "source", Version: "1", Operation: "missing", Payload: []byte("{}")}, "root", plugin.ErrGrantDenied},
	} {
		if err := r.CheckAdminReadAsOperator(invalid.request, invalid.actor); !errors.Is(err, invalid.want) {
			t.Fatalf("invalid read check: %v, want %v", err, invalid.want)
		}
	}
	if response, err := r.Invoke(request); err != nil || !response.Success {
		t.Fatalf("delegation failed: %+v %v", response, err)
	}
	if err := r.Revoke("alpha", target.ID, target.Version, "read"); err != nil {
		t.Fatal(err)
	}
	if err := r.CheckAdminReadAsOperator(request, "root"); err == nil {
		t.Fatal("delegated read check bypassed target grant")
	}
	if _, err := r.Invoke(request); err == nil {
		t.Fatal("delegation bypassed target grant")
	}
	request.ProjectID = "beta"
	if _, err := r.Invoke(request); err == nil {
		t.Fatal("delegation bypassed project installation")
	}
}
