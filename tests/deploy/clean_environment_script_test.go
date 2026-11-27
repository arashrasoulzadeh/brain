package deploy_test

import (
	"os"
	"strings"
	"testing"
)

func TestCleanEnvironmentRehearsalRequiresDeployRollbackAndRestore(t *testing.T) {
	content, err := os.ReadFile("../../scripts/clean-environment-rehearsal.sh")
	if err != nil {
		t.Fatal(err)
	}
	script := string(content)
	for _, required := range []string{"up -d --build", "start-compose-no-build.sh", "rollback-compose.sh", "ROLLBACK_COMPOSE_FILE", "restore-compose.sh", "BACKUP_INPUT", "OBJECT_MANIFEST"} {
		if !strings.Contains(script, required) {
			t.Fatalf("rehearsal script missing %q", required)
		}
	}
	for _, required := range []string{"case \"$project\" in", "case \"$restore_project\" in", "*[!A-Za-z0-9_-]*)"} {
		if !strings.Contains(script, required) {
			t.Fatalf("rehearsal script missing %q", required)
		}
	}
}

func TestTestWorkflowWiresFullRehearsalInputs(t *testing.T) {
	content, err := os.ReadFile("../../.github/workflows/test.yml")
	if err != nil {
		t.Fatal(err)
	}
	workflow := string(content)
	for _, required := range []string{"BRAIN_REHEARSAL_ENABLED", "BRAIN_ROLLBACK_COMPOSE_FILE", "BRAIN_BACKUP_INPUT", "BRAIN_RESTORE_DSN", "BRAIN_OBJECT_MANIFEST", "scripts/clean-environment-rehearsal.sh"} {
		if !strings.Contains(workflow, required) {
			t.Fatalf("test workflow missing %q", required)
		}
	}
}

func TestCleanEnvironmentRehearsalKeepsRestoreIsolatedAndCleansBothProjects(t *testing.T) {
	content, err := os.ReadFile("../../scripts/clean-environment-rehearsal.sh")
	if err != nil {
		t.Fatal(err)
	}
	script := string(content)
	for _, required := range []string{
		"docker compose -p \"$project\" -f \"$compose_file\" down --volumes --remove-orphans",
		"docker compose -p \"$restore_project\" -f \"$compose_file\" down --volumes --remove-orphans",
		"RESTORE_PROJECT=\"$restore_project\"",
		"RESTORE_DSN=\"$restore_dsn\"",
		"OBJECT_MANIFEST=\"$object_manifest\"",
	} {
		if !strings.Contains(script, required) {
			t.Fatalf("rehearsal script missing isolated restore requirement %q", required)
		}
	}
}
