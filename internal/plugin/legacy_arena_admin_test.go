package plugin

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestArenaAdminInterruptCompareAndSwap(t *testing.T) {
	s := coverageStorage(t)
	plan := func(project, operation, payload string) (WorkerRequest, func() error, error) {
		return prepareLegacyStorageRequest(s, WorkerRequest{ProjectID: project, PluginID: "arena", Operation: operation, Payload: []byte(payload)})
	}
	for _, project := range []string{"alpha", "beta"} {
		_, commit, err := plan(project, "arena.start", `{"roomId":"r","players":["a","b"]}`)
		if err != nil || commit == nil {
			t.Fatal(err)
		}
		if err := commit(); err != nil {
			t.Fatal(err)
		}
	}
	payload := `{"roomId":"r","reason":"maintenance","interruptedAt":"forged"}`
	prepared, commit, err := plan("alpha", "arena.admin.interrupt", payload)
	if err != nil || commit == nil {
		t.Fatal(err)
	}
	_, stale, err := plan("alpha", "arena.admin.interrupt", payload)
	if err != nil || stale == nil {
		t.Fatal(err)
	}
	var input map[string]any
	if json.Unmarshal(prepared.Payload, &input) != nil || input["interruptedAt"] == "forged" {
		t.Fatalf("invalid host timestamp: %s", prepared.Payload)
	}
	if err := commit(); err != nil {
		t.Fatal(err)
	}
	if err := stale(); !errors.Is(err, ErrModelConflict) {
		t.Fatalf("stale interruption: %v", err)
	}
	if _, _, err := plan("alpha", "arena.admin.interrupt", payload); !errors.Is(err, ErrModelConflict) {
		t.Fatalf("repeat interruption: %v", err)
	}
	state, err := s.State("alpha", "arena", "r")
	if err != nil || state["state"] != "interrupted" || state["result"] != nil || state["interruption"].(map[string]any)["reason"] != "maintenance" {
		t.Fatalf("state: %+v %v", state, err)
	}
	foreign, err := s.State("beta", "arena", "r")
	if err != nil || foreign["state"] != "active" {
		t.Fatalf("foreign state: %+v %v", foreign, err)
	}
	page, err := (legacyStoragePlanner{storage: s}).AdminInventory("alpha", "arena", "state_keys", "arena.snapshot", "", 10)
	if err != nil || len(page.Rows) != 1 || page.Rows[0].Status != "interrupted" || page.Rows[0].MemberCount == nil || *page.Rows[0].MemberCount != 2 {
		t.Fatalf("inventory: %+v %v", page, err)
	}
}
