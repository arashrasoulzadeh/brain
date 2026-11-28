package cortex_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"brain/internal/cortex"
	"brain/internal/execution"
)

func TestExecutionTelemetryIsBoundedRedactedAndCountsFencedResults(t *testing.T) {
	f := setup(t, 0, 0, 2, 1)
	w := register(t, f, "worker", 1, executorFunc(func(_ context.Context, req execution.Request) (execution.Result, error) {
		switch req.ID {
		case "failed":
			return execution.Result{ID: req.ID, WorkerID: "worker", Status: execution.Failed, Error: &execution.Failure{Code: "error", Message: "private-message"}}, nil
		case "broken":
			return execution.Result{}, errors.New("private-message")
		case "cancelled":
			return execution.Result{ID: req.ID, WorkerID: "worker", Status: execution.Cancelled, Error: &execution.Failure{Code: "cancelled"}}, nil
		default:
			out := result(req, "worker")
			out.Usage = execution.Usage{CPUTime: time.Millisecond, PeakMemoryBytes: 1024}
			return out, nil
		}
	}), runtime)
	for _, id := range []string{"failed", "broken", "cancelled"} {
		_, _ = f.scheduler.Execute(context.Background(), f.request(id, "private-project"))
	}
	for i := 0; i < 140; i++ {
		req := f.request(fmt.Sprintf("execution-%d", i), "private-project")
		req.Input = json.RawMessage(`"private-input"`)
		if _, err := f.scheduler.Execute(context.Background(), req); err != nil {
			t.Fatal(err)
		}
	}
	snapshot := f.scheduler.Snapshot()
	if snapshot.Executions != (cortex.ExecutionTotals{Assigned: 143, Succeeded: 140, Failed: 2, Cancelled: 1}) || len(snapshot.Recent) != 128 || snapshot.Recent[0].ExecutionID != "execution-139" || snapshot.Recent[0].CPUTimeNanos != int64(time.Millisecond) || snapshot.Recent[0].PeakMemoryBytes != 1024 {
		t.Fatalf("unexpected telemetry: %+v", snapshot.Executions)
	}
	snapshot.Recent[0].ExecutionID = "mutated"
	if f.scheduler.Snapshot().Recent[0].ExecutionID == "mutated" {
		t.Fatal("snapshot exposed telemetry storage")
	}
	raw, err := json.Marshal(f.scheduler.Snapshot())
	if err != nil || strings.Contains(string(raw), "private-") {
		t.Fatalf("private metadata exposed: %v", err)
	}
	w.Disconnect()
	started := make(chan struct{})
	w = register(t, f, "worker", 1, executorFunc(func(ctx context.Context, req execution.Request) (execution.Result, error) {
		close(started)
		<-ctx.Done()
		return result(req, "worker"), nil
	}), runtime)
	done := submit(f.scheduler, context.Background(), f.request("lost", "private-project"))
	<-started
	w.Disconnect()
	if got := receive(t, done); !errors.Is(got.err, cortex.ErrSessionLost) {
		t.Fatal(got.err)
	}
	snapshot = f.scheduler.Snapshot()
	if snapshot.Executions.Lost != 1 || snapshot.Recent[0].Outcome != "lost" || snapshot.Recent[0].CPUTimeNanos != 0 {
		t.Fatalf("lost assignment recorded incorrectly: %+v", snapshot.Recent[0])
	}
}
