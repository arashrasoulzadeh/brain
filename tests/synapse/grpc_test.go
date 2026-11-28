package synapse_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"testing"
	"time"

	"brain/internal/cortex"
	"brain/internal/execution"
	"brain/internal/neuron"
	"brain/internal/synapse"
)

type wireRuntime func(context.Context, execution.Request, []execution.BundleFile) (execution.Result, error)

func (f wireRuntime) Run(ctx context.Context, r execution.Request, files []execution.BundleFile) (execution.Result, error) {
	return f(ctx, r, files)
}

func TestGRPCFourNeuronsAndArtifactExecution(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	authority := ca(t)
	serverCert := leaf(t, authority, true)
	serverTLS, err := synapse.ServerTLS(serverCert, authority.pool)
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := execution.NewCatalog(4, 1<<20)
	if err != nil {
		t.Fatal(err)
	}
	rt := execution.Runtime{Kind: "typescript", Profile: "isolated"}
	key := execution.ArtifactKey{ID: "plugin", Version: "1.0.0"}
	bundle, err := execution.BuildBundle([]execution.BundleFile{{Path: "main.ts", Data: []byte("fixture")}}, 4096)
	if err != nil {
		t.Fatal(err)
	}
	if err = catalog.Register(execution.Registration{Key: key, Runtime: rt, Entrypoints: []string{"call"}, Bundle: bundle}); err != nil {
		t.Fatal(err)
	}
	artifact, _, err := catalog.Resolve(key, "call")
	if err != nil {
		t.Fatal(err)
	}
	scheduler, err := cortex.New(catalog, cortex.Config{MaxNeurons: 4, MaxProjectRunning: 4, HeartbeatTTL: 8 * time.Second})
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
	started := make(chan string, 4)
	release := make(chan struct{})
	workerErrors := make(chan error, 4)
	workers, stopWorkers := context.WithCancel(ctx)
	defer stopWorkers()
	for i := 0; i < 4; i++ {
		id := fmt.Sprintf("worker-%d", i)
		certificate := leaf(t, authority, false, "spiffe://brain/neurons/"+id)
		clientTLS, err := synapse.ClientTLS(certificate, authority.pool, "brain.test")
		if err != nil {
			t.Fatal(err)
		}
		adapter := wireRuntime(func(ctx context.Context, r execution.Request, files []execution.BundleFile) (execution.Result, error) {
			if len(files) != 1 || string(files[0].Data) != "fixture" || r.Auth.ProjectID != r.ProjectID {
				return execution.Result{}, errors.New("invalid artifact or identity")
			}
			started <- id
			select {
			case <-release:
				return execution.Result{Status: execution.Succeeded, Output: r.Input}, nil
			case <-ctx.Done():
				return execution.Result{}, ctx.Err()
			}
		})
		go func() {
			workerErrors <- synapse.RunNeuron(workers, listener.Addr().String(), clientTLS, neuron.Config{ID: id, Slots: 1, MaxArtifactBytes: 1 << 20, Adapters: []neuron.Adapter{{Capability: rt, Runtime: adapter}}})
		}()
	}
	for len(scheduler.Snapshot().Neurons) != 4 {
		select {
		case err := <-workerErrors:
			t.Fatal(err)
		case <-ctx.Done():
			t.Fatal("four Neurons did not register")
		case <-time.After(10 * time.Millisecond):
		}
	}
	type outcome struct {
		result execution.Result
		err    error
	}
	done := make(chan outcome, 4)
	for i := 0; i < 4; i++ {
		r := req(fmt.Sprintf("grpc-%d", i))
		r.ProjectID = fmt.Sprintf("project-%d", i%2)
		r.Auth.ProjectID = r.ProjectID
		r.Artifact, r.Runtime = artifact, rt
		r.Input = json.RawMessage("42")
		r.Deadline = time.Now().Add(10 * time.Second)
		go func() {
			result, err := scheduler.Execute(ctx, r)
			done <- outcome{result, err}
		}()
	}
	seen := map[string]bool{}
	for len(seen) != 4 {
		select {
		case id := <-started:
			seen[id] = true
		case <-ctx.Done():
			t.Fatal("executions did not distribute")
		}
	}
	close(release)
	for i := 0; i < 4; i++ {
		select {
		case v := <-done:
			if v.err != nil || v.result.Status != execution.Succeeded || string(v.result.Output) != "42" {
				t.Fatal(v)
			}
		case <-ctx.Done():
			t.Fatal(ctx.Err())
		}
	}
	stopWorkers()
	for i := 0; i < 4; i++ {
		select {
		case <-workerErrors:
		case <-ctx.Done():
			t.Fatal("worker did not disconnect")
		}
	}
	for len(scheduler.Snapshot().Neurons) != 0 {
		select {
		case <-ctx.Done():
			t.Fatal("closed sessions remained registered")
		case <-time.After(10 * time.Millisecond):
		}
	}
	r := req("zero")
	r.Artifact, r.Runtime = artifact, rt
	if _, err := scheduler.Execute(ctx, r); !errors.Is(err, execution.ErrUnavailable) {
		t.Fatal(err)
	}
}

func TestGRPCRejectsCertificateIdentitySubstitution(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	authority := ca(t)
	serverTLS, err := synapse.ServerTLS(leaf(t, authority, true), authority.pool)
	if err != nil {
		t.Fatal(err)
	}
	clientTLS, err := synapse.ClientTLS(leaf(t, authority, false, "spiffe://brain/neurons/real"), authority.pool, "brain.test")
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := execution.NewCatalog(1, 4096)
	if err != nil {
		t.Fatal(err)
	}
	scheduler, err := cortex.New(catalog, cortex.Config{MaxNeurons: 1, MaxProjectRunning: 1, HeartbeatTTL: 8 * time.Second})
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
	adapter := wireRuntime(func(context.Context, execution.Request, []execution.BundleFile) (execution.Result, error) {
		t.Error("unauthenticated execution")
		return execution.Result{}, nil
	})
	err = synapse.RunNeuron(ctx, listener.Addr().String(), clientTLS, neuron.Config{ID: "forged", Slots: 1, MaxArtifactBytes: 4096, Adapters: []neuron.Adapter{{Capability: execution.Runtime{Kind: "typescript", Profile: "isolated"}, Runtime: adapter}}})
	if err == nil || len(scheduler.Snapshot().Neurons) != 0 {
		t.Fatal("certificate identity substitution accepted", err)
	}
}
