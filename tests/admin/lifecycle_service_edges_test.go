package admin_test

import (
	"errors"
	"testing"

	"brain/internal/admin"
	"brain/internal/operator"
	"brain/internal/plugin"
	"brain/internal/project"
)

func TestLifecycleServiceFailClosedWithoutProjectOrRegistry(t *testing.T) {
	var absent *admin.LifecycleService
	actor := operator.Actor{Kind: operator.ActorOperator, ID: "admin"}
	if absent.ProjectStore() != nil || absent.ProjectURLNamespace("alpha") != "alpha" || absent.ProjectHTTPRoutes("alpha") != nil {
		t.Fatal("missing service exposed project state")
	}
	if _, ok := absent.MatchHTTPRoute("alpha", "GET", "/"); ok {
		t.Fatal("missing registry matched an HTTP route")
	}
	if _, err := absent.InvokeHTTPRoute("alpha", plugin.HTTPRouteMatch{}, nil, "admin"); !errors.Is(err, plugin.ErrNotRegistered) {
		t.Fatal(err)
	}
	operators := &operator.Store{}
	digest := func(string) string { return "digest" }
	if _, err := operators.Bootstrap("admin", "secret", digest); err != nil {
		t.Fatal(err)
	}
	if err := operators.Authenticate("admin", "secret", digest); err != nil {
		t.Fatal(err)
	}
	withoutProjects := admin.NewLifecycleService(operators, plugin.NewRegistry(1))
	if err := withoutProjects.CreateProjectWithDetails(actor, "alpha", "Alpha", "alpha-slug", nil); !errors.Is(err, operator.ErrInvalidOperatorAction) {
		t.Fatal(err)
	}
	if err := withoutProjects.ConfigureProject(actor, "alpha", "Alpha", "alpha-slug", nil); !errors.Is(err, operator.ErrInvalidOperatorAction) {
		t.Fatal(err)
	}
}

func TestLifecycleServiceConfirmedInstallAndProjectNamespace(t *testing.T) {
	operators := &operator.Store{}
	digest := func(string) string { return "digest" }
	if _, err := operators.Bootstrap("admin", "secret", digest); err != nil {
		t.Fatal(err)
	}
	if err := operators.Authenticate("admin", "secret", digest); err != nil {
		t.Fatal(err)
	}
	registry := plugin.NewRegistry(1)
	registry.BindWorkerFactory(func(_, _, _ string) (plugin.ContractWorker, error) { return &adminTestWorker{}, nil })
	projects := project.NewStore()
	service := admin.NewLifecycleServiceWithProjects(operators, registry, projects)
	actor := operator.Actor{Kind: operator.ActorOperator, ID: "admin"}
	if err := service.CreateProjectWithDetails(actor, "alpha", "Alpha", "alpha-slug", nil); err != nil {
		t.Fatal(err)
	}
	if service.ProjectURLNamespace("alpha") != "alpha-slug" {
		t.Fatal("project slug not resolved")
	}
	if err := service.ConfigureProject(actor, "alpha", "Updated", "new-slug", nil); err != nil {
		t.Fatal(err)
	}
	if service.ProjectURLNamespace("alpha") != "new-slug" || service.ProjectURLNamespace("missing") != "missing" {
		t.Fatal("project namespace not preserved")
	}
	for _, m := range []plugin.Manifest{
		{ID: "leaf", Version: "1", HostAPI: 1, Capabilities: []string{"read"}},
		{ID: "parent", Version: "1", HostAPI: 1, Capabilities: []string{"read"}, Dependencies: []string{"leaf@1"}},
	} {
		if err := registry.Register(m); err != nil {
			t.Fatal(err)
		}
	}
	if err := service.InstallAndActivatePlugin(actor, "alpha", "leaf", "1", nil); !errors.Is(err, plugin.ErrGrantDenied) {
		t.Fatal(err)
	}
	if err := service.InstallAndActivatePlugin(actor, "alpha", "leaf", "1", []string{"read"}); err != nil {
		t.Fatal(err)
	}
	if err := service.InstallPlugin(actor, "alpha", "parent", "1"); err != nil {
		t.Fatal(err)
	}
	if err := service.SetPluginReadiness(actor, "parent", "1", true); err != nil {
		t.Fatal(err)
	}
	if err := service.GrantCapability(actor, "alpha", "parent", "1", "read"); err != nil {
		t.Fatal(err)
	}
	if err := service.EnablePluginWithDependencies(actor, "alpha", "parent", "1", nil); err != nil {
		t.Fatal(err)
	}
	if err := service.ApprovePluginCapability(actor, "alpha", "parent", "1", "missing"); err == nil {
		t.Fatal("undeclared capability approved")
	}
	for _, id := range []string{"leaf", "parent"} {
		if err := service.RequireEnabledPlugin(actor, "alpha", id); err != nil {
			t.Fatal(err)
		}
	}
}
