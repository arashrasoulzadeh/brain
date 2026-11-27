package plugin

import (
	"errors"
	"testing"
)

type installBoundaryWorker struct {
	ready, started, stopped bool
}

func (w *installBoundaryWorker) Start() error {
	w.started = true
	return nil
}
func (w *installBoundaryWorker) Stop() error {
	w.stopped = true
	return nil
}
func (w *installBoundaryWorker) Ready() bool { return w.ready }
func (w *installBoundaryWorker) Handle(req WorkerRequest) WorkerResponse {
	return WorkerResponse{Protocol: req.Protocol, Success: true}
}

type lifecycleBoundaryStore struct {
	installations map[string]Installation
	readiness     map[string]bool
	loadErr       error
	saveErr       error
}

func (s *lifecycleBoundaryStore) LoadRegistryState() (map[string]Installation, map[string]bool, error) {
	return s.installations, s.readiness, s.loadErr
}

func (s *lifecycleBoundaryStore) SaveRegistryState(installations map[string]Installation, readiness map[string]bool) error {
	if s.saveErr != nil {
		return s.saveErr
	}
	s.installations = make(map[string]Installation, len(installations))
	s.readiness = make(map[string]bool, len(readiness))
	for k, v := range installations {
		s.installations[k] = cloneInstallation(v)
	}
	for k, v := range readiness {
		s.readiness[k] = v
	}
	return nil
}

func installBoundaryRegistry(t *testing.T) *Registry {
	t.Helper()
	r := NewRegistry(1)
	for _, m := range []Manifest{
		{ID: "dependency", Version: "1", HostAPI: 1},
		{ID: "target", Version: "1", HostAPI: 1, Capabilities: []string{"read", "write"}, Dependencies: []string{"dependency@1"}},
		{ID: "target", Version: "2", HostAPI: 1, Capabilities: []string{"read", "write"}},
	} {
		if err := r.Register(m); err != nil {
			t.Fatal(err)
		}
	}
	return r
}

func TestConfirmedInstallAdmissionBoundaries(t *testing.T) {
	failure := errors.New("project lookup failed")
	for _, tc := range []struct {
		name, project, id, version string
		grants                     []string
		resolve                    func(string) (ProjectSnapshot, error)
		want                       error
	}{
		{name: "invalid project", project: "bad!", id: "target", version: "2", want: ErrProjectInvalid},
		{name: "inactive project", project: "alpha", id: "target", version: "2", resolve: func(string) (ProjectSnapshot, error) { return ProjectSnapshot{ID: "alpha"}, nil }, want: ErrProjectInvalid},
		{name: "foreign project", project: "alpha", id: "target", version: "2", resolve: func(string) (ProjectSnapshot, error) { return ProjectSnapshot{ID: "beta", Active: true}, nil }, want: ErrProjectInvalid},
		{name: "failed project lookup", project: "alpha", id: "target", version: "2", resolve: func(string) (ProjectSnapshot, error) { return ProjectSnapshot{}, failure }, want: ErrProjectInvalid},
		{name: "unknown version", project: "alpha", id: "target", version: "missing", want: ErrNotRegistered},
		{name: "unapproved grant", project: "alpha", id: "target", version: "2", grants: []string{"read", "wrong"}, want: ErrGrantDenied},
		{name: "empty grant", project: "alpha", id: "target", version: "2", grants: []string{"read", ""}, want: ErrGrantDenied},
		{name: "missing dependency", project: "alpha", id: "target", version: "1", grants: []string{"read", "write"}, want: ErrDependencyMissing},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := installBoundaryRegistry(t)
			r.BindProjectResolver(tc.resolve)
			r.BindWorkerFactory(func(_, _, _ string) (ContractWorker, error) {
				t.Fatal("worker started before successful admission")
				return nil, nil
			})
			if err := r.InstallAndActivate(tc.project, tc.id, tc.version, tc.grants); !errors.Is(err, tc.want) {
				t.Fatalf("admission: got %v, want %v", err, tc.want)
			}
			if len(r.installations) != 0 {
				t.Fatal("rejected admission changed installation state")
			}
		})
	}
}

