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

func TestBackupAgeScriptValidatesTimestamps(t *testing.T) {
	content, err := os.ReadFile("../../scripts/check-backup-age.sh")
	if err != nil {
		t.Fatal(err)
	}
	script := string(content)
	for _, required := range []string{"LAST_VERIFIED_BACKUP_EPOCH must be an integer", "BACKUP_ALERT_AFTER_SECONDS must be an integer", "CURRENT_TIME_EPOCH must be an integer", "\"$age\" -lt 0", "timestamp is in the future"} {
		if !strings.Contains(script, required) {
			t.Fatalf("backup age script missing %q", required)
		}
	}
}
