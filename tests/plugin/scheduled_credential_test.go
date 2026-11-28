package plugin_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"brain/internal/cortex"
	"brain/internal/execution"
	"brain/internal/plugin"
)

func TestCredentialInvocationCannotBypassCortex(t *testing.T) {
	manifest := plugin.Manifest{
		ID: "credential-fixture", Version: "1.0.0", HostAPI: 1, Entry: "index.ts",
		SourceFiles: []string{"index.ts"}, Capabilities: []string{"write"},
		Routes: []plugin.Route{{Operation: "write", Function: "write", Capability: "write", ResourceField: "record"}},
	}
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "index.ts"), []byte("export function write() { return {}; }"), 0o600); err != nil {
		t.Fatal(err)
	}
	catalog, err := execution.NewCatalog(8, 1<<20)
	if err != nil {
		t.Fatal(err)
	}
	runtime := execution.Runtime{Kind: plugin.ScheduledTypeScriptKind, Profile: "test"}
	artifact, err := plugin.RegisterScheduledArtifact(catalog, runtime, manifest, root)
	if err != nil {
		t.Fatal(err)
	}
	scheduler, err := cortex.New(catalog, cortex.Config{MaxNeurons: 4, MaxProjectRunning: 2, HeartbeatTTL: time.Minute})
	if err != nil {
		t.Fatal(err)
	}
	registry := plugin.NewRegistry(1)
	factory := func(project, id, version string) (*plugin.ScheduledWorker, error) {
		return artifact.Worker(project, scheduler)
	}
	if err := registry.BindScheduledWorkerFactory(factory); err != nil {
		t.Fatal(err)
	}
	if err := registry.BindScheduledWorkerFactory(factory); err == nil {
		t.Fatal("scheduled factory replaced")
	}
	if err := registry.BindWorkerFactory(func(string, string, string) (plugin.ContractWorker, error) {
		return &countingWorker{}, nil
	}); err == nil {
		t.Fatal("legacy factory replaced Cortex")
	}
	planner := &cancellationPlanner{}
	registry.BindInvocationStorage(planner)
	for _, err := range []error{
		registry.Register(manifest), registry.Install("alpha", manifest.ID, manifest.Version),
		registry.SetReadiness(manifest.ID, manifest.Version, true),
		registry.GrantResource("alpha", manifest.ID, manifest.Version, "write", "one"),
		registry.EnableWithDependencies("alpha", manifest.ID, manifest.Version, nil),
	} {
		if err != nil {
			t.Fatal(err)
		}
	}
	legacy := &countingWorker{}
	if err := registry.BindWorker("alpha", manifest.ID, manifest.Version, legacy); err == nil {
		t.Fatal("legacy worker bound in scheduled mode")
	}
	wrong, err := artifact.Worker("beta", scheduler)
	if err != nil {
		t.Fatal(err)
	}
	if err := registry.BindWorker("alpha", manifest.ID, manifest.Version, wrong); err == nil {
		t.Fatal("wrong project handle bound")
	}

	allowed := true
	broker := &plugin.Broker{Registry: registry, ActorAllows: func(string, string, string) bool { return allowed }}
	now := time.Now()
	call := plugin.HostCall{ProjectID: "alpha", ActorID: "alice", PluginID: manifest.ID, Version: manifest.Version, Capability: "write", Operation: "write", Resource: "one"}
	request := plugin.WorkerRequest{Protocol: 1, ProjectID: "alpha", PluginID: manifest.ID, Version: manifest.Version, Operation: "write", Payload: []byte(`{"record":"one"}`)}
	issue := func() plugin.InvocationCredential {
		t.Helper()
		credential, err := broker.Issue(call, now, time.Minute)
		if err != nil {
			t.Fatal(err)
		}
		return credential
	}
	credential := issue()
	if _, err := plugin.InvokeWithCredential(broker, credential, legacy, request, 1, now); !errors.Is(err, plugin.ErrWorkerNotReady) {
		t.Fatalf("zero Neurons: %v", err)
	}
	if _, err := plugin.InvokeWithCredential(broker, credential, legacy, request, 1, now); !errors.Is(err, plugin.ErrInvocationExpired) {
		t.Fatalf("credential reused after capacity failure: %v", err)
	}

	calls := 0
	revoke := false
	session, err := scheduler.Register(cortex.Registration{ID: "neuron-1", Slots: 1, Runtimes: []execution.Runtime{runtime}, Executor: scheduledExecutor(func(_ context.Context, req execution.Request) (execution.Result, error) {
		calls++
		if req.ProjectID != "alpha" || req.Auth.OriginalCaller != "user:alice" || req.Auth.Caller != "user:alice" || req.Auth.Resource != "one" {
			t.Errorf("incorrect credential scope: %+v", req.Auth)
		}
		if revoke {
			allowed = false
		}
		return execution.Result{ID: req.ID, WorkerID: "neuron-1", Status: execution.Succeeded, Output: []byte("{}")}, nil
	})})
	if err != nil {
		t.Fatal(err)
	}
	defer session.Disconnect()
	credential = issue()
	tampered := request
	tampered.Payload = []byte(`{"record":"other"}`)
	if _, err := plugin.InvokeWithCredential(broker, credential, legacy, tampered, 1, now); !errors.Is(err, plugin.ErrInvocationCredentialMismatch) {
		t.Fatalf("resource substitution: %v", err)
	}
	if response, err := plugin.InvokeWithCredential(broker, credential, legacy, request, 1, now); err != nil || !response.Success {
		t.Fatalf("scheduled credential invocation: %+v, %v", response, err)
	}
	revoke = true
	credential = issue()
	if _, err := plugin.InvokeWithCredential(broker, credential, legacy, request, 1, now); !errors.Is(err, plugin.ErrActorDenied) {
		t.Fatalf("revoked credential committed: %v", err)
	}
	if calls != 2 || legacy.calls != 0 || planner.commits.Load() != 1 {
		t.Fatalf("incorrect dispatch/commits: Cortex=%d legacy=%d commits=%d", calls, legacy.calls, planner.commits.Load())
	}
}
