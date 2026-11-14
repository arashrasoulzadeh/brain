package deploy_test

import (
	"os"
	"strings"
	"testing"
)

func TestRollbackComposeRequiresPinnedRollbackDefinition(t *testing.T) {
	content, err := os.ReadFile("../../scripts/rollback-compose.sh")
	if err != nil {
		t.Fatal(err)
	}
	script := string(content)
	for _, required := range []string{"ROLLBACK_COMPOSE_FILE", "COMPOSE_PROJECT_NAME", "down --volumes", "up -d --no-build"} {
		if !strings.Contains(script, required) {
			t.Fatalf("rollback script missing %q", required)
		}
	}
}
