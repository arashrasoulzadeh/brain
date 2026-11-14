package deploy_test

import (
	"os"
	"strings"
	"testing"
)

func TestPostgresRestoreScriptRequiresExplicitTarget(t *testing.T) {
	content, err := os.ReadFile("../../scripts/restore-postgres-backup.sh")
	if err != nil {
		t.Fatal(err)
	}
	script := string(content)
	for _, required := range []string{"BACKUP_INPUT", "RESTORE_DSN", "pg_restore", "--single-transaction", "--exit-on-error", "--no-password"} {
		if !strings.Contains(script, required) {
			t.Fatalf("PostgreSQL restore script missing %q", required)
		}
	}
}
