package deploy_test

import (
	"errors"
	"testing"

	"brain/internal/deploy"
)

func TestValidateDeploymentSequencePullsBeforeDrain(t *testing.T) {
	valid := []deploy.DeploymentStep{
		deploy.StepAcquireLock,
		deploy.StepValidateRelease,
		deploy.StepValidateCompose,
		deploy.StepPullImages,
		deploy.StepPauseAssignments,
		deploy.StepDrainRooms,
	}
	if err := deploy.ValidateDeploymentSequence(valid); err != nil {
		t.Fatal(err)
	}

	invalid := append([]deploy.DeploymentStep{}, valid[:3]...)
	invalid = append(invalid, deploy.StepDrainRooms, deploy.StepPullImages)
	if !errors.Is(deploy.ValidateDeploymentSequence(invalid), deploy.ErrImagesPulledAfterDrain) {
		t.Fatal("accepted image pull after drain")
	}
}

func TestValidateDeploymentSequenceRequiresPauseBeforeDrain(t *testing.T) {
	steps := []deploy.DeploymentStep{deploy.StepAcquireLock, deploy.StepValidateRelease, deploy.StepValidateCompose, deploy.StepPullImages, deploy.StepDrainRooms}
	if err := deploy.ValidateDeploymentSequence(steps); !errors.Is(err, deploy.ErrDeploymentOrder) {
		t.Fatalf("drain without pausing assignments accepted: %v", err)
	}
}
