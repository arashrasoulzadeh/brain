package deploy_test

import (
	"errors"
	"sync"
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

func TestAssignmentGateSupportsConcurrentAdmissionChecks(t *testing.T) {
	gate := deploy.NewAssignmentGate()
	var group sync.WaitGroup
	for index := 0; index < 16; index++ {
		group.Add(1)
		go func() {
			defer group.Done()
			gate.Pause()
			_ = gate.Accept()
			gate.Resume()
			_ = gate.Accept()
		}()
	}
	group.Wait()
}
