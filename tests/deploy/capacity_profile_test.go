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
		ConservativeRoomCap:       6,
		RepresentativeAppRequests: 100,
		UploadRequests:            10,
		BackgroundJobs:            10,
		EyeExportJobs:             5,
		SimulatedRTTMs:            50,
		PacketLossPercent:         1,
		P95InputSnapshotLatencyMs: 120,
		TickDelayMs:               20,
		MeasuredCPUPercent:        45,
		MeasuredMemoryBytes:       512 << 20,
		QueueDepth:                2,
		TrafficBytes:              1 << 20,
	}
	if err := profile.Validate(); err != nil {
		t.Fatal(err)
	}
	profile.PlayersPerRoom = 0
	if !errors.Is(profile.Validate(), deploy.ErrInvalidCapacityProfile) {
		t.Fatal("empty reference workload was accepted")
	}
}

func TestCapacityProfileBoundsConservativeRoomCap(t *testing.T) {
	profile := deploy.CapacityProfile{CPUCores: 4, MemoryBytes: 1 << 30, PlayersPerRoom: 8, ConservativeRoomCap: 6, RepresentativeAppRequests: 100, UploadRequests: 10, BackgroundJobs: 10, EyeExportJobs: 5, SimulatedRTTMs: 40, PacketLossPercent: 1, P95InputSnapshotLatencyMs: 120, TickDelayMs: 20, MeasuredCPUPercent: 45, MeasuredMemoryBytes: 512 << 20, TrafficBytes: 1 << 20}
	if err := profile.Validate(); err != nil {
		t.Fatal(err)
	}
	profile.ConservativeRoomCap = 9
	if !errors.Is(profile.Validate(), deploy.ErrInvalidCapacityProfile) {
		t.Fatal("room cap exceeded measured players")
	}
}

func TestCapacityProfileRequiresRepresentativeAppRequests(t *testing.T) {
	profile := deploy.CapacityProfile{CPUCores: 4, MemoryBytes: 1 << 30, PlayersPerRoom: 8, RepresentativeAppRequests: 100, SimulatedRTTMs: 40}
	profile.ConservativeRoomCap = 6
	profile.UploadRequests = 10
	profile.BackgroundJobs = 10
	profile.EyeExportJobs = 5
	profile.PacketLossPercent = 1
	profile.P95InputSnapshotLatencyMs = 120
	profile.TickDelayMs = 20
	profile.MeasuredCPUPercent = 45
	profile.MeasuredMemoryBytes = 512 << 20
	profile.TrafficBytes = 1 << 20
	if err := profile.Validate(); err != nil {
		t.Fatal(err)
	}
	profile.RepresentativeAppRequests = 0
	if !errors.Is(profile.Validate(), deploy.ErrInvalidCapacityProfile) {
		t.Fatal("capacity profile accepted without representative app requests")
	}
}

