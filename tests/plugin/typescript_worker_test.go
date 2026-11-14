package plugin_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

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
	response, err := plugin.InvokeForProject(worker, plugin.WorkerRequest{Protocol: 1, ProjectID: "project", PluginID: "sample", Version: "1.0.0", Operation: "health"}, 1, "project")
	if err != nil {
		t.Fatal(err)
	}
	if !response.Success || response.Protocol != 1 {
		t.Fatalf("unexpected worker response: %+v", response)
	}
}

func TestTypeScriptWorkerStopsHungInvocation(t *testing.T) {
	directory := t.TempDir()
	script := filepath.Join(directory, "worker.sh")
	if err := os.WriteFile(script, []byte("#!/bin/sh\nsleep 2\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	worker := plugin.NewTypeScriptWorkerWithOptions(plugin.TypeScriptWorkerOptions{InvocationTimeout: 10 * time.Millisecond}, script)
	if err := worker.Start(); err != nil {
		t.Fatal(err)
	}
	response, err := plugin.InvokeForProject(worker, plugin.WorkerRequest{Protocol: 1, ProjectID: "project", PluginID: "sample", Version: "1.0.0", Operation: "health"}, 1, "project")
	if err != nil {
		t.Fatal(err)
	}
	if response.ErrorCode != "typescript worker invocation timed out" || worker.Ready() {
		t.Fatalf("hung worker was not stopped: %+v", response)
	}
}

func TestTypeScriptWorkerRejectsRequestOutsideHostProject(t *testing.T) {
	worker := plugin.NewTypeScriptWorkerWithOptions(plugin.TypeScriptWorkerOptions{ProjectID: "host-project"}, "true")
	if err := worker.Start(); err != nil {
		t.Fatal(err)
	}
	response := worker.Handle(plugin.WorkerRequest{Protocol: 1, ProjectID: "client-project", PluginID: "sample", Version: "1.0.0", Operation: "health"})
	if response.ErrorCode != plugin.ErrInvalidWorkerRequest.Error() {
		t.Fatalf("cross-project request reached worker: %+v", response)
	}
	_ = worker.Stop()
}
