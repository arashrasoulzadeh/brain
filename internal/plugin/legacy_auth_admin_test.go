package plugin

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestAuthAdminSearchSuspendAndReactivate(t *testing.T) {
	s := coverageStorage(t)
	p := legacyStoragePlanner{storage: s}
	now := time.Now().UnixMilli()
	for _, user := range []string{"alice", "bob"} {
		coverageState(t, s, authAccountsNamespace, user, durableAuthAccount{Username: user, UserID: "coverage:" + user, PasswordHash: "$argon2id$secret"})
		coverageState(t, s, authSessionsNamespace, authTokenDigest(user), durableAuthSession{ProjectID: "coverage", Username: user, ExpiresAt: now + 60000})
	}
	run := func(op, payload string) string {
		t.Helper()
		prepared, commit, err := p.prepareAuthStorage(coverageRequest(op, payload))
		if err != nil || commit == nil {
			t.Fatalf("%s: %v", op, err)
		}
		if strings.Contains(string(prepared.Payload), "$argon2id$") || strings.Contains(string(prepared.Payload), authTokenDigest("alice")) {
			t.Fatal("admin snapshot contains credential")
		}
		response := WorkerResponse{Success: true, Payload: []byte(`{"forged":true}`)}
		if err := commit(&response, nil); err != nil {
			t.Fatal(err)
		}
		return string(response.Payload)
	}
	var listing authAccountPage
	if err := json.Unmarshal([]byte(run("auth.admin.accounts", `{"search":" AL ","limit":1}`)), &listing); err != nil || len(listing.Rows) != 1 || listing.Rows[0].Username != "alice" {
		t.Fatalf("search: %+v %v", listing, err)
	}
	var summary authSessionSummary
	if err := json.Unmarshal([]byte(run("auth.admin.sessions.inspect", `{"username":"alice"}`)), &summary); err != nil || summary.ActiveSessions != 1 || summary.LatestExpiry != now+60000 {
		t.Fatalf("summary: %+v %v", summary, err)
	}
	run("auth.admin.accounts.suspend", `{"username":"alice"}`)
	for _, op := range []string{"auth.signin", "auth.session", "auth.identity"} {
		_, _, err := p.prepareAuthStorage(coverageRequest(op, `{"username":"alice","password":"secret","token":"alice"}`))
		if !errors.Is(err, errAuthCredentials) && !errors.Is(err, errAuthSession) {
			t.Fatalf("suspended account accepted: %s %v", op, err)
		}
	}
	if err := json.Unmarshal([]byte(run("auth.admin.accounts", `{"kind":"suspended"}`)), &listing); err != nil || len(listing.Rows) != 1 || !listing.Rows[0].Suspended {
		t.Fatalf("suspended list: %+v %v", listing, err)
	}
	run("auth.admin.accounts.reactivate", `{"username":"alice"}`)
	if _, _, err := p.prepareAuthStorage(coverageRequest("auth.signin", `{"username":"alice","password":"secret"}`)); err != nil {
		t.Fatalf("reactivated account: %v", err)
	}
	if _, _, err := p.prepareAuthStorage(coverageRequest("auth.session", `{"token":"alice"}`)); !errors.Is(err, errAuthSession) {
		t.Fatalf("revoked session revived: %v", err)
	}
	if _, _, err := p.prepareAuthStorage(coverageRequest("auth.session", `{"token":"bob"}`)); err != nil {
		t.Fatalf("other account affected: %v", err)
	}
}
