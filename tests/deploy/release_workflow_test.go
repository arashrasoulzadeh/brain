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

func TestBackendReleaseWorkflowScopesGHCRCredentialsToPublishJobs(t *testing.T) {
	workflow, err := os.ReadFile("../../.github/workflows/release-backend.yml")
	if err != nil {
		t.Fatal(err)
	}
	content := string(workflow)
	if strings.Contains(content, "permissions:\n  contents: read\n  packages: write") {
		t.Fatal("release workflow grants GHCR write permission at workflow scope")
	}
	verify := workflowJob(content, "verify")
	if strings.Contains(verify, "GITHUB_TOKEN") || strings.Contains(verify, "docker/login-action") {
		t.Fatal("verify job can access GHCR credentials")
	}
	for _, job := range []string{"publish-images", "publish-manifest"} {
		section := workflowJob(content, job)
		if !strings.Contains(section, "permissions:\n      contents: read\n      packages: write") {
			t.Fatalf("%s job does not declare scoped GHCR write permission", job)
		}
	}
}

func TestBackendReleaseWorkflowSmokeTestsDigestPinnedReferences(t *testing.T) {
	workflow, err := os.ReadFile("../../.github/workflows/release-backend.yml")
	if err != nil {
		t.Fatal(err)
	}
	verify := workflowJob(string(workflow), "verify-images")
	for _, required := range []string{
		"digest=",
		"reference=\"ghcr.io/${IMAGE_OWNER}/brain-${image}@${digest}\"",
		"docker pull --platform \"${TARGET_PLATFORM}\" \"${reference}\"",
		"--entrypoint node \"${reference}\" --version",
	} {
		if !strings.Contains(verify, required) {
			t.Fatalf("verify-images does not use digest-pinned reference %q", required)
		}
	}
}

func workflowJob(workflow, jobName string) string {
	lines := strings.Split(workflow, "\n")
	marker := "  " + jobName + ":"
	start := -1
	for index, line := range lines {
		if line == marker {
			start = index
			break
		}
	}
	if start < 0 {
		return ""
	}
	end := len(lines)
	for index := start + 1; index < len(lines); index++ {
		if strings.HasPrefix(lines[index], "  ") && !strings.HasPrefix(lines[index], "    ") {
			end = index
			break
		}
	}
	return strings.Join(lines[start:end], "\n")
}
