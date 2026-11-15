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
	if !errors.Is(deploy.ValidateComposeConfig([]byte("services:\n  brain: {}\n  postgres: {}\n")), deploy.ErrInvalidComposeConfig) {
		t.Fatal("empty required service definitions accepted")
	}
	for _, raw := range []string{
		"services:\n  brain:\n    privileged: true\n  postgres:\n    image: postgres:16\n",
		"services:\n  brain:\n    network_mode: host\n  postgres:\n    image: postgres:16\n",
		"services:\n  brain:\n    pid: host\n  postgres:\n    image: postgres:16\n",
	} {
		if !errors.Is(deploy.ValidateComposeConfig([]byte(raw)), deploy.ErrInvalidComposeConfig) {
			t.Fatal("Compose host isolation bypass was accepted")
		}
	}
}
