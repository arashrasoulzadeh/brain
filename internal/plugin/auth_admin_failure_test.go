package plugin

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestAuthAdminRejectsMalformedRequestsAndUnavailableStorage(t *testing.T) {
	s := coverageStorage(t)
	planner := legacyStoragePlanner{storage: s}
	for _, tc := range []struct {
		operation, payload string
		want               error
	}{
		{"auth.admin.accounts", `{"search":3}`, ErrInvalidWorkerRequest},
		{"auth.admin.accounts", `{"search":"a\u0000b"}`, ErrInvalidWorkerRequest},
		{"auth.admin.accounts", `{"kind":"unknown"}`, ErrInvalidWorkerRequest},
		{"auth.admin.accounts", `{"kind":true}`, ErrInvalidWorkerRequest},
		{"auth.admin.accounts", `{"limit":0}`, ErrInvalidWorkerRequest},
		{"auth.admin.accounts", `{"limit":51}`, ErrInvalidWorkerRequest},
		{"auth.admin.sessions.inspect", `{"username":"bad user"}`, ErrInvalidWorkerRequest},
		{"auth.admin.accounts.suspend", `{"username":"guest-a"}`, ErrInvalidWorkerRequest},
		{"auth.admin.accounts.reactivate", `{"username":"!"}`, ErrInvalidWorkerRequest},
		{"auth.admin.accounts.suspend", `{"username":"missing"}`, ErrStoredResourceMissing},
	} {
		if _, commit, err := planner.prepareAuthStorage(coverageRequest(tc.operation, tc.payload)); commit != nil || !errors.Is(err, tc.want) {
			t.Fatalf("%s %s: got %v, want %v", tc.operation, tc.payload, err, tc.want)
		}
	}
	coverageState(t, s, authAccountsNamespace, "corrupt", map[string]any{"username": 1})
	if _, commit, err := planner.prepareAuthStorage(coverageRequest("auth.admin.accounts.suspend", `{"username":"corrupt"}`)); commit != nil || !errors.Is(err, ErrModelSchema) {
		t.Fatalf("malformed account: %v", err)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	for _, op := range []string{"auth.admin.sessions.inspect", "auth.admin.accounts.reactivate"} {
		if _, commit, err := planner.prepareAuthStorage(coverageRequest(op, `{"username":"alice"}`)); commit != nil || err == nil {
			t.Fatalf("closed database accepted %s", op)
		}
	}
}

func TestAuthSuspensionRejectsInFlightSignInAndRollsBackOnSessionDeleteFailure(t *testing.T) {
	s := coverageStorage(t)
	planner := legacyStoragePlanner{storage: s}
	account := durableAuthAccount{Username: "alice", UserID: "coverage:alice", PasswordHash: "$argon2id$secret"}
	coverageState(t, s, authAccountsNamespace, "alice", account)
	session := durableAuthSession{ProjectID: "coverage", Username: "alice", ExpiresAt: time.Now().Add(time.Hour).UnixMilli()}
	coverageState(t, s, authSessionsNamespace, authTokenDigest("old"), session)
	plan := func(op, payload string) func(*WorkerResponse, []observationDelivery) error {
		t.Helper()
		_, commit, err := planner.prepareAuthStorage(coverageRequest(op, payload))
		if err != nil || commit == nil {
			t.Fatalf("%s: %v", op, err)
		}
		return commit
	}
	signin := plan("auth.signin", `{"username":"alice","password":"fixture"}`)
	suspend := plan("auth.admin.accounts.suspend", `{"username":"alice"}`)
	for _, statement := range []string{
		`CREATE TRIGGER reject_account_update BEFORE UPDATE ON plugin_state WHEN OLD.namespace='auth.accounts' BEGIN SELECT RAISE(ABORT,'account unavailable'); END`,
		`CREATE TRIGGER reject_session_delete BEFORE DELETE ON plugin_state WHEN OLD.namespace='auth.sessions' BEGIN SELECT RAISE(ABORT,'session unavailable'); END`,
	} {
		if _, err := s.db.Exec(statement); err != nil {
			t.Fatal(err)
		}
		if err := suspend(&WorkerResponse{Success: true}, nil); err == nil {
			t.Fatal("partial suspension accepted")
		}
		current, err := readAuthState[durableAuthAccount](s, "coverage", authAccountsNamespace, "alice")
		if err != nil || current == nil || current.Suspended {
			t.Fatalf("account update did not roll back: %+v %v", current, err)
		}
		if active, err := s.State("coverage", authSessionsNamespace, authTokenDigest("old")); err != nil || active == nil {
			t.Fatalf("existing session lost on failed suspension: %+v %v", active, err)
		}
		if _, err := s.db.Exec("DROP TRIGGER " + strings.Fields(statement)[2]); err != nil {
			t.Fatal(err)
		}
	}
	if err := suspend(&WorkerResponse{Success: true}, nil); err != nil {
		t.Fatal(err)
	}
	token := strings.Repeat("a", 64)
	response, err := json.Marshal(map[string]any{"result": map[string]string{"token": token, "username": "alice"}, "expiresAt": session.ExpiresAt})
	if err != nil {
		t.Fatal(err)
	}
	if err := signin(&WorkerResponse{Success: true, Payload: response}, nil); !errors.Is(err, errAuthCredentials) {
		t.Fatalf("in-flight signin restored a session after suspension: %v", err)
	}
	if active, err := s.State("coverage", authSessionsNamespace, authTokenDigest(token)); err != nil || active != nil {
		t.Fatalf("stale login persisted: %+v %v", active, err)
	}
	reactivate := plan("auth.admin.accounts.reactivate", `{"username":"alice"}`)
	if _, err := s.db.Exec("DELETE FROM plugin_state WHERE project_id=? AND namespace=? AND state_key=?", "coverage", authAccountsNamespace, "alice"); err != nil {
		t.Fatal(err)
	}
	if err := reactivate(&WorkerResponse{Success: true}, nil); !errors.Is(err, ErrStoredResourceMissing) {
		t.Fatalf("reactivation recreated a removed account: %v", err)
	}
}

func TestAuthSessionInspectionOnlyCountsUnexpiredProjectSessions(t *testing.T) {
	s := coverageStorage(t)
	now := time.Now().UnixMilli()
	for _, item := range []struct {
		key, project, user string
		expiry             int64
	}{
		{"expired", "coverage", "alice", now},
		{"first", "coverage", "alice", now + 10000},
		{"last", "coverage", "alice", now + 20000},
		{"foreign", "other", "alice", now + 40000},
		{"other-user", "coverage", "bob", now + 30000},
	} {
		coverageState(t, s, authSessionsNamespace, item.key, durableAuthSession{ProjectID: item.project, Username: item.user, ExpiresAt: item.expiry})
	}
	summary, err := s.authSessionSummary("coverage", "alice", now)
	if err != nil || summary.ActiveSessions != 2 || summary.EarliestExpiry != now+10000 || summary.LatestExpiry != now+20000 {
		t.Fatalf("session inspection: %+v %v", summary, err)
	}
	empty, err := s.authSessionSummary("coverage", "missing", now)
	if err != nil || empty.ActiveSessions != 0 || empty.EarliestExpiry != 0 || empty.LatestExpiry != 0 {
		t.Fatalf("empty summary: %+v %v", empty, err)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := s.authSessionSummary("coverage", "alice", now); err == nil {
		t.Fatal("inspection ignored closed storage")
	}
}
