package deploy_test

import (
	"errors"
	"testing"

	"brain/internal/deploy"
)

func TestAssignmentGatePausesNewMatches(t *testing.T) {
	gate := deploy.NewAssignmentGate()
	if err := gate.Accept(); err != nil {
		t.Fatal(err)
	}
	gate.Pause()
	if !errors.Is(gate.Accept(), deploy.ErrAssignmentsPaused) {
		t.Fatal("paused assignment gate accepted new work")
	}
	gate.Resume()
	if err := gate.Accept(); err != nil {
		t.Fatal(err)
	}
}
