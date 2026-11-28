package cortex

import (
	"context"
	"errors"
)

// OperatorDrainStore persists only host-selected Neuron identities. A pause is
// independent of session identity and does not override autonomous draining.
type OperatorDrainStore interface {
	LoadOperatorDrains(context.Context, int) ([]string, error)
	SaveOperatorDrain(context.Context, string, bool) error
}

var ErrDrainStoreAlreadyBound = errors.New("operator drain store cannot be bound after placement or management")

// BindOperatorDrainStore restores operator placement policy before serving
// requests. Neurons may already be registered, but no assignment may exist.
func (s *Scheduler) BindOperatorDrainStore(ctx context.Context, store OperatorDrainStore) error {
	if store == nil {
		return ErrInvalidConfig
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.drainStore != nil || len(s.operatorDrains) != 0 || s.telemetry.totals.Assigned != 0 || len(s.queue) != 0 {
		return ErrDrainStoreAlreadyBound
	}
	ids, err := store.LoadOperatorDrains(ctx, maxNeurons+1)
	if err != nil {
		return err
	}
	if len(ids) > maxNeurons {
		return ErrInvalidConfig
	}
	paused := make(map[string]bool, len(ids))
	for _, id := range ids {
		if !validID(id) || paused[id] {
			return ErrInvalidNeuron
		}
		paused[id] = true
	}
	s.operatorDrains, s.drainStore = paused, store
	s.signalLocked()
	return nil
}

// OperatorPaused permits clearing a persisted pause even if the enrolled
// identity is currently offline. It grants no registration or process rights.
func (s *Scheduler) OperatorPaused(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.operatorDrains[id]
}
