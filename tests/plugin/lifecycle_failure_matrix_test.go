package plugin_test

import (
	"errors"
	"maps"
	"testing"

	"brain/internal/plugin"
)

type coverageStateStore struct {
	installations map[string]plugin.Installation
	readiness     map[string]bool
	loadErr       error
	fail          bool
}

func (s *coverageStateStore) LoadRegistryState() (map[string]plugin.Installation, map[string]bool, error) {
	return s.installations, s.readiness, s.loadErr
}
func (s *coverageStateStore) SaveRegistryState(installations map[string]plugin.Installation, readiness map[string]bool) error {
	if s.fail {
		return errors.New("write unavailable")
	}
	s.installations = maps.Clone(installations)
	s.readiness = maps.Clone(readiness)
	return nil
}

func TestRegistryRestoresOnlyRegisteredStateAndRollsBackFailedWrites(t *testing.T) {
	r := plugin.NewRegistry(1)
	m := plugin.Manifest{ID: "sample", Version: "1", HostAPI: 1, Capabilities: []string{"read"}, ConfigSchema: map[string]string{"name": "string", "count": "int", "active": "bool"}}
	if err := r.Register(m); err != nil {
		t.Fatal(err)
	}
	failure := errors.New("read unavailable")
	if err := r.BindRegistryStorage(&coverageStateStore{loadErr: failure}); !errors.Is(err, failure) {
		t.Fatalf("load: %v", err)
	}
	store := &coverageStateStore{installations: map[string]plugin.Installation{"alpha:sample@1": {ProjectID: "alpha", PluginID: "sample", Version: "1"}, "alpha:unknown@1": {ProjectID: "alpha", PluginID: "unknown", Version: "1"}}, readiness: map[string]bool{"sample@1": true, "unknown@1": true}}
	if err := r.BindRegistryStorage(store); err != nil {
		t.Fatal(err)
	}
	if list := r.InstallationStatuses(""); len(list) != 1 || !list[0].Ready {
		t.Fatalf("unexpected restored state: %+v", list)
	}
	if err := r.GrantResource("alpha", "sample", "1", "read", "room"); err != nil {
		t.Fatal(err)
	}
	if err := r.Configure("alpha", "sample", "1", map[string]any{"name": "one", "active": true, "count": 1}); err != nil {
		t.Fatal(err)
	}
	if err := r.Configure("alpha", "sample", "1", map[string]any{"count": []int{1}}); !errors.Is(err, plugin.ErrInvalidConfig) {
		t.Fatalf("invalid config: %v", err)
	}
	store.fail = true
	for _, err := range []error{r.Install("beta", "sample", "1"), r.SetReadiness("sample", "1", true), r.GrantResource("alpha", "sample", "1", "read", "other"), r.Configure("alpha", "sample", "1", map[string]any{"name": "new"}), r.Drain("alpha", "sample", "1"), r.Recover("alpha", "sample", "1")} {
		if !errors.Is(err, plugin.ErrLifecycleStorage) {
			t.Fatalf("failed write: %v", err)
		}
	}
	if _, err := r.Status("beta", "sample", "1"); !errors.Is(err, plugin.ErrNotInstalled) {
		t.Fatalf("failed install visible: %v", err)
	}
	if err := r.BindSQLiteStorage(nil); err != nil {
		t.Fatal(err)
	}
	if err := r.FlushLifecycleState(); !errors.Is(err, plugin.ErrLifecycleStorage) {
		t.Fatalf("failed storage reopened: %v", err)
	}
	store.fail = false
	if err := r.BindRegistryStorage(store); err != nil {
		t.Fatal(err)
	}
	if err := r.FlushLifecycleState(); err != nil {
		t.Fatal(err)
	}
	if err := r.Grant("alpha", "sample", "1", "read"); err != nil {
		t.Fatal(err)
	}
	if err := r.Enable("alpha", "sample", "1"); err != nil {
		t.Fatal(err)
	}
	if err := r.AuthorizeResource("alpha", "sample", "1", "read", "other"); !errors.Is(err, plugin.ErrGrantDenied) {
		t.Fatalf("failed resource grant visible: %v", err)
	}
	if err := r.SaveDurableData("bad/id", "sample", "key", nil); !errors.Is(err, plugin.ErrProjectInvalid) {
		t.Fatalf("invalid save: %v", err)
	}
	if err := r.DeleteDurableData("alpha", "sample", ""); !errors.Is(err, plugin.ErrProjectInvalid) {
		t.Fatalf("invalid delete: %v", err)
	}
	db, err := plugin.OpenSQLiteStorage(t.TempDir() + "/storage.sqlite")
	if err != nil {
		t.Fatal(err)
	}
	if err = db.Close(); err != nil {
		t.Fatal(err)
	}
	if err = r.BindSQLiteStorage(db); err == nil {
		t.Fatal("closed storage bound")
	}
}

func TestLifecycleFailuresCloseAccessEvenIfCleanupFails(t *testing.T) {
	for _, operation := range []string{"disable", "stop", "revoke", "recover", "remove", "uninstall"} {
		for _, cleanup := range []string{"supervisor", "socket"} {
			t.Run(operation+"/"+cleanup, func(t *testing.T) {
				r := plugin.NewRegistry(1)
				m := plugin.Manifest{ID: "sample", Version: "1", HostAPI: 1, Capabilities: []string{"read"}}
				for _, err := range []error{r.Register(m), r.Install("alpha", m.ID, m.Version), r.SetReadiness(m.ID, m.Version, true), r.Grant("alpha", m.ID, m.Version, "read"), r.Enable("alpha", m.ID, m.Version)} {
					if err != nil {
						t.Fatal(err)
					}
				}
				failure := errors.New("cleanup failed")
				s := plugin.NewSupervisor()
				sockets := plugin.NewWebSocketRegistry()
				r.BindSupervisor(s)
				r.BindWebSocketRegistry(sockets)
				if cleanup == "supervisor" {
					if err := s.Start("alpha:sample@1", &coverageWorker{stop: failure}); err != nil {
						t.Fatal(err)
					}
				} else {
					if err := sockets.RegisterProject("one", "alpha", &failingCoverageSocket{}); err != nil {
						t.Fatal(err)
					}
				}
				var err error
				switch operation {
				case "disable":
					var preview plugin.DisablePreview
					preview, err = r.PreviewDisable("alpha", m.ID, m.Version)
					if err == nil {
						err = r.ApplyDisablePreview(preview)
					}
				case "stop":
					err = r.StopImmediately("alpha", m.ID, m.Version)
				case "revoke":
					err = r.Revoke("alpha", m.ID, m.Version, "read")
				case "recover":
					if err = r.Drain("alpha", m.ID, m.Version); err == nil {
						err = r.Recover("alpha", m.ID, m.Version)
					}
				case "remove":
					err = r.Remove("alpha", m.ID, m.Version)
				case "uninstall":
					err = r.UninstallProject("alpha")
				}
				if err == nil {
					t.Fatal("cleanup failure hidden")
				}
				if operation == "recover" {
					status, e := r.Status("alpha", m.ID, m.Version)
					if e != nil || !status.Draining {
						t.Fatalf("failed recovery resumed plugin: %+v %v", status, e)
					}
				} else if r.Authorize("alpha", m.ID, m.Version, "read") == nil {
					t.Fatal("failed cleanup restored access")
				}
			})
		}
	}
}
