package plugin_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"brain/internal/plugin"
)

func TestManifestSandboxWorkerRejectsFailedStartup(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "index.mjs"), []byte("export function health() {}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	launcher := filepath.Join(root, "launcher.sh")
	if err := os.WriteFile(launcher, []byte("#!/bin/sh\nexit 78\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	manifest := plugin.Manifest{ID: "sample", Version: "1.0.0", Entry: "index.mjs", SourceFiles: []string{"index.mjs"}, Capabilities: []string{"health"}, Routes: []plugin.Route{{Operation: "health", Function: "health", Capability: "health"}}}
	worker, err := plugin.NewManifestBubblewrapTypeScriptWorker(launcher, "worker.mjs", filepath.Join(root, "execution.json"), root, "project", manifest, plugin.IsolationPolicy{})
	if err != nil {
		t.Fatal(err)
	}
	if err := worker.Start(); err == nil || worker.Ready() {
		t.Fatal("failed launcher reported readiness")
	}
}

func TestWorkerTimeoutIncludesBlockedWrite(t *testing.T) {
	root := t.TempDir()
	launcher := filepath.Join(root, "worker.sh")
	if err := os.WriteFile(launcher, []byte("#!/bin/sh\nexec sleep 10\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	worker := plugin.NewTypeScriptWorkerWithOptions(plugin.TypeScriptWorkerOptions{InvocationTimeout: 100 * time.Millisecond}, launcher)
	if err := worker.Start(); err != nil {
		t.Fatal(err)
	}
	started := time.Now()
	response := worker.Handle(plugin.WorkerRequest{Protocol: 1, ProjectID: "project", PluginID: "sample", Version: "1.0.0", Operation: "health", Payload: []byte(`"` + strings.Repeat("x", 512<<10) + `"`)})
	if response.ErrorCode != plugin.ErrTypeScriptWorkerTimedOut.Error() || worker.Ready() || time.Since(started) > 3*time.Second {
		t.Fatalf("blocked write was not stopped: %+v", response)
	}
}
