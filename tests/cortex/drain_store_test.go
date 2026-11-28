package cortex_test

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"brain/internal/cortex"
	"brain/internal/execution"
	"brain/internal/plugin"
)

func TestOperatorDrainPersistsBeforeChangingPlacement(t *testing.T) {
	path := filepath.Join(t.TempDir(), "drains.sqlite")
	store, err := plugin.OpenSQLiteStorage(path)
	if err != nil {
		t.Fatal(err)
	}
	first := setup(t, 0, 0, 2, 1)
	if err := store.BindOperatorDrains(context.Background(), first.scheduler); err != nil {
		t.Fatal(err)
	}
	register(t, first, "worker", 1, fast("worker"), runtime)
	if err := first.scheduler.SetOperatorDrain("worker", true); err != nil {
		t.Fatal(err)
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
	if err := first.scheduler.SetOperatorDrain("worker", false); err == nil || !first.scheduler.OperatorPaused("worker") {
		t.Fatal("failed persistence cleared placement pause", err)
	}

	store, err = plugin.OpenSQLiteStorage(path)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	second := setup(t, 0, 0, 2, 1)
	register(t, second, "worker", 1, fast("worker"), runtime)
	if err := store.BindOperatorDrains(context.Background(), second.scheduler); err != nil {
		t.Fatal(err)
	}
	view := second.scheduler.Snapshot()
	if len(view.OperatorPaused) != 1 || view.OperatorPaused[0] != "worker" || !view.Neurons[0].OperatorDraining || second.scheduler.HasRuntime(runtime) {
		t.Fatalf("pause not restored: %+v", view)
	}
	if _, err := second.scheduler.Execute(context.Background(), second.request("paused", "alpha")); !errors.Is(err, execution.ErrUnavailable) {
		t.Fatal("restored pause not enforced", err)
	}
	if err := second.scheduler.SetOperatorDrain("worker", false); err != nil {
		t.Fatal(err)
	}
	if _, err := second.scheduler.Execute(context.Background(), second.request("resumed", "alpha")); err != nil {
		t.Fatal(err)
	}
	third := setup(t, 0, 0, 2, 1)
	if err := store.BindOperatorDrains(context.Background(), third.scheduler); err != nil {
		t.Fatal(err)
	}
	if third.scheduler.OperatorPaused("worker") {
		t.Fatal("cleared pause restored")
	}
	if err := store.BindOperatorDrains(context.Background(), second.scheduler); !errors.Is(err, cortex.ErrDrainStoreAlreadyBound) {
		t.Fatal(err)
	}
}
