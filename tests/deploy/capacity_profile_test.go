package deploy_test

import (
	"errors"
	"testing"

	"brain/internal/deploy"
)

func TestCapacityProfileRequiresReferenceWorkload(t *testing.T) {
	profile := deploy.CapacityProfile{
		CPUCores:                  4,
		MemoryBytes:               8 << 30,
		PlayersPerRoom:            8,
		RepresentativeAppRequests: 100,
		UploadRequests:            10,
		BackgroundJobs:            10,
		EyeExportJobs:             5,
		SimulatedRTTMs:            50,
		PacketLossPercent:         1,
		P95InputSnapshotLatencyMs: 120,
	}
	if err := profile.Validate(); err != nil {
		t.Fatal(err)
	}
	profile.PlayersPerRoom = 0
	if !errors.Is(profile.Validate(), deploy.ErrInvalidCapacityProfile) {
		t.Fatal("empty reference workload was accepted")
	}
}

func TestCapacityProfileRequiresRepresentativeAppRequests(t *testing.T) {
	profile := deploy.CapacityProfile{CPUCores: 4, MemoryBytes: 1 << 30, PlayersPerRoom: 8, RepresentativeAppRequests: 100, SimulatedRTTMs: 40}
	profile.UploadRequests = 10
	profile.BackgroundJobs = 10
	profile.EyeExportJobs = 5
	profile.PacketLossPercent = 1
	profile.P95InputSnapshotLatencyMs = 120
	if err := profile.Validate(); err != nil {
		t.Fatal(err)
	}
	profile.RepresentativeAppRequests = 0
	if !errors.Is(profile.Validate(), deploy.ErrInvalidCapacityProfile) {
		t.Fatal("capacity profile accepted without representative app requests")
	}
}

func TestCapacityProfileRequiresUploadsAndBackgroundJobs(t *testing.T) {
	profile := deploy.CapacityProfile{CPUCores: 4, MemoryBytes: 1 << 30, PlayersPerRoom: 8, RepresentativeAppRequests: 100, UploadRequests: 10, BackgroundJobs: 10, EyeExportJobs: 5, SimulatedRTTMs: 40, PacketLossPercent: 1, P95InputSnapshotLatencyMs: 120}
	if err := profile.Validate(); err != nil {
		t.Fatal(err)
	}
	profile.UploadRequests = 0
	if !errors.Is(profile.Validate(), deploy.ErrInvalidCapacityProfile) {
		t.Fatal("capacity profile accepted without uploads")
	}
	profile.UploadRequests = 10
	profile.BackgroundJobs = 0
	if !errors.Is(profile.Validate(), deploy.ErrInvalidCapacityProfile) {
		t.Fatal("capacity profile accepted without background jobs")
	}
}

func TestCapacityProfileRequiresEyeExportActivity(t *testing.T) {
	profile := deploy.CapacityProfile{CPUCores: 4, MemoryBytes: 1 << 30, PlayersPerRoom: 8, RepresentativeAppRequests: 100, UploadRequests: 10, BackgroundJobs: 10, EyeExportJobs: 5, SimulatedRTTMs: 40, PacketLossPercent: 1, P95InputSnapshotLatencyMs: 120}
	if err := profile.Validate(); err != nil {
		t.Fatal(err)
	}
	profile.EyeExportJobs = 0
	if !errors.Is(profile.Validate(), deploy.ErrInvalidCapacityProfile) {
		t.Fatal("capacity profile accepted without Eye export activity")
	}
}

func TestCapacityProfileBoundsPacketLoss(t *testing.T) {
	profile := deploy.CapacityProfile{CPUCores: 4, MemoryBytes: 1 << 30, PlayersPerRoom: 8, RepresentativeAppRequests: 100, UploadRequests: 10, BackgroundJobs: 10, EyeExportJobs: 5, SimulatedRTTMs: 40, PacketLossPercent: 101}
	if !errors.Is(profile.Validate(), deploy.ErrInvalidCapacityProfile) {
		t.Fatal("capacity profile accepted packet loss above 100 percent")
	}
}

func TestCapacityProfileRequiresP95InputSnapshotLatency(t *testing.T) {
	profile := deploy.CapacityProfile{CPUCores: 4, MemoryBytes: 1 << 30, PlayersPerRoom: 8, RepresentativeAppRequests: 100, UploadRequests: 10, BackgroundJobs: 10, EyeExportJobs: 5, SimulatedRTTMs: 40, PacketLossPercent: 1, P95InputSnapshotLatencyMs: 120}
	if err := profile.Validate(); err != nil {
		t.Fatal(err)
	}
	profile.P95InputSnapshotLatencyMs = 0
	if !errors.Is(profile.Validate(), deploy.ErrInvalidCapacityProfile) {
		t.Fatal("capacity profile accepted without p95 latency measurement")
	}
}
