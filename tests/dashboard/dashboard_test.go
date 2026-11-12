package dashboard_test

import (
	"testing"
	"time"

	"brain/internal/plugin"
	"brain/internal/project"
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

func TestDashboardLifecycleActionsUseHostRegistry(t *testing.T) {
	r, err := dashboard.Bootstrap(1)
	if err != nil {
		t.Fatal(err)
	}
	m := plugin.Manifest{ID: "rooms", Version: "1", HostAPI: 1, Capabilities: []string{"room.manage"}}
	if err := r.Register(m); err != nil {
		t.Fatal(err)
	}
	if err := r.Install("dev", m.ID, m.Version); err != nil {
		t.Fatal(err)
	}
	if err := r.SetReadiness(m.ID, m.Version, true); err != nil {
		t.Fatal(err)
	}
	if err := r.Grant("dev", m.ID, m.Version, "room.manage"); err != nil {
		t.Fatal(err)
	}
	if err := r.Enable("dev", m.ID, m.Version); err != nil {
		t.Fatal(err)
	}
	actions := dashboard.Lifecycle{Registry: r}
	if err := actions.Promote("dev", "rooms", "1"); err != nil {
		t.Fatal(err)
	}
	if err := actions.Drain("dev", "rooms", "1", time.Unix(100, 0)); err != nil {
		t.Fatal(err)
	}
	if err := actions.Rollback("dev", "rooms", "1"); err != nil {
		t.Fatal(err)
	}
	if err := actions.Remove("dev", "rooms", "1"); err == nil {
		t.Fatal("dashboard removed referenced active version")
	}
}

func TestDashboardProjectAdminActionsDelegateToHostBoundaries(t *testing.T) {
	registry, err := dashboard.Bootstrap(1)
	if err != nil {
		t.Fatal(err)
	}
	m := plugin.Manifest{ID: "auth", Version: "1", HostAPI: 1, Capabilities: []string{"auth.manage"}}
	if err := registry.Register(m); err != nil {
		t.Fatal(err)
	}
	if err := registry.SetReadiness(m.ID, m.Version, true); err != nil {
		t.Fatal(err)
	}
	projects := project.NewStore()
	admin := dashboard.ProjectAdmin{Projects: projects, Plugins: registry}
	if err := admin.CreateProject("dev"); err != nil {
		t.Fatal(err)
	}
	if err := admin.InstallPlugin("dev", m.ID, m.Version); err != nil {
		t.Fatal(err)
	}
	if err := admin.GrantPlugin("dev", m.ID, m.Version, "auth.manage"); err != nil {
		t.Fatal(err)
	}
	if err := admin.EnablePlugin("dev", m.ID, m.Version); err != nil {
		t.Fatal(err)
	}
	if _, err := projects.Resolve("dev"); err != nil {
		t.Fatal(err)
	}
}
