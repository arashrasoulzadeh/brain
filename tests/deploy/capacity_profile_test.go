package deploy_test

import (
	"errors"
	"testing"

	"brain/internal/deploy"
)

func TestCapacityProfileRequiresReferenceWorkload(t *testing.T) {
	profile := deploy.CapacityProfile{
		CPUCores:       4,
		MemoryBytes:    8 << 30,
		PlayersPerRoom: 8,
		SimulatedRTTMs: 50,
	}
	if err := profile.Validate(); err != nil {
		t.Fatal(err)
	}
	profile.PlayersPerRoom = 0
	if !errors.Is(profile.Validate(), deploy.ErrInvalidCapacityProfile) {
		t.Fatal("empty reference workload was accepted")
	}
}
