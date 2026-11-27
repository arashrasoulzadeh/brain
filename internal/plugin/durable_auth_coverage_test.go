package plugin

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestDurableCoverageAuthAdmission(t *testing.T) {
	s := coverageStorage(t)
	p := legacyStoragePlanner{storage: s}
	for _, tc := range []struct {
		op, payload string
		want        error
	}{
		{"auth.signup", "null", ErrInvalidWorkerRequest},
		{"auth.signup", `{"username":"guest-forged","password":"password"}`, errAuthCredentials},
		{"auth.signin", `{"provider":"missing","credential":"password"}`, errAuthCredentials},
		{"auth.session", `{"token":""}`, errAuthSession},
		{"auth.session", `{"token":"missing"}`, errAuthSession},
		{"auth.admin.accounts", `{"after":42}`, ErrInvalidWorkerRequest},
		{"auth.admin.sessions.revoke", `{"username":"!"}`, ErrInvalidWorkerRequest},
		{"other", `{}`, ErrInvalidWorkerRequest},
	} {
		if _, commit, err := p.prepareAuthStorage(coverageRequest(tc.op, tc.payload)); !errors.Is(err, tc.want) || commit != nil {
			t.Fatalf("%s %s: %v", tc.op, tc.payload, err)
		}
	}
	if _, _, err := (legacyStoragePlanner{}).prepareAuthStorage(coverageRequest("auth.guest", "{}")); !errors.Is(err, ErrLifecycleStorage) {
		t.Fatal(err)
	}
	coverageState(t, s, authAccountsNamespace, "bad", map[string]any{"username": 42})
	if _, _, err := p.prepareAuthStorage(coverageRequest("auth.signin", `{"username":"bad","password":"password"}`)); !errors.Is(err, ErrModelSchema) {
		t.Fatal(err)
	}
	coverageState(t, s, authAccountsNamespace, "alice", durableAuthAccount{Username: "alice", UserID: "coverage:alice", PasswordHash: "$argon2id$hash"})
	if _, _, err := p.prepareAuthStorage(coverageRequest("auth.signup", `{"username":"alice","password":"password"}`)); !errors.Is(err, errAuthCredentials) {
		t.Fatal(err)
	}
	coverageState(t, s, authSessionsNamespace, authTokenDigest("token"), durableAuthSession{ProjectID: "coverage", Username: "missing", ExpiresAt: time.Now().Add(time.Hour).UnixMilli()})
	if _, _, err := p.prepareAuthStorage(coverageRequest("auth.identity", `{"token":"token"}`)); !errors.Is(err, errAuthSession) {
		t.Fatal(err)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ op, payload string }{
		{"auth.signin", `{"username":"alice","password":"password"}`},
		{"auth.session", `{"token":"token"}`},
		{"auth.admin.accounts", `{}`},
	} {
		if _, _, err := p.prepareAuthStorage(coverageRequest(tc.op, tc.payload)); err == nil {
			t.Fatalf("%s accepted closed database", tc.op)
		}
	}
}

