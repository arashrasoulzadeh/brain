package deploy_test

import (
	"os"
	"strings"
	"testing"
)

func TestPostgresAccessScriptUsesSafeProbe(t *testing.T) {
	content, err := os.ReadFile("../../scripts/check-postgres-access.sh")
	if err != nil {
		t.Fatal(err)
	}
	script := string(content)
	for _, required := range []string{"POSTGRES_DSN", "psql", "SELECT 1", "ON_ERROR_STOP"} {
		if !strings.Contains(script, required) {
			t.Fatalf("PostgreSQL access script missing %q", required)
		}
	}
}
