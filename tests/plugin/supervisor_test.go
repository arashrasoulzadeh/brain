package plugin_test

import (
	"errors"
	"sync"
	"testing"

	"brain/internal/plugin"
)

type fakeWorker struct{ started, stopped, ready bool }

func (w *fakeWorker) Start() error { w.started = true; w.ready = true; return nil }
func (w *fakeWorker) Stop() error  { w.stopped = true; w.ready = false; return nil }
func (w *fakeWorker) Ready() bool  { return w.ready }

func TestSupervisorRecoversCrashedWorkerWithinBound(t *testing.T) {
	s := plugin.NewSupervisor()
	created := 0
	var current *fakeWorker
	factory := func() plugin.Worker {
		created++
		current = &fakeWorker{}
		return current
	}
	if err := s.StartRecovering("rooms@1", factory, plugin.RecoveryPolicy{MaxRestarts: 2}); err != nil {
		t.Fatal(err)
	}
	if created != 1 {
		t.Fatalf("expected one initial worker, got %d", created)
	}
	current.ready = false
	if err := s.Recover("rooms@1"); err != nil {
		t.Fatal(err)
	}
	if created != 2 || !s.Ready("rooms@1") {
		t.Fatalf("worker was not recovered: created=%d ready=%v", created, s.Ready("rooms@1"))
	}
}

func TestSupervisorSupportsConcurrentHealthChecks(t *testing.T) {
	s := plugin.NewSupervisor()
	if err := s.Start("rooms@1", &fakeWorker{}); err != nil {
		t.Fatal(err)
	}
	var group sync.WaitGroup
	for range 32 {
		group.Add(1)
		go func() {
			defer group.Done()
			_ = s.Ready("rooms@1")
			_ = s.CheckEnabled([]string{"rooms@1"})
		}()
	}
	group.Wait()
}

func TestSupervisorRecoversOnlyAfterCrashAndHonorsRestartLimit(t *testing.T) {
	s := plugin.NewSupervisor()
	created := 0
	workers := make([]*fakeWorker, 0, 3)
	factory := func() plugin.Worker {
		created++
		worker := &fakeWorker{}
		workers = append(workers, worker)
		return worker
	}
	if err := s.StartRecovering("rooms@1", factory, plugin.RecoveryPolicy{MaxRestarts: 1}); err != nil {
		t.Fatal(err)
	}
	if err := s.Recover("rooms@1"); err != nil {
		t.Fatal(err)
	}
	if created != 1 {
		t.Fatalf("healthy worker was restarted: %d", created)
	}
	workers[0].ready = false
	if err := s.Recover("rooms@1"); err != nil {
		t.Fatal(err)
	}
	if created != 2 || !s.Ready("rooms@1") {
		t.Fatalf("worker was not recovered: created=%d ready=%v", created, s.Ready("rooms@1"))
	}
	workers[1].ready = false
	if err := s.Recover("rooms@1"); !errors.Is(err, plugin.ErrWorkerRestartLimit) {
		t.Fatalf("expected bounded recovery, got %v", err)
	}
	if created != 2 {
		t.Fatalf("restart limit was bypassed: %d", created)
	}
}

func TestSupervisorStartsTracksAndStopsWorker(t *testing.T) {
	s := plugin.NewSupervisor()
	w := &fakeWorker{}
	if err := s.Start("guest@1", w); err != nil {
		t.Fatal(err)
	}
	if !w.started || !s.Ready("guest@1") {
		t.Fatal("worker was not ready")
	}
	if !errors.Is(s.Start("guest@1", &fakeWorker{}), plugin.ErrWorkerAlreadyRunning) {
		t.Fatal("duplicate worker accepted")
	}
	if err := s.Stop("guest@1"); err != nil || !w.stopped || s.Ready("guest@1") {
		t.Fatal("worker was not stopped")
	}
	if !errors.Is(s.Stop("guest@1"), plugin.ErrWorkerNotRunning) {
		t.Fatal("missing worker stop was accepted")
	}
}

func TestSupervisorChecksEnabledWorkerHealth(t *testing.T) {
	s := plugin.NewSupervisor()
	if err := s.Start("auth@1", &fakeWorker{}); err != nil {
		t.Fatal(err)
	}
	if err := s.CheckEnabled([]string{"auth@1"}); err != nil {
		t.Fatal(err)
	}
	if !errors.Is(s.CheckEnabled([]string{"missing@1"}), plugin.ErrEnabledWorkerUnavailable) {
		t.Fatal("missing enabled worker was accepted")
	}
	if err := s.Stop("auth@1"); err != nil {
		t.Fatal(err)
	}
	if !errors.Is(s.CheckEnabled([]string{"auth@1"}), plugin.ErrEnabledWorkerUnavailable) {
		t.Fatal("stopped enabled worker was accepted")
	}
}

func TestSupervisorStopsAllInstancesOfPluginVersion(t *testing.T) {
	s := plugin.NewSupervisor()
	first, second, unrelated := &fakeWorker{}, &fakeWorker{}, &fakeWorker{}
	if err := s.Start("dev:rooms@1.0.0", first); err != nil {
		t.Fatal(err)
	}
	if err := s.Start("prod:rooms@1.0.0", second); err != nil {
		t.Fatal(err)
	}
	if err := s.Start("dev:rooms@2.0.0", unrelated); err != nil {
		t.Fatal(err)
	}
	if err := s.StopPluginVersion("rooms", "1.0.0"); err != nil {
		t.Fatal(err)
	}
	if !first.stopped || !second.stopped || unrelated.stopped {
		t.Fatalf("unexpected worker state: first=%v second=%v unrelated=%v", first.stopped, second.stopped, unrelated.stopped)
	}
}

func TestSupervisorRejectsProjectlessStop(t *testing.T) {
	s := plugin.NewSupervisor()
	if !errors.Is(s.StopProjectPlugin("", "rooms", "1.0.0"), plugin.ErrWorkerNotRunning) {
		t.Fatal("projectless stop was accepted")
	}
}

func TestSupervisorRejectsNilWorker(t *testing.T) {
	s := plugin.NewSupervisor()
	if !errors.Is(s.Start("dev:rooms@1", nil), plugin.ErrWorkerRecoveryUnavailable) {
		t.Fatal("nil worker was accepted")
	}
}
