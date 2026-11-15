package deploy

import (
	"errors"
	"regexp"
)

type DeploymentStep string

const (
	StepAcquireLock      DeploymentStep = "acquire-lock"
	StepValidateRelease  DeploymentStep = "validate-release"
	StepValidateCompose  DeploymentStep = "validate-compose"
	StepPullImages       DeploymentStep = "pull-images"
	StepPauseAssignments DeploymentStep = "pause-assignments"
	StepDrainRooms       DeploymentStep = "drain-rooms"
)

var (
	ErrImagesPulledAfterDrain = errors.New("images must be pulled before draining")
	ErrDeploymentOrder        = errors.New("invalid deployment step order")
	ErrProjectIDRequired      = errors.New("project ID required")
)

var deploymentProjectID = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]*$`)

type AssignmentPauser interface {
	PauseAssignments(projectID string) error
}

func PauseAssignmentsForDeployment(pauser AssignmentPauser, projectID string) error {
	if pauser == nil {
		return ErrDeploymentOrder
	}
	if !deploymentProjectID.MatchString(projectID) {
		return ErrProjectIDRequired
	}
	return pauser.PauseAssignments(projectID)
}

func ValidateDeploymentSequence(steps []DeploymentStep) error {
	pullIndex := indexOfDeploymentStep(steps, StepPullImages)
	if pullIndex < 0 {
		return ErrDeploymentOrder
	}
	drainIndex := indexOfDeploymentStep(steps, StepDrainRooms)
	if drainIndex >= 0 && pullIndex > drainIndex {
		return ErrImagesPulledAfterDrain
	}
	if drainIndex >= 0 {
		pauseIndex := indexOfDeploymentStep(steps, StepPauseAssignments)
		if pauseIndex < 0 || pauseIndex > drainIndex {
			return ErrDeploymentOrder
		}
	}
	for _, required := range []DeploymentStep{StepAcquireLock, StepValidateRelease, StepValidateCompose} {
		if indexOfDeploymentStep(steps, required) < 0 {
			return ErrDeploymentOrder
		}
	}
	return nil
}

func indexOfDeploymentStep(steps []DeploymentStep, target DeploymentStep) int {
	for index, step := range steps {
		if step == target {
			return index
		}
	}
	return -1
}
