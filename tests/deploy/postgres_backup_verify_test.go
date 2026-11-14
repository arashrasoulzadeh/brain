package deploy_test

import (
	"os"
	"strings"
	"testing"
)

func TestPostgresBackupVerificationScriptInspectsCatalog(t *testing.T) {
	content, err := os.ReadFile("../../scripts/verify-postgres-backup.sh")
	if err != nil {
		t.Fatal(err)
	}
	script := string(content)
	for _, required := range []string{"BACKUP_INPUT", "pg_restore", "--list", "--no-password"} {
		if !strings.Contains(script, required) {
			t.Fatalf("backup verification script missing %q", required)
		}
	}
}
