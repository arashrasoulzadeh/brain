package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"brain/internal/plugin"
)

func startupRuntimeFixture(t *testing.T, launcher string, loader bool) {
	t.Helper()
	startupCoverageDir(t)
	bin := filepath.Join(t.TempDir(), "bin")
	if err := os.MkdirAll(bin, 0o700); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"bwrap", "node"} {
		if err := os.WriteFile(filepath.Join(bin, name), []byte("#!/bin/sh\nexit 0\n"), 0o700); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("PATH", bin)
	if loader {
		if err := os.MkdirAll("node_modules/tsx/dist", 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile("node_modules/tsx/dist/loader.mjs", nil, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.MkdirAll("scripts", 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("scripts/run-sandboxed-typescript-worker.sh", []byte(launcher), 0o700); err != nil {
		t.Fatal(err)
	}
}

const startupProbeFixture = `#!/bin/sh
printf '%s\n' '{"ready":true}'
while IFS= read -r line; do
  printf '%s\n' '{"result":{"protocol":1,"success":true,"payload":{"accepted":true}}}'
done
`

func TestStartupRuntimeCoverageFailure(t *testing.T) {
	for _, tc := range []struct {
		name, launcher string
		loader         bool
		want           string
	}{
		{"missing loader", startupProbeFixture, false, "sandbox TypeScript loader unavailable"},
		{"startup rejection", "#!/bin/sh\nexit 1\n", true, "sandbox startup probe failed"},
		{"invocation rejection", strings.Replace(startupProbeFixture, `"accepted":true`, `"accepted":false`, 1), true, "sandbox invocation probe failed"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			startupRuntimeFixture(t, tc.launcher, tc.loader)
			if err := bindSandboxedTypeScriptWorkers(plugin.NewRegistry(1), nil); err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("got %v; want %q", err, tc.want)
			}
		})
	}
	t.Run("temporary directory unavailable", func(t *testing.T) {
		startupRuntimeFixture(t, startupProbeFixture, true)
		t.Setenv("TMPDIR", filepath.Join(t.TempDir(), "missing"))
		if err := bindSandboxedTypeScriptWorkers(plugin.NewRegistry(1), nil); err == nil {
			t.Fatal("startup accepted unavailable temporary directory")
		}
	})
}

func TestProductionStartupStorageFailure(t *testing.T) {
	for _, kind := range []string{"missing manifests", "missing runtime", "invalid parent", "database directory"} {
		t.Run(kind, func(t *testing.T) {
			startupRuntimeFixture(t, startupProbeFixture, true)
			t.Setenv("OPERATOR_USERNAME", "admin")
			t.Setenv("OPERATOR_SECRET", "test-secret")
			t.Setenv("RUNTIME_DSN", "")
			t.Setenv("MIGRATION_DSN", "")
			path := filepath.Join(t.TempDir(), "projects.db")
			want := "plugin discovery unavailable"
			if kind != "missing manifests" {
				startupCoverageManifest(t, "sample", "id: sample\nversion: 1.0.0\nhost_api: 1\ndeclarative: true\n")
				want = "sandbox runtime requires bwrap"
				if kind == "missing runtime" {
					t.Setenv("PATH", t.TempDir())
				} else if kind == "invalid parent" {
					if err := os.WriteFile(path, nil, 0o600); err != nil {
						t.Fatal(err)
					}
					path = filepath.Join(path, "child.db")
					want = "project storage directory unavailable"
				} else {
					if err := os.Mkdir(path, 0o700); err != nil {
						t.Fatal(err)
					}
					want = "project storage unavailable"
				}
			}
			t.Setenv("PROJECTS_FILE", path)
			if _, err := productionServer(); err == nil || !strings.Contains(err.Error(), want) {
				t.Fatalf("got %v; want %q", err, want)
			}
		})
	}
}
