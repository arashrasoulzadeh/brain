package plugin_test

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sync"
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

func TestLoadManifestRejectsSourcePathTraversal(t *testing.T) {
	directory := t.TempDir()
	manifestPath := filepath.Join(directory, "plugin.yaml")
	unsafe := []byte("id: sample\nentrypoint: index.ts\nversion: 1.0.0\nhost_api: 1\nsource_files:\n  - index.ts\n  - ../outside.ts\nimports:\n  - ../sdk-typescript/src/index.js\n")
	if err := os.WriteFile(manifestPath, unsafe, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := plugin.LoadManifest(manifestPath); !errors.Is(err, plugin.ErrInvalidManifest) {
		t.Fatalf("source path traversal accepted: %v", err)
	}
}

func TestLoadManifestRejectsDuplicateSourceFiles(t *testing.T) {
	directory := t.TempDir()
	manifestPath := filepath.Join(directory, "plugin.yaml")
	unsafe := []byte("id: sample\nentrypoint: index.ts\nversion: 1.0.0\nhost_api: 1\nsource_files:\n  - index.ts\n  - index.ts\nimports:\n  - ../sdk-typescript/src/index.js\n")
	if err := os.WriteFile(manifestPath, unsafe, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := plugin.LoadManifest(manifestPath); !errors.Is(err, plugin.ErrInvalidManifest) {
		t.Fatalf("duplicate source file accepted: %v", err)
	}
}

func TestLoadManifestRejectsAmbiguousRoutes(t *testing.T) {
	directory := t.TempDir()
	manifestPath := filepath.Join(directory, "plugin.yaml")
	unsafe := []byte("id: sample\nentrypoint: index.ts\nversion: 1.0.0\nhost_api: 1\nsource_files:\n  - index.ts\nimports:\n  - ../sdk-typescript/src/index.js\nroutes:\n  - operation: sample.run\n    function: run\n    capability: sample.run\n  - operation: sample.run\n    function: other\n    capability: sample.run\n")
	if err := os.WriteFile(manifestPath, unsafe, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := plugin.LoadManifest(manifestPath); !errors.Is(err, plugin.ErrInvalidManifest) {
		t.Fatalf("ambiguous route accepted: %v", err)
	}
}

func TestLoadManifestRejectsUndeclaredRouteCapability(t *testing.T) {
	directory := t.TempDir()
	manifestPath := filepath.Join(directory, "plugin.yaml")
	unsafe := []byte("id: sample\nentrypoint: index.ts\nversion: 1.0.0\nhost_api: 1\ncapabilities:\n  - sample.read\nsource_files:\n  - index.ts\nimports:\n  - ../sdk-typescript/src/index.js\nroutes:\n  - operation: sample.write\n    function: write\n    capability: sample.write\n")
	if err := os.WriteFile(manifestPath, unsafe, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := plugin.LoadManifest(manifestPath); !errors.Is(err, plugin.ErrInvalidManifest) {
		t.Fatalf("undeclared route capability accepted: %v", err)
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

func TestValidatePluginSourcesRejectsUndeclaredDynamicImports(t *testing.T) {
	directory := t.TempDir()
	if err := os.WriteFile(filepath.Join(directory, "index.ts"), []byte("const dependency = require(\"node:fs\");\n"), 0600); err != nil {
		t.Fatal(err)
	}
	manifest := plugin.Manifest{ID: "sample", Entry: "index.ts", Version: "1.0.0", HostAPI: 1, SourceFiles: []string{"index.ts"}, Imports: []string{"node:path"}}
	if err := plugin.ValidatePluginSources(directory, manifest); !errors.Is(err, plugin.ErrInvalidManifest) {
		t.Fatalf("dynamic undeclared import error = %v", err)
	}
}

func TestValidatePluginSourcesRejectsSourcePathTraversal(t *testing.T) {
	directory := t.TempDir()
	manifest := plugin.Manifest{ID: "sample", Entry: "index.ts", Version: "1.0.0", HostAPI: 1, SourceFiles: []string{"index.ts", "../outside.ts"}, Imports: []string{"node:path"}}
	if err := plugin.ValidatePluginSources(directory, manifest); !errors.Is(err, plugin.ErrInvalidManifest) {
		t.Fatalf("source path traversal error = %v", err)
	}
}

func TestValidatePluginSourcesRejectsSymlinkedSource(t *testing.T) {
	directory := t.TempDir()
	outside := filepath.Join(t.TempDir(), "outside.ts")
	if err := os.WriteFile(outside, []byte("export const secret = true;\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(directory, "index.ts")); err != nil {
		t.Fatal(err)
	}
	manifest := plugin.Manifest{ID: "sample", Entry: "index.ts", Version: "1.0.0", HostAPI: 1, SourceFiles: []string{"index.ts"}}
	if err := plugin.ValidatePluginSources(directory, manifest); !errors.Is(err, plugin.ErrInvalidManifest) {
		t.Fatalf("symlinked source accepted: %v", err)
	}
}

func TestValidatePluginSourcesRejectsSymlinkedSourceDirectory(t *testing.T) {
	directory := t.TempDir()
	outside := t.TempDir()
	if err := os.WriteFile(filepath.Join(outside, "index.ts"), []byte("export const secret = true;\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(directory, "src")); err != nil {
		t.Fatal(err)
	}
	manifest := plugin.Manifest{ID: "sample", Entry: "src/index.ts", Version: "1.0.0", HostAPI: 1, SourceFiles: []string{"src/index.ts"}}
	if err := plugin.ValidatePluginSources(directory, manifest); !errors.Is(err, plugin.ErrInvalidManifest) {
		t.Fatalf("symlinked source directory accepted: %v", err)
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

func TestRegistryDurableDataSupportsConcurrentAccess(t *testing.T) {
	r := plugin.NewRegistry(1)
	var group sync.WaitGroup
	for index := 0; index < 32; index++ {
		index := index
		group.Add(1)
		go func() {
			defer group.Done()
			key := fmt.Sprintf("key-%d", index)
			if err := r.SaveDurableData("dev", "sample", key, []byte(key)); err != nil {
				t.Errorf("SaveDurableData(%q) = %v", key, err)
				return
			}
			if value, ok := r.DurableData("dev", "sample", key); !ok || string(value) != key {
				t.Errorf("DurableData(%q) = %q, %v", key, value, ok)
			}
		}()
	}
	group.Wait()
}

func TestRegistryReadinessSupportsConcurrentStatusReads(t *testing.T) {
	r := plugin.NewRegistry(1)
	manifest := plugin.Manifest{ID: "sample", Version: "1.0.0", HostAPI: 1}
	if err := r.Register(manifest); err != nil {
		t.Fatal(err)
	}
	if err := r.Install("dev", manifest.ID, manifest.Version); err != nil {
		t.Fatal(err)
	}
	var group sync.WaitGroup
	for index := 0; index < 32; index++ {
		group.Add(1)
		go func(index int) {
			defer group.Done()
			if err := r.SetReadiness(manifest.ID, manifest.Version, index%2 == 0); err != nil {
				t.Errorf("SetReadiness() = %v", err)
			}
			if _, err := r.Status("dev", manifest.ID, manifest.Version); err != nil {
				t.Errorf("Status() = %v", err)
			}
		}(index)
	}
	group.Wait()
}

func TestRegistryRejectsInstallationWithoutProjectIdentity(t *testing.T) {
	r := plugin.NewRegistry(1)
	manifest := plugin.Manifest{ID: "sample", Version: "1.0.0", HostAPI: 1}
	if err := r.Register(manifest); err != nil {
		t.Fatal(err)
	}
	if err := r.Install("", manifest.ID, manifest.Version); !errors.Is(err, plugin.ErrProjectInvalid) {
		t.Fatalf("empty project installation error = %v, want %v", err, plugin.ErrProjectInvalid)
	}
}

func TestRegistryRejectsProjectlessLifecycleMutations(t *testing.T) {
	r := plugin.NewRegistry(1)
	manifest := plugin.Manifest{ID: "sample", Version: "1.0.0", HostAPI: 1}
	if err := r.Register(manifest); err != nil {
		t.Fatal(err)
	}
	if err := r.UninstallProject(""); !errors.Is(err, plugin.ErrProjectInvalid) {
		t.Fatalf("empty project uninstall error = %v, want %v", err, plugin.ErrProjectInvalid)
	}
	if err := r.PinWork("work", "", manifest.ID, manifest.Version); !errors.Is(err, plugin.ErrProjectInvalid) {
		t.Fatalf("empty project pin error = %v, want %v", err, plugin.ErrProjectInvalid)
	}
}

func TestRegistryReadinessAndUnregisterAreConcurrentSafe(t *testing.T) {
	r := plugin.NewRegistry(1)
	manifest := plugin.Manifest{ID: "readiness-unregister", Version: "1.0.0", HostAPI: 1}
	if err := r.Register(manifest); err != nil {
		t.Fatal(err)
	}

	var group sync.WaitGroup
	group.Add(2)
	go func() {
		defer group.Done()
		_ = r.SetReadiness(manifest.ID, manifest.Version, true)
	}()
	go func() {
		defer group.Done()
		_ = r.Unregister(manifest.ID, manifest.Version)
	}()
	group.Wait()
}

func TestRegistryRejectsProjectlessRemoval(t *testing.T) {
	r := plugin.NewRegistry(1)
	if err := r.Remove("", "sample", "1.0.0"); !errors.Is(err, plugin.ErrProjectInvalid) {
		t.Fatalf("empty project removal error = %v, want %v", err, plugin.ErrProjectInvalid)
	}
}

func TestRegistryRejectsProjectlessAdmissionAndLifecycleChecks(t *testing.T) {
	r := plugin.NewRegistry(1)
	if err := r.Enable("", "sample", "1.0.0"); !errors.Is(err, plugin.ErrProjectInvalid) {
		t.Fatalf("empty project enable error = %v", err)
	}
	if err := r.Promote("", "sample", "1.0.0"); !errors.Is(err, plugin.ErrProjectInvalid) {
		t.Fatalf("empty project promote error = %v", err)
	}
	if err := r.DrainUntil("", "sample", "1.0.0", time.Now()); !errors.Is(err, plugin.ErrProjectInvalid) {
		t.Fatalf("empty project drain error = %v", err)
	}
	if err := r.StopImmediately("", "sample", "1.0.0"); !errors.Is(err, plugin.ErrProjectInvalid) {
		t.Fatalf("empty project stop error = %v", err)
	}
	if err := r.Recover("", "sample", "1.0.0"); !errors.Is(err, plugin.ErrProjectInvalid) {
		t.Fatalf("empty project recover error = %v", err)
	}
	if err := r.CanStart("", "sample", "1.0.0", time.Now()); !errors.Is(err, plugin.ErrProjectInvalid) {
		t.Fatalf("empty project can-start error = %v", err)
	}
	if err := r.CanStartNew("", "sample", "1.0.0"); !errors.Is(err, plugin.ErrProjectInvalid) {
		t.Fatalf("empty project can-start-new error = %v", err)
	}
	if err := r.CanReconnect("", "sample", "1.0.0", time.Now()); !errors.Is(err, plugin.ErrProjectInvalid) {
		t.Fatalf("empty project can-reconnect error = %v", err)
	}
	if version, ok := r.ActiveVersion("", "sample"); version != "" || ok {
		t.Fatalf("empty project active version = %q, %t", version, ok)
	}
}

func TestRegistryAuditSupportsConcurrentLifecycleWritesAndReads(t *testing.T) {
	r := plugin.NewRegistry(1)
	manifest := plugin.Manifest{ID: "audit-concurrency", Version: "1.0.0", HostAPI: 1}
	if err := r.Register(manifest); err != nil {
		t.Fatal(err)
	}
	if err := r.Install("dev", manifest.ID, manifest.Version); err != nil {
		t.Fatal(err)
	}

	var group sync.WaitGroup
	for index := 0; index < 32; index++ {
		group.Add(1)
		go func(index int) {
			defer group.Done()
			if index%2 == 0 {
				_ = r.SetReadiness(manifest.ID, manifest.Version, true)
				return
			}
			_ = r.Audit()
		}(index)
	}
	group.Wait()
}

func TestRegistryAuthorizationStateSupportsConcurrentGrantChanges(t *testing.T) {
	r := plugin.NewRegistry(1)
	manifest := plugin.Manifest{ID: "grant-concurrency", Version: "1.0.0", HostAPI: 1, Capabilities: []string{"read"}}
	if err := r.Register(manifest); err != nil {
		t.Fatal(err)
	}
	if err := r.Install("dev", manifest.ID, manifest.Version); err != nil {
		t.Fatal(err)
	}
	if err := r.Grant("dev", manifest.ID, manifest.Version, "read"); err != nil {
		t.Fatal(err)
	}
	if err := r.SetReadiness(manifest.ID, manifest.Version, true); err != nil {
		t.Fatal(err)
	}
	if err := r.Enable("dev", manifest.ID, manifest.Version); err != nil {
		t.Fatal(err)
	}

	var group sync.WaitGroup
	for index := 0; index < 32; index++ {
		group.Add(1)
		go func(index int) {
			defer group.Done()
			switch index % 3 {
			case 0:
				_ = r.Grant("dev", manifest.ID, manifest.Version, "read")
			case 1:
				_ = r.Revoke("dev", manifest.ID, manifest.Version, "read")
			default:
				_ = r.Authorize("dev", manifest.ID, manifest.Version, "read")
			}
		}(index)
	}
	group.Wait()
}

func TestRegistryInstallationAndStatusAreConcurrentSafe(t *testing.T) {
	r := plugin.NewRegistry(1)
	manifest := plugin.Manifest{ID: "installation-concurrency", Version: "1.0.0", HostAPI: 1}
	if err := r.Register(manifest); err != nil {
		t.Fatal(err)
	}

	var group sync.WaitGroup
	for index := 0; index < 32; index++ {
		group.Add(1)
		go func(index int) {
			defer group.Done()
			if index%2 == 0 {
				_ = r.Install("dev", manifest.ID, manifest.Version)
				return
			}
			_, _ = r.Status("dev", manifest.ID, manifest.Version)
		}(index)
	}
	group.Wait()
}

func TestRegistryActiveVersionSupportsConcurrentPromotionAndReads(t *testing.T) {
	r := plugin.NewRegistry(1)
	manifest := plugin.Manifest{ID: "active-concurrency", Version: "1.0.0", HostAPI: 1}
	if err := r.Register(manifest); err != nil {
		t.Fatal(err)
	}
	if err := r.Install("dev", manifest.ID, manifest.Version); err != nil {
		t.Fatal(err)
	}
	if err := r.SetReadiness(manifest.ID, manifest.Version, true); err != nil {
		t.Fatal(err)
	}
	if err := r.Enable("dev", manifest.ID, manifest.Version); err != nil {
		t.Fatal(err)
	}

	var group sync.WaitGroup
	for index := 0; index < 32; index++ {
		group.Add(1)
		go func(index int) {
			defer group.Done()
			if index%2 == 0 {
				_ = r.Promote("dev", manifest.ID, manifest.Version)
				return
			}
			_, _ = r.ActiveVersion("dev", manifest.ID)
		}(index)
	}
	group.Wait()
}

func TestRegistryPinsSupportConcurrentWritesAndReads(t *testing.T) {
	r := plugin.NewRegistry(1)
	manifest := plugin.Manifest{ID: "pin-concurrency", Version: "1.0.0", HostAPI: 1}
	if err := r.Register(manifest); err != nil {
		t.Fatal(err)
	}
	if err := r.Install("dev", manifest.ID, manifest.Version); err != nil {
		t.Fatal(err)
	}
	if err := r.SetReadiness(manifest.ID, manifest.Version, true); err != nil {
		t.Fatal(err)
	}
	if err := r.Enable("dev", manifest.ID, manifest.Version); err != nil {
		t.Fatal(err)
	}

	var group sync.WaitGroup
	for index := 0; index < 32; index++ {
		group.Add(1)
		go func(index int) {
			defer group.Done()
			if index%2 == 0 {
				_ = r.PinWork("work", "dev", manifest.ID, manifest.Version)
				return
			}
			_, _ = r.PinnedVersion("work")
		}(index)
	}
	group.Wait()
}

func TestRegistryApprovalAndArtifactStateAreConcurrentSafe(t *testing.T) {
	r := plugin.NewRegistry(1)
	manifest := plugin.Manifest{ID: "approval-concurrency", Version: "1.0.0", HostAPI: 1, Capabilities: []string{"read"}}
	if err := r.Register(manifest); err != nil {
		t.Fatal(err)
	}
	if err := r.Install("dev", manifest.ID, manifest.Version); err != nil {
		t.Fatal(err)
	}
	if err := r.SetReadiness(manifest.ID, manifest.Version, true); err != nil {
		t.Fatal(err)
	}
	if err := r.Grant("dev", manifest.ID, manifest.Version, "read"); err != nil {
		t.Fatal(err)
	}
	if err := r.Enable("dev", manifest.ID, manifest.Version); err != nil {
		t.Fatal(err)
	}

	var group sync.WaitGroup
	for index := 0; index < 32; index++ {
		group.Add(1)
		go func(index int) {
			defer group.Done()
			if index%2 == 0 {
				_ = r.ApproveCapability("dev", manifest.ID, manifest.Version, "read")
				_ = r.ApproveArtifact(manifest.ID, manifest.Version, "https://example.com/plugin.tgz")
				return
			}
			_, _ = r.ApprovedArtifact(manifest.ID, manifest.Version)
			_ = r.Promote("dev", manifest.ID, manifest.Version)
		}(index)
	}
	group.Wait()
}

func TestRegistryRemovalAndPinningAreConcurrentSafe(t *testing.T) {
	r := plugin.NewRegistry(1)
	manifest := plugin.Manifest{ID: "remove-pin-concurrency", Version: "1.0.0", HostAPI: 1}
	if err := r.Register(manifest); err != nil {
		t.Fatal(err)
	}
	if err := r.Install("dev", manifest.ID, manifest.Version); err != nil {
		t.Fatal(err)
	}
	if err := r.SetReadiness(manifest.ID, manifest.Version, true); err != nil {
		t.Fatal(err)
	}
	if err := r.Enable("dev", manifest.ID, manifest.Version); err != nil {
		t.Fatal(err)
	}

	var group sync.WaitGroup
	group.Add(2)
	go func() {
		defer group.Done()
		_ = r.PinWork("work", "dev", manifest.ID, manifest.Version)
	}()
	go func() {
		defer group.Done()
		_ = r.Remove("dev", manifest.ID, manifest.Version)
	}()
	group.Wait()
}

func TestRegistryUninstallAndPinningAreConcurrentSafe(t *testing.T) {
	r := plugin.NewRegistry(1)
	manifest := plugin.Manifest{ID: "uninstall-pin-concurrency", Version: "1.0.0", HostAPI: 1}
	if err := r.Register(manifest); err != nil {
		t.Fatal(err)
	}
	if err := r.Install("dev", manifest.ID, manifest.Version); err != nil {
		t.Fatal(err)
	}
	if err := r.SetReadiness(manifest.ID, manifest.Version, true); err != nil {
		t.Fatal(err)
	}
	if err := r.Enable("dev", manifest.ID, manifest.Version); err != nil {
		t.Fatal(err)
	}

	var group sync.WaitGroup
	group.Add(2)
	go func() {
		defer group.Done()
		_ = r.PinWork("work", "dev", manifest.ID, manifest.Version)
	}()
	go func() {
		defer group.Done()
		_ = r.UninstallProject("dev")
	}()
	group.Wait()
}

func TestRegistryUnregisterClearsCapabilityApprovals(t *testing.T) {
	r := plugin.NewRegistry(1)
	base := plugin.Manifest{ID: "approval-cleanup", Version: "1.0.0", HostAPI: 1, Capabilities: []string{"read"}}
	candidate := plugin.Manifest{ID: "approval-cleanup", Version: "2.0.0", HostAPI: 1, Capabilities: []string{"read", "write"}}
	for _, manifest := range []plugin.Manifest{base, candidate} {
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
	if err := r.Grant("dev", base.ID, base.Version, "read"); err != nil {
		t.Fatal(err)
	}
	if err := r.Grant("dev", candidate.ID, candidate.Version, "read"); err != nil {
		t.Fatal(err)
	}
	if err := r.Grant("dev", candidate.ID, candidate.Version, "write"); err != nil {
		t.Fatal(err)
	}
	if err := r.Enable("dev", base.ID, base.Version); err != nil {
		t.Fatal(err)
	}
	if err := r.Enable("dev", candidate.ID, candidate.Version); err != nil {
		t.Fatal(err)
	}
	if err := r.Promote("dev", base.ID, base.Version); err != nil {
		t.Fatal(err)
	}
	if err := r.ApproveCapability("dev", candidate.ID, candidate.Version, "write"); err != nil {
		t.Fatal(err)
	}
	if err := r.Remove("dev", candidate.ID, candidate.Version); err != nil {
		t.Fatal(err)
	}
	if err := r.Unregister(candidate.ID, candidate.Version); err != nil {
		t.Fatal(err)
	}
	if err := r.Register(candidate); err != nil {
		t.Fatal(err)
	}
	if err := r.SetReadiness(candidate.ID, candidate.Version, true); err != nil {
		t.Fatal(err)
	}
	if err := r.Install("dev", candidate.ID, candidate.Version); err != nil {
		t.Fatal(err)
	}
	if err := r.Grant("dev", candidate.ID, candidate.Version, "read"); err != nil {
		t.Fatal(err)
	}
	if err := r.Grant("dev", candidate.ID, candidate.Version, "write"); err != nil {
		t.Fatal(err)
	}
	if err := r.Enable("dev", candidate.ID, candidate.Version); err != nil {
		t.Fatal(err)
	}
	if !errors.Is(r.Promote("dev", candidate.ID, candidate.Version), plugin.ErrCapabilityApproval) {
		t.Fatal("unregistered capability approval was retained")
	}
}

func TestRegistryRegistrationAndUnregistrationAreConcurrentSafe(t *testing.T) {
	r := plugin.NewRegistry(1)
	manifest := plugin.Manifest{ID: "registration-concurrency", Version: "1.0.0", HostAPI: 1}
	if err := r.Register(manifest); err != nil {
		t.Fatal(err)
	}

	var group sync.WaitGroup
	group.Add(2)
	go func() {
		defer group.Done()
		_ = r.Register(plugin.Manifest{ID: manifest.ID, Version: "2.0.0", HostAPI: 1})
	}()
	go func() {
		defer group.Done()
		_ = r.Unregister(manifest.ID, manifest.Version)
	}()
	group.Wait()
}

func TestRevokeStopsTheProjectWorker(t *testing.T) {
	r := plugin.NewRegistry(1)
	m := plugin.Manifest{ID: "guest", Version: "1.0.0", HostAPI: 1, Capabilities: []string{"session.issue"}}
	if err := r.Register(m); err != nil {
		t.Fatal(err)
	}
	if err := r.Install("dev", m.ID, m.Version); err != nil {
		t.Fatal(err)
	}
	if err := r.Grant("dev", m.ID, m.Version, "session.issue"); err != nil {
		t.Fatal(err)
	}
	s := plugin.NewSupervisor()
	w := &supervisorTestWorker{}
	if err := s.Start("dev:guest@1.0.0", w); err != nil {
		t.Fatal(err)
	}
	r.BindSupervisor(s)
	if err := r.Revoke("dev", m.ID, m.Version, "session.issue"); err != nil {
		t.Fatal(err)
	}
	if !w.stopped || s.Ready("dev:guest@1.0.0") {
		t.Fatal("revocation left project worker running")
	}
}

func TestApplyDisablePreviewStopsTheProjectWorker(t *testing.T) {
	r := plugin.NewRegistry(1)
	m := plugin.Manifest{ID: "rooms", Version: "1.0.0", HostAPI: 1}
	if err := r.Register(m); err != nil {
		t.Fatal(err)
	}
	if err := r.Install("dev", m.ID, m.Version); err != nil {
		t.Fatal(err)
	}
	s := plugin.NewSupervisor()
	w := &supervisorTestWorker{}
	if err := s.Start("dev:rooms@1.0.0", w); err != nil {
		t.Fatal(err)
	}
	r.BindSupervisor(s)
	preview, err := r.PreviewDisable("dev", m.ID, m.Version)
	if err != nil {
		t.Fatal(err)
	}
	if err := r.ApplyDisablePreview(preview); err != nil {
		t.Fatal(err)
	}
	if !w.stopped || s.Ready("dev:rooms@1.0.0") {
		t.Fatal("disable left project worker running")
	}
}

type supervisorTestWorker struct{ stopped bool }

func (w *supervisorTestWorker) Start() error { return nil }
func (w *supervisorTestWorker) Stop() error  { w.stopped = true; return nil }
func (w *supervisorTestWorker) Ready() bool  { return !w.stopped }

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

func TestDisablePreviewApplicationIsSingleUseUnderConcurrency(t *testing.T) {
	r := plugin.NewRegistry(1)
	m := plugin.Manifest{ID: "concurrent-preview", Version: "1.0.0", HostAPI: 1}
	if err := r.Register(m); err != nil {
		t.Fatal(err)
	}
	if err := r.Install("dev", m.ID, m.Version); err != nil {
		t.Fatal(err)
	}
	preview, err := r.PreviewDisable("dev", m.ID, m.Version)
	if err != nil {
		t.Fatal(err)
	}

	results := make(chan error, 2)
	var group sync.WaitGroup
	for index := 0; index < 2; index++ {
		group.Add(1)
		go func() {
			defer group.Done()
			results <- r.ApplyDisablePreview(preview)
		}()
	}
	group.Wait()
	close(results)

	successes := 0
	stale := 0
	for result := range results {
		switch {
		case result == nil:
			successes++
		case errors.Is(result, plugin.ErrStalePreview):
			stale++
		default:
			t.Fatalf("unexpected concurrent preview result: %v", result)
		}
	}
	if successes != 1 || stale != 1 {
		t.Fatalf("preview results = successes %d, stale %d; want one each", successes, stale)
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
	if err := r.Recover("dev", m.ID, m.Version); err != nil {
		t.Fatal(err)
	}
	if err := r.CanStartNew("dev", m.ID, m.Version); err != nil {
		t.Fatalf("recovery did not restore admission: %v", err)
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
	found := false
	for _, event := range r.Audit() {
		if event.Action == "uninstall_project" && event.ProjectID == "dev" {
			found = true
		}
	}
	if !found {
		t.Fatal("project uninstall was not audited")
	}
	if err := r.DeleteDurableData("dev", "durable", "record-1"); err != nil {
		t.Fatal(err)
	}
	if _, ok := r.DurableData("dev", "durable", "record-1"); ok {
		t.Fatal("durable data deletion did not remain explicit")
	}
}

func TestUnregisterRejectsReferencedServerVersion(t *testing.T) {
	r := plugin.NewRegistry(1)
	base := plugin.Manifest{ID: "server-base", Version: "1.0.0", HostAPI: 1}
	dependent := plugin.Manifest{ID: "server-dependent", Version: "1.0.0", HostAPI: 1, Dependencies: []string{"server-base@1.0.0"}}
	if err := r.Register(base); err != nil {
		t.Fatal(err)
	}
	if err := r.Register(dependent); err != nil {
		t.Fatal(err)
	}
	if !errors.Is(r.Unregister(base.ID, base.Version), plugin.ErrVersionReferenced) {
		t.Fatal("referenced server version was unregistered")
	}
	if err := r.Unregister(dependent.ID, dependent.Version); err != nil {
		t.Fatal(err)
	}
	if err := r.Unregister(base.ID, base.Version); err != nil {
		t.Fatal(err)
	}
}

func TestManifestDeclaresAdminExtensionPoint(t *testing.T) {
	m := plugin.Manifest{ID: "admin", Version: "1.0.0", HostAPI: 1, ExtensionPoints: []string{"admin.navigation"}}
	if !m.SupportsExtensionPoint("admin.navigation") || m.SupportsExtensionPoint("admin.unknown") {
		t.Fatal("manifest extension point discovery failed")
	}
}
