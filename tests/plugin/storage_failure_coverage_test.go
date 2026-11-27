package plugin_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"math"
	"path/filepath"
	"testing"

	"brain/internal/plugin"
)

func TestLegacyStorageLifecycleAndInvalidInputs(t *testing.T) {
	storage, err := plugin.OpenSQLiteStorage(filepath.Join(t.TempDir(), "legacy.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()
	registry := plugin.NewRegistry(1)
	if err := registry.BindSQLiteStorage(storage); err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"files", "matchmaking", "rooms", "analytics", "tables"} {
		m := plugin.Manifest{ID: id, Version: "1", HostAPI: 1, Capabilities: []string{"use"}}
		operations := map[string][]string{
			"files":       {"upload", "publish", "download", "metadata"},
			"matchmaking": {"queue", "cancel", "assign", "status"},
			"rooms":       {"create", "join", "reconnect", "result"},
			"analytics":   {"capture"},
			"tables":      {"write", "read", "schema"},
		}
		for _, operation := range operations[id] {
			m.Routes = append(m.Routes, plugin.Route{Operation: id + "." + operation, Function: operation, Capability: "use"})
		}
		for _, err := range []error{registry.Register(m), registry.SetReadiness(id, "1", true), registry.Install("alpha", id, "1"), registry.Grant("alpha", id, "1", "use"), registry.Enable("alpha", id, "1")} {
			if err != nil {
				t.Fatal(err)
			}
		}
		if err := registry.BindWorker("alpha", id, "1", &storageCommitWorker{handle: func(request plugin.WorkerRequest) plugin.WorkerResponse {
			return plugin.WorkerResponse{Protocol: 1, Success: true, Payload: request.Payload}
		}}); err != nil {
			t.Fatal(err)
		}
	}
	invoke := func(id, operation, payload string) (map[string]any, error) {
		response, err := registry.Invoke(plugin.WorkerRequest{Protocol: 1, ProjectID: "alpha", PluginID: id, Version: "1", Operation: id + "." + operation, Payload: []byte(payload)})
		if err != nil {
			return nil, err
		}
		var value map[string]any
		return value, json.Unmarshal(response.Payload, &value)
	}
	for _, tc := range []struct{ id, operation, payload string }{
		{"files", "upload", `{}`}, {"files", "upload", `{"key":"one"}`}, {"files", "upload", `null`}, {"rooms", "create", `{"id":" "}`}, {"tables", "write", `{"table":"one","row":false}`},
	} {
		if _, err := invoke(tc.id, tc.operation, tc.payload); !errors.Is(err, plugin.ErrInvalidWorkerRequest) {
			t.Fatalf("invalid %s: %v", tc.operation, err)
		}
	}
	for _, tc := range []struct{ id, operation, payload string }{
		{"files", "download", `{"key":"missing"}`}, {"matchmaking", "status", `{"id":"missing"}`}, {"rooms", "result", `{"id":"missing"}`},
	} {
		if _, err := invoke(tc.id, tc.operation, tc.payload); !errors.Is(err, plugin.ErrStoredResourceMissing) {
			t.Fatalf("missing %s: %v", tc.operation, err)
		}
	}
	for _, tc := range []struct{ id, operation, payload string }{
		{"files", "upload", `{"key":"one","data":"content"}`}, {"files", "publish", `{"key":"one"}`}, {"files", "download", `{"key":"one","data":"forged"}`},
		{"matchmaking", "queue", `{"id":"ticket"}`}, {"matchmaking", "cancel", `{"id":"ticket"}`}, {"matchmaking", "assign", `{"id":"ticket"}`}, {"matchmaking", "status", `{"id":"ticket"}`},
		{"rooms", "create", `{"id":"room"}`}, {"rooms", "join", `{"id":"room","member":"player"}`}, {"rooms", "reconnect", `{"id":"room","member":"player"}`}, {"rooms", "result", `{"id":"room","result":{"winner":"player"}}`},
		{"analytics", "capture", `{"name":"example","fields":{"score":1}}`},
	} {
		if _, err := invoke(tc.id, tc.operation, tc.payload); err != nil {
			t.Fatalf("%s.%s: %v", tc.id, tc.operation, err)
		}
	}
	if _, err := invoke("rooms", "join", `{"id":"room","member":" "}`); !errors.Is(err, plugin.ErrInvalidWorkerRequest) {
		t.Fatalf("empty member: %v", err)
	}
	if data, published, err := storage.File("alpha", "one"); err != nil || data != "content" || !published {
		t.Fatalf("file: %q %t %v", data, published, err)
	}
	if state, err := storage.State("alpha", "matchmaking", "ticket"); err != nil || state["status"] != "assigned" {
		t.Fatalf("ticket: %v %v", state, err)
	}
	if state, err := storage.State("alpha", "rooms", "room"); err != nil || len(state["members"].([]any)) != 1 || state["result"] == nil {
		t.Fatalf("room: %v %v", state, err)
	}
	if state, err := storage.State("alpha", "analytics", "events"); err != nil || len(state["events"].([]any)) != 1 {
		t.Fatalf("analytics: %v %v", state, err)
	}
	if err := storage.Close(); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ id, operation, payload string }{
		{"files", "download", `{"key":"one"}`}, {"matchmaking", "status", `{"id":"ticket"}`}, {"rooms", "result", `{"id":"room"}`}, {"analytics", "capture", `{"name":"example"}`}, {"tables", "read", `{"table":"one"}`}, {"tables", "write", `{"table":"one","row":{}}`},
	} {
		if _, err := invoke(tc.id, tc.operation, tc.payload); err == nil {
			t.Fatalf("closed storage accepted %s.%s", tc.id, tc.operation)
		}
	}
}

func TestSQLiteStorageRejectsCorruptAndUnavailableData(t *testing.T) {
	if _, err := plugin.OpenSQLiteStorage(filepath.Join(t.TempDir(), "absent", "db")); err == nil {
		t.Fatal("missing parent accepted")
	}
	path := filepath.Join(t.TempDir(), "db")
	storage, err := plugin.OpenSQLiteStorage(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := storage.PutState("alpha", "test", "key", map[string]any{"bad": math.Inf(1)}); err == nil {
		t.Fatal("non-JSON state accepted")
	}
	if err := storage.Append("alpha", "", map[string]any{}); !errors.Is(err, plugin.ErrStorageTableRequired) {
		t.Fatalf("empty table: %v", err)
	}
	if err := storage.Append("alpha", "table", map[string]any{"bad": math.NaN()}); err == nil {
		t.Fatal("non-JSON row accepted")
	}
	if err := storage.PutState("", "registry", "lifecycle", map[string]any{"installations": "bad"}); err != nil {
		t.Fatal(err)
	}
	if _, _, err := storage.LoadRegistryState(); err == nil {
		t.Fatal("invalid registry state accepted")
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err := db.Exec(`INSERT INTO plugin_state VALUES ('alpha','test','bad','invalid')`); err != nil {
		t.Fatal(err)
	}
	if _, err := storage.State("alpha", "test", "bad"); err == nil {
		t.Fatal("corrupt state accepted")
	}
	if err := storage.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := storage.Rows("alpha", "table"); err == nil {
		t.Fatal("closed storage read accepted")
	}
	if _, err := storage.State("alpha", "test", "key"); err == nil {
		t.Fatal("closed state read accepted")
	}
	if storage.CheckHealth(context.Background()) == nil {
		t.Fatal("closed storage was healthy")
	}
	var absent *plugin.SQLiteStorage
	if absent.Close() != nil || absent.CheckHealth(context.Background()) == nil {
		t.Fatal("nil storage health or close failed")
	}
}
