package admin_test

import (
	"errors"
	"testing"

	"brain/internal/admin"
	"brain/internal/deploy"
	"brain/internal/operator"
	"brain/internal/plugin"
)

func TestLifecycleServiceRequiresAuthenticatedOperator(t *testing.T) {
	operators := &operator.Store{}
	digest := func(string) string { return "digest" }
	if _, err := operators.Bootstrap("admin", "secret", digest); err != nil {
		t.Fatal(err)
	}
	registry := plugin.NewRegistry(1)
	service := admin.NewLifecycleService(operators, registry)
	if err := service.EnablePlugin(operator.Actor{ID: "admin", Kind: operator.ActorOperator}, "dev", "rooms", "1.0.0"); !errors.Is(err, operator.ErrOperatorRequired) {
		t.Fatalf("unauthenticated lifecycle action = %v", err)
	}
	if events := operators.Audit(); len(events) != 1 || events[0].Action != "plugin.enable" || events[0].Success {
		t.Fatalf("authorization audit = %#v", events)
	}
}

func TestLifecycleServiceInstallsPluginAndAuditsAction(t *testing.T) {
	operators := &operator.Store{}
	digest := func(string) string { return "digest" }
	if _, err := operators.Bootstrap("admin", "secret", digest); err != nil {
		t.Fatal(err)
	}
	if err := operators.Authenticate("admin", "secret", digest); err != nil {
		t.Fatal(err)
	}
	registry := plugin.NewRegistry(1)
	if err := registry.Register(plugin.Manifest{ID: "rooms", Version: "1.0.0", HostAPI: 1, Entry: "index.ts", SourceFiles: []string{"index.ts"}}); err != nil {
		t.Fatal(err)
	}
	service := admin.NewLifecycleService(operators, registry)
	if err := service.InstallPlugin(operator.Actor{ID: "admin", Kind: operator.ActorOperator}, "dev", "rooms", "1.0.0"); err != nil {
		t.Fatal(err)
	}
	if _, err := registry.Status("dev", "rooms", "1.0.0"); err != nil {
		t.Fatal(err)
	}
	if events := operators.Audit(); len(events) != 2 || events[1].Action != "plugin.install" || !events[1].Success {
		t.Fatalf("install audit = %#v", events)
	}
}

func TestLifecycleServiceSetsPluginReadiness(t *testing.T) {
	operators := &operator.Store{}
	digest := func(string) string { return "digest" }
	_, _ = operators.Bootstrap("admin", "secret", digest)
	if err := operators.Authenticate("admin", "secret", digest); err != nil {
		t.Fatal(err)
	}
	registry := plugin.NewRegistry(1)
	if err := registry.Register(plugin.Manifest{ID: "rooms", Version: "1.0.0", HostAPI: 1, Entry: "index.ts", SourceFiles: []string{"index.ts"}}); err != nil {
		t.Fatal(err)
	}
	service := admin.NewLifecycleService(operators, registry)
	if err := service.SetPluginReadiness(operator.Actor{ID: "admin", Kind: operator.ActorOperator}, "rooms", "1.0.0", true); err != nil {
		t.Fatal(err)
	}
	if len(operators.Audit()) != 2 || operators.Audit()[1].Action != "plugin.readiness" {
		t.Fatalf("readiness audit = %#v", operators.Audit())
	}
}

func TestLifecycleServiceRoutesAuthorizedActionAndAuditsFailure(t *testing.T) {
	operators := &operator.Store{}
	digest := func(string) string { return "digest" }
	if _, err := operators.Bootstrap("admin", "secret", digest); err != nil {
		t.Fatal(err)
	}
	if err := operators.Authenticate("admin", "secret", digest); err != nil {
		t.Fatal(err)
	}
	service := admin.NewLifecycleService(operators, plugin.NewRegistry(1))
	actor := operator.Actor{ID: "admin", Kind: operator.ActorOperator}
	if err := service.RemovePlugin(actor, "dev", "rooms", "1.0.0"); !errors.Is(err, plugin.ErrNotInstalled) {
		t.Fatalf("lifecycle registry error = %v", err)
	}
	events := operators.Audit()
	if len(events) != 2 || events[1].Action != "plugin.remove" || events[1].Success {
		t.Fatalf("failure audit = %#v", events)
	}
}