func TestCapacityProfileRequiresUploadsAndBackgroundJobs(t *testing.T) {
	profile := deploy.CapacityProfile{CPUCores: 4, MemoryBytes: 1 << 30, PlayersPerRoom: 8, RepresentativeAppRequests: 100, UploadRequests: 10, BackgroundJobs: 10, EyeExportJobs: 5, SimulatedRTTMs: 40, PacketLossPercent: 1, P95InputSnapshotLatencyMs: 120, TickDelayMs: 20, MeasuredCPUPercent: 45, MeasuredMemoryBytes: 512 << 20, TrafficBytes: 1 << 20}
	profile.ConservativeRoomCap = 6
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

func TestCapacityProfileRequiresResourceAndTrafficMeasurements(t *testing.T) {
	profile := deploy.CapacityProfile{CPUCores: 4, MemoryBytes: 1 << 30, PlayersPerRoom: 8, RepresentativeAppRequests: 100, UploadRequests: 10, BackgroundJobs: 10, EyeExportJobs: 5, SimulatedRTTMs: 40, PacketLossPercent: 1, P95InputSnapshotLatencyMs: 120, TickDelayMs: 20, MeasuredCPUPercent: 45, MeasuredMemoryBytes: 512 << 20, TrafficBytes: 1 << 20}
	profile.ConservativeRoomCap = 6
	if err := profile.Validate(); err != nil {
		t.Fatal(err)
	}
	profile.MeasuredCPUPercent = 0
	if !errors.Is(profile.Validate(), deploy.ErrInvalidCapacityProfile) {
		t.Fatal("capacity profile accepted without CPU measurement")
	}
	profile.MeasuredCPUPercent = 45
	profile.TrafficBytes = 0
	if !errors.Is(profile.Validate(), deploy.ErrInvalidCapacityProfile) {
		t.Fatal("capacity profile accepted without traffic measurement")
	}
}

func TestCapacityProfileRejectsMeasuredMemoryAboveBudget(t *testing.T) {
	profile := deploy.CapacityProfile{CPUCores: 4, MemoryBytes: 1 << 30, PlayersPerRoom: 8, ConservativeRoomCap: 6, RepresentativeAppRequests: 100, UploadRequests: 10, BackgroundJobs: 10, EyeExportJobs: 5, SimulatedRTTMs: 40, PacketLossPercent: 1, P95InputSnapshotLatencyMs: 120, TickDelayMs: 20, MeasuredCPUPercent: 45, MeasuredMemoryBytes: 2 << 30, TrafficBytes: 1 << 20}
	if !errors.Is(profile.Validate(), deploy.ErrInvalidCapacityProfile) {
		t.Fatal("capacity profile accepted measured memory above budget")
	}
}

func TestCapacityProfileRequiresEyeExportActivity(t *testing.T) {
	profile := deploy.CapacityProfile{CPUCores: 4, MemoryBytes: 1 << 30, PlayersPerRoom: 8, RepresentativeAppRequests: 100, UploadRequests: 10, BackgroundJobs: 10, EyeExportJobs: 5, SimulatedRTTMs: 40, PacketLossPercent: 1, P95InputSnapshotLatencyMs: 120, TickDelayMs: 20, MeasuredCPUPercent: 45, MeasuredMemoryBytes: 512 << 20, TrafficBytes: 1 << 20}
	profile.ConservativeRoomCap = 6
	if err := profile.Validate(); err != nil {
		t.Fatal(err)
	}
	profile.EyeExportJobs = 0
	if !errors.Is(profile.Validate(), deploy.ErrInvalidCapacityProfile) {
		t.Fatal("capacity profile accepted without Eye export activity")
	}
}

func TestCapacityProfileBoundsPacketLoss(t *testing.T) {
	profile := deploy.CapacityProfile{CPUCores: 4, MemoryBytes: 1 << 30, PlayersPerRoom: 8, RepresentativeAppRequests: 100, UploadRequests: 10, BackgroundJobs: 10, EyeExportJobs: 5, SimulatedRTTMs: 40, PacketLossPercent: 101, P95InputSnapshotLatencyMs: 120, TickDelayMs: 20, MeasuredCPUPercent: 45, MeasuredMemoryBytes: 512 << 20, TrafficBytes: 1 << 20}
	profile.ConservativeRoomCap = 6
	if !errors.Is(profile.Validate(), deploy.ErrInvalidCapacityProfile) {
		t.Fatal("capacity profile accepted packet loss above 100 percent")
	}
}

func TestCapacityProfileRequiresP95InputSnapshotLatency(t *testing.T) {
	profile := deploy.CapacityProfile{CPUCores: 4, MemoryBytes: 1 << 30, PlayersPerRoom: 8, RepresentativeAppRequests: 100, UploadRequests: 10, BackgroundJobs: 10, EyeExportJobs: 5, SimulatedRTTMs: 40, PacketLossPercent: 1, P95InputSnapshotLatencyMs: 120, TickDelayMs: 20, MeasuredCPUPercent: 45, MeasuredMemoryBytes: 512 << 20, TrafficBytes: 1 << 20}
	profile.ConservativeRoomCap = 6
	if err := profile.Validate(); err != nil {
		t.Fatal(err)
	}
	profile.P95InputSnapshotLatencyMs = 0
	if !errors.Is(profile.Validate(), deploy.ErrInvalidCapacityProfile) {
		t.Fatal("capacity profile accepted without p95 latency measurement")
	}
}

func TestCapacityProfileRequiresTickDelayMeasurement(t *testing.T) {
	profile := deploy.CapacityProfile{CPUCores: 4, MemoryBytes: 1 << 30, PlayersPerRoom: 8, RepresentativeAppRequests: 100, UploadRequests: 10, BackgroundJobs: 10, EyeExportJobs: 5, SimulatedRTTMs: 40, PacketLossPercent: 1, P95InputSnapshotLatencyMs: 120, TickDelayMs: 20, MeasuredCPUPercent: 45, MeasuredMemoryBytes: 512 << 20, TrafficBytes: 1 << 20}
	profile.ConservativeRoomCap = 6
	if err := profile.Validate(); err != nil {
		t.Fatal(err)
	}
	profile.TickDelayMs = 0
	if !errors.Is(profile.Validate(), deploy.ErrInvalidCapacityProfile) {
		t.Fatal("capacity profile accepted without tick delay measurement")
	}
}
