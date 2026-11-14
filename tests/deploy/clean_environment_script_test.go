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
	for _, required := range []string{"up -d --build", "start-compose-no-build.sh", "restore-compose.sh", "BACKUP_INPUT", "OBJECT_MANIFEST"} {
		if !strings.Contains(script, required) {
			t.Fatalf("rehearsal script missing %q", required)
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
