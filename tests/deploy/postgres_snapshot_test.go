package deploy_test

import (
	"os"
	"strings"
	"testing"
)

func TestPostgresSnapshotScriptUsesConsistentDump(t *testing.T) {
	content, err := os.ReadFile("../../scripts/snapshot-postgres.sh")
	if err != nil {
		t.Fatal(err)
	}
	script := string(content)
	for _, required := range []string{"POSTGRES_DSN", "BACKUP_OUTPUT", "pg_dump", "--serializable-deferrable", "--format=custom"} {
		if !strings.Contains(script, required) {
			t.Fatalf("PostgreSQL snapshot script missing %q", required)
		}
	}
}
