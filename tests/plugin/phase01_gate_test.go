package plugin_test

import (
	"errors"
	"testing"
	"time"

	"brain/internal/plugin"
)

func TestPhase01ProjectIsolationAndRevocationGate(t *testing.T) {
	r := plugin.NewRegistry(1)
	m := plugin.Manifest{ID: "auth", Version: "1.0.0", HostAPI: 1, Capabilities: []string{"auth.guest"}, Routes: []plugin.Route{{Operation: "auth.guest", Function: "guest", Capability: "auth.guest"}}}
	if err := r.Register(m); err != nil {
		t.Fatal(err)
	}
	if err := r.SetReadiness(m.ID, m.Version, true); err != nil {
		t.Fatal(err)
	}
	for _, projectID := range []string{"dev", "prod"} {
		if err := r.Install(projectID, m.ID, m.Version); err != nil {
			t.Fatal(err)
		}
		if err := r.Grant(projectID, m.ID, m.Version, "auth.guest"); err != nil {
			t.Fatal(err)
		}
		if err := r.Enable(projectID, m.ID, m.Version); err != nil {
			t.Fatal(err)
		}
	}
	if err := r.Authorize("dev", m.ID, m.Version, "auth.guest"); err != nil {
		t.Fatal(err)
	}
	if err := r.Authorize("prod", m.ID, m.Version, "auth.guest"); err != nil {
		t.Fatal(err)
	}
	if err := r.Revoke("dev", m.ID, m.Version, "auth.guest"); err != nil {
		t.Fatal(err)
	}
	if !errors.Is(r.Authorize("dev", m.ID, m.Version, "auth.guest"), plugin.ErrGrantDenied) {
		t.Fatal("revoked grant was accepted")
	}
	if err := r.Authorize("prod", m.ID, m.Version, "auth.guest"); err != nil {
		t.Fatal("project grant was not isolated")
	}
	call := plugin.HostCall{ProjectID: "dev", ActorID: "guest", PluginID: m.ID, Version: m.Version, Operation: "auth.guest", Capability: "auth.guest", ExpiresAt: time.Now().Add(time.Minute)}
	b := plugin.Broker{Registry: r, ActorAllows: func(string, string, string) bool { return true }}
	if !errors.Is(b.Authorize(call, time.Now()), plugin.ErrGrantDenied) {
		t.Fatal("revoked in-flight effect was accepted")
	}
}
