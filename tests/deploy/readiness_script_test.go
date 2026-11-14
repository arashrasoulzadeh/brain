package deploy_test

import (
	"os"
	"strings"
	"testing"
)

func TestReadinessScriptUsesBoundedHealthRetries(t *testing.T) {
	content, err := os.ReadFile("../../scripts/check-api-readiness.sh")
	if err != nil {
		t.Fatal(err)
	}
	script := string(content)
	for _, required := range []string{"curl", "READINESS_ATTEMPTS", "READINESS_DELAY_SECONDS", "/livez"} {
		if !strings.Contains(script, required) {
			t.Fatalf("readiness script missing %q", required)
		}
	}
}
