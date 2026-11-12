package plugin_test

import (
	"os"
	"path/filepath"
	"testing"

	"brain/internal/plugin"
)

func TestDiscoverFindsFeaturePluginsWithoutHardcodingThem(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "example"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "example", "plugin.yaml"), []byte("id: example\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	locations, err := plugin.Discover(root)
	if err != nil || len(locations) != 1 || locations[0].ID != "example" {
		t.Fatalf("Discover() = %#v, %v", locations, err)
	}
}
