package plugin_test

import (
	"errors"
	"testing"

	"brain/internal/plugin"
)

type testStoragePlanner struct {
	prepare func(plugin.WorkerRequest) (plugin.WorkerRequest, func() error, error)
}

func (p testStoragePlanner) Supports(id string) bool { return id == "sample" }
func (p testStoragePlanner) Prepare(request plugin.WorkerRequest) (plugin.WorkerRequest, func() error, error) {
	return p.prepare(request)
}

func TestRegistryUsesInjectableStoragePlannerWithoutSQLite(t *testing.T) {
	registry := plugin.NewRegistry(1)
	manifest := plugin.Manifest{
		ID: "sample", Version: "1.0.0", HostAPI: 1, Capabilities: []string{"write"},
		Routes: []plugin.Route{{Operation: "write", Function: "write", Capability: "write"}},
	}
	if err := registry.Register(manifest); err != nil {
		t.Fatal(err)
	}
	if err := registry.Install("project", manifest.ID, manifest.Version); err != nil {
		t.Fatal(err)
	}
	if err := registry.SetReadiness(manifest.ID, manifest.Version, true); err != nil {
		t.Fatal(err)
	}
	if err := registry.Grant("project", manifest.ID, manifest.Version, "write"); err != nil {
		t.Fatal(err)
	}
	if err := registry.Enable("project", manifest.ID, manifest.Version); err != nil {
		t.Fatal(err)
	}
	worker := &countingWorker{}
	if err := registry.BindWorker("project", manifest.ID, manifest.Version, worker); err != nil {
		t.Fatal(err)
	}
	commits := 0
	registry.BindInvocationStorage(testStoragePlanner{prepare: func(request plugin.WorkerRequest) (plugin.WorkerRequest, func() error, error) {
		return request, func() error { commits++; return nil }, nil
	}})
	request := plugin.WorkerRequest{Protocol: 1, ProjectID: "project", PluginID: manifest.ID, Version: manifest.Version, Operation: "write"}
	if response, err := registry.Invoke(request); err != nil || !response.Success || commits != 1 {
		t.Fatalf("invoke = %#v, %v, commits = %d", response, err, commits)
	}
	registry.BindInvocationStorage(testStoragePlanner{prepare: func(request plugin.WorkerRequest) (plugin.WorkerRequest, func() error, error) {
		request.ProjectID = "other-project"
		return request, func() error { commits++; return nil }, nil
	}})
	if _, err := registry.Invoke(request); !errors.Is(err, plugin.ErrInvalidWorkerRequest) {
		t.Fatalf("adapter identity mismatch error = %v", err)
	}
	if worker.calls != 1 || commits != 1 {
		t.Fatalf("identity mismatch reached worker or storage: calls = %d, commits = %d", worker.calls, commits)
	}
}
