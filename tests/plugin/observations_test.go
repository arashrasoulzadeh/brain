package plugin_test

import (
	"encoding/json"
	"errors"
	"testing"

	"brain/internal/plugin"
)

type observationPlanner struct{ commit func() error }

func (p observationPlanner) Supports(id string) bool { return id == "rooms" }
func (p observationPlanner) Prepare(request plugin.WorkerRequest) (plugin.WorkerRequest, func() error, error) {
	return request, p.commit, nil
}

func TestObservationsUseCommittedHostProjectAndIsolateObservers(t *testing.T) {
	registry := plugin.NewRegistry(1)
	registry.BindProjectResolver(func(id string) (plugin.ProjectSnapshot, error) {
		return plugin.ProjectSnapshot{ID: id, Active: true}, nil
	})
	publisher := plugin.Manifest{ID: "rooms", Version: "1", HostAPI: 1, Capabilities: []string{"rooms.connect"}, Routes: []plugin.Route{{Operation: "rooms.connect", Function: "connect", Capability: "rooms.connect"}}, Events: []plugin.EventSpec{{Name: "connected", Fields: map[string]string{"roomId": "string"}}}}
	subscriber := plugin.Manifest{ID: "analytics", Version: "1", HostAPI: 1, Capabilities: []string{"observe.rooms.connected"}, Observers: []plugin.ObserverSpec{{Plugin: "rooms", Event: "connected", Function: "track", Capability: "observe.rooms.connected"}}}
	for _, manifest := range []plugin.Manifest{publisher, subscriber} {
		if err := registry.Register(manifest); err != nil {
			t.Fatal(err)
		}
		if err := registry.SetReadiness(manifest.ID, manifest.Version, true); err != nil {
			t.Fatal(err)
		}
		for _, projectID := range []string{"one", "two"} {
			if err := registry.Install(projectID, manifest.ID, manifest.Version); err != nil {
				t.Fatal(err)
			}
			for _, capability := range manifest.Capabilities {
				if err := registry.Grant(projectID, manifest.ID, manifest.Version, capability); err != nil {
					t.Fatal(err)
				}
			}
			if err := registry.Enable(projectID, manifest.ID, manifest.Version); err != nil {
				t.Fatal(err)
			}
		}
	}
	committed, failCommit, failPublisher, invalid, failObserver := false, false, false, false, false
	commitError := errors.New("commit failed")
	registry.BindInvocationStorage(observationPlanner{commit: func() error {
		if failCommit {
			return commitError
		}
		committed = true
		return nil
	}})
	publisherWorker := &storageCommitWorker{handle: func(request plugin.WorkerRequest) plugin.WorkerResponse {
		if request.Project == nil || request.Project.ID != "one" || !request.Project.Active {
			t.Fatalf("untrusted project reached worker: %+v", request.Project)
		}
		intent := plugin.ObservationIntent{Event: "connected", Payload: json.RawMessage(`{"roomId":"room"}`)}
		if invalid {
			intent.Payload = json.RawMessage(`{"roomId":"room","projectId":"two"}`)
		}
		return plugin.WorkerResponse{Protocol: 1, Success: !failPublisher, Observations: []plugin.ObservationIntent{intent}}
	}}
	delivered := 0
	observerWorker := &storageCommitWorker{handle: func(request plugin.WorkerRequest) plugin.WorkerResponse {
		if !committed {
			t.Fatal("event delivered before storage commit")
		}
		var event plugin.Observation
		if err := json.Unmarshal(request.Payload, &event); err != nil || event.ProjectID != "one" || event.PluginID != "rooms" || event.Event != "connected" || request.ProjectID != "one" || request.PluginID != "analytics" {
			t.Fatalf("invalid observation envelope: %+v %v", event, err)
		}
		delivered++
		return plugin.WorkerResponse{Protocol: 1, Success: !failObserver, Observations: []plugin.ObservationIntent{{Event: "connected", Payload: event.Payload}}}
	}}
	if err := registry.BindWorker("one", "rooms", "1", publisherWorker); err != nil {
		t.Fatal(err)
	}
	if err := registry.BindWorker("one", "analytics", "1", observerWorker); err != nil {
		t.Fatal(err)
	}
	if err := registry.BindWorker("two", "analytics", "1", &storageCommitWorker{handle: func(plugin.WorkerRequest) plugin.WorkerResponse {
		t.Fatal("cross-project delivery")
		return plugin.WorkerResponse{}
	}}); err != nil {
		t.Fatal(err)
	}
	request := plugin.WorkerRequest{Protocol: 1, ProjectID: "one", PluginID: "rooms", Version: "1", Operation: "rooms.connect", Project: &plugin.ProjectSnapshot{ID: "two", Active: false}}
	response, err := registry.Invoke(request)
	if err != nil || !response.Success || delivered != 1 || len(response.Observations) != 0 {
		t.Fatalf("expected one isolated delivery: %+v %v %d", response, err, delivered)
	}
	forged := plugin.WorkerRequest{Protocol: 1, ProjectID: "one", PluginID: "analytics", Version: "1", Operation: "@observe/rooms/connected"}
	if _, err := registry.Invoke(forged); !errors.Is(err, plugin.ErrGrantDenied) {
		t.Fatalf("observer exposed as public route: %v", err)
	}
	committed, failCommit = false, true
	if _, err := registry.Invoke(request); !errors.Is(err, commitError) || delivered != 1 {
		t.Fatalf("failed commit delivered event: %v %d", err, delivered)
	}
	failCommit, failPublisher = false, true
	if response, err := registry.Invoke(request); err != nil || response.Success || delivered != 1 {
		t.Fatalf("failed publisher delivered event: %+v %v %d", response, err, delivered)
	}
	failPublisher, invalid = false, true
	if _, err := registry.Invoke(request); err != nil || delivered != 1 {
		t.Fatalf("invalid payload delivered event: %v %d", err, delivered)
	}
	invalid, failObserver = false, true
	if response, err := registry.Invoke(request); err != nil || !response.Success || delivered != 2 {
		t.Fatalf("observer failure changed publisher result: %+v %v %d", response, err, delivered)
	}
	if err := registry.Revoke("one", "analytics", "1", "observe.rooms.connected"); err != nil {
		t.Fatal(err)
	}
	if _, err := registry.Invoke(request); err != nil || delivered != 2 {
		t.Fatalf("revoked observer received event: %v %d", err, delivered)
	}
	registry.BindProjectResolver(func(id string) (plugin.ProjectSnapshot, error) {
		return plugin.ProjectSnapshot{ID: id, Active: false}, nil
	})
	if _, err := registry.Invoke(request); !errors.Is(err, plugin.ErrProjectInvalid) {
		t.Fatalf("inactive project was not rejected: %v", err)
	}
}
