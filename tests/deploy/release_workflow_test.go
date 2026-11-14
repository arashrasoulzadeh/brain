package deploy_test

import (
	"os"
	"strings"
	"testing"
)

func TestBackendReleaseWorkflowPublishesAllImagesAndDigestManifest(t *testing.T) {
	workflow, err := os.ReadFile("../../.github/workflows/release-backend.yml")
	if err != nil {
		t.Fatal(err)
	}
	content := string(workflow)
	for _, image := range []string{"backend", "compiler", "runner"} {
		if !strings.Contains(content, "name: "+image) {
			t.Fatalf("release workflow does not publish %s image", image)
		}
	}
	for _, dockerfile := range []string{"Dockerfile", "Dockerfile.compiler", "Dockerfile.runner"} {
		if _, err := os.Stat("../../" + dockerfile); err != nil {
			t.Fatalf("missing release Dockerfile %s: %v", dockerfile, err)
		}
		if !strings.Contains(content, dockerfile) {
			t.Fatalf("release workflow does not select %s", dockerfile)
		}
	}
	for _, architecture := range []string{"linux/amd64", "linux/arm64"} {
		if !strings.Contains(content, architecture) {
			t.Fatalf("release workflow does not publish %s", architecture)
		}
	}
	if !strings.Contains(content, "${{ inputs.release_sha }}") {
		t.Fatal("release workflow does not tag images with the requested full commit SHA")
	}
	for _, required := range []string{
		"docker/build-push-action@v6",
		"docker buildx imagetools inspect",
		"release-manifest.json",
		"actions/upload-artifact@v4",
	} {
		if !strings.Contains(content, required) {
			t.Fatalf("release workflow is missing %q", required)
		}
	}
}
