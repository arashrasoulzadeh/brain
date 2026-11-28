package cortex_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"brain/internal/cortex"
	"brain/internal/execution"
)

func TestOperatorDrainSurvivesReconnectAndWaitsForActiveWork(t *testing.T) {
	f := setup(t, 0, 0, 4, 2)
	started, release := make(chan struct{}), make(chan struct{})
	var once sync.Once
	finish := func() { once.Do(func() { close(release) }) }
	defer finish()
	w := register(t, f, "managed", 1, executorFunc(func(_ context.Context, req execution.Request) (execution.Result, error) {
		close(started)
		<-release
		return result(req, "managed"), nil
	}), runtime)
	done := submit(f.scheduler, context.Background(), f.request("active", "alpha"))
	<-started
	if err := f.scheduler.SetOperatorDrain("managed", true); err != nil {
		t.Fatal(err)
	}
	if f.scheduler.HasRuntime(runtime) {
		t.Fatal("paused worker counted as available")
	}
	if _, err := f.scheduler.Execute(context.Background(), f.request("denied", "alpha")); !errors.Is(err, execution.ErrUnavailable) {
		t.Fatal("paused worker accepted work", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	if err := f.scheduler.WaitNeuronIdle(ctx, "managed"); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal("active assignment ignored", err)
	}
	w.Disconnect()
	replacement := register(t, f, "managed", 1, fast("managed"), runtime)
	if view, ok := f.scheduler.Neuron("managed"); !ok || !view.OperatorDraining || !view.Draining {
		t.Fatal("reconnect cleared pause", view)
	}
	oldCtx, oldCancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer oldCancel()
	if err := f.scheduler.WaitNeuronIdle(oldCtx, "managed"); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal("fenced old work ignored", err)
	}
	finish()
	if got := receive(t, done); !errors.Is(got.err, cortex.ErrSessionLost) {
		t.Fatal(got.err)
	}
	if err := f.scheduler.WaitNeuronIdle(context.Background(), "managed"); err != nil {
		t.Fatal(err)
	}
	if err := f.scheduler.SetOperatorDrain("managed", false); err != nil {
		t.Fatal(err)
	}
	if _, err := f.scheduler.Execute(context.Background(), f.request("resumed", "alpha")); err != nil {
		t.Fatal(err)
	}
	if err := replacement.Drain(); err != nil {
		t.Fatal(err)
	}
	if err := f.scheduler.SetOperatorDrain("managed", false); err != nil {
		t.Fatal(err)
	}
	if f.scheduler.HasRuntime(runtime) {
		t.Fatal("operator resume undid runtime drain")
	}
}
