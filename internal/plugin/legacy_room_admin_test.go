package plugin

import (
	"errors"
	"testing"
)

func TestRoomAdminMembershipAndClosure(t *testing.T) {
	s := coverageStorage(t)
	run := func(project, op, payload string) error {
		request := WorkerRequest{ProjectID: project, PluginID: "rooms", Operation: op, Payload: []byte(payload)}
		_, commit, err := prepareLegacyStorageRequest(s, request)
		if err != nil || commit == nil {
			return err
		}
		return commit()
	}
	for _, project := range []string{"alpha", "beta"} {
		for _, item := range []struct{ op, payload string }{
			{"rooms.create", `{"id":"room"}`},
			{"rooms.join", `{"id":"room","member":"a"}`},
			{"rooms.join", `{"id":"room","member":"b"}`},
			{"rooms.result", `{"id":"room","result":{"score":1}}`},
		} {
			if err := run(project, item.op, item.payload); err != nil {
				t.Fatal(err)
			}
		}
	}
	if err := run("alpha", "rooms.admin.remove-member", `{"id":"room","member":"a","room":{"members":["forged"]}}`); err != nil {
		t.Fatal(err)
	}
	if err := run("alpha", "rooms.admin.close", `{"id":"room"}`); err != nil {
		t.Fatal(err)
	}
	for _, op := range []string{"rooms.create", "rooms.join", "rooms.reconnect", "rooms.result", "rooms.admin.remove-member"} {
		if err := run("alpha", op, `{"id":"room","member":"a","result":2}`); !errors.Is(err, ErrRoomClosed) {
			t.Fatalf("%s admitted on closed room: %v", op, err)
		}
	}
	if err := run("alpha", "rooms.result", `{"id":"room"}`); err != nil {
		t.Fatal(err)
	}
	state, err := s.State("alpha", "rooms", "room")
	if err != nil || state["closed"] != true || len(state["members"].([]any)) != 1 || state["members"].([]any)[0] != "b" || state["result"].(map[string]any)["score"] != float64(1) {
		t.Fatalf("room state: %+v %v", state, err)
	}
	foreign, err := s.State("beta", "rooms", "room")
	if err != nil || foreign["closed"] == true || len(foreign["members"].([]any)) != 2 {
		t.Fatalf("foreign room changed: %+v %v", foreign, err)
	}
	page, err := (legacyStoragePlanner{storage: s}).AdminInventory("alpha", "rooms", "state_keys", "rooms.result", "", 10)
	if err != nil || len(page.Rows) != 1 || page.Rows[0].Status != "closed" || page.Rows[0].MemberCount == nil || *page.Rows[0].MemberCount != 1 {
		t.Fatalf("room inventory: %+v %v", page, err)
	}
}
