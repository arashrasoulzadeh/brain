package plugin_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"brain/internal/cortex"
	"brain/internal/execution"
	"brain/internal/plugin"
)

type cancellationPlanner struct {
	commits atomic.Int32
}

func (*cancellationPlanner) Supports(string) bool { return true }

func (p *cancellationPlanner) Prepare(request plugin.WorkerRequest) (plugin.WorkerRequest, func() error, error) {
	return request, func() error {
		p.commits.Add(1)
		return nil
	}, nil
}

func TestScheduledHTTPRouteCancellationPreventsCommit(t *testing.T) {
	manifest := plugin.Manifest{
		ID: "cancel-fixture", Version: "1.0.0", HostAPI: 1, Entry: "index.ts",
		SourceFiles: []string{"index.ts"}, Mount: "/cancel-fixture",
		Capabilities: []string{"write"},
		Routes: []plugin.Route{{Operation: "write", Function: "write", Capability: "write", Method: "POST", Path: "/", Auth: "operator"}},
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
	planner := &cancellationPlanner{}
	registry.BindInvocationStorage(planner)
	registry.BindWorkerFactory(func(project, id, version string) (plugin.ContractWorker, error) {
		return artifact.Worker(project, scheduler)
	})
	for _, err := range []error{
		registry.Register(manifest), registry.Install("alpha", manifest.ID, manifest.Version),
		registry.SetReadiness(manifest.ID, manifest.Version, true),
		registry.Grant("alpha", manifest.ID, manifest.Version, "write"),
		registry.EnableWithDependencies("alpha", manifest.ID, manifest.Version, nil),
	} {
		if err != nil {
			t.Fatal(err)
		}
	}
	match, ok := registry.MatchHTTPRoute("alpha", "POST", "/cancel-fixture")
	if !ok {
		t.Fatal("route not available")
	}

	var calls atomic.Int32
	started := make(chan execution.Request, 1)
	session, err := scheduler.Register(cortex.Registration{ID: "neuron-1", Slots: 1, Runtimes: []execution.Runtime{runtime}, Executor: scheduledExecutor(func(ctx context.Context, request execution.Request) (execution.Result, error) {
		if calls.Add(1) == 1 {
			started <- request
			<-ctx.Done()
		}
		// Even a successful but late response cannot commit cancelled work.
		return execution.Result{ID: request.ID, WorkerID: "neuron-1", Status: execution.Succeeded, Output: []byte("{}")}, nil
	})})
	if err != nil {
		t.Fatal(err)
	}
	defer session.Disconnect()

	cancelled, stop := context.WithCancel(context.Background())
	stop()
	if _, err := registry.InvokeHTTPRouteContext(cancelled, "alpha", match, []byte("{}"), "operator-1"); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled request admitted: %v", err)
	}
	if calls.Load() != 0 {
		t.Fatal("cancelled request reached a Neuron")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	deadline, _ := ctx.Deadline()
	finished := make(chan error, 1)
	go func() {
		_, err := registry.InvokeHTTPRouteContext(ctx, "alpha", match, []byte("{}"), "operator-1")
		finished <- err
	}()
	select {
	case request := <-started:
		if !request.Deadline.Equal(deadline) {
			t.Fatalf("caller deadline lost: %v != %v", request.Deadline, deadline)
		}
	case <-ctx.Done():
		t.Fatal("request did not reach Cortex")
	}
	cancel()
	select {
	case err := <-finished:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("cancellation lost: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("cancelled execution did not terminate")
	}
	if planner.commits.Load() != 0 {
		t.Fatal("cancelled execution committed durable effects")
	}
	response, err := registry.InvokeHTTPRoute("alpha", match, []byte("{}"), "operator-1")
	if err != nil || !response.Success || planner.commits.Load() != 1 {
		t.Fatalf("capacity was not released for the next request: %v, %v", response, err)
	}
}
