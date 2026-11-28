package plugin_test

import (
	"context"
	"errors"
	"testing"

	"brain/internal/execution"
	"brain/internal/plugin"
)

func TestManifestExecutionBoundary(t *testing.T) {
	for helper, want := range map[string]plugin.ExecutionBoundary{
		"": plugin.BoundaryNeuron, "sdk.dispatch": plugin.BoundaryNeuron,
		"sdk.request": plugin.BoundaryNeuron, "sdk.context": plugin.BoundaryNeuron,
		"sdk.response": plugin.BoundaryNeuron, "sdk.pipeline": plugin.BoundaryNeuron,
		"host.model": plugin.BoundaryStorage, "host.call": plugin.BoundaryDelegation,
		"host.admin": plugin.BoundaryRecovery, "host.unknown": plugin.BoundaryInvalid,
	} {
		if got := (plugin.Route{Helper: helper}).ExecutionBoundary(); got != want {
			t.Errorf("%q: got %q want %q", helper, got, want)
		}
	}
	locations, err := plugin.Discover("../../plugins")
	if err != nil {
		t.Fatal(err)
	}
	for _, location := range locations {
		manifest, err := plugin.LoadManifest(location.ManifestPath)
		if err != nil {
			t.Fatal(err)
		}
		for _, route := range manifest.Routes {
			if route.ExecutionBoundary() == plugin.BoundaryInvalid {
				t.Errorf("unclassified route: %s.%s", manifest.ID, route.Operation)
			}
		}
	}
}

func TestScheduledHandleRejectsHostStorageOperationBeforeDispatch(t *testing.T) {
	manifest := plugin.Manifest{
		ID: "storage-fixture", Version: "1", HostAPI: 1, Declarative: true,
		Capabilities: []string{"read"},
		Models:       map[string]plugin.ModelSpec{"item": {Version: 1, Fields: map[string]string{"name": "string"}}},
		Routes:       []plugin.Route{{Operation: "read", Function: "read", Capability: "read", Helper: "host.model", Model: &plugin.ModelRoute{Name: "item", Action: "read"}}},
	}
	catalog, err := execution.NewCatalog(4, 4096)
	if err != nil {
		t.Fatal(err)
	}
	runtime := execution.Runtime{Kind: plugin.ScheduledTypeScriptKind, Profile: "test"}
	artifact, err := plugin.RegisterScheduledArtifact(catalog, runtime, manifest, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	executor := scheduledExecutor(func(context.Context, execution.Request) (execution.Result, error) {
		t.Error("host storage operation dispatched to Neuron")
		return execution.Result{}, execution.ErrUnavailable
	})
	worker, err := artifact.Worker("alpha", executor)
	if err != nil {
		t.Fatal(err)
	}
	if err := worker.Start(); err != nil {
		t.Fatal(err)
	}
	defer worker.Stop()
	actor := plugin.InvocationActor{Kind: "operator", ID: "root", Authenticated: true, Level: "authenticated"}
	request := plugin.WorkerRequest{
		Protocol: 1, ProjectID: "alpha", PluginID: manifest.ID, Version: manifest.Version,
		Operation: "read", Payload: []byte(`{"id":"one"}`), Project: &plugin.ProjectSnapshot{ID: "alpha", Active: true},
		HostContext: &plugin.InvocationContext{InvocationID: "test", ProjectVerified: true, Auth: plugin.InvocationAuth{Actor: actor, Caller: actor, Permissions: []plugin.InvocationPermission{{Capability: "read", Operation: "read", Resource: "item"}}}},
	}
	if _, err := worker.HandleExecution(request); !errors.Is(err, plugin.ErrInvalidWorkerRequest) {
		t.Fatal("host storage operation was not rejected", err)
	}
}
