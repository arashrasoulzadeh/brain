package plugin_test

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"brain/internal/plugin"
)

func TestLoadManifestRequiresAuthorizedSourcesAndImports(t *testing.T) {
	directory := t.TempDir()
	manifestPath := filepath.Join(directory, "plugin.yaml")
	valid := []byte("id: sample\nentrypoint: index.ts\nversion: 1.0.0\nhost_api: 1\nsource_files:\n  - index.ts\nimports:\n  - ../sdk-typescript/src/index.js\nplugin_imports:\n  - rooms@0.1.0\n")
	if err := os.WriteFile(manifestPath, valid, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := plugin.LoadManifest(manifestPath); err != nil {
		t.Fatalf("valid manifest rejected: %v", err)
	}
	invalid := []byte("id: sample\nentrypoint: index.ts\nversion: 1.0.0\nhost_api: 1\nsource_files:\n  - index.ts\nimports:\n  - ../../secret.ts\n")
	if err := os.WriteFile(manifestPath, invalid, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := plugin.LoadManifest(manifestPath); !errors.Is(err, plugin.ErrInvalidManifest) {
		t.Fatalf("unsafe import accepted: %v", err)
	}
}

func TestValidatePluginSourcesRejectsUndeclaredImports(t *testing.T) {
	directory := t.TempDir()
	if err := os.WriteFile(filepath.Join(directory, "index.ts"), []byte("import \"node:crypto\";\nimport \"node:net\";\n"), 0600); err != nil {
		t.Fatal(err)
	}
	manifest := plugin.Manifest{ID: "sample", Entry: "index.ts", Version: "1.0.0", HostAPI: 1, SourceFiles: []string{"index.ts"}, Imports: []string{"node:crypto"}}
	if err := plugin.ValidatePluginSources(directory, manifest); !errors.Is(err, plugin.ErrInvalidManifest) {
		t.Fatalf("undeclared import accepted: %v", err)
	}
	if err := os.WriteFile(filepath.Join(directory, "index.ts"), []byte("import \"node:crypto\";\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := plugin.ValidatePluginSources(directory, manifest); err != nil {
		t.Fatalf("declared import rejected: %v", err)
	}
}

func TestPluginLifecycleAndScopedAuthorization(t *testing.T) {
	r := plugin.NewRegistry(1)
	m := plugin.Manifest{ID: "guest", Version: "1.0.0", HostAPI: 1, Capabilities: []string{"session.issue"}}
	if err := r.Register(m); err != nil {
		t.Fatal(err)
	}
	if err := r.Install("dev", m.ID, m.Version); err != nil {
		t.Fatal(err)
	}
	if err := r.SetReadiness(m.ID, m.Version, true); err != nil {
		t.Fatal(err)
	}
	if err := r.Enable("dev", m.ID, m.Version); err != nil {
		t.Fatal(err)
	}
	if err := r.SetReadiness(m.ID, m.Version, true); err != nil {
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
	if err := r.SetReadiness(m.ID, m.Version, true); err != nil {
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

func TestInstallationStatusSeparatesReadinessFromEnablement(t *testing.T) {
	r := plugin.NewRegistry(1)
	m := plugin.Manifest{ID: "auth", Version: "1.0.0", HostAPI: 1, Capabilities: []string{"auth.signin"}}
	if err := r.Register(m); err != nil {
		t.Fatal(err)
	}
	if err := r.Install("dev", m.ID, m.Version); err != nil {
		t.Fatal(err)
	}
	status, err := r.Status("dev", m.ID, m.Version)
	if err != nil || status.Enabled || status.Ready {
		t.Fatalf("initial status = %#v, %v", status, err)
	}
	if err := r.SetReadiness(m.ID, m.Version, true); err != nil {
		t.Fatal(err)
	}
	status, err = r.Status("dev", m.ID, m.Version)
	if err != nil || status.Enabled || !status.Ready {
		t.Fatalf("ready status = %#v, %v", status, err)
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

func TestPluginRegistrationAcceptsDeclaredExtensionPoints(t *testing.T) {
	r := plugin.NewRegistry(1)
	m := plugin.Manifest{ID: "auth", Version: "1.0.0", HostAPI: 1, ExtensionPoints: []string{"auth.provider"}}
	if err := r.Register(m); err != nil {
		t.Fatal(err)
	}
}

func TestPluginRegistrationRejectsBlankExtensionPoint(t *testing.T) {
	m := plugin.Manifest{ID: "auth", Version: "1.0.0", HostAPI: 1, ExtensionPoints: []string{""}}
	if !errors.Is(plugin.NewRegistry(1).Register(m), plugin.ErrInvalidManifest) {
		t.Fatal("blank extension point was accepted")
	}
}

func TestPluginStoresApprovedArtifactEndpoint(t *testing.T) {
	r := plugin.NewRegistry(1)
	m := plugin.Manifest{ID: "auth", Version: "1.0.0", HostAPI: 1}
	if err := r.Register(m); err != nil {
		t.Fatal(err)
	}
	if err := r.ApproveArtifact(m.ID, m.Version, "https://plugins.example/auth-1.0.0.tgz"); err != nil {
		t.Fatal(err)
	}
	approved, ok := r.ApprovedArtifact(m.ID, m.Version)
	if !ok || approved != "https://plugins.example/auth-1.0.0.tgz" {
		t.Fatalf("approved artifact = %q, %v", approved, ok)
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

func TestCapabilityRequestsAndResourceScopedGrants(t *testing.T) {
	r := plugin.NewRegistry(1)
	m := plugin.Manifest{ID: "tables", Version: "1.0.0", HostAPI: 1, Capabilities: []string{"records.read"}}
	if len(m.RequestedCapabilities()) != 1 || !m.RequestedCapabilities()[0].Required {
		t.Fatal("capability request was not exposed")
	}
	if err := r.Register(m); err != nil {
		t.Fatal(err)
	}
	if err := r.Install("dev", m.ID, m.Version); err != nil {
		t.Fatal(err)
	}
	if err := r.SetReadiness(m.ID, m.Version, true); err != nil {
		t.Fatal(err)
	}
	if err := r.GrantResource("dev", m.ID, m.Version, "records.read", "users"); err != nil {
		t.Fatal(err)
	}
	if err := r.Enable("dev", m.ID, m.Version); err != nil {
		t.Fatal(err)
	}
	if err := r.AuthorizeResource("dev", m.ID, m.Version, "records.read", "users"); err != nil {
		t.Fatal(err)
	}
	if !errors.Is(r.AuthorizeResource("dev", m.ID, m.Version, "records.read", "payments"), plugin.ErrGrantDenied) {
		t.Fatal("cross-resource access was accepted")
	}
}

func TestPluginMutationsProduceAuditEvents(t *testing.T) {
	r := plugin.NewRegistry(1)
	m := plugin.Manifest{ID: "audit", Version: "1.0.0", HostAPI: 1, Capabilities: []string{"read"}}
	if err := r.Register(m); err != nil {
		t.Fatal(err)
	}
	if err := r.Install("dev", m.ID, m.Version); err != nil {
		t.Fatal(err)
	}
	if err := r.SetReadiness(m.ID, m.Version, true); err != nil {
		t.Fatal(err)
	}
	if err := r.Grant("dev", m.ID, m.Version, "read"); err != nil {
		t.Fatal(err)
	}
	if err := r.Enable("dev", m.ID, m.Version); err != nil {
		t.Fatal(err)
	}
	audit := r.Audit()
	if len(audit) < 5 {
		t.Fatalf("audit events = %d, want at least 5", len(audit))
	}
	if audit[0].Action != "register" || audit[1].Action != "install" {
		t.Fatalf("audit order = %#v", audit[:2])
	}
}

func TestPromotionRoutesNewWorkWithoutChangingPinnedVersion(t *testing.T) {
	r := plugin.NewRegistry(1)
	for _, version := range []string{"1.0.0", "2.0.0"} {
		m := plugin.Manifest{ID: "rooms", Version: version, HostAPI: 1, Capabilities: []string{"room.run"}}
		if err := r.Register(m); err != nil {
			t.Fatal(err)
		}
		if err := r.SetReadiness(m.ID, m.Version, true); err != nil {
			t.Fatal(err)
		}
		if err := r.Install("dev", m.ID, m.Version); err != nil {
			t.Fatal(err)
		}
		if err := r.Grant("dev", m.ID, m.Version, "room.run"); err != nil {
			t.Fatal(err)
		}
		if err := r.Enable("dev", m.ID, m.Version); err != nil {
			t.Fatal(err)
		}
	}
	if err := r.Promote("dev", "rooms", "1.0.0"); err != nil {
		t.Fatal(err)
	}
	if err := r.Promote("dev", "rooms", "2.0.0"); err != nil {
		t.Fatal(err)
	}
	if got, _ := r.ActiveVersion("dev", "rooms"); got != "2.0.0" {
		t.Fatalf("active version = %q", got)
	}
	if err := r.Drain("dev", "rooms", "1.0.0"); err != nil {
		t.Fatal(err)
	}
	if err := r.Rollback("dev", "rooms", "1.0.0"); err != nil {
		t.Fatal(err)
	}
	if got, _ := r.ActiveVersion("dev", "rooms"); got != "1.0.0" {
		t.Fatalf("rollback version = %q", got)
	}
	if err := r.Authorize("dev", "rooms", "1.0.0", "room.run"); err != nil {
		t.Fatalf("rollback changed retained grants: %v", err)
	}
	deadline := time.Unix(200, 0)
	if err := r.DrainUntil("dev", "rooms", "2.0.0", deadline); err != nil {
		t.Fatal(err)
	}
	if err := r.CanStart("dev", "rooms", "2.0.0", time.Unix(199, 0)); err != nil {
		t.Fatal("work was blocked before deadline")
	}
	if err := r.CanStart("dev", "rooms", "2.0.0", deadline); !errors.Is(err, plugin.ErrDisabled) {
		t.Fatal("work was admitted after deadline")
	}
	if err := r.PinWork("room-1", "dev", "rooms", "1.0.0"); err != nil {
		t.Fatal(err)
	}
	if err := r.Promote("dev", "rooms", "2.0.0"); err != nil {
		t.Fatal(err)
	}
	if got, _ := r.PinnedVersion("room-1"); got != "1.0.0" {
		t.Fatalf("pinned version = %q", got)
	}
	if !errors.Is(r.Remove("dev", "rooms", "1.0.0"), plugin.ErrVersionReferenced) {
		t.Fatal("referenced version was removed")
	}
	if !errors.Is(r.Remove("dev", "rooms", "2.0.0"), plugin.ErrVersionReferenced) {
		t.Fatal("active version was removed")
	}
}

func TestPromotionRequiresCompletePluginConfiguration(t *testing.T) {
	r := plugin.NewRegistry(1)
	m := plugin.Manifest{ID: "configured", Version: "1.0.0", HostAPI: 1, ConfigSchema: map[string]string{"region": "string"}}
	if err := r.Register(m); err != nil {
		t.Fatal(err)
	}
	if err := r.SetReadiness(m.ID, m.Version, true); err != nil {
		t.Fatal(err)
	}
	if err := r.Install("dev", m.ID, m.Version); err != nil {
		t.Fatal(err)
	}
	if err := r.Enable("dev", m.ID, m.Version); err != nil {
		t.Fatal(err)
	}
	if !errors.Is(r.Promote("dev", m.ID, m.Version), plugin.ErrInvalidConfig) {
		t.Fatal("incomplete configuration was promoted")
	}
	if err := r.Configure("dev", m.ID, m.Version, map[string]any{"region": "eu"}); err != nil {
		t.Fatal(err)
	}
	if err := r.Promote("dev", m.ID, m.Version); err != nil {
		t.Fatalf("complete configuration rejected: %v", err)
	}
}

func TestPromotionRequiresReadyPluginVersion(t *testing.T) {
	r := plugin.NewRegistry(1)
	m := plugin.Manifest{ID: "not-ready", Version: "1.0.0", HostAPI: 1}
	if err := r.Register(m); err != nil {
		t.Fatal(err)
	}
	if err := r.Install("dev", m.ID, m.Version); err != nil {
		t.Fatal(err)
	}
	if !errors.Is(r.Promote("dev", m.ID, m.Version), plugin.ErrNotReady) {
		t.Fatal("unready version was promoted")
	}
}

func TestRollbackRejectsMissingRetainedVersion(t *testing.T) {
	r := plugin.NewRegistry(1)
	m := plugin.Manifest{ID: "rollback", Version: "1.0.0", HostAPI: 1}
	if err := r.Register(m); err != nil {
		t.Fatal(err)
	}
	if err := r.SetReadiness(m.ID, m.Version, true); err != nil {
		t.Fatal(err)
	}
	if err := r.Install("dev", m.ID, m.Version); err != nil {
		t.Fatal(err)
	}
	if err := r.Enable("dev", m.ID, m.Version); err != nil {
		t.Fatal(err)
	}
	if err := r.Promote("dev", m.ID, m.Version); err != nil {
		t.Fatal(err)
	}
	if !errors.Is(r.Rollback("dev", m.ID, "0.9.0"), plugin.ErrNotInstalled) {
		t.Fatal("rollback accepted a version that was not retained")
	}
}

func TestPromotionRechecksRequestedGrants(t *testing.T) {
	r := plugin.NewRegistry(1)
	m := plugin.Manifest{ID: "grant-recheck", Version: "1.0.0", HostAPI: 1, Capabilities: []string{"records.read"}}
	if err := r.Register(m); err != nil {
		t.Fatal(err)
	}
	if err := r.SetReadiness(m.ID, m.Version, true); err != nil {
		t.Fatal(err)
	}
	if err := r.Install("dev", m.ID, m.Version); err != nil {
		t.Fatal(err)
	}
	if err := r.Grant("dev", m.ID, m.Version, "records.read"); err != nil {
		t.Fatal(err)
	}
	if err := r.Enable("dev", m.ID, m.Version); err != nil {
		t.Fatal(err)
	}
	if err := r.Revoke("dev", m.ID, m.Version, "records.read"); err != nil {
		t.Fatal(err)
	}
	if !errors.Is(r.Promote("dev", m.ID, m.Version), plugin.ErrGrantDenied) {
		t.Fatal("promotion ignored a revoked requested grant")
	}
}

func TestPromotionRejectsIncompatibleConfigurationSchema(t *testing.T) {
	r := plugin.NewRegistry(1)
	for _, m := range []plugin.Manifest{
		{ID: "schema", Version: "1.0.0", HostAPI: 1, ConfigSchema: map[string]string{"region": "string"}},
		{ID: "schema", Version: "2.0.0", HostAPI: 1, ConfigSchema: map[string]string{"region": "bool"}},
	} {
		if err := r.Register(m); err != nil {
			t.Fatal(err)
		}
		if err := r.SetReadiness(m.ID, m.Version, true); err != nil {
			t.Fatal(err)
		}
		if err := r.Install("dev", m.ID, m.Version); err != nil {
			t.Fatal(err)
		}
		if err := r.Grant("dev", m.ID, m.Version, ""); err != nil && !errors.Is(err, plugin.ErrGrantDenied) {
			t.Fatal(err)
		}
	}
	if err := r.Configure("dev", "schema", "1.0.0", map[string]any{"region": "eu"}); err != nil {
		t.Fatal(err)
	}
	if err := r.Enable("dev", "schema", "1.0.0"); err != nil {
		t.Fatal(err)
	}
	if err := r.Promote("dev", "schema", "1.0.0"); err != nil {
		t.Fatal(err)
	}
	if err := r.Configure("dev", "schema", "2.0.0", map[string]any{"region": true}); err != nil {
		t.Fatal(err)
	}
	if err := r.Enable("dev", "schema", "2.0.0"); err != nil {
		t.Fatal(err)
	}
	if !errors.Is(r.Promote("dev", "schema", "2.0.0"), plugin.ErrSchemaIncompatible) {
		t.Fatal("incompatible configuration schema was promoted")
	}
}

func TestRollbackRejectsIncompatibleRetainedVersion(t *testing.T) {
	r := plugin.NewRegistry(1)
	versions := []plugin.Manifest{
		{ID: "rollback-schema", Version: "1.0.0", HostAPI: 1, ConfigSchema: map[string]string{"region": "string"}},
		{ID: "rollback-schema", Version: "2.0.0", HostAPI: 1, ConfigSchema: map[string]string{"region": "bool"}},
	}
	for _, manifest := range versions {
		if err := r.Register(manifest); err != nil {
			t.Fatal(err)
		}
		if err := r.SetReadiness(manifest.ID, manifest.Version, true); err != nil {
			t.Fatal(err)
		}
		if err := r.Install("dev", manifest.ID, manifest.Version); err != nil {
			t.Fatal(err)
		}
	}
	if err := r.Configure("dev", "rollback-schema", "1.0.0", map[string]any{"region": "eu"}); err != nil {
		t.Fatal(err)
	}
	if err := r.Configure("dev", "rollback-schema", "2.0.0", map[string]any{"region": true}); err != nil {
		t.Fatal(err)
	}
	if err := r.Enable("dev", "rollback-schema", "1.0.0"); err != nil {
		t.Fatal(err)
	}
	if err := r.Enable("dev", "rollback-schema", "2.0.0"); err != nil {
		t.Fatal(err)
	}
	if err := r.Promote("dev", "rollback-schema", "1.0.0"); err != nil {
		t.Fatal(err)
	}
	if !errors.Is(r.Rollback("dev", "rollback-schema", "2.0.0"), plugin.ErrSchemaIncompatible) {
		t.Fatal("rollback bypassed schema compatibility")
	}
}

func TestPromotionRequiresApprovalForNewCapabilities(t *testing.T) {
	r := plugin.NewRegistry(1)
	for _, manifest := range []plugin.Manifest{
		{ID: "approval", Version: "1.0.0", HostAPI: 1, Capabilities: []string{"read"}},
		{ID: "approval", Version: "2.0.0", HostAPI: 1, Capabilities: []string{"read", "write"}},
	} {
		if err := r.Register(manifest); err != nil {
			t.Fatal(err)
		}
		if err := r.SetReadiness(manifest.ID, manifest.Version, true); err != nil {
			t.Fatal(err)
		}
		if err := r.Install("dev", manifest.ID, manifest.Version); err != nil {
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
	if err := r.Promote("dev", "approval", "1.0.0"); err != nil {
		t.Fatal(err)
	}
	if !errors.Is(r.Promote("dev", "approval", "2.0.0"), plugin.ErrCapabilityApproval) {
		t.Fatal("new capability was promoted without approval")
	}
	if err := r.ApproveCapability("dev", "approval", "2.0.0", "write"); err != nil {
		t.Fatal(err)
	}
	if err := r.Promote("dev", "approval", "2.0.0"); err != nil {
		t.Fatalf("approved capability rejected: %v", err)
	}
}

func TestTransitiveDependentsAreCalculated(t *testing.T) {
	r := plugin.NewRegistry(1)
	manifests := []plugin.Manifest{
		{ID: "base", Version: "1.0.0", HostAPI: 1},
		{ID: "middle", Version: "1.0.0", HostAPI: 1, Dependencies: []string{"base@1.0.0"}},
		{ID: "top", Version: "1.0.0", HostAPI: 1, Dependencies: []string{"middle@1.0.0"}},
	}
	for _, manifest := range manifests {
		if err := r.Register(manifest); err != nil {
			t.Fatal(err)
		}
	}
	got := r.TransitiveDependents("base", "1.0.0")
	want := []string{"middle@1.0.0", "top@1.0.0"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("dependents = %#v, want %#v", got, want)
	}
}

func TestDisablePreviewCapturesLifecycleRevisionAndDependents(t *testing.T) {
	r := plugin.NewRegistry(1)
	base := plugin.Manifest{ID: "preview-base", Version: "1.0.0", HostAPI: 1}
	dependent := plugin.Manifest{ID: "preview-dependent", Version: "1.0.0", HostAPI: 1, Dependencies: []string{"preview-base@1.0.0"}}
	if err := r.Register(base); err != nil {
		t.Fatal(err)
	}
	if err := r.Register(dependent); err != nil {
		t.Fatal(err)
	}
	if err := r.Install("dev", base.ID, base.Version); err != nil {
		t.Fatal(err)
	}
	preview, err := r.PreviewDisable("dev", base.ID, base.Version)
	if err != nil {
		t.Fatal(err)
	}
	if preview.Revision != 0 || !reflect.DeepEqual(preview.Dependents, []string{"preview-dependent@1.0.0"}) {
		t.Fatalf("preview = %#v", preview)
	}
}

func TestDisablePreviewRejectsStaleRevision(t *testing.T) {
	r := plugin.NewRegistry(1)
	m := plugin.Manifest{ID: "stale-preview", Version: "1.0.0", HostAPI: 1}
	if err := r.Register(m); err != nil {
		t.Fatal(err)
	}
	if err := r.SetReadiness(m.ID, m.Version, true); err != nil {
		t.Fatal(err)
	}
	if err := r.Install("dev", m.ID, m.Version); err != nil {
		t.Fatal(err)
	}
	if err := r.Enable("dev", m.ID, m.Version); err != nil {
		t.Fatal(err)
	}
	preview, err := r.PreviewDisable("dev", m.ID, m.Version)
	if err != nil {
		t.Fatal(err)
	}
	if err := r.Promote("dev", m.ID, m.Version); err != nil {
		t.Fatal(err)
	}
	if !errors.Is(r.ApplyDisablePreview(preview), plugin.ErrStalePreview) {
		t.Fatal("stale disable preview was applied")
	}
}

func TestDrainBlocksNewWorkButAllowsReconnectUntilDeadline(t *testing.T) {
	r := plugin.NewRegistry(1)
	m := plugin.Manifest{ID: "drain-admission", Version: "1.0.0", HostAPI: 1}
	if err := r.Register(m); err != nil {
		t.Fatal(err)
	}
	if err := r.Install("dev", m.ID, m.Version); err != nil {
		t.Fatal(err)
	}
	deadline := time.Unix(500, 0)
	if err := r.DrainUntil("dev", m.ID, m.Version, deadline); err != nil {
		t.Fatal(err)
	}
	if !errors.Is(r.CanStartNew("dev", m.ID, m.Version), plugin.ErrDisabled) {
		t.Fatal("new work entered a draining plugin")
	}
	if err := r.CanReconnect("dev", m.ID, m.Version, time.Unix(400, 0)); err != nil {
		t.Fatalf("reconnect was blocked before deadline: %v", err)
	}
	if !errors.Is(r.CanReconnect("dev", m.ID, m.Version, deadline), plugin.ErrDisabled) {
		t.Fatal("reconnect was allowed after drain deadline")
	}
}

func TestDrainScopeAndDeadlinePersistInInstallationStatus(t *testing.T) {
	r := plugin.NewRegistry(1)
	m := plugin.Manifest{ID: "persisted-drain", Version: "1.0.0", HostAPI: 1}
	if err := r.Register(m); err != nil {
		t.Fatal(err)
	}
	if err := r.Install("dev", m.ID, m.Version); err != nil {
		t.Fatal(err)
	}
	deadline := time.Unix(900, 0)
	if err := r.DrainUntil("dev", m.ID, m.Version, deadline); err != nil {
		t.Fatal(err)
	}
	status, err := r.Status("dev", m.ID, m.Version)
	if err != nil {
		t.Fatal(err)
	}
	if !status.Draining || !status.DrainUntil.Equal(deadline) {
		t.Fatalf("drain status = %#v", status)
	}
}

func TestDrainUsesFiveMinuteDefaultWindow(t *testing.T) {
	r := plugin.NewRegistry(1)
	m := plugin.Manifest{ID: "default-drain", Version: "1.0.0", HostAPI: 1}
	if err := r.Register(m); err != nil {
		t.Fatal(err)
	}
	if err := r.Install("dev", m.ID, m.Version); err != nil {
		t.Fatal(err)
	}
	before := time.Now().Add(5 * time.Minute)
	if err := r.Drain("dev", m.ID, m.Version); err != nil {
		t.Fatal(err)
	}
	status, err := r.Status("dev", m.ID, m.Version)
	if err != nil {
		t.Fatal(err)
	}
	if status.DrainUntil.Before(before.Add(-2*time.Second)) || status.DrainUntil.After(time.Now().Add(5*time.Minute+2*time.Second)) {
		t.Fatalf("default drain deadline = %v", status.DrainUntil)
	}
}

func TestImmediateStopBlocksNewWorkAndReconnects(t *testing.T) {
	r := plugin.NewRegistry(1)
	m := plugin.Manifest{ID: "immediate-stop", Version: "1.0.0", HostAPI: 1}
	if err := r.Register(m); err != nil {
		t.Fatal(err)
	}
	if err := r.Install("dev", m.ID, m.Version); err != nil {
		t.Fatal(err)
	}
	if err := r.StopImmediately("dev", m.ID, m.Version); err != nil {
		t.Fatal(err)
	}
	if !errors.Is(r.CanStartNew("dev", m.ID, m.Version), plugin.ErrDisabled) {
		t.Fatal("immediate stop admitted new work")
	}
	if !errors.Is(r.CanReconnect("dev", m.ID, m.Version, time.Unix(1, 0)), plugin.ErrDisabled) {
		t.Fatal("immediate stop admitted reconnect")
	}
}

func TestProjectUninstallWaitsForActiveWork(t *testing.T) {
	r := plugin.NewRegistry(1)
	m := plugin.Manifest{ID: "uninstall-wait", Version: "1.0.0", HostAPI: 1}
	if err := r.Register(m); err != nil {
		t.Fatal(err)
	}
	if err := r.Install("dev", m.ID, m.Version); err != nil {
		t.Fatal(err)
	}
	if err := r.SetReadiness(m.ID, m.Version, true); err != nil {
		t.Fatal(err)
	}
	if err := r.Enable("dev", m.ID, m.Version); err != nil {
		t.Fatal(err)
	}
	if err := r.PinWork("room-1", "dev", m.ID, m.Version); err != nil {
		t.Fatal(err)
	}
	if !errors.Is(r.UninstallProject("dev"), plugin.ErrActiveWork) {
		t.Fatal("uninstall ignored active work")
	}
	if !errors.Is(r.Remove("dev", m.ID, m.Version), plugin.ErrVersionReferenced) {
		t.Fatal("active pinned installation was removed")
	}
}

func TestProjectUninstallRevokesProjectGrants(t *testing.T) {
	r := plugin.NewRegistry(1)
	m := plugin.Manifest{ID: "uninstall-grants", Version: "1.0.0", HostAPI: 1, Capabilities: []string{"records.read"}}
	if err := r.Register(m); err != nil {
		t.Fatal(err)
	}
	if err := r.Install("dev", m.ID, m.Version); err != nil {
		t.Fatal(err)
	}
	if err := r.Grant("dev", m.ID, m.Version, "records.read"); err != nil {
		t.Fatal(err)
	}
	if err := r.UninstallProject("dev"); err != nil {
		t.Fatal(err)
	}
	if !errors.Is(r.Authorize("dev", m.ID, m.Version, "records.read"), plugin.ErrNotInstalled) {
		t.Fatal("project grant survived uninstall")
	}
	found := false
	for _, event := range r.Audit() {
		if event.Action == "revoke_project_grant" && event.Detail == "records.read" {
			found = true
		}
	}
	if !found {
		t.Fatal("project grant revocation was not audited")
	}
}

func TestProjectUninstallPreservesDurablePluginData(t *testing.T) {
	r := plugin.NewRegistry(1)
	if err := r.SaveDurableData("dev", "durable", "record-1", []byte("keep-me")); err != nil {
		t.Fatal(err)
	}
	if err := r.UninstallProject("dev"); err != nil {
		t.Fatal(err)
	}
	value, ok := r.DurableData("dev", "durable", "record-1")
	if !ok || string(value) != "keep-me" {
		t.Fatalf("durable data = %q, %t", value, ok)
	}
}
