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

func TestRoomAdminClosureRejectsPreviouslyPlannedWrites(t *testing.T) {
	s := coverageStorage(t)
	plan := func(op, payload string) (func() error, error) {
		_, commit, err := prepareLegacyStorageRequest(s, WorkerRequest{ProjectID: "alpha", PluginID: "rooms", Operation: op, Payload: []byte(payload)})
		return commit, err
	}
	create, err := plan("rooms.create", `{"id":"r"}`)
	if err != nil || create == nil {
		t.Fatal(err)
	}
	if err := create(); err != nil {
		t.Fatal(err)
	}
	for _, invalid := range []struct {
		payload string
		want    error
	}{
		{`{"id":"r","member":""}`, ErrInvalidWorkerRequest},
		{`{"id":"r","member":" a "}`, ErrInvalidWorkerRequest},
		{`{"id":"r","member":"missing"}`, ErrStoredResourceMissing},
	} {
		if commit, err := plan("rooms.admin.remove-member", invalid.payload); commit != nil || !errors.Is(err, invalid.want) {
			t.Fatalf("invalid removal: %v", err)
		}
	}
	var pending []func() error
	for _, item := range []struct{ op, payload string }{
		{"rooms.create", `{"id":"r"}`},
		{"rooms.join", `{"id":"r","member":"late"}`},
		{"rooms.reconnect", `{"id":"r","member":"late"}`},
		{"rooms.result", `{"id":"r","result":1}`},
	} {
		commit, err := plan(item.op, item.payload)
		if err != nil || commit == nil {
			t.Fatal(err)
		}
		pending = append(pending, commit)
	}
	closeRoom, err := plan("rooms.admin.close", `{"id":"r"}`)
	if err != nil || closeRoom == nil {
		t.Fatal(err)
	}
	if err := closeRoom(); err != nil {
		t.Fatal(err)
	}
	for _, commit := range pending {
		if err := commit(); !errors.Is(err, ErrModelConflict) {
			t.Fatalf("stale room write after closure: %v", err)
		}
	}
	state, err := s.State("alpha", "rooms", "r")
	if err != nil || state["closed"] != true || len(state["members"].([]any)) != 0 || state["result"] != nil {
		t.Fatalf("closed room overwritten: %+v %v", state, err)
	}
	for _, op := range []string{"rooms.admin.close", "rooms.admin.remove-member"} {
		if _, commit, err := prepareLegacyStorageRequest(nil, WorkerRequest{PluginID: "rooms", Operation: op}); commit != nil || !errors.Is(err, ErrStoredResourceMissing) {
			t.Fatalf("missing durable room storage admitted %s: %v", op, err)
		}
	}
}
