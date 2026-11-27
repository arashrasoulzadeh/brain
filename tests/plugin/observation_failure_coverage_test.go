package plugin_test

import (
	"context"
	"database/sql"
	"errors"
	"math"
	"path/filepath"
	"testing"
	"time"

	"brain/internal/plugin"
)

func TestObservationQueueRetryExhaustionAndCorruptStorage(t *testing.T) {
	r := plugin.NewRegistry(1)
	now := time.Now()
	if _, err := r.DrainObservations(now, 0); !errors.Is(err, plugin.ErrInvalidWorkerRequest) {
		t.Fatalf("invalid queue limit: %v", err)
	}
	if n, err := r.DrainObservations(now, 1); err != nil || n != 0 {
		t.Fatalf("empty queue: %d %v", n, err)
	}
	r.RunObservationDelivery(context.Background(), 0)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	r.RunObservationDelivery(ctx, time.Hour)
	path := filepath.Join(t.TempDir(), "events.sqlite")
	storage, err := plugin.OpenSQLiteStorage(path)
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()
	if err := r.BindSQLiteStorage(storage); err != nil {
		t.Fatal(err)
	}
	publisher := plugin.Manifest{ID: "publisher", Version: "1", HostAPI: 1, Events: []plugin.EventSpec{{Name: "changed", Source: "host", Fields: map[string]string{"count": "number", "ready": "boolean"}}}}
	subscriber := plugin.Manifest{ID: "receiver", Version: "1", HostAPI: 1, Capabilities: []string{"observe.publisher.changed"}, Observers: []plugin.ObserverSpec{{Plugin: "publisher", Event: "changed", Function: "receive", Capability: "observe.publisher.changed", Persist: true}}}
	for _, err := range []error{r.Register(publisher), r.Register(subscriber), r.SetReadiness(subscriber.ID, subscriber.Version, true), r.Install("alpha", subscriber.ID, subscriber.Version), r.Grant("alpha", subscriber.ID, subscriber.Version, subscriber.Capabilities[0]), r.Enable("alpha", subscriber.ID, subscriber.Version)} {
		if err != nil {
			t.Fatal(err)
		}
	}
	calls := 0
	if err := r.BindWorker("alpha", subscriber.ID, subscriber.Version, &storageCommitWorker{handle: func(request plugin.WorkerRequest) plugin.WorkerResponse {
		calls++
		return plugin.WorkerResponse{Protocol: request.Protocol, ErrorCode: "unavailable"}
	}}); err != nil {
		t.Fatal(err)
	}
	for _, payload := range []map[string]any{{"count": math.NaN(), "ready": true}, {"count": "one", "ready": true}, {"count": 1, "ready": "true"}, {"count": 1, "ready": nil}} {
		r.PublishHostObservation("alpha", "publisher", "changed", payload)
	}
	r.PublishHostObservationAs("alpha", "publisher", "changed", "invalid", map[string]any{"count": 1, "ready": true})
	r.PublishHostObservation("bad/id", "publisher", "changed", map[string]any{"count": 1, "ready": true})
	if n, err := r.DrainObservations(now.Add(time.Second), 10); err != nil || n != 0 {
		t.Fatalf("invalid events enqueued: %d %v", n, err)
	}
	r.PublishHostObservation("alpha", "publisher", "changed", map[string]any{"count": 1, "ready": true})
	for _, delay := range []time.Duration{time.Second, 3 * time.Second, 6 * time.Second} {
		if n, err := r.DrainObservations(now.Add(delay), 10); err != nil || n != 1 {
			t.Fatalf("attempt at %v: %d %v", delay, n, err)
		}
	}
	if n, err := r.DrainObservations(now.Add(time.Minute), 10); err != nil || n != 0 || calls != 3 {
		t.Fatalf("retry limit bypassed: %d %d %v", n, calls, err)
	}
	if receipts, err := storage.ObservationReceipts("alpha", subscriber.ID, 10); err != nil || len(receipts) != 0 {
		t.Fatalf("failed observation got receipt: %v %v", receipts, err)
	}
	if _, err := storage.ObservationReceipts("alpha", subscriber.ID, 0); !errors.Is(err, plugin.ErrInvalidWorkerRequest) {
		t.Fatalf("invalid receipt limit: %v", err)
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err := db.Exec(`INSERT INTO plugin_observation_receipts VALUES ('alpha','receiver','bad',0,'invalid')`); err != nil {
		t.Fatal(err)
	}
	if _, err := storage.ObservationReceipts("alpha", subscriber.ID, 10); err == nil {
		t.Fatal("corrupt receipt accepted")
	}
	r.PublishHostObservation("alpha", "publisher", "changed", map[string]any{"count": 2, "ready": false})
	if _, err := db.Exec(`UPDATE plugin_observation_queue SET delivery_json='invalid'`); err != nil {
		t.Fatal(err)
	}
	if _, err := r.DrainObservations(now.Add(time.Minute), 10); err == nil {
		t.Fatal("corrupt queue accepted")
	}
	if _, err := db.Exec(`DROP TABLE plugin_observation_queue; DROP TABLE plugin_observation_receipts`); err != nil {
		t.Fatal(err)
	}
	r.PublishHostObservation("alpha", "publisher", "changed", map[string]any{"count": 2, "ready": false})
	if _, err := r.DrainObservations(now.Add(time.Minute), 10); err == nil {
		t.Fatal("missing queue accepted")
	}
	if _, err := storage.ObservationReceipts("alpha", subscriber.ID, 10); err == nil {
		t.Fatal("missing receipts accepted")
	}
}
