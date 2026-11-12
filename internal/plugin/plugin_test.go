package plugin_test

import (
	"errors"
	"testing"

	"brain/internal/plugin"
)

func TestPluginLifecycleAndScopedAuthorization(t *testing.T) {
	r := plugin.NewRegistry(1)
	m := plugin.Manifest{ID: "guest", Version: "1.0.0", HostAPI: 1, Capabilities: []string{"session.issue"}}
	if err := r.Register(m); err != nil {
		t.Fatal(err)
	}
	if err := r.Install("dev", m.ID, m.Version); err != nil {
		t.Fatal(err)
	}
	if err := r.Enable("dev", m.ID, m.Version); !errors.Is(err, plugin.ErrGrantDenied) {
		t.Fatalf("Enable() = %v, want grant denial", err)
	}
	if err := r.Grant("dev", m.ID, m.Version, "session.issue"); err != nil {
		t.Fatal(err)
	}
	if err := r.Enable("dev", m.ID, m.Version); err != nil {
		t.Fatal(err)
	}
	if err := r.Authorize("dev", m.ID, m.Version, "session.issue"); err != nil {
		t.Fatal(err)
	}
	if err := r.Authorize("prod", m.ID, m.Version, "session.issue"); !errors.Is(err, plugin.ErrNotInstalled) {
		t.Fatalf("cross-project authorization = %v", err)
	}
}

func TestPluginRegistrationRejectsDuplicatesAndWrongHostAPI(t *testing.T) {
	r := plugin.NewRegistry(1)
	m := plugin.Manifest{ID: "arena", Version: "1.0.0", HostAPI: 1}
	if err := r.Register(m); err != nil {
		t.Fatal(err)
	}
	if !errors.Is(r.Register(m), plugin.ErrDuplicate) {
		t.Fatal("duplicate registration was accepted")
	}
	if !errors.Is(r.Register(plugin.Manifest{ID: "bad", Version: "1", HostAPI: 2}), plugin.ErrIncompatible) {
		t.Fatal("wrong API was accepted")
	}
}
