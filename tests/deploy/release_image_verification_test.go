package deploy_test

import (
	"os"
	"strings"
	"testing"
)

func TestBackendReleaseWorkflowPullsAndSmokeTestsBothArchitectures(t *testing.T) {
	workflow, err := os.ReadFile("../../.github/workflows/release-backend.yml")
	if err != nil {
		t.Fatal(err)
	}
	content := string(workflow)

	for _, architecture := range []string{"linux/amd64", "linux/arm64"} {
		if !strings.Contains(content, architecture) {
			t.Fatalf("release workflow does not verify %s", architecture)
		}
	}
	for _, required := range []string{
		"verify-images:",
		"docker pull --platform",
		"docker buildx imagetools inspect",
		"docker run --rm --platform",
		"--entrypoint node",
		"brain-${image}@${digest}",
		"^sha256:[0-9a-f]{64}$",
	} {
		if !strings.Contains(content, required) {
			t.Fatalf("release workflow is missing %q", required)
		}
	}
}
