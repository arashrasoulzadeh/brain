package distributed_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"brain/internal/cortex"
	"brain/internal/execution"
	"brain/internal/neuron"
	"brain/internal/synapse"
)

type source func(context.Context, execution.Artifact) ([]byte, error)

func (f source) Fetch(ctx context.Context, a execution.Artifact) ([]byte, error) { return f(ctx, a) }

type runtime func(context.Context, execution.Request, []execution.BundleFile) (execution.Result, error)

func (f runtime) Run(ctx context.Context, r execution.Request, files []execution.BundleFile) (execution.Result, error) {
	return f(ctx, r, files)
}

// Exercise the actual scheduler, session, artifact transfer and Neuron engine
// together. This is intentionally in-process, not a gRPC or sandbox test.
func TestDistributedPipeline(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	rt := execution.Runtime{Kind: "typescript", Profile: "isolated"}
	cat, err := execution.NewCatalog(8, 4096)
	if err != nil {
		t.Fatal(err)
	}
	key := execution.ArtifactKey{ProjectID: "alpha", ID: "plugin", Version: "1.0.0"}
	bundle, err := execution.BuildBundle([]execution.BundleFile{{Path: "main.ts", Data: []byte("source")}}, 4096)
	if err != nil {
		t.Fatal(err)
	}
	if err = cat.Register(execution.Registration{Key: key, Runtime: rt, Entrypoints: []string{"call"}, Bundle: bundle}); err != nil {
		t.Fatal(err)
	}
	artifact, _, err := cat.Resolve(key, "call")
	if err != nil {
		t.Fatal(err)
	}
	req := execution.Request{ID: "id", TraceID: "trace", ProjectID: "alpha", Artifact: artifact, Runtime: rt, Input: json.RawMessage(`{"value":42}`), Deadline: time.Now().Add(4 * time.Second), Limits: execution.Limits{CPUTime: time.Second, MemoryBytes: 1024, InputBytes: 128, OutputBytes: 128}, Auth: execution.Authorization{ProjectID: "alpha", OriginalCaller: "user:1", Caller: "user:1", Capability: "call"}}
	scheduler, err := cortex.New(cat, cortex.Config{MaxNeurons: 4, MaxProjectRunning: 4, HeartbeatTTL: time.Minute})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = scheduler.Execute(ctx, req); !errors.Is(err, execution.ErrUnavailable) {
		t.Fatal(err)
	}
	session, err := synapse.NewSession(synapse.Config{WorkerID: "worker", SessionID: "session", MaxPending: 1, QueueSize: 2})
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	registration, err := scheduler.Register(cortex.Registration{ID: "worker", Runtimes: []execution.Runtime{rt}, Slots: 1, Executor: session})
	if err != nil {
		t.Fatal(err)
	}
	defer registration.Disconnect()
	type outcome struct {
		result execution.Result
		err    error
	}
	done := make(chan outcome, 1)
	go func() { v, e := scheduler.Execute(ctx, req); done <- outcome{v, e} }()
	command, err := session.Next(ctx)
	if err != nil || command.Assignment == nil {
		t.Fatal(command, err)
	}
	assignment := command.Assignment
	if err = session.Accept(assignment.Lease); err != nil {
		t.Fatal(err)
	}
	worker, err := neuron.New(neuron.Config{ID: "worker", Slots: 1, MaxArtifactBytes: 4096, Source: source(func(ctx context.Context, a execution.Artifact) ([]byte, error) {
		assembler, e := synapse.NewAssembler(assignment.Lease, a, 4096)
		if e != nil {
			return nil, e
		}
		if e = synapse.StreamArtifact(ctx, session, cat, assignment.Lease, assembler.Push); e != nil {
			return nil, e
		}
		return assembler.Finish()
	}), Adapters: []neuron.Adapter{{Capability: rt, Runtime: runtime(func(_ context.Context, r execution.Request, files []execution.BundleFile) (execution.Result, error) {
		if r.ProjectID != "alpha" || len(files) != 1 || string(files[0].Data) != "source" {
			return execution.Result{}, errors.New("incorrect dispatch")
		}
		return execution.Result{Status: execution.Succeeded, Output: r.Input}, nil
	})}}})
	if err != nil {
		t.Fatal(err)
	}
	defer worker.Close()
	result, err := worker.Execute(ctx, assignment.Request)
	if err != nil {
		t.Fatal(err)
	}
	if err = session.Complete(assignment.Lease, result); err != nil {
		t.Fatal(err)
	}
	select {
	case out := <-done:
		if out.err != nil || out.result.WorkerID != "worker" || string(out.result.Output) != string(req.Input) {
			t.Fatal(out)
		}
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}
	if v := scheduler.Snapshot(); v.Running != 0 || v.Queued != 0 {
		t.Fatal(v)
	}
	if v := worker.Snapshot(); v.Available != 1 || v.Running != 0 {
		t.Fatal(v)
	}
	if err = session.Complete(assignment.Lease, result); err == nil {
		t.Fatal("duplicate completion accepted")
	}
	registration.Disconnect()
	if _, err = scheduler.Execute(ctx, req); !errors.Is(err, execution.ErrUnavailable) {
		t.Fatal(err)
	}
}
