package cortex_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"brain/internal/cortex"
	"brain/internal/execution"
)

var runtime = execution.Runtime{Kind: "typescript", Profile: "isolated"}
var otherRuntime = execution.Runtime{Kind: "javascript", Profile: "isolated"}

type executorFunc func(context.Context, execution.Request) (execution.Result, error)

func (f executorFunc) Execute(ctx context.Context, r execution.Request) (execution.Result, error) {
	return f(ctx, r)
}

type clock struct {
	mu  sync.Mutex
	now time.Time
}

func (c *clock) read() time.Time         { c.mu.Lock(); defer c.mu.Unlock(); return c.now }
func (c *clock) advance(d time.Duration) { c.mu.Lock(); defer c.mu.Unlock(); c.now = c.now.Add(d) }

type fixture struct {
	scheduler *cortex.Scheduler
	catalog   *execution.Catalog
	artifact  execution.Artifact
	clock     *clock
}

func setup(t *testing.T, queue, projectQueue, quota, maxNeurons int) fixture {
	t.Helper()
	cat, err := execution.NewCatalog(4, 4096)
	if err != nil {
		t.Fatal(err)
	}
	bundle, err := execution.BuildBundle([]execution.BundleFile{{Path: "main.ts", Data: []byte("export {}")}}, 4096)
	if err != nil {
		t.Fatal(err)
	}
	key := execution.ArtifactKey{ID: "plugin", Version: "1.0.0"}
	if err = cat.Register(execution.Registration{Key: key, Runtime: runtime, Entrypoints: []string{"call"}, Bundle: bundle}); err != nil {
		t.Fatal(err)
	}
	art, _, err := cat.Resolve(key, "call")
	if err != nil {
		t.Fatal(err)
	}
	c := &clock{now: time.Now()}
	s, err := cortex.New(cat, cortex.Config{MaxNeurons: maxNeurons, MaxProjectRunning: quota, MaxQueued: queue, MaxProjectQueued: projectQueue, HeartbeatTTL: time.Minute, Clock: c.read})
	if err != nil {
		t.Fatal(err)
	}
	return fixture{s, cat, art, c}
}
func (f fixture) request(id, project string) execution.Request {
	return execution.Request{ID: id, TraceID: id, ProjectID: project, Artifact: f.artifact, Runtime: runtime, Input: json.RawMessage(`null`), Deadline: time.Now().Add(10 * time.Second), Limits: execution.Limits{CPUTime: time.Second, MemoryBytes: 1024, InputBytes: 1024, OutputBytes: 1024}, Auth: execution.Authorization{ProjectID: project, OriginalCaller: "user:1", Caller: "user:1", Capability: "call"}}
}
func result(r execution.Request, id string) execution.Result {
	return execution.Result{ID: r.ID, WorkerID: id, Status: execution.Succeeded, Output: json.RawMessage(`null`)}
}
func fast(id string) executorFunc {
	return func(_ context.Context, r execution.Request) (execution.Result, error) { return result(r, id), nil }
}
func register(t *testing.T, f fixture, id string, slots int, exec execution.Executor, rts ...execution.Runtime) *cortex.Session {
	t.Helper()
	w, err := f.scheduler.Register(cortex.Registration{ID: id, Slots: slots, Executor: exec, Runtimes: rts})
	if err != nil {
		t.Fatal(err)
	}
	return w
}
func waitState(t *testing.T, s *cortex.Scheduler, running, queued int) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		v := s.Snapshot()
		if v.Running == running && v.Queued == queued {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatalf("unexpected state: %+v", s.Snapshot())
}

type outcome struct {
	result execution.Result
	err    error
}

