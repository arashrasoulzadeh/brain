package dashboard_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
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

func TestDashboardProjectFormDelegatesAndReturnsBoundedErrors(t *testing.T) {
	admin := dashboard.ProjectAdmin{Projects: project.NewStore(), Plugins: plugin.NewRegistry(1)}
	handler := dashboard.Handler(admin)
	req := httptest.NewRequest(http.MethodPost, "/admin/projects", strings.NewReader("id=dev"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusCreated {
		t.Fatalf("create status = %d", res.Code)
	}
	duplicate := httptest.NewRequest(http.MethodPost, "/admin/projects", strings.NewReader("id=dev"))
	duplicate.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	duplicateRes := httptest.NewRecorder()
	handler.ServeHTTP(duplicateRes, duplicate)
	if duplicateRes.Code != http.StatusBadRequest {
		t.Fatalf("duplicate status = %d", duplicateRes.Code)
	}
}

func TestDashboardPluginFormsInstallGrantAndEnable(t *testing.T) {
	r := plugin.NewRegistry(1)
	m := plugin.Manifest{ID: "auth", Version: "1", HostAPI: 1, Capabilities: []string{"auth.manage"}}
	if err := r.Register(m); err != nil {
		t.Fatal(err)
	}
	if err := r.SetReadiness(m.ID, m.Version, true); err != nil {
		t.Fatal(err)
	}
	admin := dashboard.ProjectAdmin{Projects: project.NewStore(), Plugins: r}
	if err := admin.CreateProject("dev"); err != nil {
		t.Fatal(err)
	}
	handler := dashboard.Handler(admin)
	post := func(path, form string) int {
		req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(form))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		res := httptest.NewRecorder()
		handler.ServeHTTP(res, req)
		return res.Code
	}
	if got := post("/admin/plugins/install", "project=dev&plugin=auth&version=1"); got != http.StatusCreated {
		t.Fatalf("install status = %d", got)
	}
	if got := post("/admin/plugins/grant", "project=dev&plugin=auth&version=1&capability=auth.manage"); got != http.StatusNoContent {
		t.Fatalf("grant status = %d", got)
	}
	if got := post("/admin/plugins/enable", "project=dev&plugin=auth&version=1"); got != http.StatusNoContent {
		t.Fatalf("enable status = %d", got)
	}
}
