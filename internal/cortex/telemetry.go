package cortex

import (
	"context"
	"errors"
	"time"

	"brain/internal/execution"
)

const executionHistoryLimit = 128

// ExecutionTotals count assigned work during this Brain process lifetime.
// A succeeded runtime result does not imply that Brain committed its effects.
type ExecutionTotals struct {
	Assigned, Succeeded, Failed, Cancelled, Lost uint64
}

// ExecutionTrace deliberately excludes project, principal, resource, source,
// request/response payloads, effect intents and raw failure messages.
type ExecutionTrace struct {
	ExecutionID, TraceID, ParentID, NeuronID string
	Runtime                                  execution.Runtime
	Outcome                                  string
	FinishedAt                               time.Time
	WaitMillis, DurationMillis               int64
	CPUTimeNanos, PeakMemoryBytes            int64
}

type executionTelemetry struct {
	totals ExecutionTotals
	ring   [executionHistoryLimit]ExecutionTrace
	next   int
	count  int
}

// finish is called once under the scheduler lock after result fencing.
func (t *executionTelemetry) finish(req execution.Request, a *assignment, result execution.Result, err, cause error) {
	outcome := "failed"
	switch {
	case errors.Is(cause, ErrSessionLost) || errors.Is(err, ErrSessionLost):
		outcome = "lost"
		t.totals.Lost++
	case errors.Is(cause, context.Canceled) || errors.Is(cause, context.DeadlineExceeded) || errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) || (cause == nil && err == nil && result.Status == execution.Cancelled):
		outcome = "cancelled"
		t.totals.Cancelled++
	case cause == nil && err == nil && result.Status == execution.Succeeded:
		outcome = "succeeded"
		t.totals.Succeeded++
	default:
		t.totals.Failed++
	}
	now := time.Now()
	trace := ExecutionTrace{ExecutionID: req.ID, TraceID: req.TraceID, ParentID: req.ParentID, NeuronID: a.session.id, Runtime: req.Runtime, Outcome: outcome, FinishedAt: now.UTC(), WaitMillis: a.wait.Milliseconds(), DurationMillis: now.Sub(a.started).Milliseconds()}
	// Usage is reported by the authenticated runtime, not proof of enforcement.
	// Zero means the runtime did not report a measurement.
	if cause == nil && err == nil {
		trace.CPUTimeNanos = int64(result.Usage.CPUTime)
		trace.PeakMemoryBytes = result.Usage.PeakMemoryBytes
	}
	t.ring[t.next] = trace
	t.next = (t.next + 1) % executionHistoryLimit
	if t.count < executionHistoryLimit {
		t.count++
	}
}

func (t *executionTelemetry) recent() []ExecutionTrace {
	rows := make([]ExecutionTrace, 0, t.count)
	for i := 0; i < t.count; i++ {
		rows = append(rows, t.ring[(t.next-1-i+executionHistoryLimit)%executionHistoryLimit])
	}
	return rows
}