func submit(s *cortex.Scheduler, ctx context.Context, req execution.Request) <-chan outcome {
	done := make(chan outcome, 1)
	go func() { v, e := s.Execute(ctx, req); done <- outcome{v, e} }()
	return done
}
func receive(t *testing.T, ch <-chan outcome) outcome {
	t.Helper()
	select {
	case v := <-ch:
		return v
	case <-time.After(3 * time.Second):
		t.Fatal("execution did not finish")
		return outcome{}
	}
}
func TestConfigurationAndRegistration(t *testing.T) {
	f := setup(t, 0, 0, 2, 1)
	if _, err := cortex.New(nil, cortex.Config{}); !errors.Is(err, cortex.ErrInvalidConfig) {
		t.Fatal(err)
	}
	for _, cfg := range []cortex.Config{{MaxNeurons: 1, MaxProjectRunning: 1, HeartbeatTTL: time.Second, MaxQueued: 1}, {MaxNeurons: 1, MaxProjectRunning: 1, HeartbeatTTL: time.Second, MaxProjectQueued: 1}} {
		if _, err := cortex.New(f.catalog, cfg); !errors.Is(err, cortex.ErrInvalidConfig) {
			t.Fatal(err)
		}
	}
	if _, err := cortex.New(f.catalog, cortex.Config{MaxNeurons: 1, MaxProjectRunning: 1, HeartbeatTTL: time.Second}); err != nil {
		t.Fatal(err)
	}
	for _, reg := range []cortex.Registration{{}, {ID: "n", Slots: 1, Executor: fast("n"), Runtimes: []execution.Runtime{runtime, runtime}}, {ID: "n", Slots: 1, Executor: fast("n"), Runtimes: []execution.Runtime{{}}}} {
		if _, err := f.scheduler.Register(reg); !errors.Is(err, cortex.ErrInvalidNeuron) {
			t.Fatal(err)
		}
	}
	w := register(t, f, "a", 1, fast("a"), runtime, otherRuntime, execution.Runtime{Kind: "typescript", Profile: "another"})
	if _, err := f.scheduler.Register(cortex.Registration{ID: "a", Slots: 1, Executor: fast("a"), Runtimes: []execution.Runtime{runtime}}); !errors.Is(err, cortex.ErrDuplicateNeuron) {
		t.Fatal(err)
	}
	if _, err := f.scheduler.Register(cortex.Registration{ID: "b", Slots: 1, Executor: fast("b"), Runtimes: []execution.Runtime{runtime}}); !errors.Is(err, execution.ErrUnavailable) {
		t.Fatal(err)
	}
	if err := w.Pulse(-1); !errors.Is(err, cortex.ErrInvalidNeuron) {
		t.Fatal(err)
	}
	if err := w.Pulse(2); err != nil {
		t.Fatal(err)
	}
	v := f.scheduler.Snapshot()
	if len(v.Neurons) != 1 || v.Neurons[0].Slots != 2 || len(v.Neurons[0].Runtimes) != 3 {
		t.Fatal(v)
	}
	if err := w.Drain(); err != nil {
		t.Fatal(err)
	}
	if !f.scheduler.Snapshot().Neurons[0].Draining {
		t.Fatal("not draining")
	}
	w.Disconnect()
	w.Disconnect()
	if err := w.Pulse(1); !errors.Is(err, cortex.ErrSessionLost) {
		t.Fatal(err)
	}
	if err := w.Drain(); !errors.Is(err, cortex.ErrSessionLost) {
		t.Fatal(err)
	}
}
func TestZeroNeuronAndPlacement(t *testing.T) {
	f := setup(t, 0, 0, 2, 4)
	ctx := context.Background()
	r := f.request("a", "alpha")
	if _, err := f.scheduler.Execute(ctx, r); !errors.Is(err, execution.ErrUnavailable) {
		t.Fatal(err)
	}
	incompatible := register(t, f, "0", 1, fast("0"), otherRuntime)
	if _, err := f.scheduler.Execute(ctx, r); !errors.Is(err, execution.ErrUnavailable) {
		t.Fatal(err)
	}
	defer incompatible.Disconnect()
	release := make(chan struct{})
	started := make(chan string, 2)
	block := func(id string) executorFunc {
		return func(ctx context.Context, r execution.Request) (execution.Result, error) {
			started <- id
			select {
			case <-release:
				return result(r, id), nil
			case <-ctx.Done():
				return execution.Result{}, ctx.Err()
			}
		}
	}
	a := register(t, f, "a", 2, block("a"), runtime)
	b := register(t, f, "b", 2, block("b"), runtime)
	if v := f.scheduler.Snapshot(); len(v.Neurons) != 3 || v.Neurons[0].ID != "0" || v.Neurons[1].ID != "a" || v.Neurons[2].ID != "b" {
		t.Fatal(v)
	}
	one := submit(f.scheduler, ctx, r)
	if got := <-started; got != "a" {
		t.Fatal(got)
	}
	two := submit(f.scheduler, ctx, f.request("b", "beta"))
	if got := <-started; got != "b" {
		t.Fatal(got)
	}
	if _, err := f.scheduler.Execute(ctx, r); !errors.Is(err, cortex.ErrDuplicateExecution) {
		t.Fatal(err)
	}
	if err := a.Pulse(0); err != nil {
		t.Fatal(err)
	}
	if err := b.Drain(); err != nil {
		t.Fatal(err)
	}
	if _, err := f.scheduler.Execute(ctx, f.request("c", "gamma")); !errors.Is(err, execution.ErrUnavailable) {
		t.Fatal(err)
	}
	close(release)
	for _, done := range []<-chan outcome{one, two} {
		if v := receive(t, done); v.err != nil {
			t.Fatal(v.err)
		}
	}
	a.Disconnect()
	b.Disconnect()
	waitState(t, f.scheduler, 0, 0)
	invalid := r
	invalid.Artifact.Digest = "bad"
	if _, err := f.scheduler.Execute(ctx, invalid); !errors.Is(err, execution.ErrInvalidRequest) {
		t.Fatal(err)
	}
	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	if _, err := f.scheduler.Execute(cancelled, r); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}
