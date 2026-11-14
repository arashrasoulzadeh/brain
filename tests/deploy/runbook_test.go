package deploy_test

import (
	"os"
	"strings"
	"testing"
)

func TestDeploymentAndRestoreRunbooksDescribeCleanEnvironmentGate(t *testing.T) {
	deployment, err := os.ReadFile("../../docs/deployment-runbook.md")
	if err != nil {
		t.Fatal(err)
	}
	restore, err := os.ReadFile("../../docs/restore-runbook.md")
	if err != nil {
		t.Fatal(err)
	}
	combined := string(deployment) + "\n" + string(restore)
	for _, required := range []string{"fresh isolated environment", "rollback", "worker health", "project-isolation", "plugin-revocation", "object manifest"} {
		if !strings.Contains(strings.ToLower(combined), strings.ToLower(required)) {
			t.Fatalf("runbooks missing clean-environment requirement %q", required)
		}
	}
}
