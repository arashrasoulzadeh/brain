package synapse_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"brain/internal/cortex"
	"brain/internal/execution"
	"brain/internal/synapse"
)

func req(id string) execution.Request {
	return execution.Request{ID: id, TraceID: id, ProjectID: "alpha", Artifact: execution.Artifact{ID: "plugin", Version: "1.0.0", Digest: "sha256:" + strings.Repeat("a", 64), Entrypoint: "call"}, Runtime: execution.Runtime{Kind: "typescript", Profile: "isolated"}, Input: json.RawMessage(`null`), Context: json.RawMessage(`{}`), Deadline: time.Now().Add(5 * time.Second), Limits: execution.Limits{CPUTime: time.Second, MemoryBytes: 1024, InputBytes: 128, OutputBytes: 128}, Auth: execution.Authorization{ProjectID: "alpha", OriginalCaller: "user:1", Caller: "user:1", Capability: "call"}}
}
func newSession(t *testing.T, pending, queue int, attempts uint64, grace time.Duration) *synapse.Session {
	t.Helper()
	s, err := synapse.NewSession(synapse.Config{WorkerID: "neuron", SessionID: "session", MaxPending: pending, QueueSize: queue, MaxAssignments: attempts, CancelGrace: grace})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(s.Close)
	return s
}

type outcome struct {
	result execution.Result
	err    error
}

