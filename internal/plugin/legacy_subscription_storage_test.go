package plugin

import (
	"encoding/json"
	"errors"
	"path/filepath"
	"strings"
	"testing"
)

func TestSubscriptionStorageRecoveryAndDeduplication(t *testing.T) {
	path := filepath.Join(t.TempDir(), "subscriptions.sqlite")
	storage, err := OpenSQLiteStorage(path)
	if err != nil {
		t.Fatal(err)
	}
	plan := func(project, operation, payload string) (WorkerRequest, func() error, error) {
		return prepareLegacyStorageRequest(storage, WorkerRequest{Protocol: 1, ProjectID: project, PluginID: "subscriptions", Operation: operation, Payload: []byte(payload)})
	}
	prepared, commit, err := plan("alpha", "subscriptions.open", `{"channel":"rooms.connected"}`)
	if err != nil || commit == nil {
		t.Fatalf("open: %v", err)
	}
	var open struct {
		Snapshot struct {
			Allocation struct {
				Ticket string
			}
		} `json:"_brainSubscriptions"`
	}
	if err := json.Unmarshal(prepared.Payload, &open); err != nil || open.Snapshot.Allocation.Ticket == "" {
		t.Fatalf("allocation: %s %v", prepared.Payload, err)
	}
	token := open.Snapshot.Allocation.Ticket
	if err := commit(); err != nil {
		t.Fatal(err)
	}
	record, err := storage.State("alpha", subscriptionTickets, authTokenDigest(token))
	if err != nil || record == nil || record["ticket"] != nil {
		t.Fatalf("ticket digest storage: %+v %v", record, err)
	}
	observation := `{"id":"one","projectId":"alpha","pluginId":"rooms","event":"connected","payload":{"roomId":"room","reconnected":false}}`
	_, commit, err = plan("alpha", "@observe/rooms/connected", observation)
	if err != nil || commit == nil {
		t.Fatalf("observation: %v", err)
	}
	if err := commit(); err != nil {
		t.Fatal(err)
	}
	if err := storage.Close(); err != nil {
		t.Fatal(err)
	}
	storage, err = OpenSQLiteStorage(path)
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()
	_, duplicate, err := plan("alpha", "@observe/rooms/connected", observation)
	if err != nil || duplicate != nil {
		t.Fatalf("duplicate observation planned: %v", err)
	}
	state, _, _, err := subscriptionStream(storage, "alpha")
	if err != nil || state.Sequence != 1 || len(state.Events) != 1 || state.Events[0].Value.RoomID != "room" {
		t.Fatalf("restored event stream: %+v %v", state, err)
	}
	payload, _ := json.Marshal(map[string]any{"channel": "rooms.connected", "ticket": token, "cursor": 0})
	if _, _, err := plan("beta", "subscriptions.receive", string(payload)); !errors.Is(err, errSubscriptionTicket) {
		t.Fatalf("foreign ticket admitted: %v", err)
	}
	_, read, err := plan("alpha", "subscriptions.receive", string(payload))
	if err != nil || read == nil {
		t.Fatalf("read: %v", err)
	}
	inspect, _, err := plan("alpha", "subscriptions.inspect", `{"channel":"rooms.connected"}`)
	if err != nil || strings.Contains(string(inspect.Payload), token) || strings.Contains(string(inspect.Payload), authTokenDigest(token)) {
		t.Fatalf("inspection leaked token material: %s %v", inspect.Payload, err)
	}
	_, revoke, err := plan("alpha", "subscriptions.revoke", `{"channel":"rooms.connected"}`)
	if err != nil || revoke == nil {
		t.Fatalf("revoke: %v", err)
	}
	if err := revoke(); err != nil {
		t.Fatal(err)
	}
	if err := read(); !errors.Is(err, errSubscriptionTicket) {
		t.Fatalf("revoked in-flight ticket accepted: %v", err)
	}
	if _, _, err := plan("alpha", "subscriptions.receive", string(payload)); !errors.Is(err, errSubscriptionTicket) {
		t.Fatalf("revoked ticket admitted: %v", err)
	}
}
