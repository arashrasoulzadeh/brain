package plugin

import (
	"errors"
	"testing"
)

func TestCoverageLegacyStoragePlannerAdmission(t *testing.T) {
	s := coverageStorage(t)
	planner := legacyStoragePlanner{storage: s}
	request := coverageRequest("auth.guest", "{}")
	request.PluginID = "auth"
	if _, commit, err := planner.Prepare(request); commit != nil || !errors.Is(err, ErrInvalidWorkerRequest) {
		t.Fatalf("auth admitted without response-aware commit: %v", err)
	}
	request.PluginID = "unknown"
	if _, commit, err := planner.Prepare(request); commit != nil || err != nil {
		t.Fatalf("unsupported operation: %v", err)
	}
	if _, commit, err := prepareLegacyStorageEffect(nil, request); commit != nil || err != nil {
		t.Fatalf("missing adapter: %v", err)
	}
	request.PluginID = "analytics"
	request.Operation = "analytics.capture"
	request.Payload = []byte(`{"projectId":"another","name":"event"}`)
	if _, commit, err := planner.Prepare(request); commit != nil || !errors.Is(err, ErrInvalidWorkerRequest) {
		t.Fatalf("forged analytics identity: %v", err)
	}
	request.Operation = "@observe/rooms/connected"
	if _, commit, err := planner.Prepare(request); commit != nil || !errors.Is(err, ErrInvalidWorkerRequest) {
		t.Fatalf("forged observation: %v", err)
	}
	request.Payload = []byte(`{"projectId":"coverage","pluginId":"rooms","event":"connected","id":"same","payload":{}}`)
	coverageState(t, s, "analytics", "events", map[string]any{"events": []any{map[string]any{"observationId": "same"}}})
	if _, commit, err := planner.Prepare(request); commit != nil || err != nil {
		t.Fatalf("duplicate event was not ignored: %v", err)
	}
}

func TestCoverageLegacyStorageAtomicRollback(t *testing.T) {
	s := coverageStorage(t)
	planner := legacyStoragePlanner{storage: s}
	request := coverageRequest("rooms.create", `{"id":"room"}`)
	request.PluginID = "rooms"
	_, commit, err := planner.prepareWithObservations(request)
	if err != nil || commit == nil {
		t.Fatalf("room admission: %v", err)
	}
	if _, err := s.db.Exec("CREATE TRIGGER reject_room BEFORE INSERT ON plugin_state BEGIN SELECT RAISE(ABORT,'test failure'); END"); err != nil {
		t.Fatal(err)
	}
	delivery := observationDelivery{ID: "event", Target: WorkerRequest{ProjectID: "coverage", PluginID: "observer", Version: "1"}}
	if err := commit([]observationDelivery{delivery}); err == nil {
		t.Fatal("failed room write accepted")
	}
	var count int
	if err := s.db.QueryRow("SELECT COUNT(*) FROM plugin_observation_queue").Scan(&count); err != nil || count != 0 {
		t.Fatalf("failed mutation retained outbox: %d %v", count, err)
	}
	if value, err := s.State("coverage", "rooms", "room"); err != nil || value != nil {
		t.Fatalf("failed mutation retained room: %v %v", value, err)
	}
}
