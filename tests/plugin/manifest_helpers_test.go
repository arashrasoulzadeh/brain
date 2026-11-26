package plugin_test

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"brain/internal/plugin"
)

func TestManifestRouteHelperAllowlist(t *testing.T) {
	manifest := plugin.Manifest{
		SourceFiles:  []string{"index.ts"},
		Capabilities: []string{"sample.read"},
		Routes: []plugin.Route{
			{Operation: "sample.read", Function: "read", Capability: "sample.read", Helper: "sdk.dispatch"},
		},
	}
	if err := plugin.ValidateManifestContract(manifest); err != nil {
		t.Fatalf("declared SDK helper rejected: %v", err)
	}
	manifest.Routes[0].Helper = "sdk.request"
	if err := plugin.ValidateManifestContract(manifest); err != nil {
		t.Fatalf("declared request helper rejected: %v", err)
	}
	manifest.Routes[0].Helper = "arbitrary.module"
	if err := plugin.ValidateManifestContract(manifest); !errors.Is(err, plugin.ErrInvalidManifest) {
		t.Fatalf("unknown helper accepted: %v", err)
	}
}

func TestManifestRouteHelperReachesWorkerDescriptor(t *testing.T) {
	directory := t.TempDir()
	manifestPath := filepath.Join(directory, "plugin.yaml")
	contents := "id: sample\nversion: 1.0.0\nsource_files:\n  - index.ts\ncapabilities:\n  - sample.read\nroutes:\n  - operation: sample.read\n    function: read\n    capability: sample.read\n    helper: sdk.request\n"
	if err := os.WriteFile(manifestPath, []byte(contents), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, "index.ts"), []byte("export function dispatchRequest() {}\n"), 0600); err != nil {
		t.Fatal(err)
	}
	manifest, err := plugin.LoadManifest(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	descriptorPath := filepath.Join(directory, "descriptor.json")
	if _, err := plugin.NewManifestTypeScriptWorker("node", "worker.mjs", descriptorPath, directory, "project", manifest, plugin.IsolationPolicy{}); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(descriptorPath)
	if err != nil {
		t.Fatal(err)
	}
	var descriptor struct {
		Routes []struct {
			Helper string `json:"helper"`
		} `json:"routes"`
	}
	if err := json.Unmarshal(data, &descriptor); err != nil {
		t.Fatal(err)
	}
	if len(descriptor.Routes) != 1 || descriptor.Routes[0].Helper != "sdk.request" {
		t.Fatalf("helper missing from descriptor: %s", data)
	}
}