func TestConfirmedInstallRejectsUnreadyDependenciesAndImplicitVersionSwitch(t *testing.T) {
	r := installBoundaryRegistry(t)
	const project = "alpha"
	grants := []string{"read", "write"}
	if err := r.Install(project, "dependency", "1"); err != nil {
		t.Fatal(err)
	}
	if err := r.SetReadiness("dependency", "1", true); err != nil {
		t.Fatal(err)
	}
	if err := r.Enable(project, "dependency", "1"); err != nil {
		t.Fatal(err)
	}
	if err := r.SetReadiness("dependency", "1", false); err != nil {
		t.Fatal(err)
	}
	if err := r.InstallAndActivate(project, "target", "1", grants); !errors.Is(err, ErrDependencyMissing) {
		t.Fatalf("unready dependency: %v", err)
	}
	if err := r.SetReadiness("dependency", "1", true); err != nil {
		t.Fatal(err)
	}
	if err := r.Drain(project, "dependency", "1"); err != nil {
		t.Fatal(err)
	}
	if err := r.InstallAndActivate(project, "target", "1", grants); !errors.Is(err, ErrDependencyMissing) {
		t.Fatalf("draining dependency: %v", err)
	}
	if err := r.Recover(project, "dependency", "1"); err != nil {
		t.Fatal(err)
	}
	r.BindWorkerFactory(func(_, _, _ string) (ContractWorker, error) {
		return &confirmedInstallWorker{}, nil
	})
	if err := r.InstallAndActivate(project, "target", "1", grants); err != nil {
		t.Fatal(err)
	}
	if err := r.InstallAndActivate(project, "target", "2", grants); !errors.Is(err, ErrVersionReferenced) {
		t.Fatalf("implicit upgrade of enabled version: %v", err)
	}
	if err := r.Promote(project, "target", "1"); err != nil {
		t.Fatal(err)
	}
	if err := r.InstallAndActivate(project, "target", "2", grants); !errors.Is(err, ErrVersionReferenced) {
		t.Fatalf("implicit switch of promoted version: %v", err)
	}
	if got, ok := r.ActiveVersion(project, "target"); !ok || got != "1" {
		t.Fatalf("promotion changed: %q %v", got, ok)
	}
}

func TestConfirmedInstallWorkerAndCommitBoundaries(t *testing.T) {
	failure := errors.New("injected failure")
	grants := []string{"read", "write"}
	notReady := &installBoundaryWorker{}
	for _, tc := range []struct {
		name    string
		factory func(string, string, string) (ContractWorker, error)
		want    error
	}{
		{name: "missing factory", want: ErrWorkerUnavailable},
		{name: "factory error", factory: func(_, _, _ string) (ContractWorker, error) { return nil, failure }, want: failure},
		{name: "nil worker", factory: func(_, _, _ string) (ContractWorker, error) { return nil, nil }, want: ErrWorkerUnavailable},
		{name: "not ready", factory: func(_, _, _ string) (ContractWorker, error) { return notReady, nil }, want: ErrWorkerNotReady},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := installBoundaryRegistry(t)
			r.BindWorkerFactory(tc.factory)
			if err := r.InstallAndActivate("alpha", "target", "2", grants); !errors.Is(err, tc.want) {
				t.Fatalf("activation: got %v, want %v", err, tc.want)
			}
			if len(r.installations) != 0 || len(r.workers) != 0 || r.readiness["target@2"] {
				t.Fatal("failed activation left partial state")
			}
		})
	}
	if !notReady.started || !notReady.stopped {
		t.Fatal("unready worker was not stopped")
	}
	r := installBoundaryRegistry(t)
	store := &lifecycleBoundaryStore{}
	if err := r.BindRegistryStorage(store); err != nil {
		t.Fatal(err)
	}
	worker := &installBoundaryWorker{ready: true}
	r.BindWorkerFactory(func(_, _, _ string) (ContractWorker, error) { return worker, nil })
	store.saveErr = failure
	if err := r.InstallAndActivate("alpha", "target", "2", grants); !errors.Is(err, ErrLifecycleStorage) {
		t.Fatalf("persistence failure: %v", err)
	}
	if !worker.started || !worker.stopped || len(r.installations) != 0 || len(r.workers) != 0 || r.readiness["target@2"] {
		t.Fatal("failed persistence did not roll back activation")
	}
	worker.started = false
	if err := r.InstallAndActivate("alpha", "target", "2", grants); !errors.Is(err, ErrLifecycleStorage) || worker.started {
		t.Fatal("unhealthy storage admitted another worker", err)
	}
	store.saveErr = nil
	if err := r.FlushLifecycleState(); err != nil {
		t.Fatal(err)
	}
	if err := r.InstallAndActivate("alpha", "target", "2", grants); err != nil {
		t.Fatal(err)
	}
	if got := store.installations["alpha:target@2"]; !got.Enabled || !got.Grants["write"] {
		t.Fatalf("successful retry not persisted: %+v", got)
	}
}

func TestVersionRestoreRejectsAmbiguousOrMissingPromotedManifest(t *testing.T) {
	failure := errors.New("unavailable lifecycle store")
	for _, tc := range []struct {
		name  string
		store *lifecycleBoundaryStore
		want  error
	}{
		{name: "load error", store: &lifecycleBoundaryStore{loadErr: failure}, want: failure},
		{name: "missing manifest", store: &lifecycleBoundaryStore{installations: map[string]Installation{
			"alpha:target@missing": {ProjectID: "alpha", PluginID: "target", Version: "missing", Promoted: true},
		}}, want: ErrLifecycleStorage},
		{name: "ambiguous promotion", store: &lifecycleBoundaryStore{installations: map[string]Installation{
			"alpha:target@1": {ProjectID: "alpha", PluginID: "target", Version: "1", Promoted: true},
			"alpha:target@2": {ProjectID: "alpha", PluginID: "target", Version: "2", Promoted: true},
		}}, want: ErrLifecycleStorage},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := installBoundaryRegistry(t)
			if err := r.BindRegistryStorage(tc.store); !errors.Is(err, tc.want) {
				t.Fatalf("restore: %v", err)
			}
			if _, ok := r.ActiveVersion("alpha", "target"); ok {
				t.Fatal("invalid persisted promotion was exposed")
			}
		})
	}
}
