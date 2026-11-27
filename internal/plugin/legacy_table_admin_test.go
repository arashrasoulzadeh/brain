package plugin

import (
	"encoding/json"
	"errors"
	"path/filepath"
	"testing"
)

func TestTableAdminPageAndRevisionCheckedMutation(t *testing.T) {
	storage, err := OpenSQLiteStorage(filepath.Join(t.TempDir(), "tables.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()
	for _, projectID := range []string{"alpha", "beta"} {
		for _, name := range []string{"first", "second", "third"} {
			if err := storage.Append(projectID, "messages", map[string]any{"name": name, "project": projectID}); err != nil {
				t.Fatal(err)
			}
		}
	}
	request := WorkerRequest{ProjectID: "alpha", PluginID: "tables", Operation: "tables.read", Payload: []byte(`{"table":"messages","offset":1,"limit":1,"rows":[{"forged":true}]}`)}
	prepared, commit, err := prepareLegacyStorageRequest(storage, request)
	if err != nil || commit != nil {
		t.Fatalf("read page: %v", err)
	}
	var result struct {
		Rows []struct {
			Index int               `json:"row_index"`
			Row   map[string]string `json:"row"`
		} `json:"rows"`
		Page struct {
			Revision string `json:"revision"`
			Total    int    `json:"total_rows"`
			Next     int    `json:"next_offset"`
		} `json:"page"`
	}
	if err := json.Unmarshal(prepared.Payload, &result); err != nil || len(result.Rows) != 1 || result.Rows[0].Index != 1 || result.Rows[0].Row["project"] != "alpha" || result.Page.Total != 3 || result.Page.Next != 2 || result.Page.Revision == "" {
		t.Fatalf("unexpected page: %s (%v)", prepared.Payload, err)
	}
	mutate := func(operation, revision string, index int) WorkerRequest {
		data, err := json.Marshal(map[string]any{"table": "messages", "revision": revision, "row_index": index, "row": map[string]any{"name": "updated"}})
		if err != nil {
			t.Fatal(err)
		}
		return WorkerRequest{ProjectID: "alpha", PluginID: "tables", Operation: operation, Payload: data}
	}
	update := mutate("tables.update", result.Page.Revision, 1)
	_, commit, err = prepareLegacyStorageRequest(storage, update)
	if err != nil || commit == nil {
		t.Fatalf("update plan: %v", err)
	}
	if err := storage.Append("alpha", "messages", map[string]any{"name": "concurrent"}); err != nil {
		t.Fatal(err)
	}
	if err := commit(); !errors.Is(err, ErrTableRevision) {
		t.Fatalf("stale commit: %v", err)
	}
	if _, _, err := prepareLegacyStorageRequest(storage, update); !errors.Is(err, ErrTableRevision) {
		t.Fatalf("stale plan: %v", err)
	}
	encoded, _, err := tableSnapshot(storage, "alpha", "messages")
	if err != nil {
		t.Fatal(err)
	}
	update = mutate("tables.update", tableRevision("alpha", "messages", encoded), 1)
	_, commit, err = prepareLegacyStorageRequest(storage, update)
	if err != nil || commit == nil {
		t.Fatalf("fresh update plan: %v", err)
	}
	if err := commit(); err != nil {
		t.Fatal(err)
	}
	rows, err := storage.Rows("alpha", "messages")
	if err != nil || len(rows) != 4 || rows[1]["name"] != "updated" || rows[2]["name"] != "third" {
		t.Fatalf("updated rows: %+v %v", rows, err)
	}
	encoded, _, err = tableSnapshot(storage, "alpha", "messages")
	if err != nil {
		t.Fatal(err)
	}
	remove := mutate("tables.delete", tableRevision("alpha", "messages", encoded), 1)
	_, commit, err = prepareLegacyStorageRequest(storage, remove)
	if err != nil || commit == nil {
		t.Fatalf("delete plan: %v", err)
	}
	if err := commit(); err != nil {
		t.Fatal(err)
	}
	rows, err = storage.Rows("alpha", "messages")
	if err != nil || len(rows) != 3 || rows[1]["name"] != "third" {
		t.Fatalf("deleted rows: %+v %v", rows, err)
	}
	foreign, err := storage.Rows("beta", "messages")
	if err != nil || len(foreign) != 3 || foreign[1]["name"] != "second" {
		t.Fatalf("foreign rows changed: %+v %v", foreign, err)
	}
}
