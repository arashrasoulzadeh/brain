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
