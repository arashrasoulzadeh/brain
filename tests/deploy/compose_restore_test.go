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
	for _, required := range []string{"COMPOSE_FILE", "RESTORE_PROJECT", "case \"$restore_project\" in", "*[!A-Za-z0-9_-]*)", "--no-build", "restore-postgres-backup.sh", "verify-object-manifest.sh"} {
		if !strings.Contains(script, required) {
			t.Fatalf("Compose restore script missing %q", required)
		}
	}
	if strings.Contains(script, "tr -cd") {
		t.Fatal("restore script silently rewrites invalid project names")
	}
}
