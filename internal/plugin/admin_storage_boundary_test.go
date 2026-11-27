package plugin

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestTableAdminValidationAndResponseCeiling(t *testing.T) {
	s := coverageStorage(t)
	request := func(op, payload string) WorkerRequest {
		return WorkerRequest{ProjectID: "coverage", PluginID: "tables", Operation: op, Payload: []byte(payload)}
	}
	if _, _, err := prepareTableAdmin(nil, request("tables.read", "{}")); !errors.Is(err, ErrStoredResourceMissing) {
		t.Fatal(err)
	}
	for _, payload := range []string{
		"null", "{", "{}", `{"table":null}`, `{"table":3}`, `{"table":" "}`, `{"table":"bad\u0000name"}`,
		`{"table":"r","offset":-1}`, `{"table":"r","offset":"oops"}`, `{"table":"r","offset":true}`,
		`{"table":"r","offset":"-1"}`, `{"table":"r","limit":101}`, `{"table":"r","limit":null}`,
	} {
		if _, effect, err := prepareTableAdmin(s, request("tables.read", payload)); !errors.Is(err, ErrInvalidWorkerRequest) || effect != nil {
			t.Fatalf("invalid table page %s: %v", payload, err)
		}
	}
	empty, effect, err := prepareTableAdmin(s, request("tables.read", `{"table":"empty","limit":"1","offset":"0"}`))
	if err != nil || effect != nil || !strings.Contains(string(empty.Payload), `"total_rows":0`) {
		t.Fatalf("empty page: %s %v", empty.Payload, err)
	}
	for _, value := range []string{strings.Repeat("x", 120*1024), strings.Repeat("y", 120*1024), strings.Repeat("z", 210*1024)} {
		if err := s.Append("coverage", "r", map[string]any{"value": value}); err != nil {
			t.Fatal(err)
		}
	}
	page, effect, err := prepareTableAdmin(s, request("tables.read", `{"table":"r","limit":100}`))
	if err != nil || effect != nil {
		t.Fatal(err)
	}
	var result struct {
		Rows []json.RawMessage `json:"rows"`
		Page struct {
			Next     int    `json:"next_offset"`
			Revision string `json:"revision"`
		} `json:"page"`
	}
	if err := json.Unmarshal(page.Payload, &result); err != nil || len(result.Rows) != 1 || result.Page.Next != 1 {
		t.Fatalf("page ceiling: rows=%d next=%d err=%v", len(result.Rows), result.Page.Next, err)
	}
	if _, _, err := prepareTableAdmin(s, request("tables.read", `{"table":"r","offset":2}`)); !errors.Is(err, ErrInvalidWorkerRequest) {
		t.Fatalf("oversized first row: %v", err)
	}
	for _, row := range []string{"null", "[]", "false"} {
		payload := `{"table":"r","revision":"` + result.Page.Revision + `","row_index":0,"row":` + row + "}"
		if _, effect, err := prepareTableAdmin(s, request("tables.update", payload)); !errors.Is(err, ErrInvalidWorkerRequest) || effect != nil {
			t.Fatalf("invalid replacement %s: %v", row, err)
		}
	}
	for _, index := range []string{"-1", "3", "null", "1.5"} {
		payload := `{"table":"r","revision":"` + result.Page.Revision + `","row_index":` + index + "}"
		if _, effect, err := prepareTableAdmin(s, request("tables.delete", payload)); !errors.Is(err, ErrInvalidWorkerRequest) || effect != nil {
			t.Fatalf("invalid index %s: %v", index, err)
		}
	}
	payload := `{"table":"r","revision":"` + result.Page.Revision + `","row_index":0}`
	if _, _, err := prepareTableAdmin(s, request("unknown", payload)); !errors.Is(err, ErrInvalidWorkerRequest) {
		t.Fatal(err)
	}
	_, remove, err := prepareTableAdmin(s, request("tables.delete", payload))
	if err != nil || remove == nil {
		t.Fatal(err)
	}
	failure := errors.New("injected SQL failure")
	for _, executor := range []coverageSQLExecutor{{execErr: failure}, {resultErr: failure}} {
		if err := remove(executor); !errors.Is(err, failure) {
			t.Fatalf("delete commit error: %v", err)
		}
	}
	if err := remove(coverageSQLExecutor{}); !errors.Is(err, ErrTableRevision) {
		t.Fatalf("delete conflict: %v", err)
	}
	if _, err := s.db.Exec("UPDATE plugin_table_rows SET rows_json='invalid' WHERE project_id='coverage' AND table_name='r'"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := prepareTableAdmin(s, request("tables.read", `{"table":"r"}`)); err == nil {
		t.Fatal("corrupt table snapshot accepted")
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	if _, _, err := prepareTableAdmin(s, request("tables.read", `{"table":"r"}`)); err == nil {
		t.Fatal("closed table store accepted")
	}
}

func TestAnalyticsHistoryValidationAndResponseCeiling(t *testing.T) {
	s := coverageStorage(t)
	request := func(op, payload string) WorkerRequest {
		return WorkerRequest{ProjectID: "coverage", PluginID: "analytics", Operation: op, Payload: []byte(payload)}
	}
	if _, _, err := prepareAnalyticsRead(nil, request("analytics.history", "{}")); !errors.Is(err, ErrStoredResourceMissing) {
		t.Fatal(err)
	}
	for _, payload := range []string{"null", "{", `{}`, `{"name":"!"}`, `{"name":"event","from":"invalid"}`, `{"name":"event","limit":21}`, `{"name":"event","offset":"bad"}`} {
		if _, effect, err := prepareAnalyticsRead(s, request("analytics.history", payload)); !errors.Is(err, ErrInvalidWorkerRequest) || effect != nil {
			t.Fatalf("invalid analytics query %s: %v", payload, err)
		}
	}
	event := func(name string, size int) map[string]any {
		return map[string]any{"projectId": "coverage", "name": name, "fields": map[string]any{"data": strings.Repeat("x", size)}}
	}
	coverageState(t, s, "analytics", "events", map[string]any{"events": []any{
		event("large", 210*1024), event("event", 120*1024), event("event", 120*1024), event("event", 0), event("other", 0),
	}})
	page, effect, err := prepareAnalyticsRead(s, request("analytics.history", `{"name":"event","limit":"20"}`))
	if err != nil || effect != nil {
		t.Fatal(err)
	}
	var envelope struct {
		Result struct {
			Total   int               `json:"total_events"`
			Next    int               `json:"next_offset"`
			Undated int               `json:"undated_events"`
			Rows    []json.RawMessage `json:"rows"`
		} `json:"result"`
	}
	if err := json.Unmarshal(page.Payload, &envelope); err != nil || len(envelope.Result.Rows) != 2 || envelope.Result.Total != 3 || envelope.Result.Next != 2 || envelope.Result.Undated != 3 {
		t.Fatalf("history ceiling: %+v %v", envelope.Result.Total, err)
	}
	if _, _, err := prepareAnalyticsRead(s, request("analytics.history", `{"name":"large"}`)); !errors.Is(err, ErrInvalidWorkerRequest) {
		t.Fatalf("oversized event accepted: %v", err)
	}
	summary, _, err := prepareAnalyticsRead(s, request("analytics.summary", `{"limit":1}`))
	if err != nil || !strings.Contains(string(summary.Payload), `"next_offset":1`) || !strings.Contains(string(summary.Payload), `"total_names":3`) {
		t.Fatalf("summary pagination: %s %v", summary.Payload, err)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	if _, _, err := prepareAnalyticsRead(s, request("analytics.summary", "{}")); err == nil {
		t.Fatal("closed analytics store accepted")
	}
}

func TestInventoryTablePaginationAndCorruptMetadata(t *testing.T) {
	s := coverageStorage(t)
	planner := legacyStoragePlanner{storage: s}
	for _, table := range []string{"a", "b", "c"} {
		if err := s.Append("coverage", table, map[string]any{"value": true}); err != nil {
			t.Fatal(err)
		}
	}
	if err := s.Append("other", "foreign", map[string]any{"value": true}); err != nil {
		t.Fatal(err)
	}
	page, err := planner.AdminInventorySearch("coverage", "tables", "table_names", "tables.read", "", "", 2, nil)
	if err != nil || len(page.Rows) != 2 || page.Rows[0].Key != "a" || page.Rows[1].Key != "b" || page.Next != "b" {
		t.Fatalf("table inventory: %+v %v", page, err)
	}
	page, err = planner.AdminInventorySearch("coverage", "tables", "table_names", "tables.read", page.Next, "", 2, nil)
	if err != nil || len(page.Rows) != 1 || page.Rows[0].Key != "c" || page.Next != "" {
		t.Fatalf("table inventory second page: %+v %v", page, err)
	}
	if err := s.PutFile("coverage", "broken", "content", false); err != nil {
		t.Fatal(err)
	}
	if _, err := s.db.Exec("UPDATE plugin_file_rows SET metadata_json='invalid' WHERE project_id='coverage' AND file_key='broken'"); err != nil {
		t.Fatal(err)
	}
	if page, err := planner.AdminInventorySearch("coverage", "files", "file_keys", "files.metadata", "", "broken", 2, nil); err == nil || len(page.Rows) != 0 {
		t.Fatalf("corrupt file metadata exposed: %+v %v", page, err)
	}
}

func TestArenaInterruptRejectsInvalidReasonAndForgedUser(t *testing.T) {
	s := coverageStorage(t)
	coverageState(t, s, "arena", "room", map[string]any{"state": "active", "sequences": map[string]any{}})
	for _, reason := range []string{"", " ", strings.Repeat("x", 257), "bad\x00reason"} {
		data, err := json.Marshal(map[string]any{"roomId": "room", "reason": reason})
		if err != nil {
			t.Fatal(err)
		}
		if _, effect, err := prepareArenaStorageEffect(s, WorkerRequest{ProjectID: "coverage", Operation: "arena.admin.interrupt", Payload: data}); !errors.Is(err, ErrInvalidWorkerRequest) || effect != nil {
			t.Fatalf("invalid interrupt reason accepted: %q %v", reason, err)
		}
	}
	for _, principal := range []string{"host", "operator:alice", "user:other:alice", "user:coverage:bob"} {
		request := WorkerRequest{ProjectID: "coverage", Operation: "arena.player.input", Payload: []byte(`{"roomId":"room","playerId":"coverage:alice","sequence":1}`), accessPrincipal: principal}
		if _, effect, err := prepareArenaStorageEffect(s, request); !errors.Is(err, ErrActorDenied) || effect != nil {
			t.Fatalf("forged player accepted: %s %v", principal, err)
		}
	}
}
