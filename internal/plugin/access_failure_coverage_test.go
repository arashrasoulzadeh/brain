package plugin

import (
	"errors"
	"testing"
	"time"
)

func TestCoverageAccessProviderFailure(t *testing.T) {
	r := NewRegistry(1)
	if err := r.AuthorizeHostOperation("coverage", "user:alice", "rooms", "rooms.read", "room"); !errors.Is(err, ErrActorDenied) {
		t.Fatal(err)
	}
	if err := r.AuthorizeHostOperation("coverage", "host", "rooms", "rooms.read", "room"); err != nil {
		t.Fatal(err)
	}
	key := accessPluginID + "@1"
	r.registered[key] = Registered{Manifest: Manifest{ID: accessPluginID, Version: "1", Models: map[string]ModelSpec{accessRuleModel: accessRuleSpec}}}
	r.readiness[key] = true
	install := Installation{ProjectID: "coverage", PluginID: accessPluginID, Version: "1", Enabled: true, Grants: map[string]bool{"access.rules.read": true, "access.rules.manage": true}}
	r.installations["coverage:"+key] = install
	if installed, available := r.accessProviderLocked("coverage"); !installed || !available {
		t.Fatal("configured provider unavailable")
	}
	if err := r.authorizeAccessLocked("coverage", "", "rooms", "1", "rooms.read", "rooms.read", "room", nil); !errors.Is(err, ErrActorDenied) {
		t.Fatal(err)
	}
	for _, principal := range []string{"invalid", "operator:alice"} {
		if err := r.AuthorizeHostOperation("coverage", principal, accessPluginID, "read", ""); !errors.Is(err, ErrActorDenied) {
			t.Fatal(err)
		}
	}
	install.Grants = map[string]bool{"access.rules.read": true}
	r.installations["coverage:"+key] = install
	if installed, available := r.accessProviderLocked("coverage"); !installed || available {
		t.Fatal("provider with missing grant accepted")
	}
	if err := r.authorizeAccessLocked("coverage", "operator:alice", "rooms", "1", "rooms.read", "rooms.read", "room", nil); !errors.Is(err, ErrActorDenied) {
		t.Fatal(err)
	}
}

func TestCoverageAccessCredentialAndWorkerValidation(t *testing.T) {
	r := NewRegistry(1)
	req := WorkerRequest{Protocol: 1, ProjectID: "coverage", PluginID: "rooms", Version: "1", Operation: "rooms.read", Payload: []byte("{}")}
	if _, err := r.InvokeAsOperator(req, ""); !errors.Is(err, ErrActorDenied) {
		t.Fatal(err)
	}
	if err := r.CheckAdminReadAsOperator(req, ""); !errors.Is(err, ErrActorDenied) {
		t.Fatal(err)
	}
	if err := r.CheckAdminReadAsOperator(req, "alice"); !errors.Is(err, ErrGrantDenied) {
		t.Fatal(err)
	}
	if _, err := r.InvokeWithCredential(nil, InvocationCredential{}, req, time.Now()); !errors.Is(err, ErrInvalidWorkerRequest) {
		t.Fatal(err)
	}
	if err := r.validateCredentialResource(req, HostCall{}); !errors.Is(err, ErrInvocationCredentialMismatch) {
		t.Fatal(err)
	}
	req.requireAccessPolicy = true
	if err := r.authorizeInvocation(req, false); !errors.Is(err, ErrActorDenied) {
		t.Fatal(err)
	}
	req.requireAccessPolicy = false
	r.storageFailed.Store(true)
	if err := r.StartWorker("coverage", "rooms", "1"); !errors.Is(err, ErrLifecycleStorage) {
		t.Fatal(err)
	}
	if err := r.authorizeInvocation(req, false); !errors.Is(err, ErrLifecycleStorage) {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		id, payload string
	}{
		{"invites", `{"_brainInvites":{}}`},
		{"subscriptions", `{"_brainSubscriptions":{}}`},
	} {
		req.PluginID, req.Payload = tc.id, []byte(tc.payload)
		if err := validatePayloadProject(req); !errors.Is(err, ErrInvalidWorkerRequest) {
			t.Fatalf("%s reserved snapshot accepted: %v", tc.id, err)
		}
	}
}

func TestCoverageInventoryDenialAndCorruptStorage(t *testing.T) {
	s := coverageStorage(t)
	p := legacyStoragePlanner{storage: s}
	if _, err := (legacyStoragePlanner{}).AdminInventory("coverage", "rooms", "state_keys", "rooms.result", "", 10); !errors.Is(err, ErrAdminInventoryUnavailable) {
		t.Fatal(err)
	}
	if _, err := p.AdminInventory("coverage", "unknown", "state_keys", "rooms.result", "", 10); !errors.Is(err, ErrGrantDenied) {
		t.Fatal(err)
	}
	coverageState(t, s, "rooms", "*", map[string]any{})
	coverageState(t, s, "rooms", "room", map[string]any{})
	page, err := p.AdminInventoryScoped("coverage", "rooms", "state_keys", "rooms.result", "", 10, AdminInventoryFilter{AllowAll: true})
	if err != nil || len(page.Rows) != 1 || page.Rows[0].Key != "room" {
		t.Fatalf("invalid inventory key exposed: %+v %v", page, err)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := p.AdminInventory("coverage", "rooms", "state_keys", "rooms.result", "", 10); err == nil {
		t.Fatal("closed inventory accepted")
	}
	if _, err := s.accessRuleAllows("coverage", "operator:alice", "rooms", "1", "rooms.read", "rooms.read", "room"); err == nil {
		t.Fatal("closed policy store accepted")
	}
	if _, err := s.accessInventoryFilter("coverage", "operator:alice", "rooms", "1", Route{Operation: "rooms.read", Capability: "rooms.read", ResourceField: "roomId"}); err == nil {
		t.Fatal("closed policy inventory accepted")
	}
}