func TestLifecycleServiceAuditsDurableDataDeletion(t *testing.T) {
	operators := &operator.Store{}
	digest := func(string) string { return "digest" }
	_, _ = operators.Bootstrap("admin", "secret", digest)
	if err := operators.Authenticate("admin", "secret", digest); err != nil {
		t.Fatal(err)
	}
	registry := plugin.NewRegistry(1)
	if err := registry.SaveDurableData("dev", "rooms", "snapshot", []byte("state")); err != nil {
		t.Fatal(err)
	}
	service := admin.NewLifecycleService(operators, registry)
	actor := operator.Actor{ID: "admin", Kind: operator.ActorOperator}
	if err := service.DeleteDurableData(actor, "dev", "rooms", "snapshot"); err != nil {
		t.Fatal(err)
	}
	if _, ok := registry.DurableData("dev", "rooms", "snapshot"); ok {
		t.Fatal("durable data was not deleted")
	}
	events := operators.Audit()
	if len(events) != 2 || events[1].Action != "durable-data.delete" || !events[1].Success {
		t.Fatalf("deletion audit = %#v", events)
	}
}

func TestLifecycleServiceAuditsRecoveryOperation(t *testing.T) {
	operators := &operator.Store{}
	digest := func(string) string { return "digest" }
	_, _ = operators.Bootstrap("admin", "secret", digest)
	if err := operators.Authenticate("admin", "secret", digest); err != nil {
		t.Fatal(err)
	}
	service := admin.NewLifecycleService(operators, plugin.NewRegistry(1))
	actor := operator.Actor{ID: "admin", Kind: operator.ActorOperator}
	if err := service.RecoverPlugin(actor, "dev", "rooms", "1.0.0"); !errors.Is(err, plugin.ErrNotInstalled) {
		t.Fatalf("recovery registry error = %v", err)
	}
	events := operators.Audit()
	if len(events) != 2 || events[1].Action != "plugin.recover" || events[1].Success {
		t.Fatalf("recovery audit = %#v", events)
	}
}

func TestLifecycleServiceAuditsMigrationAndRollback(t *testing.T) {
	operators := &operator.Store{}
	digest := func(string) string { return "digest" }
	_, _ = operators.Bootstrap("admin", "secret", digest)
	if err := operators.Authenticate("admin", "secret", digest); err != nil {
		t.Fatal(err)
	}
	service := admin.NewLifecycleService(operators, plugin.NewRegistry(1))
	actor := operator.Actor{ID: "admin", Kind: operator.ActorOperator}
	if err := service.RunMigration(actor, "42", func() error { return nil }); err != nil {
		t.Fatal(err)
	}
	manifest := deploy.ReleaseManifest{
		CommitSHA:                  "0123456789abcdef0123456789abcdef01234567",
		Backend:                    "backend@sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		Compiler:                   "compiler@sha256:cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc",
		Runner:                     "runner@sha256:dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd",
		PackagedIsolationVerified:  true,
		PackagedRevocationVerified: true,
	}
	if _, err := service.SelectRollbackManifest(actor, manifest, false); !errors.Is(err, deploy.ErrIncompatibleRollbackSchema) {
		t.Fatalf("rollback result = %v", err)
	}
	events := operators.Audit()
	if len(events) != 3 || events[1].Action != "migration.apply" || !events[1].Success || events[2].Action != "release.rollback" || events[2].Success {
		t.Fatalf("migration/rollback audit = %#v", events)
	}
}

func TestLifecycleServiceAuditsPermissionAndPromotionActions(t *testing.T) {
	operators := &operator.Store{}
	digest := func(string) string { return "digest" }
	_, _ = operators.Bootstrap("admin", "secret", digest)
	if err := operators.Authenticate("admin", "secret", digest); err != nil {
		t.Fatal(err)
	}
	service := admin.NewLifecycleService(operators, plugin.NewRegistry(1))
	actor := operator.Actor{ID: "admin", Kind: operator.ActorOperator}
	if err := service.GrantResource(actor, "dev", "rooms", "1.0.0", "rooms.read", "room-1"); !errors.Is(err, plugin.ErrNotInstalled) {
		t.Fatalf("resource grant error = %v", err)
	}
	if err := service.ConfigurePlugin(actor, "dev", "rooms", "1.0.0", map[string]any{"limit": 1}); !errors.Is(err, plugin.ErrNotInstalled) {
		t.Fatalf("configuration error = %v", err)
	}
	if err := service.PromotePlugin(actor, "dev", "rooms", "1.0.0"); !errors.Is(err, plugin.ErrNotInstalled) {
		t.Fatalf("promotion error = %v", err)
	}
	events := operators.Audit()
	if len(events) != 4 || events[1].Action != "plugin.grant-resource" || events[2].Action != "plugin.configure" || events[3].Action != "plugin.promote" {
		t.Fatalf("permission/promotion audit = %#v", events)
	}
}
