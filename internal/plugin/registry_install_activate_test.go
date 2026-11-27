package plugin

import (
	"errors"
	"path/filepath"
	"testing"
)

type confirmedInstallWorker struct {
	ready bool
	fail  error
}

func (w *confirmedInstallWorker) Start() error {
	if w.fail != nil {
		return w.fail
	}
	w.ready = true
	return nil
}
func (w *confirmedInstallWorker) Stop() error { w.ready = false; return nil }
func (w *confirmedInstallWorker) Ready() bool { return w.ready }
func (w *confirmedInstallWorker) Handle(req WorkerRequest) WorkerResponse {
	return WorkerResponse{Protocol: req.Protocol, Success: true}
}

func TestConfirmedInstallIsAtomicAndProjectScoped(t *testing.T) {
	r := NewRegistry(1)
	m := Manifest{ID: "sample", Version: "1", HostAPI: 1, Capabilities: []string{"read", "write"}}
	if err := r.Register(m); err != nil {
		t.Fatal(err)
	}
	storage, err := OpenSQLiteStorage(filepath.Join(t.TempDir(), "install.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()
	if err := r.BindSQLiteStorage(storage); err != nil {
		t.Fatal(err)
	}
	calls := 0
	worker := &confirmedInstallWorker{}
	r.BindWorkerFactory(func(project, id, version string) (ContractWorker, error) {
		calls++
		if project != "one" || id != "sample" || version != "1" {
			t.Fatalf("wrong worker scope: %s %s %s", project, id, version)
		}
		return worker, nil
	})
	for _, grants := range [][]string{nil, {"read"}, {"read", "read"}, {"read", "unexpected"}, {"read", "write", "extra"}} {
		if err := r.InstallAndActivate("one", "sample", "1", grants); !errors.Is(err, ErrGrantDenied) {
			t.Fatalf("bad consent accepted: %v %v", grants, err)
		}
	}
	if calls != 0 {
		t.Fatal("worker started before complete consent")
	}
	if err := r.InstallAndActivate("one", "sample", "1", []string{"write", "read"}); err != nil {
		t.Fatal(err)
	}
	status, err := r.Status("one", "sample", "1")
	if err != nil || !status.Enabled || !status.Ready || !status.WorkerReady || len(status.Grants) != 2 {
		t.Fatalf("not active: %#v %v", status, err)
	}
	if _, err := r.Status("two", "sample", "1"); !errors.Is(err, ErrNotInstalled) {
		t.Fatal("cross-project installation")
	}
	if err := r.InstallAndActivate("one", "sample", "1", m.Capabilities); !errors.Is(err, ErrDuplicate) {
		t.Fatalf("reinstall: %v", err)
	}
	persisted, readiness, err := storage.LoadRegistryState()
	if err != nil || !persisted["one:sample@1"].Enabled || !persisted["one:sample@1"].Grants["write"] || !readiness["sample@1"] {
		t.Fatal("activation was not persisted", err)
	}
}

func TestConfirmedInstallDoesNotLeavePartialGrantsOnFailure(t *testing.T) {
	r := NewRegistry(1)
	m := Manifest{ID: "sample", Version: "1", HostAPI: 1, Capabilities: []string{"read"}}
	if err := r.Register(m); err != nil {
		t.Fatal(err)
	}
	failure := errors.New("worker start failed")
	w := &confirmedInstallWorker{fail: failure}
	r.BindWorkerFactory(func(_, _, _ string) (ContractWorker, error) { return w, nil })
	if err := r.InstallAndActivate("one", m.ID, m.Version, m.Capabilities); !errors.Is(err, failure) {
		t.Fatalf("startup failure not reported: %v", err)
	}
	if _, err := r.Status("one", m.ID, m.Version); !errors.Is(err, ErrNotInstalled) {
		t.Fatal("partial installation left behind")
	}
	if r.readiness["sample@1"] || w.Ready() {
		t.Fatal("failed worker marked ready")
	}
}
