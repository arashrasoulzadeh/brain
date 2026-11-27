package plugin

import (
	"bytes"
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"testing"
	"time"
)

type coverageFailEntropy struct{}

func (coverageFailEntropy) Read([]byte) (int, error) { return 0, io.ErrUnexpectedEOF }

func TestStorageCoverageEntropyFailure(t *testing.T) {
	s := coverageStorage(t)
	original := rand.Reader
	t.Cleanup(func() { rand.Reader = original })
	request := coverageRequest("invites.create", `{"roomId":"room","targetPlayerId":"player"}`)
	for _, reader := range []io.Reader{coverageFailEntropy{}, io.MultiReader(bytes.NewReader(make([]byte, 16)), coverageFailEntropy{})} {
		rand.Reader = reader
		if _, effect, err := prepareInviteStorageEffect(s, request); !errors.Is(err, io.ErrUnexpectedEOF) || effect != nil {
			t.Fatalf("invite entropy failure: %v", err)
		}
	}
	rand.Reader = coverageFailEntropy{}
	if _, effect, err := prepareSubscriptionStorageEffect(s, coverageRequest("subscriptions.open", `{"channel":"rooms.connected"}`)); !errors.Is(err, io.ErrUnexpectedEOF) || effect != nil {
		t.Fatalf("subscription entropy failure: %v", err)
	}
	spec := ModelSpec{Version: 1, Fields: map[string]string{"name": "string"}}
	req := coverageRequest("create", `{"id":"item","fields":{"name":"test"}}`)
	req.PluginID = "example"
	if _, err := s.executeModel(req, spec, ModelRoute{Name: "item", Action: "create"}); !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("model entropy failure: %v", err)
	}
}

type coverageQueryFailure struct{ db *sql.DB }

func (e coverageQueryFailure) Exec(string, ...any) (sql.Result, error) {
	return coverageSQLResult{}, nil
}
func (e coverageQueryFailure) QueryRow(string, ...any) *sql.Row {
	return e.db.QueryRow("SELECT * FROM missing_coverage_table")
}

func TestStorageCoverageInviteCommitFailures(t *testing.T) {
	s := coverageStorage(t)
	req := coverageRequest("invites.create", `{"roomId":"room","targetPlayerId":"player"}`)
	req.ProjectID = strings.Repeat("p", 256)
	if _, _, err := prepareInviteStorageEffect(s, req); !errors.Is(err, ErrInvalidWorkerRequest) {
		t.Fatal(err)
	}
	req.ProjectID = "coverage"
	_, effect, err := prepareInviteStorageEffect(s, req)
	if err != nil || effect == nil {
		t.Fatalf("create admission: %v", err)
	}
	if err := effect(coverageQueryFailure{s.db}); err == nil {
		t.Fatal("failed capacity query accepted")
	}
	record := storedInvite{ID: "id", ProjectID: "coverage", RoomID: "room", TargetPlayerID: "player", ExpiresAt: time.Now().Add(time.Hour).UnixMilli()}
	coverageState(t, s, inviteNamespace, "id", record)
	_, cancel, err := prepareInviteStorageEffect(s, coverageRequest("invites.cancel", `{"roomId":"room","inviteId":"id"}`))
	if err != nil || cancel == nil {
		t.Fatalf("cancel admission: %v", err)
	}
	failure := errors.New("injected SQL failure")
	for _, executor := range []coverageSQLExecutor{{execErr: failure}, {resultErr: failure}} {
		if err := cancel(executor); !errors.Is(err, failure) {
			t.Fatal(err)
		}
	}
	if err := cancel(s.db); err != nil {
		t.Fatal(err)
	}
	if value, err := s.State("coverage", inviteNamespace, "id"); err != nil || value != nil {
		t.Fatalf("cancelled invite retained: %v %v", value, err)
	}
	coverageState(t, s, inviteNamespace, "bad", map[string]any{"roomId": "room"})
	if _, _, err := prepareInviteStorageEffect(s, coverageRequest("invites.inspect", `{"roomId":"room"}`)); err == nil {
		t.Fatal("malformed inventory record accepted")
	}
}

