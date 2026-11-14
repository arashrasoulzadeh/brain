package plugin_test

import (
	"os"
	"path/filepath"
	"testing"

	"brain/internal/plugin"
)

func TestEveryPluginManifestAuthorizesItsSources(t *testing.T) {
	entries, err := os.ReadDir("../../plugins")
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if !entry.IsDir() || entry.Name() == "sdk-typescript" {
			continue
		}
		pluginDirectory := filepath.Join("../../plugins", entry.Name())
		manifestPath := filepath.Join(pluginDirectory, "plugin.yaml")
		if _, err := os.Stat(manifestPath); err != nil {
			t.Errorf("plugin %q has no manifest: %v", entry.Name(), err)
			continue
		}
		manifest, err := plugin.LoadManifest(manifestPath)
		if err != nil {
			t.Errorf("plugin %q manifest: %v", entry.Name(), err)
			continue
		}
		if err := plugin.ValidatePluginSources(pluginDirectory, manifest); err != nil {
			t.Errorf("plugin %q sources: %v", entry.Name(), err)
		}
	}
}
