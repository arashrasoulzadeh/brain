package cortex_test

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"testing"

	"brain/internal/cortex"
	"brain/internal/plugin"
)

func TestSQLiteExecutionHistorySurvivesRestartAndIsBounded(t *testing.T) {
	path := filepath.Join(t.TempDir(), "history.sqlite")
	storage, err := plugin.OpenSQLiteStorage(path)
	if err != nil {
		t.Fatal(err)
	}
	f := setup(t, 0, 0, 2, 1)
	if err := storage.BindExecutionHistory(context.Background(), f.scheduler); err != nil {
		t.Fatal(err)
	}
	register(t, f, "worker", 1, fast("worker"), runtime)
	for i := 0; i < 132; i++ {
		if _, err := f.scheduler.Execute(context.Background(), f.request(fmt.Sprintf("execution-%03d", i), "alpha")); err != nil {
			t.Fatal(err)
		}
	}
	if err := storage.Close(); err != nil {
		t.Fatal(err)
	}
	storage, err = plugin.OpenSQLiteStorage(path)
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()
	next := setup(t, 0, 0, 2, 1)
	if err := storage.BindExecutionHistory(context.Background(), next.scheduler); err != nil {
		t.Fatal(err)
	}
	view := next.scheduler.Snapshot()
	if !view.HistoryPersistent || view.HistoryDegraded || view.Executions.Assigned != 0 || len(view.Recent) != 128 || view.Recent[0].ExecutionID != "execution-131" || view.Recent[127].ExecutionID != "execution-004" {
		t.Fatalf("unexpected restored history: %+v", view)
	}
	if err := next.scheduler.BindExecutionHistory(context.Background(), storage); !errors.Is(err, cortex.ErrHistoryAlreadyBound) {
		t.Fatal(err)
	}
	register(t, next, "worker", 1, fast("worker"), runtime)
	_ = storage.Close()
	if _, err := next.scheduler.Execute(context.Background(), next.request("after-failure", "alpha")); err != nil {
		t.Fatal("telemetry storage must not reinterpret a successful execution", err)
	}
	view = next.scheduler.Snapshot()
	if !view.HistoryDegraded || view.Executions.Succeeded != 1 || view.Recent[0].ExecutionID != "after-failure" {
		t.Fatal("history failure not surfaced", view)
	}
}
