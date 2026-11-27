package neuron_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"brain/internal/execution"
	"brain/internal/neuron"
)

type sourceFunc func(context.Context, execution.Artifact) ([]byte, error)

func (f sourceFunc) Fetch(ctx context.Context, a execution.Artifact) ([]byte, error) {
	return f(ctx, a)
}

type runtimeFunc func(context.Context, execution.Request, []execution.BundleFile) (execution.Result, error)

func (f runtimeFunc) Run(ctx context.Context, r execution.Request, files []execution.BundleFile) (execution.Result, error) {
	return f(ctx, r, files)
}

var rt = execution.Runtime{Kind: "typescript", Profile: "isolated"}

func fixture(t *testing.T, runner runtimeFunc) (neuron.Config, execution.Request, []byte) {
	t.Helper()
	bundle, err := execution.BuildBundle([]execution.BundleFile{{Path: "main.ts", Data: []byte("source")}}, 4096)
	if err != nil {
		t.Fatal(err)
	}
	r := execution.Request{ID: "id", TraceID: "trace", ProjectID: "alpha", Artifact: execution.Artifact{ID: "plugin", Version: "1.0.0", Digest: execution.ArtifactDigest(bundle), Entrypoint: "call"}, Runtime: rt, Input: json.RawMessage(`null`), Context: json.RawMessage(`{}`), Deadline: time.Now().Add(5 * time.Second), Limits: execution.Limits{CPUTime: time.Second, MemoryBytes: 1024, InputBytes: 128, OutputBytes: 128}, Auth: execution.Authorization{ProjectID: "alpha", OriginalCaller: "user:1", Caller: "user:1", Capability: "call"}}
	cfg := neuron.Config{ID: "neuron", Slots: 1, MaxArtifactBytes: 4096, Adapters: []neuron.Adapter{{Capability: rt, Runtime: runner}}, Source: sourceFunc(func(_ context.Context, a execution.Artifact) ([]byte, error) {
		if a != r.Artifact {
			t.Error("unexpected artifact")
		}
		return bundle, nil
	})}
	return cfg, r, bundle
}
func newEngine(t *testing.T, cfg neuron.Config) *neuron.Engine {
	t.Helper()
	e, err := neuron.New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(e.Close)
	return e
}
func success() execution.Result {
	return execution.Result{Status: execution.Succeeded, Output: json.RawMessage(`null`)}
}
func TestEngineConfigurationAndVerifiedExecution(t *testing.T) {
	cfg, r, _ := fixture(t, func(ctx context.Context, req execution.Request, files []execution.BundleFile) (execution.Result, error) {
		if req.ProjectID != "alpha" || len(files) != 1 || string(files[0].Data) != "source" {
			t.Fatal(req, files)
		}
		files[0].Data[0] = 'X'
		req.Input[0] = 'X'
		req.Context[0] = 'X'
		v := success()
		v.ID = "forged"
		v.WorkerID = "forged"
		v.Effects = []json.RawMessage{json.RawMessage(`{}`)}
		return v, nil
	})
	for _, bad := range []neuron.Config{{}, {ID: "n", Slots: 1, MaxArtifactBytes: 4096, Source: cfg.Source, Adapters: []neuron.Adapter{{Capability: rt}}}, {ID: "n", Slots: 1, MaxArtifactBytes: 4096, Source: cfg.Source, Adapters: append(cfg.Adapters, cfg.Adapters[0])}} {
		if _, err := neuron.New(bad); !errors.Is(err, neuron.ErrInvalidConfig) {
			t.Fatal(err)
		}
	}
	cfg.Adapters = append(cfg.Adapters, neuron.Adapter{Capability: execution.Runtime{Kind: "javascript", Profile: "isolated"}, Runtime: cfg.Adapters[0].Runtime}, neuron.Adapter{Capability: execution.Runtime{Kind: "typescript", Profile: "another"}, Runtime: cfg.Adapters[0].Runtime})
	e := newEngine(t, cfg)
	v := e.Snapshot()
	if v.Available != 1 || len(v.Runtimes) != 3 || v.Runtimes[0].Kind != "javascript" {
		t.Fatal(v)
	}
	for i := 0; i < 2; i++ {
		res, err := e.Execute(context.Background(), r)
		if err != nil || res.ID != r.ID || res.WorkerID != "neuron" || res.Duration < 0 || r.Input[0] != 'n' || r.Context[0] != '{' {
			t.Fatal(res, err)
		}
	}
	bad := r
	bad.ID = ""
	if _, err := e.Execute(context.Background(), bad); !errors.Is(err, execution.ErrInvalidRequest) {
		t.Fatal(err)
	}
	bad = r
	bad.Runtime.Profile = "unsupported"
	if _, err := e.Execute(context.Background(), bad); !errors.Is(err, execution.ErrUnavailable) {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := e.Execute(ctx, r); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if err := e.Wait(context.Background()); err != nil {
		t.Fatal(err)
	}
	e.Drain()
	if _, err := e.Execute(context.Background(), r); !errors.Is(err, neuron.ErrDraining) {
		t.Fatal(err)
	}
	if e.Snapshot().Available != 0 {
		t.Fatal("draining capacity advertised")
	}
	e.Close()
	if _, err := e.Execute(context.Background(), r); !errors.Is(err, neuron.ErrClosed) {
		t.Fatal(err)
	}
}
func TestEngineRejectsTamperedArtifactsAndResults(t *testing.T) {
	failure := errors.New("runtime error")
	cfg, r, bundle := fixture(t, func(context.Context, execution.Request, []execution.BundleFile) (execution.Result, error) {
		return execution.Result{}, failure
	})
	e := newEngine(t, cfg)
	if _, err := e.Execute(context.Background(), r); !errors.Is(err, failure) {
		t.Fatal(err)
	}
	cfg.Source = sourceFunc(func(context.Context, execution.Artifact) ([]byte, error) { return nil, failure })
	if _, err := newEngine(t, cfg).Execute(context.Background(), r); !errors.Is(err, failure) {
		t.Fatal(err)
	}
	cfg.Source = sourceFunc(func(context.Context, execution.Artifact) ([]byte, error) {
		bad := append([]byte(nil), bundle...)
		bad[0] ^= 1
		return bad, nil
	})
	if _, err := newEngine(t, cfg).Execute(context.Background(), r); !errors.Is(err, execution.ErrArtifactMismatch) {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cfg.Source = sourceFunc(func(context.Context, execution.Artifact) ([]byte, error) { cancel(); return bundle, nil })
	if _, err := newEngine(t, cfg).Execute(ctx, r); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	cfg.Source = sourceFunc(func(context.Context, execution.Artifact) ([]byte, error) { return bundle, nil })
	cfg.Adapters[0].Runtime = runtimeFunc(func(context.Context, execution.Request, []execution.BundleFile) (execution.Result, error) {
		return execution.Result{Status: execution.Succeeded, Output: []byte("bad")}, nil
	})
	if _, err := newEngine(t, cfg).Execute(context.Background(), r); !errors.Is(err, execution.ErrInvalidResult) {
		t.Fatal(err)
	}
	cfg.Adapters[0].Runtime = runtimeFunc(func(context.Context, execution.Request, []execution.BundleFile) (execution.Result, error) {
		return execution.Result{Status: execution.Failed, Error: &execution.Failure{Code: "failed"}}, nil
	})
	if v, err := newEngine(t, cfg).Execute(context.Background(), r); err != nil || v.Error.Code != "failed" {
		t.Fatal(v, err)
	}
}
func TestEngineCapacityDrainAndCancel(t *testing.T) {
	started := make(chan struct{})
	stopped := make(chan struct{})
	release := make(chan struct{})
	cfg, r, _ := fixture(t, func(ctx context.Context, _ execution.Request, _ []execution.BundleFile) (execution.Result, error) {
		close(started)
		<-ctx.Done()
		close(stopped)
		<-release
		return success(), nil
	})
	e := newEngine(t, cfg)
	done := make(chan error, 1)
	go func() { _, err := e.Execute(context.Background(), r); done <- err }()
	<-started
	if v := e.Snapshot(); v.Running != 1 || v.Available != 0 {
		t.Fatal(v)
	}
	if _, err := e.Execute(context.Background(), r); !errors.Is(err, neuron.ErrDuplicate) {
		t.Fatal(err)
	}
	other := r
	other.ID = "other"
	if _, err := e.Execute(context.Background(), other); !errors.Is(err, execution.ErrUnavailable) {
		t.Fatal(err)
	}
	e.Drain()
	if _, err := e.Execute(context.Background(), other); !errors.Is(err, neuron.ErrDraining) {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	if err := e.Wait(ctx); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal(err)
	}
	e.Close()
	<-stopped
	close(release)
	select {
	case err := <-done:
		if !errors.Is(err, neuron.ErrClosed) {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("not cancelled")
	}
	if err := e.Wait(context.Background()); err != nil {
		t.Fatal(err)
	}
	if v := e.Snapshot(); v.Running != 0 || !v.Closed {
		t.Fatal(v)
	}
}
