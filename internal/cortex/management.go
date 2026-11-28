package cortex

import (
	"context"

	"brain/internal/execution"
)

// SetOperatorDrain changes host-side placement, not the Neuron's autonomous
// runtime drain. The bounded set survives reconnects for this Brain lifetime.
// Only trusted operator/control-plane adapters may call this method.
func (s *Scheduler) SetOperatorDrain(id string, draining bool) error {
	if !validID(id) {
		return ErrInvalidNeuron
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if draining {
		if !s.operatorDrains[id] && len(s.operatorDrains) >= s.config.MaxNeurons {
			return execution.ErrUnavailable
		}
		s.operatorDrains[id] = true
	} else {
		delete(s.operatorDrains, id)
	}
	s.signalLocked()
	return nil
}

// WaitNeuronIdle waits for all assignments owned by the paused identity,
// including a fenced old session still terminating. It never waits while
// holding the scheduler lock, and cannot be used to drain other identities.
func (s *Scheduler) WaitNeuronIdle(ctx context.Context, id string) error {
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		s.mu.Lock()
		s.expireLocked()
		if !s.operatorDrains[id] {
			s.mu.Unlock()
			return ErrInvalidNeuron
		}
		busy := false
		for _, assignment := range s.active {
			if assignment.session.id == id {
				busy = true
				break
			}
		}
		changed := s.changed
		s.mu.Unlock()
		if !busy {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-changed:
		}
	}
}

// Neuron returns a snapshot of a live identity, without exposing its executor.
func (s *Scheduler) Neuron(id string) (NeuronSnapshot, bool) {
	for _, neuron := range s.Snapshot().Neurons {
		if neuron.ID == id {
			return neuron, true
		}
	}
	return NeuronSnapshot{}, false
}