func TestStorageCoverageSubscriptionQueryFailures(t *testing.T) {
	s := coverageStorage(t)
	_, effect, err := prepareSubscriptionStorageEffect(s, coverageRequest("subscriptions.open", `{"channel":"rooms.connected"}`))
	if err != nil || effect == nil {
		t.Fatalf("open admission: %v", err)
	}
	failure := errors.New("injected SQL failure")
	if err := effect(coverageSQLExecutor{execErr: failure}); !errors.Is(err, failure) {
		t.Fatal(err)
	}
	if err := effect(coverageQueryFailure{s.db}); err == nil {
		t.Fatal("failed ticket capacity query accepted")
	}
	// A malformed ticket affects JSON queries without corrupting the stream.
	if _, err := s.db.Exec("INSERT INTO plugin_state(project_id,namespace,state_key,value_json) VALUES(?,?,?,?)", "coverage", subscriptionTickets, "bad", "{"); err != nil {
		t.Fatal(err)
	}
	for _, op := range []string{"subscriptions.inspect", "subscriptions.revoke"} {
		if _, _, err := prepareSubscriptionStorageEffect(s, coverageRequest(op, `{"channel":"rooms.connected"}`)); err == nil {
			t.Fatalf("%s accepted corrupt tickets", op)
		}
	}
}

func TestStorageCoverageAuthCommitAndRollback(t *testing.T) {
	s := coverageStorage(t)
	p := legacyStoragePlanner{storage: s}
	token := strings.Repeat("a", 64)
	_, commit, err := p.prepareAuthStorage(coverageRequest("auth.guest", "{}"))
	if err != nil || commit == nil {
		t.Fatalf("guest admission: %v", err)
	}
	payload, err := json.Marshal(map[string]any{
		"result": map[string]any{"token": token, "username": "guest-test", "userId": "coverage:guest-test", "guest": true},
		"expiresAt": time.Now().Add(time.Hour).UnixMilli(),
	})
	if err != nil {
		t.Fatal(err)
	}
	response := &WorkerResponse{Payload: payload}
	if err := commit(response, nil); err != nil {
		t.Fatal(err)
	}
	key := authTokenDigest(token)
	_, read, err := p.prepareAuthStorage(coverageRequest("auth.session", `{"token":"`+token+`"}`))
	if err != nil || read == nil {
		t.Fatalf("session admission: %v", err)
	}
	if err := read(&WorkerResponse{Payload: []byte(`{"username":"guest-test"}`)}, nil); err != nil {
		t.Fatal(err)
	}
	coverageState(t, s, authSessionsNamespace, key, map[string]any{"username": 42})
	if err := read(&WorkerResponse{Payload: []byte(`{"username":"guest-test"}`)}, nil); !errors.Is(err, ErrModelSchema) {
		t.Fatalf("corrupt session accepted: %v", err)
	}
	coverageState(t, s, authAccountsNamespace, "bad", map[string]any{"username": 42})
	coverageState(t, s, authSessionsNamespace, key, durableAuthSession{ProjectID: "coverage", Username: "bad", ExpiresAt: time.Now().Add(time.Hour).UnixMilli()})
	if _, _, err := p.prepareAuthStorage(coverageRequest("auth.identity", `{"token":"`+token+`"}`)); !errors.Is(err, ErrModelSchema) {
		t.Fatalf("corrupt account accepted: %v", err)
	}
	if _, err := s.authAccountPage("coverage", "", time.Now().UnixMilli()); err == nil {
		t.Fatal("malformed account page accepted")
	}
	_, revoke, err := p.prepareAuthStorage(coverageRequest("auth.admin.sessions.revoke", `{"username":"bad"}`))
	if err != nil || revoke == nil {
		t.Fatalf("revoke admission: %v", err)
	}
	if _, err := s.db.Exec("CREATE TRIGGER reject_revoke BEFORE DELETE ON plugin_state BEGIN SELECT RAISE(ABORT,'test failure'); END"); err != nil {
		t.Fatal(err)
	}
	if err := revoke(&WorkerResponse{}, nil); err == nil {
		t.Fatal("failed revocation accepted")
	}
	if _, err := s.db.Exec("DROP TRIGGER reject_revoke"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.db.Exec("DROP TABLE plugin_observation_queue"); err != nil {
		t.Fatal(err)
	}
	if err := revoke(&WorkerResponse{}, []observationDelivery{{ID: "test"}}); err == nil {
		t.Fatal("revocation with failed observation accepted")
	}
	if value, err := s.State("coverage", authSessionsNamespace, key); err != nil || value == nil {
		t.Fatalf("failed transaction did not roll back: %v %v", value, err)
	}
}
