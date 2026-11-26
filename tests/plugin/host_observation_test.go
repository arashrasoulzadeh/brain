package plugin_test

import (
	"encoding/json"
	"errors"
	"testing"

	"brain/internal/plugin"
)

func TestHostObservationsCannotBePublishedThroughWorkerRoutes(t *testing.T) {
	r := plugin.NewRegistry(1)
	publisher := plugin.Manifest{ID: "dashboard", Version: "1", HostAPI: 1, Capabilities: []string{"plugin.enable", "dashboard.read"}, Routes: []plugin.Route{
		{Operation: "plugin.enable", Function: "enablePlugin", Helper: "host.admin", Capability: "plugin.enable"},
		{Operation: "dashboard.read", Function: "read", Capability: "dashboard.read"},
	}, Events: []plugin.EventSpec{{Name: "lifecycleChanged", Source: "host", Fields: map[string]string{"action": "string"}}}}
	subscriber := plugin.Manifest{ID: "observer", Version: "1", HostAPI: 1, Capabilities: []string{"observe.dashboard.lifecycleChanged"}, Observers: []plugin.ObserverSpec{{Plugin: "dashboard", Event: "lifecycleChanged", Function: "record", Capability: "observe.dashboard.lifecycleChanged"}}}
	for _, m := range []plugin.Manifest{publisher, subscriber} {
		if err := r.Register(m); err != nil {
			t.Fatal(err)
		}
		if err := r.SetReadiness(m.ID, m.Version, true); err != nil {
			t.Fatal(err)
		}
		if err := r.Install("project", m.ID, m.Version); err != nil {
			t.Fatal(err)
		}
		for _, capability := range m.Capabilities {
			if err := r.Grant("project", m.ID, m.Version, capability); err != nil {
				t.Fatal(err)
			}
		}
		if err := r.Enable("project", m.ID, m.Version); err != nil {
			t.Fatal(err)
		}
	}
	delivered := 0
	if err := r.BindWorker("project", "observer", "1", &storageCommitWorker{handle: func(request plugin.WorkerRequest) plugin.WorkerResponse {
		var event plugin.Observation
		if err := json.Unmarshal(request.Payload, &event); err != nil || event.PluginID != "dashboard" || event.ProjectID != "project" {
			t.Fatalf("unexpected event: %+v %v", event, err)
		}
		delivered++
		return plugin.WorkerResponse{Protocol: 1, Success: true}
	}}); err != nil {
		t.Fatal(err)
	}
	if err := r.BindWorker("project", "dashboard", "1", &storageCommitWorker{handle: func(plugin.WorkerRequest) plugin.WorkerResponse {
		return plugin.WorkerResponse{Protocol: 1, Success: true, Observations: []plugin.ObservationIntent{{Event: "lifecycleChanged", Payload: json.RawMessage(`{"action":"forged"}`)}}}
	}}); err != nil {
		t.Fatal(err)
	}
	request := plugin.WorkerRequest{Protocol: 1, ProjectID: "project", PluginID: "dashboard", Version: "1", Operation: "plugin.enable"}
	if _, err := r.Invoke(request); !errors.Is(err, plugin.ErrGrantDenied) {
		t.Fatalf("host route exposed to workers: %v", err)
	}
	request.Operation = "dashboard.read"
	if _, err := r.Invoke(request); err != nil || delivered != 0 {
		t.Fatalf("forged host observation delivered: %v %d", err, delivered)
	}
	r.PublishHostObservation("project", "dashboard", "lifecycleChanged", map[string]any{"action": "plugin.enable"})
	if delivered != 1 {
		t.Fatalf("host event not delivered: %d", delivered)
	}
}
