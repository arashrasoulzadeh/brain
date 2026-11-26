package plugin

import (
	"encoding/json"
	"errors"
	"path/filepath"
	"testing"
)

func TestLegacyPlannerRejectsForgedSnapshotsForMissingRecords(t *testing.T) {
	storage, err := OpenSQLiteStorage(filepath.Join(t.TempDir(), "storage.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()
	cases := []struct {
		plugin, operation, payload string
	}{
		{"rooms", "rooms.join", `{"id":"one","member":"p","room":{"members":["forged"]}}`},
		{"matchmaking", "matchmaking.assign", `{"id":"one","ticket":{"status":"assigned"}}`},
		{"files", "files.download", `{"key":"one","data":"forged","published":true}`},
		{"files", "files.publish", `{"key":"one","published":true}`},
	}
	for _, tc := range cases {
		t.Run(tc.operation, func(t *testing.T) {
			request := WorkerRequest{ProjectID: "alpha", PluginID: tc.plugin, Operation: tc.operation, Payload: []byte(tc.payload)}
			_, commit, err := prepareLegacyStorageRequest(storage, request)
			if !errors.Is(err, ErrStoredResourceMissing) || commit != nil {
				t.Fatalf("forged snapshot accepted: %v", err)
			}
		})
	}
}

func TestLegacyPlannerNormalizesKeysAndUsesDurableState(t *testing.T) {
	storage, err := OpenSQLiteStorage(filepath.Join(t.TempDir(), "storage.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()
	request := WorkerRequest{ProjectID: "alpha", PluginID: "rooms", Operation: "rooms.create", Payload: []byte(`{"id":" one ","room":{"members":["forged"]}}`)}
	_, commit, err := prepareLegacyStorageRequest(storage, request)
	if err != nil || commit == nil {
		t.Fatalf("create plan: %v", err)
	}
	if state, err := storage.State("alpha", "rooms", "one"); err != nil || state != nil {
		t.Fatalf("prepare wrote state: %v %v", state, err)
	}
	if err := commit(); err != nil {
		t.Fatal(err)
	}
	request.Operation = "rooms.join"
	request.Payload = []byte(`{"id":"one","member":" player ","room":{"members":["forged"]}}`)
	prepared, commit, err := prepareLegacyStorageRequest(storage, request)
	if err != nil || commit == nil {
		t.Fatalf("join plan: %v", err)
	}
	var payload struct {
		Member string
		Room   struct {
			Members []string
		}
	}
	if err := json.Unmarshal(prepared.Payload, &payload); err != nil || payload.Member != "player" || len(payload.Room.Members) != 1 || payload.Room.Members[0] != "player" {
		t.Fatalf("untrusted state: %s %v", prepared.Payload, err)
	}
	request.ProjectID = "beta"
	if _, _, err := prepareLegacyStorageRequest(storage, request); !errors.Is(err, ErrStoredResourceMissing) {
		t.Fatalf("cross-project access: %v", err)
	}
}

func TestLegacySchemaReadsDurableRows(t *testing.T) {
	storage, err := OpenSQLiteStorage(filepath.Join(t.TempDir(), "storage.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()
	if err := storage.Append("alpha", "messages", map[string]any{"id": "one"}); err != nil {
		t.Fatal(err)
	}
	request := WorkerRequest{ProjectID: "alpha", PluginID: "tables", Operation: "tables.schema", Payload: []byte(`{"table":"messages","rows":[{"forged":true}]}`)}
	prepared, commit, err := prepareLegacyStorageRequest(storage, request)
	if err != nil || commit != nil {
		t.Fatalf("schema plan: %v", err)
	}
	var payload struct {
		Rows []map[string]any
	}
	if err := json.Unmarshal(prepared.Payload, &payload); err != nil || len(payload.Rows) != 1 || payload.Rows[0]["id"] != "one" {
		t.Fatalf("schema snapshot: %s %v", prepared.Payload, err)
	}
}
