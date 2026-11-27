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

func TestTableSchemaValidationAndCommitFailures(t *testing.T) {
	s := coverageStorage(t)
	request := func(op, payload string) WorkerRequest {
		return WorkerRequest{ProjectID: "coverage", PluginID: "tables", Operation: op, Payload: []byte(payload)}
	}
	if _, _, err := prepareTableSchema(nil, request("tables.schema.profile", "{}")); !errors.Is(err, ErrStoredResourceMissing) {
		t.Fatal(err)
	}
	for _, payload := range []string{"null", `{}`, `{"table":" "}`, `{"table":"r","offset":-1}`, `{"table":"r","limit":101}`} {
		if _, _, err := prepareTableSchema(s, request("tables.schema.profile", payload)); !errors.Is(err, ErrInvalidWorkerRequest) {
			t.Fatalf("invalid schema profile %s: %v", payload, err)
		}
	}
	_, create, err := prepareTableSchema(s, request("tables.admin.create", `{"table":"r"}`))
	if err != nil || create == nil {
		t.Fatal(err)
	}
	failure := errors.New("injected SQL failure")
	for _, executor := range []coverageSQLExecutor{{execErr: failure}, {resultErr: failure}} {
		if err := create(executor); !errors.Is(err, failure) {
			t.Fatalf("create commit: %v", err)
		}
	}
	if err := create(coverageSQLExecutor{}); !errors.Is(err, ErrModelConflict) {
		t.Fatalf("create conflict: %v", err)
	}
	if err := create(s.db); err != nil {
		t.Fatal(err)
	}
	if _, _, err := prepareTableSchema(s, request("tables.admin.create", `{"table":"r"}`)); !errors.Is(err, ErrModelConflict) {
		t.Fatalf("duplicate create: %v", err)
	}
	if err := s.Append("coverage", "r", map[string]any{"x": "value", "n": nil, "a": []any{1}, "o": map[string]any{"k": true}}); err != nil {
		t.Fatal(err)
	}
	if err := s.Append("coverage", "r", map[string]any{"x": "other"}); err != nil {
		t.Fatal(err)
	}
	raw, _, err := tableSnapshot(s, "coverage", "r")
	if err != nil {
		t.Fatal(err)
	}
	revision := tableRevision("coverage", "r", raw)
	payload := func(column, target string) string {
		p, err := json.Marshal(map[string]any{"table": "r", "revision": revision, "column": column, "new_column": target})
		if err != nil {
			t.Fatal(err)
		}
		return string(p)
	}
	for _, tc := range []struct {
		op, payload string
		want        error
	}{
		{"tables.admin.add-column", `{"table":"r","revision":"stale","column":"x"}`, ErrTableRevision},
		{"tables.admin.add-column", payload("x", ""), ErrModelConflict},
		{"tables.admin.rename-column", payload("x", "x"), ErrInvalidWorkerRequest},
		{"tables.admin.rename-column", payload("n", "x"), ErrModelConflict},
		{"tables.admin.drop-column", payload("absent", ""), ErrStoredResourceMissing},
		{"unknown", `{"table":"r"}`, ErrInvalidWorkerRequest},
	} {
		if _, _, err := prepareTableSchema(s, request(tc.op, tc.payload)); !errors.Is(err, tc.want) {
			t.Fatalf("%s %s: %v", tc.op, tc.payload, err)
		}
	}
	_, migration, err := prepareTableSchema(s, request("tables.admin.rename-column", payload("n", "nullable")))
	if err != nil || migration == nil {
		t.Fatal(err)
	}
	for _, executor := range []coverageSQLExecutor{{execErr: failure}, {resultErr: failure}} {
		if err := migration(executor); !errors.Is(err, failure) {
			t.Fatalf("migration failure: %v", err)
		}
	}
	if err := migration(coverageSQLExecutor{}); !errors.Is(err, ErrTableRevision) {
		t.Fatalf("migration conflict: %v", err)
	}
	if err := migration(s.db); err != nil {
		t.Fatal(err)
	}
	profile, _, err := prepareTableSchema(s, request("tables.schema.profile", `{"table":"r","limit":2}`))
	if err != nil {
		t.Fatal(err)
	}
	var result struct {
		Result struct {
			Columns []tableColumnProfile `json:"columns"`
			Next    int                  `json:"next_offset"`
		} `json:"result"`
	}
	if err := json.Unmarshal(profile.Payload, &result); err != nil || result.Result.Next != 2 || result.Result.Columns[0].Types[0] != "array" || result.Result.Columns[1].Types[0] != "null" {
		t.Fatalf("profile: %s %v", profile.Payload, err)
	}
	if _, err := s.db.Exec("UPDATE plugin_table_rows SET rows_json='[null]' WHERE project_id='coverage' AND table_name='r'"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := prepareTableSchema(s, request("tables.schema.profile", `{"table":"r"}`)); !errors.Is(err, ErrModelSchema) {
		t.Fatalf("corrupt record: %v", err)
	}
}
