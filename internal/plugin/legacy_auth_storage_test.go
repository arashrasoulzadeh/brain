package plugin

import (
	"encoding/json"
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestDurableAuthStorageRedactsAndRevokes(t *testing.T) {
	path := filepath.Join(t.TempDir(), "auth.sqlite")
	storage, err := OpenSQLiteStorage(path)
	if err != nil {
		t.Fatal(err)
	}
	planner := legacyStoragePlanner{storage: storage}
	plan := func(project, operation, payload string) (WorkerRequest, func(*WorkerResponse, []observationDelivery) error, error) {
		return planner.prepareAuthStorage(WorkerRequest{Protocol: 1, ProjectID: project, PluginID: "auth", Version: "0.1.0", Operation: operation, Payload: []byte(payload)})
	}
	execute := func(project, operation, payload, output string) WorkerResponse {
		t.Helper()
		_, commit, err := plan(project, operation, payload)
		if err != nil || commit == nil {
			t.Fatalf("%s plan: %v", operation, err)
		}
		response := WorkerResponse{Protocol: 1, Success: true, Payload: []byte(output)}
		if err := commit(&response, nil); err != nil {
			t.Fatalf("%s commit: %v", operation, err)
		}
		return response
	}
	const hash = "$argon2id$fixture-hash"
	const signupPayload = `{"username":"alice","password":"fixture-password"}`
	const signupResult = `{"result":{"username":"alice","userId":"alpha:alice"},"passwordHash":"$argon2id$fixture-hash"}`
	_, invalidCommit, err := plan("alpha", "auth.signup", signupPayload)
	if err != nil {
		t.Fatal(err)
	}
	invalid := WorkerResponse{Success: true, Payload: []byte(`{"result":{"username":"alice","userId":"beta:alice"},"passwordHash":"$argon2id$fixture-hash"}`)}
	if err := invalidCommit(&invalid, nil); !errors.Is(err, ErrInvalidWorkerRequest) {
		t.Fatalf("cross-project signup result accepted: %v", err)
	}
	if record, err := storage.State("alpha", authAccountsNamespace, "alice"); err != nil || record != nil {
		t.Fatalf("invalid result persisted: %+v %v", record, err)
	}
	signup := execute("alpha", "auth.signup", signupPayload, signupResult)
	if strings.Contains(string(signup.Payload), hash) || strings.Contains(string(signup.Payload), "password") {
		t.Fatalf("signup exposed a credential: %s", signup.Payload)
	}
	if _, _, err := plan("alpha", "auth.signup", signupPayload); !errors.Is(err, errAuthCredentials) {
		t.Fatalf("duplicate signup admitted: %v", err)
	}
	if _, _, err := plan("beta", "auth.signin", signupPayload); !errors.Is(err, errAuthCredentials) {
		t.Fatalf("foreign account admitted: %v", err)
	}
	const token = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	expires := time.Now().Add(time.Hour).UnixMilli()
	result, _ := json.Marshal(map[string]any{"result": map[string]string{"token": token, "username": "alice"}, "expiresAt": expires})
	execute("alpha", "auth.signin", signupPayload, string(result))
	if raw, err := storage.State("alpha", authSessionsNamespace, token); err != nil || raw != nil {
		t.Fatalf("raw token stored: %+v %v", raw, err)
	}
	session, err := storage.State("alpha", authSessionsNamespace, authTokenDigest(token))
	if err != nil || session == nil || session["username"] != "alice" {
		t.Fatalf("digest record: %+v %v", session, err)
	}
	if err := storage.Close(); err != nil {
		t.Fatal(err)
	}
	storage, err = OpenSQLiteStorage(path)
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()
	planner.storage = storage
	payload, _ := json.Marshal(map[string]string{"token": token})
	identity := execute("alpha", "auth.identity", string(payload), `{"userId":"alpha:alice"}`)
	if string(identity.Payload) != `{"userId":"alpha:alice"}` {
		t.Fatalf("identity: %s", identity.Payload)
	}
	prepared, commit, err := plan("alpha", "auth.admin.accounts", "{}")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(prepared.Payload), hash) || strings.Contains(string(prepared.Payload), authTokenDigest(token)) {
		t.Fatalf("credential in admin worker snapshot: %s", prepared.Payload)
	}
	page := WorkerResponse{Success: true, Payload: []byte("{}")}
	if err := commit(&page, nil); err != nil {
		t.Fatal(err)
	}
	var listing authAccountPage
	if err := json.Unmarshal(page.Payload, &listing); err != nil || len(listing.Rows) != 1 || listing.Rows[0].Username != "alice" || listing.Rows[0].ActiveSessions != 1 {
		t.Fatalf("account page: %s %v", page.Payload, err)
	}
	if _, _, err := plan("beta", "auth.identity", string(payload)); !errors.Is(err, errAuthSession) {
		t.Fatalf("foreign session admitted: %v", err)
	}
	revoked := execute("alpha", "auth.admin.sessions.revoke", `{"username":"alice"}`, "{}")
	if string(revoked.Payload) != `{"revoked":1,"username":"alice"}` {
		t.Fatalf("revocation count: %s", revoked.Payload)
	}
	if _, _, err := plan("alpha", "auth.session", string(payload)); !errors.Is(err, errAuthSession) {
		t.Fatalf("revoked session admitted: %v", err)
	}
	forged := WorkerRequest{ProjectID: "alpha", PluginID: "auth", Payload: []byte(`{"_brainAuth":{}}`)}
	if err := validatePayloadProject(forged); !errors.Is(err, ErrInvalidWorkerRequest) {
		t.Fatalf("forged credential state admitted: %v", err)
	}
}
