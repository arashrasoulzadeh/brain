package plugin_test

import (
	"errors"
	"testing"

	"brain/internal/plugin"
)

type fakeWorker struct{ started, stopped, ready bool }

func (w *fakeWorker) Start() error { w.started = true; w.ready = true; return nil }
func (w *fakeWorker) Stop() error  { w.stopped = true; w.ready = false; return nil }
func (w *fakeWorker) Ready() bool  { return w.ready }

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
