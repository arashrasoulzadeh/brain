package plugin_test

import (
	"errors"
	"testing"

	"brain/internal/plugin"
)

func TestPromotedVersionAndApprovalSurviveRegistryRestart(t *testing.T) {
	storage, err := plugin.OpenSQLiteStorage(t.TempDir() + "/versions.sqlite")
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()
	one := plugin.Manifest{ID: "versions", Version: "1", HostAPI: 1, Capabilities: []string{"read"}, Mount: "/versions", Routes: []plugin.Route{{Operation: "versions.read", Function: "read", Capability: "read", Method: "GET", Path: "/", Auth: "operator"}}}
	two := one
	two.Version = "2"
	two.Capabilities = []string{"read", "write"}
	restore := func() *plugin.Registry {
		t.Helper()
		r := plugin.NewRegistry(1)
		for _, m := range []plugin.Manifest{one, two} {
			if err := r.Register(m); err != nil {
				t.Fatal(err)
			}
		}
		if err := r.BindSQLiteStorage(storage); err != nil {
			t.Fatal(err)
		}
		return r
	}
	r := restore()
	for _, m := range []plugin.Manifest{one, two} {
		if err := r.Install("alpha", m.ID, m.Version); err != nil {
			t.Fatal(err)
		}
		if err := r.SetReadiness(m.ID, m.Version, true); err != nil {
			t.Fatal(err)
		}
		for _, capability := range m.Capabilities {
			if err := r.Grant("alpha", m.ID, m.Version, capability); err != nil {
				t.Fatal(err)
			}
		}
		if err := r.Enable("alpha", m.ID, m.Version); err != nil {
			t.Fatal(err)
		}
	}
	if err := r.Promote("alpha", one.ID, one.Version); err != nil {
		t.Fatal(err)
	}
	if err := r.Promote("alpha", two.ID, two.Version); !errors.Is(err, plugin.ErrCapabilityApproval) {
		t.Fatalf("unapproved upgrade: %v", err)
	}
	if err := r.ApproveCapability("alpha", two.ID, two.Version, "write"); err != nil {
		t.Fatal(err)
	}
	r = restore()
	if match, ok := r.MatchHTTPRoute("alpha", "GET", "/versions"); !ok || match.Version != "1" {
		t.Fatalf("active route after restart: %+v %v", match, ok)
	}
	status, err := r.Status("alpha", two.ID, two.Version)
	if err != nil || status.Promoted || len(status.ApprovedCapabilities) != 1 || status.ApprovedCapabilities[0] != "write" {
		t.Fatalf("approval lost: %+v %v", status, err)
	}
	if err := r.Promote("alpha", two.ID, two.Version); err != nil {
		t.Fatal(err)
	}
	r = restore()
	if match, ok := r.MatchHTTPRoute("alpha", "GET", "/versions"); !ok || match.Version != "2" {
		t.Fatalf("new active route after restart: %+v %v", match, ok)
	}
	status, err = r.Status("alpha", one.ID, one.Version)
	if err != nil || status.Promoted {
		t.Fatalf("old version still promoted: %+v %v", status, err)
	}
	if _, ok := r.ActiveVersion("beta", two.ID); ok {
		t.Fatal("promotion escaped project")
	}
}
