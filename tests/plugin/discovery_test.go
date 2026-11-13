package plugin_test

import (
	"os"
	"path/filepath"
	"strings"
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

func TestLoadManifestDiscoversRoutesAndMetadata(t *testing.T) {
	path := filepath.Join(t.TempDir(), "plugin.yaml")
	if err := os.WriteFile(path, []byte("id: demo\ntitle: Demo\ndescription: test\nversion: 1.0.0\nhost_api: 1\nentrypoint: index.ts\nsource_files:\n  - index.ts\nimports: []\nplugin_imports: []\nroutes:\n  - operation: demo.run\n    function: run\n    capability: demo.run\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	m, err := plugin.LoadManifest(path)
	if err != nil {
		t.Fatal(err)
	}
	route, ok := m.Route("demo.run")
	if !ok || m.Title != "Demo" || route.Function != "run" || route.Capability != "demo.run" {
		t.Fatalf("manifest = %#v route = %#v", m, route)
	}
	if doc := plugin.Documentation(m); !strings.Contains(doc, "# Demo") || !strings.Contains(doc, "demo.run") {
		t.Fatalf("documentation = %q", doc)
	}
}

func TestRepositoryPluginManifestsDeclareSourceAndImportGates(t *testing.T) {
	paths, err := filepath.Glob("../../plugins/*/plugin.yaml")
	if err != nil || len(paths) == 0 {
		t.Fatalf("plugin manifests = %v, %v", paths, err)
	}
	for _, manifestPath := range paths {
		manifest, loadErr := plugin.LoadManifest(manifestPath)
		if loadErr != nil {
			t.Fatalf("LoadManifest(%s) = %v", manifestPath, loadErr)
		}
		if len(manifest.SourceFiles) == 0 || len(manifest.Imports) == 0 {
			t.Fatalf("manifest %s does not declare source and import gates", manifestPath)
		}
	}
}
