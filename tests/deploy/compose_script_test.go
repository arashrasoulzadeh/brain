package deploy_test

import (
	"os"
	"strings"
	"testing"
)

func TestComposeDeploymentScriptUsesNoBuild(t *testing.T) {
	content, err := os.ReadFile("../../scripts/start-compose-no-build.sh")
	if err != nil {
		t.Fatal(err)
	}
	script := string(content)
	if !strings.Contains(script, "up -d --no-build") {
		t.Fatal("deployment script does not use Compose --no-build")
	}
	if strings.Contains(script, "up -d --build") {
		t.Fatal("deployment script rebuilds images")
	}
}
