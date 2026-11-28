package cortex

import (
	"context"
	"errors"
	"time"
)

// ExecutionHistory is a trusted host metadata store. It must not store inputs,
// outputs, caller credentials, or raw execution errors.
type ExecutionHistory interface {
	LoadRecentExecutions(context.Context, int) ([]ExecutionTrace, error)
	AppendExecutionTrace(context.Context, ExecutionTrace, int) error
}

var ErrHistoryAlreadyBound = errors.New("execution history cannot be bound after assignment")

// BindExecutionHistory is a startup operation, before admitting any execution.
// A failed load leaves the scheduler unchanged. Counters remain process-local.
func (s *Scheduler) BindExecutionHistory(ctx context.Context, store ExecutionHistory) error {
	if store == nil {
		return ErrInvalidConfig
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.telemetry.store != nil || s.telemetry.totals.Assigned != 0 {
		return ErrHistoryAlreadyBound
	}
	rows, err := store.LoadRecentExecutions(ctx, executionHistoryLimit)
	if err != nil {
		return err
	}
	if len(rows) > executionHistoryLimit {
		return ErrInvalidConfig
	}
	for i := len(rows) - 1; i >= 0; i-- {
		s.telemetry.record(rows[i])
	}
	s.telemetry.store = store
	return nil
}

// Persistence is outside the scheduler lock. A metadata storage failure must
// never reinterpret or retry a completed execution, especially a mutation.
// Sticky degradation makes the possible history gap visible to the operator.
func (s *Scheduler) persistTrace(store ExecutionHistory, trace ExecutionTrace) {
	if store == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := store.AppendExecutionTrace(ctx, trace, executionHistoryLimit); err != nil {
		s.mu.Lock()
		s.telemetry.degraded = true
		s.mu.Unlock()
	}
}
