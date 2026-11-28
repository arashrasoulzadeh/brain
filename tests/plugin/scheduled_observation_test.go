package plugin_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"brain/internal/cortex"
	"brain/internal/execution"
	"brain/internal/plugin"
)

func TestScheduledObservationSurvivesNeuronOutage(t *testing.T) {
	storage, err := plugin.OpenSQLiteStorage(filepath.Join(t.TempDir(), "observations.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()
	publisher := plugin.Manifest{ID: "publisher", Version: "1", HostAPI: 1, Capabilities: []string{"read"},
		Routes: []plugin.Route{{Operation: "read", Function: "read", Capability: "read"}},
		Events: []plugin.EventSpec{{Name: "completed", Source: "host", Fields: map[string]string{"value": "string"}}}}
	subscriber := plugin.Manifest{ID: "subscriber", Version: "1", HostAPI: 1, Entry: "index.ts",
		SourceFiles: []string{"index.ts"}, Capabilities: []string{"observe.publisher.completed"},
		Observers: []plugin.ObserverSpec{{Plugin: "publisher", Event: "completed", Function: "record", Capability: "observe.publisher.completed", Persist: true}}}
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "index.ts"), []byte("export function record() { return {}; }"), 0o600); err != nil {
		t.Fatal(err)
	}
	catalog, err := execution.NewCatalog(8, 1<<20)
	if err != nil {
		t.Fatal(err)
	}
	runtime := execution.Runtime{Kind: plugin.ScheduledTypeScriptKind, Profile: "test"}
	artifact, err := plugin.RegisterScheduledArtifact(catalog, runtime, subscriber, root)
	if err != nil {
		t.Fatal(err)
	}
	scheduler, err := cortex.New(catalog, cortex.Config{MaxNeurons: 4, MaxProjectRunning: 2, HeartbeatTTL: time.Minute})
	if err != nil {
		t.Fatal(err)
	}
	registry := plugin.NewRegistry(1)
	for _, m := range []plugin.Manifest{publisher, subscriber} {
		if err := registry.Register(m); err != nil {
			t.Fatal(err)
		}
	}
	if err := registry.BindSQLiteStorage(storage); err != nil {
		t.Fatal(err)
	}
	for _, m := range []plugin.Manifest{publisher, subscriber} {
		if err := registry.Install("project", m.ID, m.Version); err != nil {
			t.Fatal(err)
		}
		if err := registry.SetReadiness(m.ID, m.Version, true); err != nil {
			t.Fatal(err)
		}
		for _, capability := range m.Capabilities {
			if err := registry.Grant("project", m.ID, m.Version, capability); err != nil {
				t.Fatal(err)
			}
		}
		if err := registry.Enable("project", m.ID, m.Version); err != nil {
			t.Fatal(err)
		}
	}
	worker, err := artifact.Worker("project", scheduler)
	if err != nil {
		t.Fatal(err)
	}
	if err := worker.Start(); err != nil {
		t.Fatal(err)
	}
	defer worker.Stop()
	if err := registry.BindWorker("project", subscriber.ID, subscriber.Version, worker); err != nil {
		t.Fatal(err)
	}
	registry.PublishHostObservationAs("project", publisher.ID, "completed", "operator:root", map[string]any{"value": "test"})
	now := time.Now().Add(time.Second)
	// Outage exceeds the ordinary three-attempt failure budget.
	for i := 0; i < 5; i++ {
		count, err := registry.DrainObservations(now.Add(time.Duration(i)*6*time.Second), 10)
		if err != nil || count != 1 {
			t.Fatalf("outage drain %d: %d, %v", i, count, err)
		}
	}
	if count, err := registry.DrainObservations(now.Add(24*time.Second), 10); err != nil || count != 0 {
		t.Fatalf("capacity deferral did not delay delivery: %d, %v", count, err)
	}
	calls := 0
	session, err := scheduler.Register(cortex.Registration{ID: "neuron-1", Slots: 1, Runtimes: []execution.Runtime{runtime}, Executor: scheduledExecutor(func(_ context.Context, req execution.Request) (execution.Result, error) {
		calls++
		decoded, err := plugin.DecodeScheduledRequest(req, subscriber)
		if err != nil {
			return execution.Result{}, err
		}
		if req.Auth.OriginalCaller != "operator:root" || req.Auth.Caller != "plugin:publisher@1" || req.Auth.Resource != "publisher.completed" {
			t.Errorf("incorrect observer identity: %+v", req.Auth)
		}
		var event plugin.Observation
		if err := json.Unmarshal(decoded.Payload, &event); err != nil || event.ProjectID != "project" || event.Event != "completed" {
			t.Errorf("incorrect observation: %+v, %v", event, err)
		}
		return execution.Result{ID: req.ID, WorkerID: "neuron-1", Status: execution.Succeeded, Output: []byte("{}")}, nil
	})})
	if err != nil {
		t.Fatal(err)
	}
	defer session.Disconnect()
	if count, err := registry.DrainObservations(now.Add(30*time.Second), 10); err != nil || count != 1 || calls != 1 {
		t.Fatalf("recovered delivery: %d, %v, calls=%d", count, err, calls)
	}
	receipts, err := storage.ObservationReceipts("project", subscriber.ID, 10)
	if err != nil || len(receipts) != 1 || receipts[0].Event != "completed" {
		t.Fatalf("receipts: %+v, %v", receipts, err)
	}
	if count, err := registry.DrainObservations(now.Add(time.Minute), 10); err != nil || count != 0 || calls != 1 {
		t.Fatalf("acknowledged observation replayed: %d, %v, calls=%d", count, err, calls)
	}
}
