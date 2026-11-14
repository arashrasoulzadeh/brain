package deploy_test

import (
	"os"
	"strings"
	"testing"
)

func TestIncompatibleMigrationRecoveryDocumentDefinesSafePaths(t *testing.T) {
	content, err := os.ReadFile("../../docs/incompatible-migration-recovery.md")
	if err != nil {
		t.Fatal(err)
	}
	document := string(content)
	for _, required := range []string{"Compatible schema", "verified backup", "keep traffic stopped", "--no-build"} {
		if !strings.Contains(document, required) {
			t.Fatalf("recovery document missing %q", required)
		}
	}
}
