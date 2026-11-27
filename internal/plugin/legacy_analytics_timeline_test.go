package plugin

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestAnalyticsTimelineBoundariesAndMissingDates(t *testing.T) {
	s := coverageStorage(t)
	events := []any{
		map[string]any{"projectId": "coverage", "name": "test", "occurred_at": "2026-01-01T00:00:00Z"},
		map[string]any{"projectId": "coverage", "name": "test", "occurred_at": "2026-01-01T00:59:59Z"},
		map[string]any{"projectId": "coverage", "name": "test", "occurred_at": "2026-01-01T02:00:00Z"},
		map[string]any{"projectId": "coverage", "name": "test"},
		map[string]any{"projectId": "foreign", "name": "test", "occurred_at": "2026-01-01T00:30:00Z"},
		map[string]any{"projectId": "coverage", "name": "other", "occurred_at": "2026-01-01T00:30:00Z"},
	}
	coverageState(t, s, "analytics", "events", map[string]any{"events": events})
	query := func(op, payload string) map[string]any {
		t.Helper()
		prepared, commit, err := prepareLegacyStorageRequest(s, WorkerRequest{ProjectID: "coverage", PluginID: "analytics", Operation: op, Payload: []byte(payload)})
		if err != nil || commit != nil {
			t.Fatalf("%s: %v", op, err)
		}
		var response struct {
			Result map[string]any `json:"result"`
		}
		if err := json.Unmarshal(prepared.Payload, &response); err != nil {
			t.Fatal(err)
		}
		return response.Result
	}
	window := `{"name":"test","from":"2026-01-01T00:00:00Z","until":"2026-01-01T02:00:00Z","bucket":"hour"}`
	timeline := query("analytics.timeline", window)
	rows := timeline["rows"].([]any)
	if len(rows) != 2 || rows[0].(map[string]any)["count"] != float64(2) || rows[1].(map[string]any)["count"] != float64(0) || timeline["total_events"] != float64(2) || timeline["undated_events"] != float64(1) {
		t.Fatalf("timeline: %+v", timeline)
	}
	history := query("analytics.history", window)
	if history["total_events"] != float64(2) || history["undated_events"] != float64(1) {
		t.Fatalf("history: %+v", history)
	}
	summary := query("analytics.summary", window)
	if summary["total_events"] != float64(3) || summary["undated_events"] != float64(1) {
		t.Fatalf("summary: %+v", summary)
	}
	for _, payload := range []string{`{}`, `{"from":"bad","until":"2026-01-01T00:00:00Z","bucket":"day"}`, `{"from":"2026-01-01T00:00:00Z","until":"2026-01-20T00:00:00Z","bucket":"hour"}`, `{"from":"2026-01-01T00:00:00Z","until":"2026-01-02T00:00:00Z","bucket":"week"}`} {
		if _, _, err := prepareLegacyStorageRequest(s, WorkerRequest{ProjectID: "coverage", PluginID: "analytics", Operation: "analytics.timeline", Payload: []byte(payload)}); !errors.Is(err, ErrInvalidWorkerRequest) {
			t.Fatalf("invalid timeline accepted: %s %v", payload, err)
		}
	}
}
