package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"brain/internal/plugin"
)

func startupFactoryRegistry(t *testing.T) *plugin.Registry {
	t.Helper()
	r := plugin.NewRegistry(1)
	m := plugin.Manifest{ID: "sample", Version: "1.0.0", HostAPI: 1, Declarative: true}
	for _, err := range []error{r.Register(m), r.Install("coverage", m.ID, m.Version), r.SetReadiness(m.ID, m.Version, true), r.Enable("coverage", m.ID, m.Version)} {
		if err != nil {
			t.Fatal(err)
		}
	}
	return r
}

func TestStartupWorkerFactoryFailures(t *testing.T) {
	for _, kind := range []string{"unknown location", "missing manifest", "missing root", "invalid source", "temporary directory unavailable"} {
		t.Run(kind, func(t *testing.T) {
			startupRuntimeFixture(t, startupProbeFixture, true)
			r := startupFactoryRegistry(t)
			startupCoverageManifest(t, "sample", "id: sample\nversion: 1.0.0\nhost_api: 1\ndeclarative: true\n")
			location := plugin.PluginLocation{ID: "sample", Directory: "plugins/sample", ManifestPath: "plugins/sample/plugin.yaml"}
			locations := map[string]plugin.PluginLocation{"sample@1.0.0": location}
			switch kind {
			case "unknown location":
				delete(locations, "sample@1.0.0")
			case "missing manifest":
				location.ManifestPath = "missing.yaml"
				locations["sample@1.0.0"] = location
			case "missing root":
				location.Directory = "missing"
				locations["sample@1.0.0"] = location
			case "invalid source":
				body := "id: sample\nversion: 1.0.0\nhost_api: 1\nentrypoint: missing.ts\nsource_files: [missing.ts]\n"
				if err := os.WriteFile(location.ManifestPath, []byte(body), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			if err := bindSandboxedTypeScriptWorkers(r, locations); err != nil {
				t.Fatal(err)
			}
			if kind == "temporary directory unavailable" {
				t.Setenv("TMPDIR", filepath.Join(t.TempDir(), "missing"))
			}
			if err := r.StartWorker("coverage", "sample", "1.0.0"); err == nil {
				t.Fatalf("%s accepted", kind)
			} else if kind == "unknown location" && !strings.Contains(err.Error(), "plugin location unavailable") {
				t.Fatal(err)
			}
		})
	}
}
