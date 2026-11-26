package plugin_test

import (
	"os"
	"testing"

	"brain/internal/plugin"
)

func TestGeneratedManifestSDKOperationsAreCurrent(t *testing.T) {
	expected, err := plugin.ManifestSDKTypes("../../plugins")
	if err != nil {
		t.Fatal(err)
	}
	actual, err := os.ReadFile("../../plugins/sdk-typescript/src/manifest-operations.generated.ts")
	if err != nil {
		t.Fatal(err)
	}
	if string(actual) != expected {
		t.Fatal("manifest SDK types are stale; run go run ./cmd/manifest-sdk")
	}
}
