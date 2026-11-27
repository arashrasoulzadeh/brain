package plugin

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestMacOSProfileConstructionCoverage(t *testing.T) {
	root := t.TempDir()
	root, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatal(err)
	}
	scripts := filepath.Join(root, "scripts")
	pluginRoot := filepath.Join(root, "plugins", "sample")
	bin := filepath.Join(root, "bin")
	for _, directory := range []string{scripts, pluginRoot, bin, filepath.Join(root, "node_modules", "@esbuild", "darwin-arm64", "bin")} {
		if err := os.MkdirAll(directory, 0o700); err != nil {
			t.Fatal(err)
		}
	}
	workerScript := filepath.Join(scripts, "worker.mjs")
	node := filepath.Join(bin, "node")
	descriptor := filepath.Join(root, "descriptor.json")
	esbuild := filepath.Join(root, "node_modules", "@esbuild", "darwin-arm64", "bin", "esbuild")
	for _, file := range []string{workerScript, node, esbuild, filepath.Join(pluginRoot, "index.ts")} {
		if err := os.WriteFile(file, []byte("# fixture"), 0o700); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("PATH", bin)
	manifest := Manifest{ID: "sample", Version: "1.0.0", HostAPI: 1, Entry: "index.ts", SourceFiles: []string{"index.ts"}}
	policy := IsolationPolicy{FilesystemReadOnly: true, CPUMillis: 30000, MemoryBytes: 2 << 30, ProcessLimit: 512, WallTime: time.Second}
	construct := func(platform, script, desc, directory string, m Manifest, p IsolationPolicy) (*TypeScriptWorker, error) {
		return newManifestMacOSTypeScriptWorker(platform, "/bin/false", script, desc, directory, "coverage", m, p)
	}
	if _, err := construct("linux", workerScript, descriptor, pluginRoot, manifest, policy); !errors.Is(err, ErrInvalidIsolationPolicy) {
		t.Fatal(err)
	}
	if _, err := construct("darwin", workerScript, descriptor, pluginRoot, manifest, IsolationPolicy{}); !errors.Is(err, ErrInvalidIsolationPolicy) {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		script, descriptor, directory string
	}{
		{filepath.Join(root, "missing-worker"), descriptor, pluginRoot},
		{workerScript, descriptor, filepath.Join(root, "missing-plugin")},
		{workerScript, root, pluginRoot},
	} {
		if _, err := construct("darwin", tc.script, tc.descriptor, tc.directory, manifest, policy); err == nil {
			t.Fatalf("invalid path accepted: %+v", tc)
		}
	}
	worker, err := construct("darwin", workerScript, descriptor, pluginRoot, manifest, policy)
	if err != nil {
		t.Fatal(err)
	}
	profile := ""
	for _, value := range worker.env {
		if strings.HasPrefix(value, "BRAIN_MACOS_PROFILE=") {
			profile = strings.TrimPrefix(value, "BRAIN_MACOS_PROFILE=")
		}
	}
	for _, required := range []string{"(deny default)", filepath.Join(pluginRoot, "index.ts"), node, descriptor, esbuild} {
		if !strings.Contains(profile, required) {
			t.Fatalf("missing sandbox rule %q", required)
		}
	}
	for _, required := range []string{
		`(allow file-read* (literal "/opt/homebrew/etc/openssl@3/openssl.cnf") (literal "/usr/local/etc/openssl@3/openssl.cnf"))`,
	} {
		if !strings.Contains(profile, required) {
			t.Fatalf("missing Homebrew OpenSSL read-only rule %q", required)
		}
	}
	if strings.Contains(profile, `(subpath "/opt/homebrew/etc`) || strings.Contains(profile, `(subpath "/usr/local/etc`) {
		t.Fatal("profile grants broad Homebrew configuration access")
	}
	if strings.Contains(profile, "(allow network") || strings.Contains(profile, "(allow default)") {
		t.Fatal("profile grants network or default access")
	}
	if err := os.Remove(esbuild); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(node, esbuild); err != nil {
		t.Fatal(err)
	}
	if _, err := construct("darwin", workerScript, descriptor, pluginRoot, manifest, policy); !errors.Is(err, ErrInvalidManifest) {
		t.Fatalf("escaped esbuild path accepted: %v", err)
	}
	t.Setenv("PATH", t.TempDir())
	if _, err := construct("darwin", workerScript, descriptor, pluginRoot, manifest, policy); err == nil {
		t.Fatal("missing node accepted")
	}
}

func TestMacOSWorkerPublicConstructorRejectsUnsupportedPlatform(t *testing.T) {
	if runtime.GOOS == "darwin" {
		t.Skip("non-Darwin rejection")
	}
	if _, err := NewManifestMacOSTypeScriptWorker("", "", "", "", "alpha", Manifest{}, IsolationPolicy{}); !errors.Is(err, ErrInvalidIsolationPolicy) {
		t.Fatalf("unexpected constructor error: %v", err)
	}
}