func TestDurableCoverageAuthWorkerOutput(t *testing.T) {
	s := coverageStorage(t)
	p := legacyStoragePlanner{storage: s}
	_, signup, err := p.prepareAuthStorage(coverageRequest("auth.signup", `{"username":"alice","password":"password"}`))
	if err != nil || signup == nil {
		t.Fatalf("signup: %v", err)
	}
	for _, payload := range []string{"null", `{"result":{"username":"alice","userId":"other:alice"},"passwordHash":"$argon2id$hash"}`} {
		if err := signup(&WorkerResponse{Payload: []byte(payload)}, nil); !errors.Is(err, ErrInvalidWorkerRequest) {
			t.Fatal(err)
		}
	}
	if value, err := s.State("coverage", authAccountsNamespace, "alice"); err != nil || value != nil {
		t.Fatalf("invalid signup persisted: %v %v", value, err)
	}
	coverageState(t, s, authAccountsNamespace, "alice", durableAuthAccount{Username: "alice", UserID: "coverage:alice", PasswordHash: "$argon2id$hash"})
	for _, op := range []string{"auth.guest", "auth.signin"} {
		_, commit, err := p.prepareAuthStorage(coverageRequest(op, `{"username":"alice","password":"password"}`))
		if err != nil || commit == nil {
			t.Fatalf("%s: %v", op, err)
		}
		validToken := strings.Repeat("a", 64)
		for _, tc := range []struct {
			token, username, userID string
			guest                   bool
			expires                 int64
		}{
			{"short", "alice", "coverage:alice", false, time.Now().Add(time.Hour).UnixMilli()},
			{strings.Repeat("z", 64), "alice", "coverage:alice", false, time.Now().Add(time.Hour).UnixMilli()},
			{validToken, "wrong", "other:wrong", false, time.Now().Add(time.Hour).UnixMilli()},
			{validToken, "alice", "coverage:alice", false, time.Now().Add(-time.Hour).UnixMilli()},
		} {
			payload, err := json.Marshal(map[string]any{"result": map[string]any{"token": tc.token, "username": tc.username, "userId": tc.userID, "guest": tc.guest}, "expiresAt": tc.expires})
			if err != nil {
				t.Fatal(err)
			}
			if err := commit(&WorkerResponse{Payload: payload}, nil); !errors.Is(err, ErrInvalidWorkerRequest) {
				t.Fatalf("%s accepted malformed worker response: %v", op, err)
			}
		}
		if err := commit(&WorkerResponse{Payload: []byte("{")}, nil); !errors.Is(err, ErrInvalidWorkerRequest) {
			t.Fatal(err)
		}
	}
	if value, err := s.State("coverage", authSessionsNamespace, authTokenDigest(strings.Repeat("a", 64))); err != nil || value != nil {
		t.Fatalf("invalid session persisted: %v %v", value, err)
	}
}

func TestDurableCoverageAuthSessionCommitAndPaging(t *testing.T) {
	s := coverageStorage(t)
	p := legacyStoragePlanner{storage: s}
	for i := 0; i < 51; i++ {
		user := fmt.Sprintf("user-%02d", i)
		coverageState(t, s, authAccountsNamespace, user, durableAuthAccount{Username: user, UserID: "coverage:" + user, PasswordHash: "$argon2id$secret"})
	}
	page, err := s.authAccountPage("coverage", "", time.Now().UnixMilli())
	if err != nil || len(page.Rows) != 50 || page.Next != "user-49" {
		t.Fatalf("first page: %+v %v", page, err)
	}
	page, err = s.authAccountPage("coverage", page.Next, time.Now().UnixMilli())
	if err != nil || len(page.Rows) != 1 || page.Rows[0].Username != "user-50" || page.Next != "" {
		t.Fatalf("second page: %+v %v", page, err)
	}
	key := authTokenDigest("token")
	session := durableAuthSession{ProjectID: "coverage", Username: "user-00", ExpiresAt: time.Now().Add(time.Hour).UnixMilli()}
	coverageState(t, s, authSessionsNamespace, key, session)
	for _, op := range []string{"auth.session", "auth.identity"} {
		_, commit, err := p.prepareAuthStorage(coverageRequest(op, `{"token":"token"}`))
		if err != nil || commit == nil {
			t.Fatalf("%s: %v", op, err)
		}
		for _, payload := range []string{"{", `{"username":"wrong"}`, `{"userId":"other:user-00"}`} {
			if err := commit(&WorkerResponse{Payload: []byte(payload)}, nil); !errors.Is(err, ErrInvalidWorkerRequest) {
				t.Fatalf("%s %s: %v", op, payload, err)
			}
		}
		session.ExpiresAt++
		coverageState(t, s, authSessionsNamespace, key, session)
		if err := commit(&WorkerResponse{Payload: []byte(`{"username":"user-00"}`)}, nil); !errors.Is(err, errAuthSession) {
			t.Fatalf("changed credential accepted: %v", err)
		}
	}
	_, commit, err := p.prepareAuthStorage(coverageRequest("auth.admin.sessions.revoke", `{"username":"user-00"}`))
	if err != nil || commit == nil {
		t.Fatalf("revoke: %v", err)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	if err := commit(&WorkerResponse{}, nil); err == nil {
		t.Fatal("closed database accepted revocation")
	}
}
