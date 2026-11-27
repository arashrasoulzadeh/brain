package plugin

import (
	"encoding/json"
	"errors"
	"path/filepath"
	"testing"
)

func TestAnalyticsHistoryAndSummaryUseOnlyHostProjectState(t *testing.T) {
	storage, err := OpenSQLiteStorage(filepath.Join(t.TempDir(), "analytics.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()
	for _, projectID := range []string{"alpha", "beta"} {
		for _, name := range []string{"opened", "closed", "opened"} {
			payload, err := json.Marshal(map[string]any{"name": name, "fields": map[string]any{"project": projectID}})
			if err != nil {
				t.Fatal(err)
			}
			_, commit, err := prepareLegacyStorageRequest(storage, WorkerRequest{ProjectID: projectID, PluginID: "analytics", Operation: "analytics.capture", Payload: payload})
			if err != nil || commit == nil {
				t.Fatalf("capture plan: %v", err)
			}
			if err := commit(); err != nil {
				t.Fatal(err)
			}
		}
	}
	request := WorkerRequest{ProjectID: "alpha", PluginID: "analytics", Operation: "analytics.history", Payload: []byte(`{"name":"opened","limit":1,"result":{"rows":[{"forged":true}]}}`)}
	prepared, commit, err := prepareLegacyStorageRequest(storage, request)
	if err != nil || commit != nil {
		t.Fatalf("history plan: %v", err)
	}
	var snapshot struct {
		ProjectID string `json:"projectId"`
		Result struct {
			Total int `json:"total_events"`
			Next  int `json:"next_offset"`
			Rows  []struct {
				Fields     map[string]string `json:"fields"`
				OccurredAt string `json:"occurred_at"`
			} `json:"rows"`
		} `json:"result"`
	}
	if err := json.Unmarshal(prepared.Payload, &snapshot); err != nil || snapshot.ProjectID != "alpha" || snapshot.Result.Total != 2 || snapshot.Result.Next != 1 || len(snapshot.Result.Rows) != 1 || snapshot.Result.Rows[0].Fields["project"] != "alpha" || snapshot.Result.Rows[0].OccurredAt == "" {
		t.Fatalf("unexpected history: %s (%v)", prepared.Payload, err)
	}
	request.Operation = "analytics.summary"
	request.Payload = []byte(`{}`)
	prepared, commit, err = prepareLegacyStorageRequest(storage, request)
	if err != nil || commit != nil {
		t.Fatalf("summary plan: %v", err)
	}
	var summary struct {
		Result struct {
			Total int `json:"total_events"`
			Names int `json:"total_names"`
			Rows []struct {
				Name  string `json:"name"`
				Count int `json:"count"`
			} `json:"rows"`
		} `json:"result"`
	}
	if err := json.Unmarshal(prepared.Payload, &summary); err != nil || summary.Result.Total != 3 || summary.Result.Names != 2 || len(summary.Result.Rows) != 2 || summary.Result.Rows[0].Name != "closed" || summary.Result.Rows[1].Count != 2 {
		t.Fatalf("unexpected summary: %s (%v)", prepared.Payload, err)
	}
	for _, payload := range []string{`{"projectId":"beta"}`, `{"limit":101}`, `{"offset":-1}`, `{"limit":1.5}`} {
		request.Payload = []byte(payload)
		if _, _, err := prepareLegacyStorageRequest(storage, request); !errors.Is(err, ErrInvalidWorkerRequest) {
			t.Fatalf("invalid input accepted: %s: %v", payload, err)
		}
	}
}
