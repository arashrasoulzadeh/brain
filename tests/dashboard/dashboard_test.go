package dashboard_test

import (
	"testing"

	"brain/internal/plugin"
	"brain/plugins/dashboard/implementation"
)

func TestDashboardIsAHostValidatedPlugin(t *testing.T) {
	r := plugin.NewRegistry(1)
	manifest := dashboard.Manifest()
	if err := r.Register(manifest); err != nil {
		t.Fatal(err)
	}
	if err := r.Install("dev", manifest.ID, manifest.Version); err != nil {
		t.Fatal(err)
	}
	if err := r.Configure("dev", manifest.ID, manifest.Version, map[string]any{"title": "Brain Admin"}); err != nil {
		t.Fatal(err)
	}
	if err := r.SetReadiness(manifest.ID, manifest.Version, true); err != nil {
		t.Fatal(err)
	}
	for _, capability := range manifest.Capabilities {
		if err := r.Grant("dev", manifest.ID, manifest.Version, capability); err != nil {
			t.Fatal(err)
		}
	}
	if err := r.Enable("dev", manifest.ID, manifest.Version); err != nil {
		t.Fatal(err)
	}
}

func TestDashboardBootstrapUsesPluginRegistry(t *testing.T) {
	r, err := dashboard.Bootstrap(1)
	if err != nil {
		t.Fatal(err)
	}
	if err := r.Install("dev", "dashboard", "0.1.0"); err != nil {
		t.Fatal(err)
	}
	for _, capability := range dashboard.Manifest().Capabilities {
		if err := r.Grant("dev", "dashboard", "0.1.0", capability); err != nil {
			t.Fatal(err)
		}
	}
	if err := r.Enable("dev", "dashboard", "0.1.0"); err != nil {
		t.Fatal(err)
	}
}
