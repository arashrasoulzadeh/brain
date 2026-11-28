package synapse_test

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"testing"
	"time"

	"brain/internal/cortex"
	"brain/internal/execution"
	"brain/internal/neuron"
	"brain/internal/synapse"
)

func TestGRPCNestedReadOnSameTwoSlotNeuron(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	a := ca(t)
	serverTLS, err := synapse.ServerTLS(leaf(t, a, true), a.pool)
	if err != nil {
		t.Fatal(err)
	}
	clientTLS, err := synapse.ClientTLS(leaf(t, a, false, "spiffe://brain/neurons/worker"), a.pool, "brain.test")
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := execution.NewCatalog(4, 4096)
	if err != nil {
		t.Fatal(err)
	}
	rt := execution.Runtime{Kind: execution.UserFunctionRuntimeKind, Profile: "isolated"}
	childRT := execution.Runtime{Kind: "read-only", Profile: "isolated"}
	bundle, err := execution.BuildBundle([]execution.BundleFile{{Path: "source.mjs", Data: []byte("fixture")}}, 4096)
	if err != nil {
		t.Fatal(err)
	}
	parentKey := execution.ArtifactKey{ProjectID: "alpha", ID: "parent", Version: "1"}
	childKey := execution.ArtifactKey{ID: "read", Version: "1"}
	for _, reg := range []execution.Registration{
		{Key: parentKey, Runtime: rt, Entrypoints: []string{"functions.invoke", "functions.call"}, Bundle: bundle},
		{Key: childKey, Runtime: childRT, Entrypoints: []string{"read"}, Bundle: bundle},
	} {
		if err := catalog.Register(reg); err != nil {
			t.Fatal(err)
		}
	}
	parentArtifact, _, err := catalog.Resolve(parentKey, "functions.invoke")
	if err != nil {
		t.Fatal(err)
	}
	childArtifact, _, err := catalog.Resolve(childKey, "read")
	if err != nil {
		t.Fatal(err)
	}
	scheduler, err := cortex.New(catalog, cortex.Config{MaxNeurons: 1, MaxProjectRunning: 2, MaxQueued: 4, MaxProjectQueued: 2, HeartbeatTTL: 8 * time.Second})
	if err != nil {
		t.Fatal(err)
	}
	server, err := synapse.NewGRPCServer(serverTLS, scheduler, catalog, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	go server.Serve(listener)
	defer server.Stop()
	parentRuntime := wireRuntime(func(ctx context.Context, req execution.Request, _ []execution.BundleFile) (execution.Result, error) {
		output, err := execution.CallHost(ctx, "read", json.RawMessage(`{"key":"one"}`))
		if req.Artifact.Entrypoint == "functions.call" {
			if !errors.Is(err, execution.ErrHostCallDenied) {
				t.Error("public bridge not denied", err)
			}
			return execution.Result{Status: execution.Succeeded, Output: json.RawMessage(`{"denied":true}`)}, nil
		}
		if err != nil {
			return execution.Result{}, err
		}
		return execution.Result{Status: execution.Succeeded, Output: output}, nil
	})
	childRuntime := wireRuntime(func(_ context.Context, req execution.Request, _ []execution.BundleFile) (execution.Result, error) {
		if req.ProjectID != "alpha" || req.ParentID != "root" || req.TraceID != "root" || req.Auth.Caller != "operator:root" {
			t.Error("wrong child scope", req)
		}
		return execution.Result{Status: execution.Succeeded, Output: req.Input}, nil
	})
	done := make(chan error, 1)
	go func() {
		done <- synapse.RunNeuron(ctx, listener.Addr().String(), clientTLS, neuron.Config{ID: "worker", Slots: 2, MaxArtifactBytes: 4096, Adapters: []neuron.Adapter{{Capability: rt, Runtime: parentRuntime}, {Capability: childRT, Runtime: childRuntime}}})
	}()
	for len(scheduler.Snapshot().Neurons) != 1 {
		select {
		case err := <-done:
			t.Fatal(err)
		case <-ctx.Done():
			t.Fatal(ctx.Err())
		case <-time.After(10 * time.Millisecond):
		}
	}
	root := req("root")
	root.Artifact, root.Runtime = parentArtifact, rt
	root.Auth.OriginalCaller, root.Auth.Caller, root.Auth.Capability, root.Auth.Resource = "operator:root", "operator:root", "functions.invoke", "sample"
	root.Deadline = time.Now().Add(10 * time.Second)
	host := execution.WithHostHandler(ctx, func(ctx context.Context, operation string, input json.RawMessage) (json.RawMessage, error) {
		parent, ok := execution.ParentFromContext(ctx)
		if !ok || parent != execution.ScopeOf(root) || operation != "read" {
			return nil, execution.ErrHostCallDenied
		}
		child := req("child")
		child.Artifact, child.Runtime, child.Input = childArtifact, childRT, input
		child.ParentID, child.TraceID = parent.ID, parent.TraceID
		child.Auth.OriginalCaller, child.Auth.Caller = parent.OriginalCaller, parent.Caller
		child.Deadline, _ = ctx.Deadline()
		result, err := scheduler.Execute(ctx, child)
		return result.Output, err
	})
	result, err := scheduler.Execute(host, root)
	if err != nil || string(result.Output) != `{"key":"one"}` {
		t.Fatal(result, err)
	}
	snapshot := scheduler.Snapshot()
	if snapshot.Running != 0 || snapshot.Queued != 0 || len(snapshot.Recent) != 2 || snapshot.Recent[1].ParentID != "root" {
		t.Fatalf("unexpected nested trace: %+v", snapshot)
	}
	public := req("anonymous")
	public.Artifact, _, err = catalog.Resolve(parentKey, "functions.call")
	if err != nil {
		t.Fatal(err)
	}
	public.Runtime = rt
	public.Auth.OriginalCaller, public.Auth.Caller, public.Auth.Capability, public.Auth.Resource = "public", "public", "functions.call", "sample"
	result, err = scheduler.Execute(host, public)
	if err != nil || string(result.Output) != `{"denied":true}` {
		t.Fatal("public call", result, err)
	}
	cancel()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("Neuron did not disconnect")
	}
}
