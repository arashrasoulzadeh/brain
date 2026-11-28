package plugin_test

import (
	"errors"
	"testing"

	"brain/internal/plugin"
)

type unreadyDependencyWorker struct{ dependencyEnableWorker }

func (w *unreadyDependencyWorker) Start() error { return nil }

func dependencyFixture(t *testing.T) (*plugin.Registry, *failingLifecycleStore) {
	t.Helper()
	r := plugin.NewRegistry(1)
	store := &failingLifecycleStore{}
	if err := r.BindRegistryStorage(store); err != nil {
		t.Fatal(err)
	}
	for _, m := range []plugin.Manifest{
		{ID: "leaf", Version: "1", HostAPI: 1, Capabilities: []string{"read"}},
		{ID: "leaf", Version: "2", HostAPI: 1, Capabilities: []string{"read"}},
		{ID: "child", Version: "1", HostAPI: 1, Capabilities: []string{"read"}, Dependencies: []string{"leaf@1"}},
		{ID: "parent", Version: "1", HostAPI: 1, Capabilities: []string{"read"}, Dependencies: []string{"child@1", "leaf@1"}},
		{ID: "conflict", Version: "1", HostAPI: 1, Capabilities: []string{"read"}, Dependencies: []string{"leaf@1", "leaf@2"}},
	} {
		for _, err := range []error{r.Register(m), r.Install("alpha", m.ID, m.Version), r.SetReadiness(m.ID, m.Version, true), r.Grant("alpha", m.ID, m.Version, "read")} {
			if err != nil {
				t.Fatal(err)
			}
		}
	}
	return r, store
}

func TestDependencyActivationAdmissionAndVersionConflicts(t *testing.T) {
	r, _ := dependencyFixture(t)
	plan := []string{"leaf@1", "child@1"}
	if err := r.EnableWithDependencies("bad/project", "parent", "1", plan); !errors.Is(err, plugin.ErrProjectInvalid) {
		t.Fatal(err)
	}
	r.BindProjectResolver(func(id string) (plugin.ProjectSnapshot, error) {
		return plugin.ProjectSnapshot{ID: id, Active: false}, nil
	})
	if err := r.EnableWithDependencies("alpha", "parent", "1", plan); !errors.Is(err, plugin.ErrProjectInvalid) {
		t.Fatal(err)
	}
	r.BindProjectResolver(func(id string) (plugin.ProjectSnapshot, error) {
		return plugin.ProjectSnapshot{ID: id, Active: true}, nil
	})
	if err := r.EnableWithDependencies("alpha", "conflict", "1", []string{"leaf@1", "leaf@2"}); !errors.Is(err, plugin.ErrVersionReferenced) {
		t.Fatal(err)
	}
	if err := r.Revoke("alpha", "child", "1", "read"); err != nil {
		t.Fatal(err)
	}
	if err := r.EnableWithDependencies("alpha", "parent", "1", plan); !errors.Is(err, plugin.ErrGrantDenied) {
		t.Fatal(err)
	}
	if err := r.Grant("alpha", "child", "1", "read"); err != nil {
		t.Fatal(err)
	}
	if err := r.Enable("alpha", "leaf", "2"); err != nil {
		t.Fatal(err)
	}
	if err := r.EnableWithDependencies("alpha", "parent", "1", plan); !errors.Is(err, plugin.ErrVersionReferenced) {
		t.Fatal(err)
	}
	if err := r.Promote("alpha", "leaf", "2"); err != nil {
		t.Fatal(err)
	}
	if err := r.EnableWithDependencies("alpha", "parent", "1", plan); !errors.Is(err, plugin.ErrVersionReferenced) {
		t.Fatal(err)
	}
}

func TestDependencyActivationRollsBackOnFactoryAndPersistenceFailures(t *testing.T) {
	r, store := dependencyFixture(t)
	plan := []string{"leaf@1", "child@1"}
	enable := func() error { return r.EnableWithDependencies("alpha", "parent", "1", plan) }
	if err := enable(); !errors.Is(err, plugin.ErrWorkerUnavailable) {
		t.Fatal(err)
	}
	failure := errors.New("factory failed")
	r.BindWorkerFactory(func(_, _, _ string) (plugin.ContractWorker, error) { return nil, failure })
	if err := enable(); !errors.Is(err, failure) {
		t.Fatal(err)
	}
	r.BindWorkerFactory(func(_, _, _ string) (plugin.ContractWorker, error) { return nil, nil })
	if err := enable(); !errors.Is(err, plugin.ErrWorkerUnavailable) {
		t.Fatal(err)
	}
	r.BindWorkerFactory(func(_, _, _ string) (plugin.ContractWorker, error) { return &unreadyDependencyWorker{}, nil })
	if err := enable(); !errors.Is(err, plugin.ErrWorkerNotReady) {
		t.Fatal(err)
	}
	var workers []*dependencyEnableWorker
	r.BindWorkerFactory(func(_, _, _ string) (plugin.ContractWorker, error) {
		w := &dependencyEnableWorker{}
		workers = append(workers, w)
		return w, nil
	})
	store.fail = true
	if err := enable(); !errors.Is(err, plugin.ErrLifecycleStorage) {
		t.Fatal(err)
	}
	if len(workers) != 3 {
		t.Fatal("expected to start the exact dependency plan", len(workers))
	}
	for _, w := range workers {
		if w.Ready() {
			t.Fatal("failed persistence left a worker running")
		}
	}
	for _, id := range []string{"leaf", "child", "parent"} {
		s, err := r.Status("alpha", id, "1")
		if err != nil || s.Enabled {
			t.Fatal(s, err)
		}
	}
	if err := enable(); !errors.Is(err, plugin.ErrLifecycleStorage) {
		t.Fatal("unhealthy lifecycle storage admitted activation", err)
	}
	store.fail = false
	if err := r.FlushLifecycleState(); err != nil {
		t.Fatal(err)
	}
	prebound := &dependencyEnableWorker{ready: true}
	if err := r.BindWorker("alpha", "leaf", "1", prebound); err != nil {
		t.Fatal(err)
	}
	workers = nil
	if err := enable(); err != nil {
		t.Fatal(err)
	}
	if len(workers) != 2 || !prebound.Ready() {
		t.Fatal("healthy prebound dependency was unnecessarily replaced")
	}
	if err := r.EnableWithDependencies("alpha", "parent", "1", nil); err != nil {
		t.Fatal("already-enabled dependency plan must be empty", err)
	}
}
