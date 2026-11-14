package deploy_test

import (
	"os"
	"strings"
	"testing"
)

func TestDataLossTargetScriptChecksTwentyFourHours(t *testing.T) {
	content, err := os.ReadFile("../../scripts/check-data-loss-target.sh")
	if err != nil {
		t.Fatal(err)
	}
	script := string(content)
	for _, required := range []string{"LAST_VERIFIED_BACKUP_EPOCH", "DATA_LOSS_TARGET_SECONDS", "86400", "exit 1"} {
		if !strings.Contains(script, required) {
			t.Fatalf("data-loss target script missing %q", required)
		}
	}
}
