package deploy_test

import (
	"os"
	"strings"
	"testing"
)

func TestComposeRestoreScriptUsesFreshProjectAndNoBuild(t *testing.T) {
	content, err := os.ReadFile("../../scripts/restore-compose.sh")
	if err != nil {
		t.Fatal(err)
	}
	script := string(content)
	for _, required := range []string{"COMPOSE_FILE", "RESTORE_PROJECT", "--no-build", "restore-postgres-backup.sh", "verify-object-manifest.sh"} {
		if !strings.Contains(script, required) {
			t.Fatalf("Compose restore script missing %q", required)
		}
	}
}
