package plugin

import (
	"encoding/json"
	"errors"
	"path/filepath"
	"testing"
)

func TestArenaPlannerDurabilityAndStaleWriteProtection(t *testing.T) {
	path := filepath.Join(t.TempDir(), "arena.sqlite")
	storage, err := OpenSQLiteStorage(path)
	if err != nil {
		t.Fatal(err)
	}
	plan := func(operation, payload string) (WorkerRequest, func() error, error) {
		return prepareLegacyStorageRequest(storage, WorkerRequest{Protocol: 1, ProjectID: "alpha", PluginID: "arena", Operation: operation, Payload: []byte(payload)})
	}
	startPayload := `{"roomId":"room","players":["one","two"],"_brainArena":{"state":"finished"}}`
	prepared, commit, err := plan("arena.start", startPayload)
	if err != nil || commit == nil {
		t.Fatalf("start plan: %v", err)
	}
	var input map[string]any
	if json.Unmarshal(prepared.Payload, &input) != nil || input["_brainArena"] != nil {
		t.Fatalf("caller supplied a snapshot: %s", prepared.Payload)
	}
	_, duplicate, err := plan("arena.start", startPayload)
	if err != nil || duplicate == nil {
		t.Fatal(err)
	}
	if err := commit(); err != nil {
		t.Fatal(err)
	}
	if err := duplicate(); !errors.Is(err, ErrModelConflict) {
		t.Fatalf("duplicate start committed: %v", err)
	}
	_, first, err := plan("arena.input", `{"roomId":"room","playerId":"one","sequence":1}`)
	if err != nil || first == nil {
		t.Fatal(err)
	}
	_, stale, err := plan("arena.input", `{"roomId":"room","playerId":"two","sequence":1}`)
	if err != nil || stale == nil {
		t.Fatal(err)
	}
	if err := first(); err != nil {
		t.Fatal(err)
	}
	if err := stale(); !errors.Is(err, ErrModelConflict) {
		t.Fatalf("stale mutation committed: %v", err)
	}
	_, finish, err := plan("arena.result", `{"roomId":"room","winnerIds":["one"],"completedAt":"2026-01-01T00:00:00Z"}`)
	if err != nil || finish == nil {
		t.Fatal(err)
	}
	if err := finish(); err != nil {
		t.Fatal(err)
	}
	if err := storage.Close(); err != nil {
		t.Fatal(err)
	}
	storage, err = OpenSQLiteStorage(path)
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()
	prepared, commit, err = plan("arena.snapshot", `{"roomId":"room"}`)
	if err != nil || commit != nil {
		t.Fatalf("snapshot plan: %v", err)
	}
	var snapshot struct {
		State struct {
			State     string
			Sequences map[string]int
			Result    struct {
				WinnerIDs []string
			}
		} `json:"_brainArena"`
	}
	if err := json.Unmarshal(prepared.Payload, &snapshot); err != nil || snapshot.State.State != "finished" || snapshot.State.Sequences["one"] != 1 || len(snapshot.State.Result.WinnerIDs) != 1 || snapshot.State.Result.WinnerIDs[0] != "one" {
		t.Fatalf("restored snapshot: %s %v", prepared.Payload, err)
	}
	foreign := WorkerRequest{ProjectID: "beta", PluginID: "arena", Operation: "arena.snapshot", Payload: []byte(`{"roomId":"room"}`)}
	if _, _, err := prepareLegacyStorageRequest(storage, foreign); !errors.Is(err, ErrStoredResourceMissing) {
		t.Fatalf("cross-project state disclosed: %v", err)
	}
	forged := WorkerRequest{Protocol: 1, ProjectID: "alpha", PluginID: "arena", Version: "0.1.0", Operation: "arena.snapshot", Payload: []byte(startPayload)}
	if err := validatePayloadProject(forged); !errors.Is(err, ErrInvalidWorkerRequest) {
		t.Fatalf("forged snapshot admitted: %v", err)
	}
	page, err := (legacyStoragePlanner{storage: storage}).AdminInventory("alpha", "arena", "state_keys", "arena.snapshot", "", 10)
	if err != nil || len(page.Rows) != 1 || page.Rows[0].Key != "room" {
		t.Fatalf("inventory: %+v %v", page, err)
	}
}
