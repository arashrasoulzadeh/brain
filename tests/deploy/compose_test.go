package deploy_test

import (
	"errors"
	"os"
	"testing"

	"brain/internal/deploy"
)

func TestValidateComposeConfigChecksRequiredBackendServices(t *testing.T) {
	raw, err := os.ReadFile("../../compose.test.yml")
	if err != nil {
		t.Fatal(err)
	}
	if err := deploy.ValidateComposeConfig(raw); err != nil {
		t.Fatal(err)
	}
	if !errors.Is(deploy.ValidateComposeConfig([]byte("services: {}")), deploy.ErrInvalidComposeConfig) {
		t.Fatal("incomplete compose config accepted")
	}
}
