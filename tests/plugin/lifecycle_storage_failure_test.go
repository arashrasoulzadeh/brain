package plugin_test

import (
	"errors"
	"testing"

	"brain/internal/plugin"
)

type failingLifecycleStore struct {
	fail bool
}

func (s *failingLifecycleStore) LoadRegistryState() (map[string]plugin.Installation, map[string]bool, error) {
	return nil, nil, nil
}

func (s *failingLifecycleStore) SaveRegistryState(map[string]plugin.Installation, map[string]bool) error {
	if s.fail {
		return errors.New("private database error")
	}
	return nil
}

func TestLifecycleWriteFailureBlocksAuthorizationAndRollsBackGrants(t *testing.T) {
	r := plugin.NewRegistry(1)
	m := plugin.Manifest{ID: "sample", Version: "1.0.0", HostAPI: 1, Capabilities: []string{"read", "write"}, Routes: []plugin.Route{{Operation: "read", Function: "read", Capability: "read"}}}
	if err := r.Register(m); err != nil {
		t.Fatal(err)
	}
	store := &failingLifecycleStore{}
	if err := r.BindRegistryStorage(store); err != nil {
		t.Fatal(err)
	}
	if err := r.Install("project", m.ID, m.Version); err != nil {
		t.Fatal(err)
	}
	if err := r.SetReadiness(m.ID, m.Version, true); err != nil {
		t.Fatal(err)
	}
	if err := r.Grant("project", m.ID, m.Version, "read"); err != nil {
		t.Fatal(err)
	}
	store.fail = true
	if err := r.Grant("project", m.ID, m.Version, "write"); !errors.Is(err, plugin.ErrLifecycleStorage) {
		t.Fatalf("failed grant = %v", err)
	}
	status, err := r.Status("project", m.ID, m.Version)
	if err != nil || len(status.Grants) != 1 || status.Grants[0] != "read" {
		t.Fatalf("failed grant became visible: %#v, %v", status, err)
	}
	if !errors.Is(r.LifecycleStorageHealth(), plugin.ErrLifecycleStorage) {
		t.Fatal("failed storage reported healthy")
	}
	store.fail = false
	if err := r.Grant("project", m.ID, m.Version, "write"); err != nil {
		t.Fatal(err)
	}
	if err := r.Enable("project", m.ID, m.Version); err != nil {
		t.Fatal(err)
	}
	store.fail = true
	if err := r.Revoke("project", m.ID, m.Version, "read"); !errors.Is(err, plugin.ErrLifecycleStorage) {
		t.Fatalf("failed revoke = %v", err)
	}
	if err := r.Authorize("project", m.ID, m.Version, "write"); !errors.Is(err, plugin.ErrLifecycleStorage) {
		t.Fatalf("authorization remained open: %v", err)
	}
	status, err = r.Status("project", m.ID, m.Version)
	if err != nil || len(status.Grants) != 1 || status.Grants[0] != "write" {
		t.Fatalf("failed revoke restored access: %#v, %v", status, err)
	}
	store.fail = false
	if err := r.FlushLifecycleState(); err != nil {
		t.Fatal(err)
	}
	if err := r.Authorize("project", m.ID, m.Version, "read"); !errors.Is(err, plugin.ErrGrantDenied) {
		t.Fatalf("revoked grant restored after flush: %v", err)
	}
}

func TestFailedWorkerStartDoesNotPersistEnabledInstallation(t *testing.T) {
	storage, err := plugin.OpenSQLiteStorage(t.TempDir() + "/lifecycle.sqlite")
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()
	m := plugin.Manifest{ID: "sample", Version: "1.0.0", HostAPI: 1, Capabilities: []string{"read"}}
	r := plugin.NewRegistry(1)
	if err := r.Register(m); err != nil {
		t.Fatal(err)
	}
	if err := r.BindSQLiteStorage(storage); err != nil {
		t.Fatal(err)
	}
	if err := r.Install("project", m.ID, m.Version); err != nil {
		t.Fatal(err)
	}
	if err := r.SetReadiness(m.ID, m.Version, true); err != nil {
		t.Fatal(err)
	}
	if err := r.Grant("project", m.ID, m.Version, "read"); err != nil {
		t.Fatal(err)
	}
	r.BindWorkerFactory(func(_, _, _ string) (plugin.ContractWorker, error) {
		return nil, errors.New("worker unavailable")
	})
	if err := r.Enable("project", m.ID, m.Version); err == nil {
		t.Fatal("enable unexpectedly succeeded")
	}
	reloaded := plugin.NewRegistry(1)
	if err := reloaded.Register(m); err != nil {
		t.Fatal(err)
	}
	if err := reloaded.BindSQLiteStorage(storage); err != nil {
		t.Fatal(err)
	}
	status, err := reloaded.Status("project", m.ID, m.Version)
	if err != nil || status.Enabled {
		t.Fatalf("failed enable was persisted: %#v, %v", status, err)
	}
}
