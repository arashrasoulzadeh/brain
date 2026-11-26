package plugin_test

import (
	"encoding/json"
	"path/filepath"
	"testing"
	"time"

	"brain/internal/plugin"
)

func TestDurableObservationDeliverySurvivesRestartAndRetries(t *testing.T) {
	path := filepath.Join(t.TempDir(), "observations.sqlite")
	publisher := plugin.Manifest{ID: "rooms", Version: "1", HostAPI: 1, Capabilities: []string{"rooms.join"}, Routes: []plugin.Route{{Operation: "rooms.join", Function: "join", Capability: "rooms.join"}}, Events: []plugin.EventSpec{{Name: "connected", Fields: map[string]string{"roomId": "string"}}}}
	subscriber := plugin.Manifest{ID: "analytics", Version: "1", HostAPI: 1, Capabilities: []string{"observe.rooms.connected"}, Observers: []plugin.ObserverSpec{{Plugin: "rooms", Event: "connected", Function: "track", Capability: "observe.rooms.connected", Persist: true}}}
	makeRegistry := func(storage *plugin.SQLiteStorage) *plugin.Registry {
		r := plugin.NewRegistry(1)
		for _, m := range []plugin.Manifest{publisher, subscriber} {
			if err := r.Register(m); err != nil {
				t.Fatal(err)
			}
		}
		if err := r.BindSQLiteStorage(storage); err != nil {
			t.Fatal(err)
		}
		return r
	}
	storage, err := plugin.OpenSQLiteStorage(path)
	if err != nil {
		t.Fatal(err)
	}
	r := makeRegistry(storage)
	for _, m := range []plugin.Manifest{publisher, subscriber} {
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
	if err := r.BindWorker("project", "rooms", "1", &storageCommitWorker{handle: func(plugin.WorkerRequest) plugin.WorkerResponse {
		return plugin.WorkerResponse{Protocol: 1, Success: true, Observations: []plugin.ObservationIntent{{Event: "connected", Payload: json.RawMessage(`{"roomId":"room"}`)}}}
	}}); err != nil {
		t.Fatal(err)
	}
	request := plugin.WorkerRequest{Protocol: 1, ProjectID: "project", PluginID: "rooms", Version: "1", Operation: "rooms.join"}
	if response, err := r.Invoke(request); err != nil || !response.Success {
		t.Fatalf("publisher blocked by missing observer worker: %+v %v", response, err)
	}
	if err := storage.Close(); err != nil {
		t.Fatal(err)
	}
	storage, err = plugin.OpenSQLiteStorage(path)
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()
	r = makeRegistry(storage)
	now := time.Now().Add(time.Second)
	if count, err := r.DrainObservations(now, 10); err != nil || count != 1 {
		t.Fatalf("first attempt = %d, %v", count, err)
	}
	calls := 0
	if err := r.BindWorker("project", "analytics", "1", &storageCommitWorker{handle: func(request plugin.WorkerRequest) plugin.WorkerResponse {
		calls++
		var event plugin.Observation
		if err := json.Unmarshal(request.Payload, &event); err != nil || event.ID == "" || event.ProjectID != "project" {
			t.Fatalf("invalid persisted event: %+v %v", event, err)
		}
		return plugin.WorkerResponse{Protocol: 1, Success: true}
	}}); err != nil {
		t.Fatal(err)
	}
	if count, err := r.DrainObservations(now, 10); err != nil || count != 0 {
		t.Fatalf("retry was not delayed: %d %v", count, err)
	}
	if count, err := r.DrainObservations(now.Add(2*time.Second), 10); err != nil || count != 1 || calls != 1 {
		t.Fatalf("retry = %d, %v, calls=%d", count, err, calls)
	}
	if count, err := r.DrainObservations(now.Add(10*time.Second), 10); err != nil || count != 0 {
		t.Fatalf("acknowledged event was delivered again: %d %v", count, err)
	}
	receipts, err := storage.ObservationReceipts("project", "analytics", 10)
	if err != nil || len(receipts) != 1 || receipts[0].Event != "connected" {
		t.Fatalf("receipts = %+v, %v", receipts, err)
	}
	other, err := storage.ObservationReceipts("other", "analytics", 10)
	if err != nil || len(other) != 0 {
		t.Fatalf("receipt leaked across projects: %+v %v", other, err)
	}
}
