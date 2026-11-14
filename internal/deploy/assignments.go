package deploy

import "errors"

var ErrAssignmentsPaused = errors.New("new match assignments are paused")

type AssignmentGate struct {
	paused bool
}

func NewAssignmentGate() *AssignmentGate {
	return &AssignmentGate{}
}

func (g *AssignmentGate) Pause() {
	g.paused = true
}

func (g *AssignmentGate) Resume() {
	g.paused = false
}

func (g *AssignmentGate) Accept() error {
	if g.paused {
		return ErrAssignmentsPaused
	}
	return nil
}
