package deploy

import (
	"errors"
	"strings"
	"time"
)

const MaxManualRoomDrainDuration = 30 * time.Minute

var (
	ErrInvalidRoomDrainRequest  = errors.New("invalid manual room drain request")
	ErrRoomDrainDeadlineExpired = errors.New("manual room drain deadline expired")
	ErrRoomDrainDeadlineTooLong = errors.New("manual room drain deadline exceeds maximum")
)

type ManualRoomDrainRequest struct {
	DeploymentID string
	RequestedAt  time.Time
	Deadline     time.Time
	RoomCount    int
}

type RoomDrainStatus string

const (
	RoomDrainInProgress      RoomDrainStatus = "in-progress"
	RoomDrainComplete        RoomDrainStatus = "complete"
	RoomDrainDeadlineReached RoomDrainStatus = "deadline-reached"
)

func ValidateManualRoomDrain(request ManualRoomDrainRequest, now time.Time) error {
	if strings.TrimSpace(request.DeploymentID) == "" || request.RoomCount < 0 || request.RequestedAt.IsZero() || request.Deadline.IsZero() {
		return ErrInvalidRoomDrainRequest
	}
	if !request.Deadline.After(request.RequestedAt) || request.Deadline.Sub(request.RequestedAt) > MaxManualRoomDrainDuration {
		return ErrRoomDrainDeadlineTooLong
	}
	if now.After(request.Deadline) {
		return ErrRoomDrainDeadlineExpired
	}
	return nil
}

func EvaluateManualRoomDrain(request ManualRoomDrainRequest, drainedRooms int, now time.Time) (RoomDrainStatus, error) {
	if err := ValidateManualRoomDrain(request, now); err != nil && !errors.Is(err, ErrRoomDrainDeadlineExpired) {
		return "", err
	}
	if drainedRooms < 0 || drainedRooms > request.RoomCount {
		return "", ErrInvalidRoomDrainRequest
	}
	if drainedRooms == request.RoomCount {
		return RoomDrainComplete, nil
	}
	if !now.Before(request.Deadline) {
		return RoomDrainDeadlineReached, nil
	}
	return RoomDrainInProgress, nil
}
