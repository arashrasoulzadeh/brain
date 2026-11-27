package plugin

import (
	"encoding/json"
	"errors"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestLegacyStorageReadWriteContracts(t *testing.T) {
	storage, err := OpenSQLiteStorage(filepath.Join(t.TempDir(), "storage.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()
	plan := func(plugin, operation, payload string, writes bool) map[string]any {
		t.Helper()
		req := WorkerRequest{ProjectID: "alpha", PluginID: plugin, Operation: operation, Payload: []byte(payload)}
		prepared, commit, err := prepareLegacyStorageRequest(storage, req)
		if err != nil || (commit != nil) != writes {
			t.Fatalf("%s: commit=%v err=%v", operation, commit != nil, err)
		}
		if commit != nil {
			if err := commit(); err != nil {
				t.Fatal(err)
			}
		}
		var result map[string]any
		if err := json.Unmarshal(prepared.Payload, &result); err != nil {
			t.Fatal(err)
		}
		return result
	}
	plan("files", "files.upload", `{"key":"file","data":"hello","published":true}`, true)
	data, published, err := storage.File("alpha", "file")
	if err != nil || data != "hello" || published {
		t.Fatalf("upload: %q %v %v", data, published, err)
	}
	plan("files", "files.publish", `{"key":"file","data":"forged"}`, true)
	for _, op := range []string{"files.download", "files.metadata"} {
		result := plan("files", op, `{"key":"file","data":"forged","published":false}`, false)
		if result["data"] != "hello" || result["published"] != true {
			t.Fatalf("%s: %v", op, result)
		}
	}
	plan("matchmaking", "matchmaking.queue", `{"id":"ticket"}`, true)
	for _, tc := range []struct{ op, status string }{{"matchmaking.assign", "assigned"}, {"matchmaking.cancel", "cancelled"}} {
		plan("matchmaking", tc.op, `{"id":"ticket","ticket":{"status":"forged"}}`, true)
		result := plan("matchmaking", "matchmaking.status", `{"id":"ticket"}`, false)
		if result["ticket"].(map[string]any)["status"] != tc.status {
			t.Fatalf("ticket status: %v", result)
		}
	}
	plan("rooms", "rooms.create", `{"id":"room"}`, true)
	plan("rooms", "rooms.join", `{"id":"room","member":"player"}`, true)
	result := plan("rooms", "rooms.reconnect", `{"id":"room","member":"player"}`, true)
	if !reflect.DeepEqual(result["room"].(map[string]any)["members"], []any{"player"}) {
		t.Fatalf("reconnect duplicated member: %v", result)
	}
	plan("rooms", "rooms.result", `{"id":"room","result":{"score":5}}`, true)
	result = plan("rooms", "rooms.result", `{"id":"room"}`, false)
	if !reflect.DeepEqual(result["room"].(map[string]any)["result"], map[string]any{"score": float64(5)}) {
		t.Fatalf("room result: %v", result)
	}
	plan("tables", "tables.write", `{"table":"messages","row":{"id":"one"}}`, true)
	result = plan("tables", "tables.read", `{"table":"messages","rows":[]}`, false)
	if !reflect.DeepEqual(result["rows"], []any{map[string]any{"id": "one"}}) {
		t.Fatalf("table snapshot: %v", result)
	}
	plan("analytics", "analytics.capture", `{"projectId":"alpha","name":"click","fields":{}}`, true)
	event := `{"id":"ob-1","projectId":"alpha","pluginId":"rooms","event":"connected","payload":{"member":"player"}}`
	plan("analytics", "@observe/rooms/connected", event, true)
	plan("analytics", "@observe/rooms/connected", event, false)
	state, err := storage.State("alpha", "analytics", "events")
	if err != nil || len(state["events"].([]any)) != 2 {
		t.Fatalf("observation deduplication: %v %v", state, err)
	}
}

func TestLegacyStorageInvalidAndUnavailableRequests(t *testing.T) {
	storage, err := OpenSQLiteStorage(filepath.Join(t.TempDir(), "storage.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ plugin, op, payload string }{
		{"files", "files.upload", `null`},
		{"files", "files.upload", `{"key":" "}`},
		{"files", "files.upload", `{"key":"one","data":false}`},
		{"tables", "tables.write", `{"table":"one","row":null}`},
		{"analytics", "analytics.capture", `{"projectId":"forged"}`},
		{"analytics", "@observe/rooms/connected", `{"projectId":"alpha","pluginId":"forged"}`},
	} {
		req := WorkerRequest{ProjectID: "alpha", PluginID: tc.plugin, Operation: tc.op, Payload: []byte(tc.payload)}
		if _, commit, err := prepareLegacyStorageRequest(storage, req); !errors.Is(err, ErrInvalidWorkerRequest) || commit != nil {
			t.Fatalf("%s invalid request: %v", tc.op, err)
		}
	}
	if err := storage.PutState("alpha", "rooms", "room", map[string]any{"members": []any{}}); err != nil {
		t.Fatal(err)
	}
	req := WorkerRequest{ProjectID: "alpha", PluginID: "rooms", Operation: "rooms.join", Payload: []byte(`{"id":"room","member":""}`)}
	if _, _, err := prepareLegacyStorageRequest(storage, req); !errors.Is(err, ErrInvalidWorkerRequest) {
		t.Fatalf("missing member: %v", err)
	}
	for _, id := range []string{"files", "matchmaking", "rooms", "analytics", "tables"} {
		if !(legacyStoragePlanner{storage: storage}).Supports(id) {
			t.Fatalf("missing adapter for %s", id)
		}
	}
	if (legacyStoragePlanner{storage: storage}).Supports("other") {
		t.Fatal("unknown plugin supported")
	}
	if got, commit, err := prepareLegacyStorageRequest(nil, req); err != nil || commit != nil || !reflect.DeepEqual(got, req) {
		t.Fatalf("nil adapter changed request: %v", err)
	}
	if err := storage.Close(); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ plugin, op, payload string }{
		{"files", "files.download", `{"key":"one"}`},
		{"matchmaking", "matchmaking.status", `{"id":"one"}`},
		{"rooms", "rooms.result", `{"id":"one"}`},
		{"tables", "tables.read", `{"table":"one"}`},
		{"analytics", "analytics.capture", `{"name":"click"}`},
	} {
		req := WorkerRequest{ProjectID: "alpha", PluginID: tc.plugin, Operation: tc.op, Payload: []byte(tc.payload)}
		if _, commit, err := prepareLegacyStorageRequest(storage, req); err == nil || commit != nil {
			t.Fatalf("%s ignored database failure", tc.op)
		}
	}
}

func TestModelStorageValidationAndCorruption(t *testing.T) {
	storage, err := OpenSQLiteStorage(filepath.Join(t.TempDir(), "models.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()
	spec := ModelSpec{Version: 1, Fields: map[string]string{"name": "string", "count": "integer", "active": "boolean"}}
	valid := map[string]any{"name": "example", "count": float64(2), "active": true}
	if !modelFieldsValid(spec, valid) {
		t.Fatal("valid fields rejected")
	}
	for _, fields := range []map[string]any{
		{"name": "example"},
		{"name": true, "count": float64(2), "active": true},
		{"name": strings.Repeat("x", 1025), "count": float64(2), "active": true},
		{"name": "example", "count": float64(2.5), "active": true},
		{"name": "example", "count": float64(9007199254740992), "active": true},
		{"name": "example", "count": float64(2), "active": "true"},
		{"name": "example", "count": float64(2), "active": nil},
	} {
		if modelFieldsValid(spec, fields) {
			t.Fatalf("invalid model accepted: %v", fields)
		}
	}
	req := WorkerRequest{ProjectID: "alpha", PluginID: "sample", Payload: []byte(`{"id":"record","fields":{"name":"example","count":2,"active":true}}`)}
	created, err := storage.executeModel(req, spec, ModelRoute{Name: "example", Action: "create"})
	if err != nil {
		t.Fatal(err)
	}
	record := created.(modelRecord)
	if record.ID != "record" || record.Revision == "" {
		t.Fatalf("invalid record: %v", record)
	}
	if _, err := storage.executeModel(req, spec, ModelRoute{Name: "example", Action: "create"}); !errors.Is(err, ErrModelConflict) {
		t.Fatalf("duplicate create: %v", err)
	}
	for _, tc := range []struct{ payload, action string }{
		{`{`, "read"},
		{`{"id":"bad id"}`, "read"},
		{`{"after":"bad id"}`, "list"},
		{`{"id":"other","fields":{}}`, "create"},
	} {
		req.Payload = []byte(tc.payload)
		if _, err := storage.executeModel(req, spec, ModelRoute{Name: "example", Action: tc.action}); !errors.Is(err, ErrInvalidWorkerRequest) {
			t.Fatalf("invalid %s: %v", tc.action, err)
		}
	}
	req.Payload = []byte(`{"id":"record","revision":"stale"}`)
	if _, err := storage.executeModel(req, spec, ModelRoute{Name: "example", Action: "delete"}); !errors.Is(err, ErrModelConflict) {
		t.Fatalf("stale revision: %v", err)
	}
	req.Payload = []byte(`{"id":"record"}`)
	for _, tc := range []struct {
		version int
		fields  string
		want    error
	}{
		{2, `{}`, ErrModelSchema},
		{1, `{}`, ErrModelSchema},
		{1, `not-json`, nil},
	} {
		if _, err := storage.db.Exec("UPDATE plugin_model_records SET schema_version=?,fields_json=? WHERE record_id=?", tc.version, tc.fields, "record"); err != nil {
			t.Fatal(err)
		}
		for _, action := range []string{"read", "list"} {
			_, err := storage.executeModel(req, spec, ModelRoute{Name: "example", Action: action})
			if err == nil || (tc.want != nil && !errors.Is(err, tc.want)) {
				t.Fatalf("corrupt model %s accepted: %v", action, err)
			}
		}
	}
	if err := storage.Close(); err != nil {
		t.Fatal(err)
	}
	for _, action := range []string{"read", "list", "update"} {
		if _, err := storage.executeModel(req, spec, ModelRoute{Name: "example", Action: action}); err == nil {
			t.Fatalf("closed model storage accepted %s", action)
		}
	}
}