func submit(s execution.Executor, ctx context.Context, r execution.Request) <-chan outcome {
	done := make(chan outcome, 1)
	go func() { v, e := s.Execute(ctx, r); done <- outcome{v, e} }()
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
func next(t *testing.T, s *synapse.Session) synapse.Command {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	c, err := s.Next(ctx)
	if err != nil {
		t.Fatal(err)
	}
	return c
}
func waitPending(t *testing.T, s *synapse.Session, count int) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if s.Pending() == count {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatalf("pending=%d want %d", s.Pending(), count)
}
func success(id string) execution.Result {
	return execution.Result{ID: id, WorkerID: "neuron", Status: execution.Succeeded, Output: json.RawMessage(`null`)}
}
func TestVersionAndEnvelope(t *testing.T) {
	if v, err := synapse.Negotiate([]uint32{99, 1}); err != nil || v != 1 {
		t.Fatal(v, err)
	}
	for _, versions := range [][]uint32{nil, {2}, make([]uint32, 17)} {
		if _, err := synapse.Negotiate(versions); !errors.Is(err, synapse.ErrProtocol) {
			t.Fatal(err)
		}
	}
	if _, err := synapse.NewReceiver(""); !errors.Is(err, synapse.ErrProtocol) {
		t.Fatal(err)
	}
	r, err := synapse.NewReceiver("session")
	if err != nil {
		t.Fatal(err)
	}
	if err = r.Check(1, "session", 1, 128); err != nil {
		t.Fatal(err)
	}
	for _, v := range []struct {
		version  uint32
		session  string
		sequence uint64
		size     int
	}{{2, "session", 2, 128}, {1, "other", 2, 128}, {1, "session", 1, 128}, {1, "session", 0, 128}, {1, "session", 3, 128}, {1, "session", 2, 0}, {1, "session", 2, synapse.MaxMessageBytes + 1}} {
		if err = r.Check(v.version, v.session, v.sequence, v.size); !errors.Is(err, synapse.ErrProtocol) {
			t.Fatal(err)
		}
	}
	if err = r.Check(1, "session", 2, 128); err != nil {
		t.Fatal(err)
	}
	if _, err = synapse.NewSession(synapse.Config{}); !errors.Is(err, synapse.ErrProtocol) {
		t.Fatal(err)
	}
}
func TestAssignmentsAreFencedAndResultsAreCopied(t *testing.T) {
	s := newSession(t, 2, 2, 0, 0)
	ctx := context.Background()
	r := req("id")
	done := submit(s, ctx, r)
	waitPending(t, s, 1)
	guess := synapse.Lease{SessionID: "session", Attempt: 1, ExecutionID: "id"}
	if err := s.Accept(guess); !errors.Is(err, synapse.ErrProtocol) {
		t.Fatal(err)
	}
	if _, err := s.ArtifactFor(guess); !errors.Is(err, synapse.ErrProtocol) {
		t.Fatal(err)
	}
	a := next(t, s).Assignment
	if a == nil || a.Lease != guess {
		t.Fatal(a)
	}
	a.Request.Input[0] = 'X'
	a.Request.Artifact.ID = "tampered"
	if art, err := s.ArtifactFor(a.Lease); err != nil || art != r.Artifact {
		t.Fatal(art, err)
	}
	wrong := a.Lease
	wrong.SessionID = "old"
	if err := s.Accept(wrong); !errors.Is(err, synapse.ErrProtocol) {
		t.Fatal(err)
	}
	wrong = a.Lease
	wrong.ExecutionID = "other"
	if err := s.Accept(wrong); !errors.Is(err, synapse.ErrProtocol) {
		t.Fatal(err)
	}
	if err := s.Complete(a.Lease, success("id")); !errors.Is(err, synapse.ErrProtocol) {
		t.Fatal(err)
	}
	if err := s.Accept(a.Lease); err != nil {
		t.Fatal(err)
	}
	if err := s.Accept(a.Lease); !errors.Is(err, synapse.ErrProtocol) {
		t.Fatal(err)
	}
	bad := success("id")
	bad.WorkerID = "different"
	if err := s.Complete(a.Lease, bad); !errors.Is(err, execution.ErrInvalidResult) {
		t.Fatal(err)
	}
	v := success("id")
	v.Effects = []json.RawMessage{json.RawMessage(`{}`)}
	if err := s.Complete(a.Lease, v); err != nil {
		t.Fatal(err)
	}
	v.Output[0] = 'X'
	v.Effects[0][0] = 'X'
	if out := receive(t, done); out.err != nil || string(out.result.Output) != "null" || string(out.result.Effects[0]) != "{}" {
		t.Fatal(out)
	}
	if err := s.Complete(a.Lease, v); !errors.Is(err, synapse.ErrProtocol) {
		t.Fatal(err)
	}
	again := submit(s, ctx, r)
	b := next(t, s).Assignment
	if b.Lease.Attempt != 2 {
		t.Fatal(b)
	}
	if err := s.Accept(a.Lease); !errors.Is(err, synapse.ErrProtocol) {
		t.Fatal(err)
	}
	if err := s.Accept(b.Lease); err != nil {
		t.Fatal(err)
	}
	failure := execution.Result{ID: "id", WorkerID: "neuron", Status: execution.Failed, Error: &execution.Failure{Code: "failed"}}
	if err := s.Complete(b.Lease, failure); err != nil {
		t.Fatal(err)
	}
	failure.Error.Code = "tampered"
	if out := receive(t, again); out.err != nil || out.result.Error.Code != "failed" {
		t.Fatal(out)
	}
}
func TestLimitsAndClose(t *testing.T) {
	ctx := context.Background()
	s := newSession(t, 1, 1, 1, 0)
	r := req("one")
	bad := r
	bad.ID = ""
	if _, err := s.Execute(ctx, bad); !errors.Is(err, execution.ErrInvalidRequest) {
		t.Fatal(err)
	}
	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	if _, err := s.Execute(cancelled, r); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if _, err := s.Next(cancelled); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	done := submit(s, ctx, r)
	a := next(t, s).Assignment
	if _, err := s.Execute(ctx, r); !errors.Is(err, synapse.ErrDuplicate) {
		t.Fatal(err)
	}
	if _, err := s.Execute(ctx, req("two")); !errors.Is(err, synapse.ErrBackpressure) {
		t.Fatal(err)
	}
	if err := s.Accept(a.Lease); err != nil {
		t.Fatal(err)
	}
	if err := s.Complete(a.Lease, success("one")); err != nil {
		t.Fatal(err)
	}
	if out := receive(t, done); out.err != nil {
		t.Fatal(out)
	}
	if _, err := s.Execute(ctx, req("two")); !errors.Is(err, synapse.ErrExhausted) {
		t.Fatal(err)
	}
	s.Close()
	s.Close()
	<-s.Done()
	if _, err := s.Execute(ctx, r); !errors.Is(err, synapse.ErrClosed) {
		t.Fatal(err)
	}
	if _, err := s.Next(ctx); !errors.Is(err, synapse.ErrClosed) {
		t.Fatal(err)
	}
	s2 := newSession(t, 2, 1, 0, 0)
	blocked, stop := context.WithTimeout(ctx, 10*time.Millisecond)
	defer stop()
	if _, err := s2.Next(blocked); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal(err)
	}
	pending := submit(s2, ctx, r)
	waitPending(t, s2, 1)
	if _, err := s2.Execute(ctx, req("two")); !errors.Is(err, synapse.ErrBackpressure) {
		t.Fatal(err)
	}
	s2.Close()
	if out := receive(t, pending); !errors.Is(out.err, synapse.ErrClosed) {
		t.Fatal(out)
	}
}
func TestCancellationRetainsSlotUntilAcknowledged(t *testing.T) {
	s := newSession(t, 1, 2, 0, time.Second)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := submit(s, ctx, req("one"))
	a := next(t, s).Assignment
	cancel()
	c := next(t, s)
	if c.Cancel == nil || *c.Cancel != a.Lease {
		t.Fatal(c)
	}
	if s.Pending() != 1 {
		t.Fatal("released before acknowledgement")
	}
	if _, err := s.ArtifactFor(a.Lease); !errors.Is(err, synapse.ErrProtocol) {
		t.Fatal(err)
	}
	v := execution.Result{ID: "one", WorkerID: "neuron", Status: execution.Cancelled, Error: &execution.Failure{Code: "cancelled"}}
	if err := s.Complete(a.Lease, v); err != nil {
		t.Fatal(err)
	}
	if out := receive(t, done); !errors.Is(out.err, context.Canceled) || out.result.ID != "" {
		t.Fatal(out)
	}
	waitPending(t, s, 0)
}
func TestUndeliveredCancelAndGraceTimeout(t *testing.T) {
	s := newSession(t, 2, 2, 0, 30*time.Millisecond)
	ctx, cancel := context.WithCancel(context.Background())
	done := submit(s, ctx, req("one"))
	waitPending(t, s, 1)
	cancel()
	if out := receive(t, done); !errors.Is(out.err, context.Canceled) {
		t.Fatal(out)
	}
	ctx2, cancel2 := context.WithCancel(context.Background())
	defer cancel2()
	two := submit(s, ctx2, req("two"))
	a := next(t, s).Assignment
	if a.Lease.ExecutionID != "two" {
		t.Fatal("cancelled assignment dispatched")
	}
	cancel2()
	if next(t, s).Cancel == nil {
		t.Fatal("no cancel")
	}
	if out := receive(t, two); !errors.Is(out.err, context.Canceled) {
		t.Fatal(out)
	}
	<-s.Done()
}
func TestCancellationWhenOutboundFullClosesStream(t *testing.T) {
	s := newSession(t, 2, 1, 0, time.Second)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	one := submit(s, ctx, req("one"))
	next(t, s)
	two := submit(s, context.Background(), req("two"))
	waitPending(t, s, 2)
	cancel()
	if out := receive(t, one); !errors.Is(out.err, context.Canceled) {
		t.Fatal(out)
	}
	if out := receive(t, two); !errors.Is(out.err, synapse.ErrClosed) {
		t.Fatal(out)
	}
	<-s.Done()
}
func TestCancellationInterruptedByDisconnect(t *testing.T) {
	s := newSession(t, 1, 2, 0, time.Second)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	one := submit(s, ctx, req("one"))
	next(t, s)
	cancel()
	next(t, s)
	s.Close()
	if out := receive(t, one); !errors.Is(out.err, context.Canceled) {
		t.Fatal(out)
	}
}
func TestCortexSessionAdapter(t *testing.T) {
	catalog, err := execution.NewCatalog(1, 4096)
	if err != nil {
		t.Fatal(err)
	}
	bundle, err := execution.BuildBundle([]execution.BundleFile{{Path: "main.ts"}}, 4096)
	if err != nil {
		t.Fatal(err)
	}
	r := req("one")
	key := execution.ArtifactKey{ID: r.Artifact.ID, Version: r.Artifact.Version}
	if err = catalog.Register(execution.Registration{Key: key, Runtime: r.Runtime, Entrypoints: []string{r.Artifact.Entrypoint}, Bundle: bundle}); err != nil {
		t.Fatal(err)
	}
	r.Artifact, _, err = catalog.Resolve(key, r.Artifact.Entrypoint)
	if err != nil {
		t.Fatal(err)
	}
	scheduler, err := cortex.New(catalog, cortex.Config{MaxNeurons: 1, MaxProjectRunning: 1, HeartbeatTTL: time.Minute})
	if err != nil {
		t.Fatal(err)
	}
	s := newSession(t, 1, 2, 0, time.Second)
	w, err := scheduler.Register(cortex.Registration{ID: "neuron", Slots: 1, Runtimes: []execution.Runtime{r.Runtime}, Executor: s})
	if err != nil {
		t.Fatal(err)
	}
	defer w.Disconnect()
	done := submit(scheduler, context.Background(), r)
	a := next(t, s).Assignment
	if a.Request.Artifact != r.Artifact {
		t.Fatal(a)
	}
	if err := s.Accept(a.Lease); err != nil {
		t.Fatal(err)
	}
	if err := s.Complete(a.Lease, success(r.ID)); err != nil {
		t.Fatal(err)
	}
	if out := receive(t, done); out.err != nil {
		t.Fatal(out)
	}
}
