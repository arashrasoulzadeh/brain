package plugin_test

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"brain/internal/cortex"
	"brain/internal/execution"
	"brain/internal/plugin"
)

type scheduledExecutor func(context.Context, execution.Request) (execution.Result, error)

func (f scheduledExecutor) Execute(ctx context.Context, req execution.Request) (execution.Result, error) {
	return f(ctx, req)
}

func TestRegistryScheduledWorkerUsesCortex(t *testing.T) {
	manifest := plugin.Manifest{
		ID: "scheduled-fixture", Version: "1.0.0", HostAPI: 1, Entry: "index.ts",
		SourceFiles: []string{"index.ts"}, Capabilities: []string{"read"},
		Routes: []plugin.Route{{Operation: "read", Function: "read", Capability: "read"}},
		Events: []plugin.EventSpec{{Name: "completed", Fields: map[string]string{"ok": "boolean"}}},
	}
	root := t.TempDir()
	path := filepath.Join(root, "index.ts")
	if err := os.WriteFile(path, []byte("export function read() { return { ok: true }; }"), 0o600); err != nil {
		t.Fatal(err)
	}
	catalog, err := execution.NewCatalog(8, 1<<20)
	if err != nil {
		t.Fatal(err)
	}
	rt := execution.Runtime{Kind: plugin.ScheduledTypeScriptKind, Profile: "test"}
	artifact, err := plugin.RegisterScheduledArtifact(catalog, rt, manifest, root)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(path, []byte("export function read() { return {}; }"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err = plugin.RegisterScheduledArtifact(catalog, rt, manifest, root); !errors.Is(err, execution.ErrArtifactConflict) {
		t.Fatal("changed content replaced a pinned version", err)
	}
	scheduler, err := cortex.New(catalog, cortex.Config{MaxNeurons: 4, MaxProjectRunning: 2, HeartbeatTTL: time.Minute})
	if err != nil {
		t.Fatal(err)
	}
	r := plugin.NewRegistry(1)
	r.BindProjectResolver(func(id string) (plugin.ProjectSnapshot, error) {
		return plugin.ProjectSnapshot{ID: id, Active: true}, nil
	})
	r.BindWorkerFactory(func(project, id, version string) (plugin.ContractWorker, error) {
		if id != manifest.ID || version != manifest.Version {
			return nil, plugin.ErrInvalidWorkerRequest
		}
		return artifact.Worker(project, scheduler)
	})
	for _, err := range []error{
		r.Register(manifest), r.Install("alpha", manifest.ID, manifest.Version),
		r.SetReadiness(manifest.ID, manifest.Version, true),
		r.Grant("alpha", manifest.ID, manifest.Version, "read"),
		r.EnableWithDependencies("alpha", manifest.ID, manifest.Version, nil),
	} {
		if err != nil {
			t.Fatal(err)
		}
	}
	request := plugin.WorkerRequest{Protocol: 1, ProjectID: "alpha", PluginID: manifest.ID, Version: manifest.Version, Operation: "read", Payload: []byte("{}")}
	if _, err = r.InvokeAsOperator(request, "operator-1"); !errors.Is(err, plugin.ErrWorkerNotReady) {
		t.Fatal("zero Neurons must fail closed", err)
	}
	var calls int
	var captured execution.Request
	session, err := scheduler.Register(cortex.Registration{ID: "neuron-1", Slots: 1, Runtimes: []execution.Runtime{rt}, Executor: scheduledExecutor(func(_ context.Context, assigned execution.Request) (execution.Result, error) {
		calls++
		captured = assigned
		decoded, err := plugin.DecodeScheduledRequest(assigned, manifest)
		if err != nil {
			return execution.Result{}, err
		}
		if decoded.ProjectID != "alpha" || decoded.HostContext.Auth.Actor.ID != "operator-1" || decoded.HostContext.Auth.Caller.ID != "operator-1" || !decoded.HostContext.ProjectVerified {
			return execution.Result{}, errors.New("trusted host context was not preserved")
		}
		result, err := plugin.EncodeScheduledResult(decoded, plugin.WorkerResponse{Protocol: 1, Success: true, Payload: []byte(`{"ok":true}`), Observations: []plugin.ObservationIntent{{Event: "completed", Payload: json.RawMessage(`{"ok":true}`)}}})
		result.ID, result.WorkerID = assigned.ID, "neuron-1"
		return result, err
	})})
	if err != nil {
		t.Fatal(err)
	}
	defer session.Disconnect()
	response, err := r.InvokeAsOperator(request, "operator-1")
	if err != nil || !response.Success || string(response.Payload) != `{"ok":true}` || calls != 1 {
		t.Fatal(response, err, calls)
	}
	if len(response.Observations) != 0 {
		t.Fatal("registry leaked internal observation intents")
	}
	if captured.Idempotent || captured.Artifact.Version != manifest.Version || captured.Auth.OriginalCaller != "operator:operator-1" {
		t.Fatal("assignment lost version, original identity or retry policy")
	}
	tampered := captured
	tampered.Auth.Caller = "host"
	if _, err := plugin.DecodeScheduledRequest(tampered, manifest); err == nil {
		t.Fatal("mismatched assignment identity accepted")
	}
	tampered = captured
	tampered.Auth.ProjectID = "beta"
	if _, err := plugin.DecodeScheduledRequest(tampered, manifest); err == nil {
		t.Fatal("mismatched project accepted")
	}
	tampered = captured
	tampered.Artifact.Entrypoint = "not-declared"
	if _, err := plugin.DecodeScheduledRequest(tampered, manifest); err == nil {
		t.Fatal("undeclared operation accepted")
	}
	mismatch := request
	mismatch.Payload = []byte(`{"projectId":"beta"}`)
	if _, err := r.InvokeAsOperator(mismatch, "operator-1"); !errors.Is(err, plugin.ErrInvalidWorkerRequest) {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Fatal("denied invocation reached a Neuron")
	}
	session.Disconnect()
	if _, err := r.InvokeAsOperator(request, "operator-1"); !errors.Is(err, plugin.ErrWorkerNotReady) {
		t.Fatal("disconnected Neuron was still used", err)
	}
}

// All shipped manifests must be snapshotable without mounting other plugin
// source directories. Dependency metadata is not permission to load code.
func TestScheduledArtifactsForShippedPlugins(t *testing.T) {
	locations, err := plugin.Discover("../../plugins")
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := execution.NewCatalog(128, execution.MaxBundleBytes)
	if err != nil {
		t.Fatal(err)
	}
	rt := execution.Runtime{Kind: plugin.ScheduledTypeScriptKind, Profile: "test"}
	for _, location := range locations {
		t.Run(location.ID, func(t *testing.T) {
			manifest, err := plugin.LoadManifest(location.ManifestPath)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := plugin.RegisterScheduledArtifact(catalog, rt, manifest, location.Directory); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestScheduledManifestRejectsDependencySourceAndUndeclaredMetadata(t *testing.T) {
	manifest := plugin.Manifest{
		ID: "sample", Version: "1.0.0", HostAPI: 1, Entry: "index.ts",
		SourceFiles: []string{"index.ts"}, Capabilities: []string{"read"},
		Routes:               []plugin.Route{{Operation: "read", Function: "read", Capability: "read"}},
		OptionalDependencies: []string{"other@1.0.0"}, PluginImports: []string{"other@1.0.0"},
		Imports: []string{"../sdk-typescript/src/index.js"},
	}
	if err := plugin.ValidateScheduledManifest(manifest); err != nil {
		t.Fatal(err)
	}
	manifest.Imports = append(manifest.Imports, "../other/index.js")
	if !errors.Is(plugin.ValidateScheduledManifest(manifest), plugin.ErrInvalidWorkerRequest) {
		t.Fatal("dependency source import accepted")
	}
	manifest.Imports = manifest.Imports[:1]
	manifest.PluginImports = []string{"unknown@1.0.0"}
	if !errors.Is(plugin.ValidateScheduledManifest(manifest), plugin.ErrInvalidWorkerRequest) {
		t.Fatal("undeclared dependency metadata accepted")
	}
}
