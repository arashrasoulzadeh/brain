package plugin_test

import (
	"errors"
	"testing"
	"time"

	"brain/internal/plugin"
)

func TestBrokerAuthorizesTypedHostCall(t *testing.T) {
	r := plugin.NewRegistry(1)
	m := plugin.Manifest{ID: "tables", Version: "1.0.0", HostAPI: 1, Capabilities: []string{"records.read"}}
	if err := r.Register(m); err != nil {
		t.Fatal(err)
	}
	if err := r.Install("dev", m.ID, m.Version); err != nil {
		t.Fatal(err)
	}
	if err := r.GrantResource("dev", m.ID, m.Version, "records.read", "users"); err != nil {
		t.Fatal(err)
	}
	if err := r.SetReadiness(m.ID, m.Version, true); err != nil {
		t.Fatal(err)
	}
	if err := r.Enable("dev", m.ID, m.Version); err != nil {
		t.Fatal(err)
	}
	b := plugin.Broker{Registry: r, ActorAllows: func(actor, capability, resource string) bool {
		return actor == "alice" && capability == "records.read" && resource == "users"
	}}
	call := plugin.HostCall{ProjectID: "dev", ActorID: "alice", PluginID: m.ID, Version: m.Version, Capability: "records.read", Resource: "users", ExpiresAt: time.Unix(100, 0)}
	if err := b.Authorize(call, time.Unix(99, 0)); err != nil {
		t.Fatal(err)
	}
}

func TestBrokerChecksMatchmakingScopeWhenPresent(t *testing.T) {
	b := plugin.Broker{
		ActorAllows:       func(string, string, string) bool { return true },
		MatchmakingAllows: func(projectID, scope string) bool { return projectID == "dev" && scope == "ranked" },
	}
	call := plugin.HostCall{ProjectID: "prod", ActorID: "alice", MatchmakingScope: "ranked", ExpiresAt: time.Unix(100, 0)}
	if err := b.Authorize(call, time.Unix(99, 0)); !errors.Is(err, plugin.ErrPluginScopeDenied) {
		t.Fatalf("unapproved matchmaking scope = %v", err)
	}
}

func TestBrokerRejectsActorAndExpiredInvocation(t *testing.T) {
	b := plugin.Broker{ActorAllows: func(string, string, string) bool { return false }}
	call := plugin.HostCall{ProjectID: "dev", ActorID: "alice", ExpiresAt: time.Unix(1, 0)}
	if !errors.Is(b.Authorize(call, time.Unix(0, 0)), plugin.ErrActorDenied) {
		t.Fatal("actor denial was not enforced")
	}
	if !errors.Is(b.Authorize(call, time.Unix(2, 0)), plugin.ErrInvocationExpired) {
		t.Fatal("expiry was not enforced")
	}
}

func TestBrokerIssuesBoundedCredential(t *testing.T) {
	r := plugin.NewRegistry(1)
	m := plugin.Manifest{ID: "plugin", Version: "1", HostAPI: 1, Capabilities: []string{"read"}}
	if err := r.Register(m); err != nil {
		t.Fatal(err)
	}
	if err := r.Install("dev", m.ID, m.Version); err != nil {
		t.Fatal(err)
	}
	if err := r.Grant("dev", m.ID, m.Version, "read"); err != nil {
		t.Fatal(err)
	}
	if err := r.SetReadiness(m.ID, m.Version, true); err != nil {
		t.Fatal(err)
	}
	if err := r.Enable("dev", m.ID, m.Version); err != nil {
		t.Fatal(err)
	}
	now := time.Unix(100, 0)
	b := plugin.Broker{Registry: r, ActorAllows: func(string, string, string) bool { return true }}
	credential, err := b.Issue(plugin.HostCall{ProjectID: "dev", ActorID: "a", PluginID: m.ID, Version: m.Version, Capability: "read"}, now, time.Minute)
	if err != nil || credential.Token == "" || !credential.Call.ExpiresAt.Equal(now.Add(time.Minute)) {
		t.Fatalf("credential = %#v, %v", credential, err)
	}
}

func TestBrokerRechecksGrantAtCommit(t *testing.T) {
	r := plugin.NewRegistry(1)
	m := plugin.Manifest{ID: "plugin", Version: "1", HostAPI: 1, Capabilities: []string{"write"}}
	if err := r.Register(m); err != nil {
		t.Fatal(err)
	}
	if err := r.Install("dev", m.ID, m.Version); err != nil {
		t.Fatal(err)
	}
	if err := r.Grant("dev", m.ID, m.Version, "write"); err != nil {
		t.Fatal(err)
	}
	if err := r.SetReadiness(m.ID, m.Version, true); err != nil {
		t.Fatal(err)
	}
	if err := r.Enable("dev", m.ID, m.Version); err != nil {
		t.Fatal(err)
	}
	b := plugin.Broker{Registry: r, ActorAllows: func(string, string, string) bool { return true }}
	call := plugin.HostCall{ProjectID: "dev", ActorID: "a", PluginID: m.ID, Version: m.Version, Capability: "write", ExpiresAt: time.Unix(100, 0)}
	if err := b.Authorize(call, time.Unix(99, 0)); err != nil {
		t.Fatal(err)
	}
	if err := r.Revoke("dev", m.ID, m.Version, "write"); err != nil {
		t.Fatal(err)
	}
	if !errors.Is(b.Commit(call, time.Unix(99, 0)), plugin.ErrGrantDenied) {
		t.Fatal("revoked grant committed")
	}
}

func TestBrokerReturnsBoundedWorkerUnavailableError(t *testing.T) {
	b := plugin.Broker{WorkerAvailable: func(string, string) bool { return false }, ActorAllows: func(string, string, string) bool { return true }}
	err := b.Authorize(plugin.HostCall{ProjectID: "dev", ActorID: "a", PluginID: "p", Version: "1", ExpiresAt: time.Unix(100, 0)}, time.Unix(99, 0))
	if !errors.Is(err, plugin.ErrWorkerUnavailable) {
		t.Fatalf("Authorize() = %v", err)
	}
}
