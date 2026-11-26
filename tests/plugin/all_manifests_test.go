package plugin_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
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
		declaredSources := map[string]bool{}
		for _, source := range manifest.SourceFiles {
			declaredSources[source] = true
		}
		if err := filepath.WalkDir(pluginDirectory, func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".ts") {
				return nil
			}
			relative, err := filepath.Rel(pluginDirectory, path)
			if err != nil {
				return err
			}
			if !declaredSources[filepath.ToSlash(relative)] {
				t.Errorf("plugin %q has undeclared TypeScript source %q", manifest.ID, relative)
			}
			return nil
		}); err != nil {
			t.Errorf("plugin %q source inventory: %v", entry.Name(), err)
		}
		if strings.TrimSpace(manifest.Title) == "" || strings.TrimSpace(manifest.Description) == "" {
			t.Errorf("plugin %q must declare title and description", entry.Name())
		}
		capabilities := map[string]bool{}
		for _, capability := range manifest.Capabilities {
			capabilities[capability] = true
		}
		for _, route := range manifest.Routes {
			if !capabilities[route.Capability] {
				t.Errorf("plugin %q route %q lacks declared capability %q", entry.Name(), route.Operation, route.Capability)
			}
		}
	}
}
