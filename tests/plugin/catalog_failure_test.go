package plugin_test

import (
	"errors"
	"testing"

	"brain/internal/plugin"
)

func TestRegistryCatalogOrderingAndReferenceCleanup(t *testing.T) {
	r := plugin.NewRegistry(1)
	first := plugin.Manifest{ID: "sample", Version: "1", HostAPI: 1, Capabilities: []string{"read"}}
	second := first
	second.Version = "2"
	second.Capabilities = []string{"read", "write"}
	invalid := first
	invalid.ExtensionPoints = []string{" "}
	if err := r.Register(invalid); !errors.Is(err, plugin.ErrInvalidManifest) {
		t.Fatalf("invalid extension: %v", err)
	}
	for _, m := range []plugin.Manifest{second, first} {
		if err := r.Register(m); err != nil {
			t.Fatal(err)
		}
		installAccessTestPlugin(t, r, "alpha", m, &contractWorker{ready: true})
	}
	manifests := r.RegisteredManifests()
	if len(manifests) != 2 || manifests[0].Version != "1" || manifests[1].Version != "2" {
		t.Fatalf("manifest ordering: %+v", manifests)
	}
	statuses := r.InstallationStatuses("alpha")
	if len(statuses) != 2 || statuses[0].Version != "1" || statuses[1].Version != "2" {
		t.Fatalf("installation ordering: %+v", statuses)
	}
	if _, err := r.Status("bad/id", first.ID, first.Version); !errors.Is(err, plugin.ErrProjectInvalid) {
		t.Fatalf("invalid status: %v", err)
	}
	if err := r.Promote("alpha", first.ID, first.Version); err != nil {
		t.Fatal(err)
	}
	if err := r.ApproveCapability("alpha", second.ID, second.Version, "write"); err != nil {
		t.Fatal(err)
	}
	if err := r.ApproveArtifact(second.ID, second.Version, "https://example.test/fixture"); err != nil {
		t.Fatal(err)
	}
	if err := r.Remove("alpha", second.ID, second.Version); err != nil {
		t.Fatal(err)
	}
	if err := r.Unregister(second.ID, second.Version); err != nil {
		t.Fatal(err)
	}
	if _, ok := r.ApprovedArtifact(second.ID, second.Version); ok {
		t.Fatal("unregistered artifact retained")
	}
	if err := r.Register(second); err != nil {
		t.Fatal(err)
	}
	installAccessTestPlugin(t, r, "alpha", second, &contractWorker{ready: true})
	if err := r.Promote("alpha", second.ID, second.Version); !errors.Is(err, plugin.ErrCapabilityApproval) {
		t.Fatalf("unregistered approval retained: %v", err)
	}
	if err := r.Revoke("alpha", second.ID, second.Version, "write"); err != nil {
		t.Fatal(err)
	}
	if err := r.Promote("alpha", second.ID, second.Version); !errors.Is(err, plugin.ErrGrantDenied) {
		t.Fatalf("missing grant: %v", err)
	}
	if err := r.Revoke("alpha", second.ID, second.Version, "write"); !errors.Is(err, plugin.ErrGrantDenied) {
		t.Fatalf("duplicate revoke: %v", err)
	}
	if err := r.GrantResource("alpha", second.ID, second.Version, "missing", "key"); !errors.Is(err, plugin.ErrGrantDenied) {
		t.Fatalf("undeclared resource grant: %v", err)
	}
	if err := r.AuthorizeResource("alpha", second.ID, second.Version, "write", "key"); !errors.Is(err, plugin.ErrGrantDenied) {
		t.Fatalf("revoked resource: %v", err)
	}
	if err := r.PinWork("work", "beta", second.ID, second.Version); !errors.Is(err, plugin.ErrNotInstalled) {
		t.Fatalf("missing pin target: %v", err)
	}
}
