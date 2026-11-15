package plugin_test

import (
	"testing"

	"brain/internal/plugin"
)

func TestRegistryRestoresInstallationLifecycleFromSQLite(t *testing.T) {
	path := t.TempDir() + "/lifecycle.sqlite"
	storage, err := plugin.OpenSQLiteStorage(path)
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()
	manifest := plugin.Manifest{ID: "sample", Title: "Sample", Version: "1.0.0", HostAPI: 1, Capabilities: []string{"sample.read"}}
	first := plugin.NewRegistry(1)
	if err := first.Register(manifest); err != nil {
		t.Fatal(err)
	}
	first.BindSQLiteStorage(storage)
	if err := first.Install("dev", manifest.ID, manifest.Version); err != nil {
		t.Fatal(err)
	}
	if err := first.SetReadiness(manifest.ID, manifest.Version, true); err != nil {
		t.Fatal(err)
	}
	if err := first.Grant("dev", manifest.ID, manifest.Version, "sample.read"); err != nil {
		t.Fatal(err)
	}
	if err := first.Enable("dev", manifest.ID, manifest.Version); err != nil {
		t.Fatal(err)
	}

	second := plugin.NewRegistry(1)
	if err := second.Register(manifest); err != nil {
		t.Fatal(err)
	}
	second.BindSQLiteStorage(storage)
	status, err := second.Status("dev", manifest.ID, manifest.Version)
	if err != nil {
		t.Fatal(err)
	}
	if !status.Ready || !status.Enabled || len(status.Grants) != 1 || status.Grants[0] != "sample.read" {
		t.Fatalf("restored status = %#v", status)
	}
}
