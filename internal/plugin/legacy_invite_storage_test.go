package plugin

import (
	"encoding/json"
	"errors"
	"path/filepath"
	"strings"
	"testing"
)

func TestDurableInvitationScopeAndSingleUse(t *testing.T) {
	path := filepath.Join(t.TempDir(), "invites.sqlite")
	storage, err := OpenSQLiteStorage(path)
	if err != nil {
		t.Fatal(err)
	}
	plan := func(project, operation, payload string) (WorkerRequest, func() error, error) {
		return prepareLegacyStorageRequest(storage, WorkerRequest{Protocol: 1, ProjectID: project, PluginID: "invites", Operation: operation, Payload: []byte(payload)})
	}
	prepared, commit, err := plan("alpha", "invites.create", `{"roomId":"room","targetPlayerId":"alice"}`)
	if err != nil || commit == nil {
		t.Fatalf("create: %v", err)
	}
	var allocation struct {
		Snapshot struct {
			IDPart   string `json:"idPart"`
			CodePart string `json:"codePart"`
		} `json:"_brainInvites"`
	}
	if err := json.Unmarshal(prepared.Payload, &allocation); err != nil {
		t.Fatal(err)
	}
	id := "alpha:invite-" + allocation.Snapshot.IDPart
	code := "INVITE-alpha-" + allocation.Snapshot.CodePart
	if err := commit(); err != nil {
		t.Fatal(err)
	}
	record, err := storage.State("alpha", inviteNamespace, id)
	if err != nil || record["codeDigest"] != authTokenDigest(code) || record["code"] != nil {
		t.Fatalf("credential persisted incorrectly: %+v %v", record, err)
	}
	if err := storage.Close(); err != nil {
		t.Fatal(err)
	}
	storage, err = OpenSQLiteStorage(path)
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()
	payload, _ := json.Marshal(map[string]string{"roomId": "room", "playerId": "alice", "code": code})
	if _, _, err := plan("beta", "invites.accept", string(payload)); !errors.Is(err, errInviteUnavailable) {
		t.Fatalf("foreign invite accepted: %v", err)
	}
	listed, _, err := plan("alpha", "invites.inspect", `{"roomId":"room"}`)
	if err != nil || strings.Contains(string(listed.Payload), code) || strings.Contains(string(listed.Payload), authTokenDigest(code)) || !strings.Contains(string(listed.Payload), id) {
		t.Fatalf("redacted listing: %s %v", listed.Payload, err)
	}
	_, accept, err := plan("alpha", "invites.accept", string(payload))
	if err != nil || accept == nil {
		t.Fatal(err)
	}
	_, stale, err := plan("alpha", "invites.accept", string(payload))
	if err != nil || stale == nil {
		t.Fatal(err)
	}
	if err := accept(); err != nil {
		t.Fatal(err)
	}
	if err := stale(); !errors.Is(err, ErrModelConflict) {
		t.Fatalf("stale acceptance committed: %v", err)
	}
	if _, _, err := plan("alpha", "invites.accept", string(payload)); !errors.Is(err, errInviteUnavailable) {
		t.Fatalf("consumed invitation accepted: %v", err)
	}
}
