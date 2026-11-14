package deploy

import "errors"

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
)

func ValidateDeploymentSequence(steps []DeploymentStep) error {
	pullIndex := indexOfDeploymentStep(steps, StepPullImages)
	if pullIndex < 0 {
		return ErrDeploymentOrder
	}
	drainIndex := indexOfDeploymentStep(steps, StepDrainRooms)
	if drainIndex >= 0 && pullIndex > drainIndex {
		return ErrImagesPulledAfterDrain
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
