package deploy_test

import (
	"errors"
	"testing"
	"time"

	"brain/internal/deploy"
)

func TestManualRoomDrainRequiresControlledDeadline(t *testing.T) {
	now := time.Date(2026, 11, 15, 10, 0, 0, 0, time.UTC)
	request := deploy.ManualRoomDrainRequest{
		DeploymentID: "deploy-1",
		RequestedAt:  now,
		Deadline:     now.Add(15 * time.Minute),
		RoomCount:    3,
	}

	if err := deploy.ValidateManualRoomDrain(request, now); err != nil {
		t.Fatalf("valid drain request rejected: %v", err)
	}
	if err := deploy.ValidateManualRoomDrain(request, now.Add(16*time.Minute)); !errors.Is(err, deploy.ErrRoomDrainDeadlineExpired) {
		t.Fatalf("expired drain request returned %v", err)
	}
}

func TestManualRoomDrainCannotExceedMaximumDeadline(t *testing.T) {
	now := time.Date(2026, 11, 15, 10, 0, 0, 0, time.UTC)
	request := deploy.ManualRoomDrainRequest{
		DeploymentID: "deploy-1",
		RequestedAt:  now,
		Deadline:     now.Add(deploy.MaxManualRoomDrainDuration + time.Second),
		RoomCount:    1,
	}

	if err := deploy.ValidateManualRoomDrain(request, now); !errors.Is(err, deploy.ErrRoomDrainDeadlineTooLong) {
		t.Fatalf("overlong drain request returned %v", err)
	}
}

func TestManualRoomDrainReportsCompletionOnlyAfterAllRoomsDrain(t *testing.T) {
	now := time.Date(2026, 11, 15, 10, 0, 0, 0, time.UTC)
	request := deploy.ManualRoomDrainRequest{
		DeploymentID: "deploy-1",
		RequestedAt:  now,
		Deadline:     now.Add(15 * time.Minute),
		RoomCount:    3,
	}

	status, err := deploy.EvaluateManualRoomDrain(request, 2, now.Add(5*time.Minute))
	if err != nil {
		t.Fatalf("partial drain evaluation failed: %v", err)
	}
	if status != deploy.RoomDrainInProgress {
		t.Fatalf("partial drain status = %q", status)
	}

	status, err = deploy.EvaluateManualRoomDrain(request, 3, now.Add(5*time.Minute))
	if err != nil {
		t.Fatalf("completed drain evaluation failed: %v", err)
	}
	if status != deploy.RoomDrainComplete {
		t.Fatalf("completed drain status = %q", status)
	}
}

func TestManualRoomDrainDeadlineProducesControlledExpiry(t *testing.T) {
	now := time.Date(2026, 11, 15, 10, 0, 0, 0, time.UTC)
	request := deploy.ManualRoomDrainRequest{
		DeploymentID: "deploy-1",
		RequestedAt:  now,
		Deadline:     now.Add(15 * time.Minute),
		RoomCount:    3,
	}

	status, err := deploy.EvaluateManualRoomDrain(request, 2, request.Deadline)
	if err != nil {
		t.Fatalf("deadline evaluation failed: %v", err)
	}
	if status != deploy.RoomDrainDeadlineReached {
		t.Fatalf("deadline status = %q", status)
	}
}