func TestQueueQuotasAndCancellation(t *testing.T) {
	f := setup(t, 2, 1, 1, 2)
	ctx := context.Background()
	a := submit(f.scheduler, ctx, f.request("a", "alpha"))
	waitState(t, f.scheduler, 0, 1)
	if _, err := f.scheduler.Execute(ctx, f.request("a", "alpha")); !errors.Is(err, cortex.ErrDuplicateExecution) {
		t.Fatal(err)
	}
	if _, err := f.scheduler.Execute(ctx, f.request("b", "alpha")); !errors.Is(err, cortex.ErrProjectQuota) {
		t.Fatal(err)
	}
	cancelled, cancel := context.WithCancel(ctx)
	b := submit(f.scheduler, cancelled, f.request("b", "beta"))
	waitState(t, f.scheduler, 0, 2)
	if _, err := f.scheduler.Execute(ctx, f.request("c", "gamma")); !errors.Is(err, cortex.ErrQueueFull) {
		t.Fatal(err)
	}
	cancel()
	if v := receive(t, b); !errors.Is(v.err, context.Canceled) {
		t.Fatal(v.err)
	}
	waitState(t, f.scheduler, 0, 1)
	w := register(t, f, "n", 1, fast("n"), runtime)
	if v := receive(t, a); v.err != nil {
		t.Fatal(v.err)
	}
	waitState(t, f.scheduler, 0, 0)
	w.Disconnect()
}
func TestQueueOrderAndProjectRunningLimit(t *testing.T) {
	f := setup(t, 4, 2, 1, 1)
	ctx := context.Background()
	started := make(chan string, 4)
	release := make(chan struct{})
	w := register(t, f, "n", 2, executorFunc(func(ctx context.Context, r execution.Request) (execution.Result, error) {
		started <- r.ID
		select {
		case <-release:
			return result(r, "n"), nil
		case <-ctx.Done():
			return execution.Result{}, ctx.Err()
		}
	}), runtime)
	one := submit(f.scheduler, ctx, f.request("one", "a"))
	if <-started != "one" {
		t.Fatal("first")
	}
	two := submit(f.scheduler, ctx, f.request("two", "a"))
	waitState(t, f.scheduler, 1, 1)
	three := submit(f.scheduler, ctx, f.request("three", "b"))
	if <-started != "three" {
		t.Fatal("quota-blocked head stalled other project")
	}
	waitState(t, f.scheduler, 2, 1)
	four := submit(f.scheduler, ctx, f.request("four", "c"))
	waitState(t, f.scheduler, 2, 2)
	release <- struct{}{}
	if id := <-started; id != "two" && id != "four" {
		t.Fatal(id)
	}
	close(release)
	for _, done := range []<-chan outcome{one, two, three, four} {
		if v := receive(t, done); v.err != nil {
			t.Fatal(v.err)
		}
	}
	waitState(t, f.scheduler, 0, 0)
	w.Disconnect()
}
func TestFailFastProjectQuota(t *testing.T) {
	f := setup(t, 0, 0, 1, 1)
	started := make(chan struct{})
	release := make(chan struct{})
	register(t, f, "n", 2, executorFunc(func(_ context.Context, r execution.Request) (execution.Result, error) {
		close(started)
		<-release
		return result(r, "n"), nil
	}), runtime)
	done := submit(f.scheduler, context.Background(), f.request("one", "a"))
	<-started
	if _, err := f.scheduler.Execute(context.Background(), f.request("two", "a")); !errors.Is(err, cortex.ErrProjectQuota) {
		t.Fatal(err)
	}
	close(release)
	if v := receive(t, done); v.err != nil {
		t.Fatal(v.err)
	}
}
func TestExpiryFencesOldSessionAndLateResult(t *testing.T) {
	f := setup(t, 0, 0, 4, 1)
	started := make(chan struct{})
	cancelled := make(chan struct{})
	release := make(chan struct{})
	old := register(t, f, "n", 1, executorFunc(func(ctx context.Context, r execution.Request) (execution.Result, error) {
		close(started)
		<-ctx.Done()
		close(cancelled)
		<-release
		return result(r, "n"), nil
	}), runtime)
	done := submit(f.scheduler, context.Background(), f.request("old", "a"))
	<-started
	f.clock.advance(time.Minute)
	f.scheduler.Expire()
	<-cancelled
	if len(f.scheduler.Snapshot().Neurons) != 0 {
		t.Fatal("expired capacity remains")
	}
	if err := old.Pulse(1); !errors.Is(err, cortex.ErrSessionLost) {
		t.Fatal(err)
	}
	newSession := register(t, f, "n", 1, fast("n"), runtime)
	old.Disconnect()
	if v, err := f.scheduler.Execute(context.Background(), f.request("new", "b")); err != nil || v.WorkerID != "n" {
		t.Fatal(v, err)
	}
	close(release)
	if v := receive(t, done); !errors.Is(v.err, cortex.ErrSessionLost) || v.result.ID != "" {
		t.Fatal(v)
	}
	newSession.Disconnect()
	waitState(t, f.scheduler, 0, 0)
}
func TestFourNeuronsAcrossProjects(t *testing.T) {
	f := setup(t, 0, 0, 4, 4)
	started := make(chan string, 4)
	release := make(chan struct{})
	for i := 0; i < 4; i++ {
		id := fmt.Sprintf("n%d", i)
		register(t, f, id, 1, executorFunc(func(_ context.Context, r execution.Request) (execution.Result, error) {
			started <- id
			<-release
			return result(r, id), nil
		}), runtime)
	}
	var calls []<-chan outcome
	for i := 0; i < 4; i++ {
		calls = append(calls, submit(f.scheduler, context.Background(), f.request(fmt.Sprint(i), fmt.Sprint(i%2))))
	}
	seen := map[string]bool{}
	for range calls {
		select {
		case id := <-started:
			seen[id] = true
		case <-time.After(time.Second):
			t.Fatal("not parallel")
		}
	}
	if len(seen) != 4 {
		t.Fatal(seen)
	}
	close(release)
	for _, call := range calls {
		if v := receive(t, call); v.err != nil {
			t.Fatal(v.err)
		}
	}
}
