package plugin_test

import (
	"os"
	"path/filepath"
	"testing"

	"brain/internal/plugin"
)

func TestBubblewrapWorkerUsesExplicitSandboxEnvironment(t *testing.T) {
	directory := t.TempDir()
	launcher := filepath.Join(directory, "launcher.sh")
	if err := os.WriteFile(launcher, []byte("#!/bin/sh\nif [ -n \"$SECRET\" ]; then exit 1; fi\nprintf '%s\\n' '{\"jsonrpc\":\"2.0\",\"id\":\"health\",\"result\":{\"protocol\":1,\"success\":true}}'\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	worker := plugin.NewBubblewrapTypeScriptWorker(launcher, "/plugin/worker.mjs", "/plugin/execution.json", "/plugin", plugin.IsolationPolicy{FilesystemReadOnly: true, CPUMillis: 1000, MemoryBytes: 64 << 20, ProcessLimit: 8})
	if err := worker.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = worker.Stop() }()
	response, err := plugin.Invoke(worker, plugin.WorkerRequest{Protocol: 1, ProjectID: "project", PluginID: "sample", Version: "1.0.0", Operation: "health"}, 1)
	if err != nil {
		t.Fatal(err)
	}
	if !response.Success {
		t.Fatalf("sandbox launcher worker failed: %+v", response)
	}
}
