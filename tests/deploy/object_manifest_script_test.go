package deploy_test

import (
	"os"
	"strings"
	"testing"
)

func TestObjectManifestVerificationScriptUsesHostValidator(t *testing.T) {
	content, err := os.ReadFile("../../scripts/verify-object-manifest.sh")
	if err != nil {
		t.Fatal(err)
	}
	script := string(content)
	for _, required := range []string{"OBJECT_MANIFEST", "BRAIN_BIN", "verify-object-manifest"} {
		if !strings.Contains(script, required) {
			t.Fatalf("object manifest verification script missing %q", required)
		}
	}
}
