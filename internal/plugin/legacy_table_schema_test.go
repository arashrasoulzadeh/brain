package plugin

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestTableSchemaProfileAndAtomicColumnMigration(t *testing.T) {
	s := coverageStorage(t)
	plan := func(project, op string, fields map[string]any) (map[string]any, func() error, error) {
		fields["table"] = "r"
		payload, err := json.Marshal(fields)
		if err != nil {
			t.Fatal(err)
		}
		prepared, commit, err := prepareLegacyStorageRequest(s, WorkerRequest{ProjectID: project, PluginID: "tables", Operation: op, Payload: payload})
		if err != nil {
			return nil, nil, err
		}
		var envelope struct {
			Result map[string]any `json:"result"`
		}
		if err := json.Unmarshal(prepared.Payload, &envelope); err != nil {
			t.Fatal(err)
		}
		return envelope.Result, commit, nil
	}
	run := func(project, op string, fields map[string]any) map[string]any {
		t.Helper()
		result, commit, err := plan(project, op, fields)
		if err != nil {
			t.Fatal(err)
		}
		if commit != nil {
			if err := commit(); err != nil {
				t.Fatal(err)
			}
		}
		return result
	}
	for _, project := range []string{"alpha", "beta"} {
		run(project, "tables.admin.create", map[string]any{})
		for _, row := range []map[string]any{{"name": "one", "value": 1}, {"name": "two", "value": true}, {"name": "three"}} {
			if err := s.Append(project, "r", row); err != nil {
				t.Fatal(err)
			}
		}
	}
	profile := run("alpha", "tables.schema.profile", map[string]any{"limit": 1, "offset": 1})
	column := profile["columns"].([]any)[0].(map[string]any)
	if column["name"] != "value" || column["present"] != float64(2) || column["missing"] != float64(1) || len(column["types"].([]any)) != 2 {
		t.Fatalf("profile: %+v", profile)
	}
	revision := profile["revision"]
	if _, _, err := plan("alpha", "tables.admin.rename-column", map[string]any{"revision": revision, "column": "name", "new_column": "value"}); !errors.Is(err, ErrModelConflict) {
		t.Fatalf("collision: %v", err)
	}
	rename := map[string]any{"revision": revision, "column": "name", "new_column": "label"}
	_, stale, err := plan("alpha", "tables.admin.rename-column", rename)
	if err != nil {
		t.Fatal(err)
	}
	changed := run("alpha", "tables.admin.rename-column", rename)
	if err := stale(); !errors.Is(err, ErrTableRevision) {
		t.Fatalf("stale migration: %v", err)
	}
	added := run("alpha", "tables.admin.add-column", map[string]any{"revision": changed["revision"], "column": "count", "default": 0})
	run("alpha", "tables.admin.drop-column", map[string]any{"revision": added["revision"], "column": "value"})
	rows, err := s.Rows("alpha", "r")
	if err != nil || len(rows) != 3 || rows[0]["label"] != "one" || rows[0]["count"] != float64(0) || len(rows[0]) != 2 {
		t.Fatalf("rows: %+v %v", rows, err)
	}
	foreign, err := s.Rows("beta", "r")
	if err != nil || foreign[0]["name"] != "one" || foreign[0]["value"] != float64(1) {
		t.Fatalf("foreign rows: %+v %v", foreign, err)
	}
}
