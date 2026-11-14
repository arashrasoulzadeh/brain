package deploy

import (
	"errors"
	"sync"
)

var ErrAssignmentsPaused = errors.New("new match assignments are paused")

type AssignmentGate struct {
	mu     sync.RWMutex
	paused bool
}

func NewAssignmentGate() *AssignmentGate {
	return &AssignmentGate{}
}

func (g *AssignmentGate) Pause() {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.paused = true
}

func (g *AssignmentGate) Resume() {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.paused = false
}

func (g *AssignmentGate) Accept() error {
	g.mu.RLock()
	defer g.mu.RUnlock()
	if g.paused {
		return ErrAssignmentsPaused
	}
	return nil
}
