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

func TestProjectsCanInstallSameVersionIndependently(t *testing.T) {
	r := plugin.NewRegistry(1)
	m := plugin.Manifest{ID: "guest", Version: "1.0.0", HostAPI: 1, Capabilities: []string{"session.issue"}}
	if err := r.Register(m); err != nil {
		t.Fatal(err)
	}
	for _, projectID := range []string{"dev", "prod"} {
		if err := r.Install(projectID, m.ID, m.Version); err != nil {
			t.Fatal(err)
		}
		if err := r.Grant(projectID, m.ID, m.Version, "session.issue"); err != nil {
			t.Fatal(err)
		}
		if err := r.Enable(projectID, m.ID, m.Version); err != nil {
			t.Fatal(err)
		}
	}
	if err := r.Authorize("dev", m.ID, m.Version, "session.issue"); err != nil {
		t.Fatal(err)
	}
	if err := r.Authorize("prod", m.ID, m.Version, "session.issue"); err != nil {
		t.Fatal(err)
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

func TestPluginRegistrationRequiresDependencies(t *testing.T) {
	r := plugin.NewRegistry(1)
	child := plugin.Manifest{ID: "room", Version: "1.0.0", HostAPI: 1, Dependencies: []string{"core@1.0.0"}}
	if !errors.Is(r.Register(child), plugin.ErrDependencyMissing) {
		t.Fatal("missing dependency was accepted")
	}
	if err := r.Register(plugin.Manifest{ID: "core", Version: "1.0.0", HostAPI: 1}); err != nil {
		t.Fatal(err)
	}
	if err := r.Register(child); err != nil {
		t.Fatal(err)
	}
}

func TestPluginRegistrationRejectsIncompleteManifest(t *testing.T) {
	if !errors.Is(plugin.NewRegistry(1).Register(plugin.Manifest{HostAPI: 1}), plugin.ErrInvalidManifest) {
		t.Fatal("incomplete manifest was accepted")
	}
}

func TestInstallationConfigurationUsesManifestSchema(t *testing.T) {
	r := plugin.NewRegistry(1)
	m := plugin.Manifest{ID: "dashboard", Version: "1.0.0", HostAPI: 1, ConfigSchema: map[string]string{"label": "string", "enabled": "bool"}}
	if err := r.Register(m); err != nil {
		t.Fatal(err)
	}
	if err := r.Install("dev", m.ID, m.Version); err != nil {
		t.Fatal(err)
	}
	if err := r.Configure("dev", m.ID, m.Version, map[string]any{"label": "local", "enabled": true}); err != nil {
		t.Fatal(err)
	}
	if err := r.Configure("dev", m.ID, m.Version, map[string]any{"enabled": "yes"}); !errors.Is(err, plugin.ErrInvalidConfig) {
		t.Fatalf("Configure() = %v", err)
	}
}
