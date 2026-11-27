package plugin_test

import (
	"errors"
	"testing"
	"time"

	"brain/internal/plugin"
)

type lifecycleCoverageWorker struct{ coverageWorker }

func (w *lifecycleCoverageWorker) Handle(req plugin.WorkerRequest) plugin.WorkerResponse {
	return plugin.WorkerResponse{Protocol: req.Protocol, Success: true, Payload: req.Payload}
}

func TestRegistryRestartsWorkersOnlyForAuthorizedInstallations(t *testing.T) {
	r := plugin.NewRegistry(1)
	m := plugin.Manifest{ID: "runtime", Version: "1", HostAPI: 1, Capabilities: []string{"read"}, Routes: []plugin.Route{{Operation: "read", Function: "read", Capability: "read"}}}
	if err := r.Register(m); err != nil {
		t.Fatal(err)
	}
	start := func() error { return r.StartWorker("alpha", m.ID, m.Version) }
	if !errors.Is(start(), plugin.ErrNotInstalled) {
		t.Fatal("missing installation accepted")
	}
	if err := r.Install("alpha", m.ID, m.Version); err != nil {
		t.Fatal(err)
	}
	if !errors.Is(start(), plugin.ErrDisabled) {
		t.Fatal("disabled installation accepted")
	}
	for _, err := range []error{r.SetReadiness(m.ID, m.Version, true), r.Grant("alpha", m.ID, m.Version, "read"), r.Enable("alpha", m.ID, m.Version)} {
		if err != nil {
			t.Fatal(err)
		}
	}
	if !errors.Is(start(), plugin.ErrWorkerUnavailable) {
		t.Fatal("missing worker factory accepted")
	}
	failure := errors.New("worker unavailable")
	r.BindWorkerFactory(func(_, _, _ string) (plugin.ContractWorker, error) { return nil, failure })
	if !errors.Is(start(), failure) {
		t.Fatal("worker factory failure ignored")
	}
	r.BindWorkerFactory(func(_, _, _ string) (plugin.ContractWorker, error) {
		return &lifecycleCoverageWorker{coverageWorker{start: failure}}, nil
	})
	if !errors.Is(start(), failure) {
		t.Fatal("worker start failure ignored")
	}
	calls := 0
	worker := &lifecycleCoverageWorker{}
	r.BindWorkerFactory(func(project, id, version string) (plugin.ContractWorker, error) {
		if project != "alpha" || id != m.ID || version != m.Version {
			t.Fatal("worker factory lost installation identity")
		}
		calls++
		return worker, nil
	})
	if err := start(); err != nil {
		t.Fatal(err)
	}
	if err := start(); err != nil || calls != 1 {
		t.Fatalf("healthy worker restarted: %d %v", calls, err)
	}
	if err := r.SetReadiness(m.ID, m.Version, false); err != nil {
		t.Fatal(err)
	}
	if !errors.Is(start(), plugin.ErrNotReady) {
		t.Fatal("unready plugin started")
	}
	if err := r.SetReadiness(m.ID, m.Version, true); err != nil {
		t.Fatal(err)
	}
	if err := r.Revoke("alpha", m.ID, m.Version, "read"); err != nil {
		t.Fatal(err)
	}
	if !errors.Is(start(), plugin.ErrGrantDenied) {
		t.Fatal("revoked installation started")
	}
	if err := r.Grant("alpha", m.ID, m.Version, "read"); err != nil {
		t.Fatal(err)
	}
	if err := start(); err != nil || calls != 2 {
		t.Fatalf("worker did not restart: %d %v", calls, err)
	}
	if err := r.Drain("alpha", m.ID, m.Version); err != nil {
		t.Fatal(err)
	}
	if !errors.Is(start(), plugin.ErrDisabled) {
		t.Fatal("draining worker started")
	}
	if err := r.Recover("alpha", m.ID, m.Version); err != nil {
		t.Fatal(err)
	}
	worker.stop = failure
	if err := r.StopImmediately("alpha", m.ID, m.Version); !errors.Is(err, failure) {
		t.Fatalf("shutdown failure hidden: %v", err)
	}
	if _, err := r.Invoke(plugin.WorkerRequest{Protocol: 1, ProjectID: "alpha", PluginID: m.ID, Version: m.Version, Operation: "read"}); !errors.Is(err, plugin.ErrDisabled) {
		t.Fatalf("stopped plugin admitted invocation: %v", err)
	}
}

func TestRegistryLifecycleRejectsInvalidProjectsAndMissingInstalls(t *testing.T) {
	r := plugin.NewRegistry(1)
	for _, id := range []string{"bad/id", "missing"} {
		want := plugin.ErrNotInstalled
		if id == "bad/id" {
			want = plugin.ErrProjectInvalid
		}
		for _, err := range []error{r.Drain(id, "rooms", "1"), r.StopImmediately(id, "rooms", "1"), r.Recover(id, "rooms", "1"), r.CanStart(id, "rooms", "1", time.Now()), r.CanStartNew(id, "rooms", "1"), r.CanReconnect(id, "rooms", "1", time.Now()), r.Remove(id, "rooms", "1"), r.Configure(id, "rooms", "1", nil), r.GrantResource(id, "rooms", "1", "read", "room"), r.ApproveCapability(id, "rooms", "1", "read")} {
			if !errors.Is(err, want) {
				t.Fatalf("invalid lifecycle operation for %q: %v, want %v", id, err, want)
			}
		}
	}
	if err := r.GrantResource("alpha", "rooms", "1", "read", ""); !errors.Is(err, plugin.ErrGrantDenied) {
		t.Fatalf("empty resource: %v", err)
	}
	if err := r.AuthorizeResource("alpha", "rooms", "1", "read", ""); !errors.Is(err, plugin.ErrGrantDenied) {
		t.Fatalf("empty resource authorization: %v", err)
	}
	if err := r.Revoke("alpha", "rooms", "1", ""); !errors.Is(err, plugin.ErrGrantDenied) {
		t.Fatalf("empty revocation: %v", err)
	}
	if err := r.Unregister("missing", "1"); !errors.Is(err, plugin.ErrNotRegistered) {
		t.Fatalf("missing unregister: %v", err)
	}
	if err := r.BindWorker("bad/id", "rooms", "1", nil); !errors.Is(err, plugin.ErrInvalidWorkerRequest) {
		t.Fatalf("invalid worker binding: %v", err)
	}
	r.UnbindWorker("alpha", "rooms", "1")
}
