package admin

import (
	"errors"
	"testing"

	"brain/internal/operator"
	"brain/internal/plugin"
	"brain/internal/project"
)

func TestCoverageLifecycleUnavailableDependencies(t *testing.T) {
	actor := operator.Actor{Kind: operator.ActorOperator, ID: "admin"}
	var missing *LifecycleService
	if _, err := missing.AdminPageRevision(actor, "page", PluginInvocation{}); !errors.Is(err, operator.ErrInvalidOperatorAction) {
		t.Fatal(err)
	}
	if err := missing.RunMigration(actor, "", func() error { return nil }); !errors.Is(err, operator.ErrInvalidOperatorAction) {
		t.Fatal(err)
	}
	if err := missing.RunMigration(actor, "v1", func() error { return nil }); !errors.Is(err, operator.ErrInvalidOperatorAction) {
		t.Fatal(err)
	}
	operators := &operator.Store{}
	digest := func(value string) string { return value }
	if _, err := operators.Bootstrap("admin", "secret", digest); err != nil {
		t.Fatal(err)
	}
	if err := operators.Authenticate("admin", "secret", digest); err != nil {
		t.Fatal(err)
	}
	r := plugin.NewRegistry(1)
	service := NewLifecycleService(operators, r)
	if err := service.CreateProject(actor, "coverage"); !errors.Is(err, operator.ErrInvalidOperatorAction) {
		t.Fatal(err)
	}
	if err := service.SetProjectActive(actor, "coverage", true); !errors.Is(err, operator.ErrInvalidOperatorAction) {
		t.Fatal(err)
	}
	if _, err := service.adminPageManifest(actor, "page", PluginInvocation{}); !errors.Is(err, operator.ErrInvalidOperatorAction) {
		t.Fatal(err)
	}
	service = NewLifecycleServiceWithProjects(operators, r, project.NewStore(project.Project{ID: "coverage", Active: true}))
	if err := service.SetProjectActive(actor, "coverage", false); err != nil {
		t.Fatal(err)
	}
	if _, err := service.AdminPageRevision(operator.Actor{Kind: operator.ActorUser, ID: "admin"}, "page", PluginInvocation{}); !errors.Is(err, operator.ErrOperatorRequired) {
		t.Fatal(err)
	}
}

func TestCoverageFunctionValidationAndInactiveProject(t *testing.T) {
	actor := operator.Actor{Kind: operator.ActorOperator, ID: "admin"}
	service := NewLifecycleServiceWithProjects(nil, plugin.NewRegistry(1), project.NewStore())
	source := `export const config = {method:"POST"}; export default () => ({});`
	if _, err := normalizeFunctionMethod("DELETE"); err == nil {
		t.Fatal("unsupported method accepted")
	}
	if method, err := normalizeFunctionMethod(""); err != nil || method != "POST" {
		t.Fatalf("default method: %s %v", method, err)
	}
	if _, err := service.CreateFunctionDraftWithSettings(actor, "missing", "test", source, -1, false, nil, "POST"); !errors.Is(err, ErrFunctionDraftInvalid) {
		t.Fatal(err)
	}
	if _, err := service.CreateFunctionDraftWithSettings(actor, "missing", "test", source, 1, false, nil, "POST"); !errors.Is(err, project.ErrNotFound) {
		t.Fatal(err)
	}
	if _, err := service.ListFunctionDrafts(actor, "missing"); !errors.Is(err, project.ErrNotFound) {
		t.Fatal(err)
	}
	if _, err := service.PublishFunction(actor, "missing", "test", "POST"); !errors.Is(err, project.ErrNotFound) {
		t.Fatal(err)
	}
	if _, err := service.validateFunctionDraft(actor, "missing", "test", source); !errors.Is(err, project.ErrNotFound) {
		t.Fatal(err)
	}
}
