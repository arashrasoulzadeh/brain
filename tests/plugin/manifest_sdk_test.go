package plugin_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"brain/internal/plugin"
)

func TestGeneratedManifestSDKOperationsAreCurrent(t *testing.T) {
	expected, err := plugin.ManifestSDKTypes("../../plugins")
	if err != nil {
		t.Fatal(err)
	}
	actual, err := os.ReadFile("../../plugins/sdk-typescript/src/manifest-operations.generated.d.ts")
	if err != nil {
		t.Fatal(err)
	}
	if string(actual) != expected {
		t.Fatal("manifest SDK types are stale; run go run ./cmd/manifest-sdk")
	}
}

func TestManifestSDKTypesRejectsInvalidManifests(t *testing.T) {
	root := t.TempDir()
	if _, err := plugin.ManifestSDKTypes(filepath.Join(root, "missing")); err == nil {
		t.Fatal("missing root accepted")
	}
	dir := filepath.Join(root, "sample")
	if err := os.Mkdir(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "plugin.yaml")
	if err := os.WriteFile(path, []byte("["), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := plugin.ManifestSDKTypes(root); err == nil {
		t.Fatal("invalid manifest accepted")
	}
	if err := os.WriteFile(path, []byte("id: sample\ndeclarative: true\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	source, err := plugin.ManifestSDKTypes(root)
	if err != nil || !strings.Contains(source, `"sample": never;`) {
		t.Fatalf("empty routes: %s (%v)", source, err)
	}
}
