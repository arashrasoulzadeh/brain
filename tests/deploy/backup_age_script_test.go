package deploy_test

import (
	"os"
	"strings"
	"testing"
)

func TestBackupAgeScriptFailsForStaleVerification(t *testing.T) {
	content, err := os.ReadFile("../../scripts/check-backup-age.sh")
	if err != nil {
		t.Fatal(err)
	}
	script := string(content)
	for _, required := range []string{"LAST_VERIFIED_BACKUP_EPOCH", "BACKUP_ALERT_AFTER_SECONDS", "stale", "exit 1"} {
		if !strings.Contains(script, required) {
			t.Fatalf("backup age script missing %q", required)
		}
	}
}
