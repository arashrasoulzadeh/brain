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
}
