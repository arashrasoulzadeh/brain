package plugin_test

import (
	"os"
	"path/filepath"
	"testing"

	"brain/internal/plugin"
)

func TestTypeScriptWorkerExecutesThroughHostProcess(t *testing.T) {
	directory := t.TempDir()
	script := filepath.Join(directory, "worker.sh")
	if err := os.WriteFile(script, []byte("#!/bin/sh\nwhile IFS= read -r line; do printf '%s\\n' '{\"jsonrpc\":\"2.0\",\"id\":\"invoke\",\"result\":{\"protocol\":1,\"success\":true}}'; done\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	worker := plugin.NewTypeScriptWorker(script)
	if err := worker.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = worker.Stop() }()
	response, err := plugin.Invoke(worker, plugin.WorkerRequest{Protocol: 1, ProjectID: "project", PluginID: "sample", Version: "1.0.0", Operation: "health"}, 1)
	if err != nil {
		t.Fatal(err)
	}
	if !response.Success || response.Protocol != 1 {
		t.Fatalf("unexpected worker response: %+v", response)
	}
}
