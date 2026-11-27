package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"brain/internal/plugin"
)

func startupCoverageDir(t *testing.T) {
	t.Helper()
	original, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	if err := os.Chdir(directory); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(original); err != nil {
			t.Error(err)
		}
	})
}

func startupCoverageManifest(t *testing.T, directory, body string) {
	t.Helper()
	root := filepath.Join("plugins", directory)
	if err := os.MkdirAll(root, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "plugin.yaml"), []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestStartupCoverageDiscoveryErrors(t *testing.T) {
	const valid = "id: sample\nversion: 1.0.0\nhost_api: 1\ndeclarative: true\n"
	for _, tc := range []struct {
		name      string
		manifests []string
		want      string
	}{
		{"missing root", nil, "plugin discovery unavailable"},
		{"empty root", []string{}, "no plugins discovered"},
		{"invalid yaml", []string{"["}, "plugin a rejected"},
		{"missing source", []string{"id: sample\nversion: 1.0.0\nhost_api: 1\nentrypoint: index.ts\nsource_files: [index.ts]\n"}, "sources rejected"},
		{"duplicate", []string{valid, valid}, "duplicate plugin"},
		{"incompatible", []string{strings.Replace(valid, "host_api: 1", "host_api: 2", 1)}, "plugin host API incompatible"},
		{"missing dependency", []string{valid + "dependencies: [missing@1.0.0]\n"}, "plugin dependency missing"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			startupCoverageDir(t)
			if tc.manifests != nil {
				if err := os.MkdirAll("plugins", 0o700); err != nil {
					t.Fatal(err)
				}
			}
			for i, manifest := range tc.manifests {
				startupCoverageManifest(t, string(rune('a'+i)), manifest)
			}
			if _, err := registerDiscoveredPlugins(plugin.NewRegistry(1)); err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("got %v; want %q", err, tc.want)
			}
		})
	}
}

func TestStartupCoverageMissingSandboxRuntime(t *testing.T) {
	startupCoverageDir(t)
	t.Setenv("PATH", t.TempDir())
	if err := bindSandboxedTypeScriptWorkers(plugin.NewRegistry(1), nil); err == nil || !strings.Contains(err.Error(), "sandbox runtime requires bwrap") {
		t.Fatalf("missing sandbox runtime: %v", err)
	}
}

func TestStartupCoverageRejectedOperatorBootstrap(t *testing.T) {
	t.Setenv("OPERATOR_USERNAME", " ")
	t.Setenv("OPERATOR_SECRET", "test-secret")
	t.Setenv("RUNTIME_DSN", "")
	t.Setenv("MIGRATION_DSN", "")
	t.Setenv("PROJECTS_FILE", filepath.Join(t.TempDir(), "projects.db"))
	if _, err := productionServer(); err == nil || !strings.Contains(err.Error(), "operator bootstrap failed") {
		t.Fatalf("invalid bootstrap: %v", err)
	}
}
