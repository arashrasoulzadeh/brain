package admin_test

import (
	"errors"
	"testing"

	"brain/internal/admin"
	"brain/internal/operator"
	"brain/internal/plugin"
	"brain/internal/project"
)

func TestLifecycleInventoryAndGrantAudit(t *testing.T) {
	operators := &operator.Store{}
	digest := func(string) string { return "digest" }
	if _, err := operators.Bootstrap("admin", "secret", digest); err != nil {
		t.Fatal(err)
	}
	if err := operators.Authenticate("admin", "secret", digest); err != nil {
		t.Fatal(err)
	}
	registry := plugin.NewRegistry(1)
	if err := registry.Register(plugin.Manifest{ID: "sample", Version: "1.0.0", HostAPI: 1, Entry: "index.ts", SourceFiles: []string{"index.ts"}, Capabilities: []string{"sample.read"}}); err != nil {
		t.Fatal(err)
	}
	projects := project.NewStore()
	service := admin.NewLifecycleServiceWithProjects(operators, registry, projects)
	actor := operator.Actor{ID: "admin", Kind: operator.ActorOperator}
	if err := service.CreateProject(actor, "dev"); err != nil {
		t.Fatal(err)
	}
	if got, err := service.ListProjects(actor); err != nil || len(got) != 1 || got[0].ID != "dev" {
		t.Fatalf("projects = %+v, %v", got, err)
	}
	if err := service.InstallPlugin(actor, "dev", "sample", "1.0.0"); err != nil {
		t.Fatal(err)
	}
	if err := service.GrantCapability(actor, "dev", "sample", "1.0.0", "sample.read"); err != nil {
		t.Fatal(err)
	}
	if err := service.GrantResource(actor, "dev", "sample", "1.0.0", "z"); err != nil {
		t.Fatal(err)
	}
	if err := service.GrantResource(actor, "dev", "sample", "1.0.0", "a"); err != nil {
		t.Fatal(err)
	}
	if manifests, installations, err := service.ListPlugins(actor, "dev"); err != nil || len(manifests) != 1 || len(installations) != 1 || len(installations[0].ResourceGrants["sample.read"]) != 2 || installations[0].ResourceGrants["sample.read"][0] != "a" {
		t.Fatalf("plugins = %+v, %+v, %v", manifests, installations, err)
	}
	if capabilities, err := service.SystemCapabilities(actor); err != nil || len(capabilities) != 1 || capabilities[0].ID != "sample" {
		t.Fatalf("capabilities = %+v, %v", capabilities, err)
	}
	if err := service.RequireEnabledPlugin(actor, "dev", "sample"); !errors.Is(err, admin.ErrPluginUnavailable) {
		t.Fatalf("disabled plugin gate: %v", err)
	}
	if err := service.SetProjectActive(actor, "dev", false); err != nil {
		t.Fatal(err)
	}
	if got, err := service.ListProjects(actor); err != nil || len(got) != 1 || got[0].Active {
		t.Fatalf("deactivated projects = %+v, %v", got, err)
	}
	records, err := service.AuditRecords(actor)
	if err != nil {
		t.Fatal(err)
	}
	var grant, pluginRecord bool
	for _, record := range records {
		if record.OccurredAt == "" {
			t.Fatalf("missing audit timestamp: %+v", record)
		}
		if record.Source == "operator" && record.Action == "plugin.grant" && record.Success {
			grant = true
		}
		if record.Source == "plugin" {
			pluginRecord = true
		}
	}
	if !grant || !pluginRecord {
		t.Fatalf("missing audit entries: %+v", records)
	}
	unauthorized := operator.Actor{}
	if _, err := service.ListProjects(unauthorized); !errors.Is(err, operator.ErrInvalidOperatorAction) {
		t.Fatalf("unauthorized project listing: %v", err)
	}
	if _, _, err := service.ListPlugins(unauthorized, "dev"); !errors.Is(err, operator.ErrInvalidOperatorAction) {
		t.Fatalf("unauthorized plugin listing: %v", err)
	}
	if _, err := service.SystemCapabilities(unauthorized); !errors.Is(err, operator.ErrInvalidOperatorAction) {
		t.Fatalf("unauthorized capability listing: %v", err)
	}
	if _, err := service.AuditRecords(unauthorized); !errors.Is(err, operator.ErrInvalidOperatorAction) {
		t.Fatalf("unauthorized audit listing: %v", err)
	}
}
