package synapse_test

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"math"
	"net"
	"testing"
	"time"

	"brain/internal/cortex"
	"brain/internal/execution"
	"brain/internal/neuron"
	"brain/internal/synapse"
	pb "brain/proto/synapse/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
)

type grpcFixture struct {
	address   string
	tls       *tls.Config
	scheduler *cortex.Scheduler
	artifact  execution.Artifact
	runtime   execution.Runtime
}

func newGRPCFixture(t *testing.T) grpcFixture {
	t.Helper()
	a := ca(t)
	serverTLS, err := synapse.ServerTLS(leaf(t, a, true), a.pool)
	if err != nil {
		t.Fatal(err)
	}
	clientTLS, err := synapse.ClientTLS(leaf(t, a, false, "spiffe://brain/neurons/worker"), a.pool, "brain.test")
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
	if _, err := synapse.NewGRPCServer(nil, scheduler, catalog, time.Second); !errors.Is(err, synapse.ErrProtocol) {
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
	t.Cleanup(server.Stop)
	return grpcFixture{listener.Addr().String(), clientTLS, scheduler, artifact, rt}
}

func (f grpcFixture) stream(t *testing.T) pb.Synapse_ConnectClient {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	t.Cleanup(cancel)
	conn, err := grpc.NewClient(f.address, grpc.WithTransportCredentials(credentials.NewTLS(f.tls)))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	stream, err := pb.NewSynapseClient(conn).Connect(ctx)
	if err != nil {
		t.Fatal(err)
	}
	return stream
}

func registration() *pb.NeuronMessage {
	return &pb.NeuronMessage{Body: &pb.NeuronMessage_Register{Register: &pb.Register{NeuronId: "worker", SupportedVersions: []uint32{1}, Slots: 1, Runtimes: []*pb.Runtime{{Kind: "typescript", IsolationProfile: "isolated"}}}}}
}

func registerWire(t *testing.T, stream pb.Synapse_ConnectClient) string {
	t.Helper()
	if err := stream.Send(registration()); err != nil {
		t.Fatal(err)
	}
	m, err := stream.Recv()
	if err != nil || m.GetRegistered() == nil || m.GetRegistered().SessionId == "" {
		t.Fatal(m, err)
	}
	return m.GetRegistered().SessionId
}

func expectWireError(t *testing.T, stream pb.Synapse_ConnectClient, code codes.Code) {
	t.Helper()
	if _, err := stream.Recv(); status.Code(err) != code {
		t.Fatalf("status %v, want %v", err, code)
	}
}

func (f grpcFixture) waitNeurons(t *testing.T, n int) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if len(f.scheduler.Snapshot().Neurons) == n {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("registered Neurons: %d, want %d", len(f.scheduler.Snapshot().Neurons), n)
}

func TestGRPCRegistrationAndEnvelopeRejections(t *testing.T) {
	f := newGRPCFixture(t)
	for _, tc := range []struct {
		name   string
		mutate func(*pb.NeuronMessage)
		code   codes.Code
	}{
		{"missing register", func(m *pb.NeuronMessage) { m.Body = nil }, codes.InvalidArgument},
		{"wrong identity", func(m *pb.NeuronMessage) { m.GetRegister().NeuronId = "other" }, codes.InvalidArgument},
		{"version in envelope", func(m *pb.NeuronMessage) { m.Version = 1 }, codes.InvalidArgument},
		{"zero slots", func(m *pb.NeuronMessage) { m.GetRegister().Slots = 0 }, codes.InvalidArgument},
		{"unsupported version", func(m *pb.NeuronMessage) { m.GetRegister().SupportedVersions = []uint32{2} }, codes.FailedPrecondition},
		{"invalid runtime", func(m *pb.NeuronMessage) { m.GetRegister().Runtimes[0].IsolationProfile = "" }, codes.InvalidArgument},
	} {
		t.Run(tc.name, func(t *testing.T) {
			stream := f.stream(t)
			m := registration()
			tc.mutate(m)
			if err := stream.Send(m); err != nil {
				t.Fatal(err)
			}
			expectWireError(t, stream, tc.code)
			f.waitNeurons(t, 0)
		})
	}
	stream := f.stream(t)
	session := registerWire(t, stream)
	f.waitNeurons(t, 1)
	duplicate := f.stream(t)
	if err := duplicate.Send(registration()); err != nil {
		t.Fatal(err)
	}
	expectWireError(t, duplicate, codes.ResourceExhausted)
	f.waitNeurons(t, 1)
	pulse := &pb.NeuronMessage{Version: 1, SessionId: session, Sequence: 1, Body: &pb.NeuronMessage_Pulse{Pulse: &pb.Pulse{Slots: 1}}}
	if err := stream.Send(pulse); err != nil {
		t.Fatal(err)
	}
	if err := stream.Send(pulse); err != nil {
		t.Fatal(err)
	}
	expectWireError(t, stream, codes.InvalidArgument)
	f.waitNeurons(t, 0)
}

func TestGRPCRejectsMalformedResultsAndUnauthorizedArtifactRequests(t *testing.T) {
	f := newGRPCFixture(t)
	for _, tc := range []struct {
		name string
		msg  *pb.NeuronMessage
	}{
		{"inflated capacity", &pb.NeuronMessage{Body: &pb.NeuronMessage_Pulse{Pulse: &pb.Pulse{Slots: 2}}}},
		{"missing accepted lease", &pb.NeuronMessage{Body: &pb.NeuronMessage_Accepted{Accepted: &pb.Accepted{}}}},
		{"missing result lease", &pb.NeuronMessage{Body: &pb.NeuronMessage_Result{Result: &pb.Result{Status: pb.Result_SUCCEEDED}}}},
		{"overflow result", &pb.NeuronMessage{Body: &pb.NeuronMessage_Result{Result: &pb.Result{DurationNanos: math.MaxUint64}}}},
		{"missing artifact lease", &pb.NeuronMessage{Body: &pb.NeuronMessage_ArtifactRequest{ArtifactRequest: &pb.ArtifactRequest{}}}},
		{"unassigned artifact", &pb.NeuronMessage{Body: &pb.NeuronMessage_ArtifactRequest{ArtifactRequest: &pb.ArtifactRequest{Lease: &pb.Lease{SessionId: "other", ExecutionId: "guess", Attempt: 1}}}}},
		{"unknown message", &pb.NeuronMessage{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			stream := f.stream(t)
			session := registerWire(t, stream)
			tc.msg.Version, tc.msg.SessionId, tc.msg.Sequence = 1, session, 1
			if err := stream.Send(tc.msg); err != nil {
				t.Fatal(err)
			}
			expectWireError(t, stream, codes.InvalidArgument)
			f.waitNeurons(t, 0)
		})
	}
	stream := f.stream(t)
	session := registerWire(t, stream)
	message := &pb.NeuronMessage{Version: 1, SessionId: session, Sequence: 1, Body: &pb.NeuronMessage_Drained{Drained: &pb.Drained{}}}
	if err := stream.Send(message); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		snapshot := f.scheduler.Snapshot()
		if len(snapshot.Neurons) == 1 && snapshot.Neurons[0].Draining {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("draining status was not recorded")
}

func TestGRPCHeartbeatExpiryAndRegistrationTimeout(t *testing.T) {
	f := newGRPCFixture(t)
	t.Run("missing registration", func(t *testing.T) {
		stream := f.stream(t)
		// Send a header to initiate the RPC without an application message.
		expectWireError(t, stream, codes.DeadlineExceeded)
		f.waitNeurons(t, 0)
	})
	t.Run("missing heartbeat", func(t *testing.T) {
		stream := f.stream(t)
		registerWire(t, stream)
		f.waitNeurons(t, 1)
		expectWireError(t, stream, codes.DeadlineExceeded)
		f.waitNeurons(t, 0)
	})
}

func TestGRPCFailureEffectsAndCancellation(t *testing.T) {
	f := newGRPCFixture(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	started := make(chan struct{})
	stopped := make(chan struct{})
	adapter := wireRuntime(func(ctx context.Context, r execution.Request, _ []execution.BundleFile) (execution.Result, error) {
		switch r.ID {
		case "error":
			return execution.Result{}, errors.New("private runtime failure")
		case "failure":
			return execution.Result{Status: execution.Failed, Error: &execution.Failure{Code: "denied", Message: "safe failure"}}, nil
		case "cancel":
			close(started)
			<-ctx.Done()
			close(stopped)
			return execution.Result{}, ctx.Err()
		default:
			return execution.Result{Status: execution.Succeeded, Output: json.RawMessage("42"), Effects: []json.RawMessage{json.RawMessage("{}")}}, nil
		}
	})
	cfg := neuron.Config{ID: "worker", Slots: 1, MaxArtifactBytes: 1 << 20, Adapters: []neuron.Adapter{{Capability: f.runtime, Runtime: adapter}}}
	if err := synapse.RunNeuron(ctx, f.address, nil, cfg); !errors.Is(err, synapse.ErrProtocol) {
		t.Fatal(err)
	}
	invalid := cfg
	invalid.Slots = 0
	if err := synapse.RunNeuron(ctx, f.address, f.tls, invalid); err == nil {
		t.Fatal("invalid configuration accepted")
	}
	done := make(chan error, 1)
	go func() { done <- synapse.RunNeuron(ctx, f.address, f.tls, cfg) }()
	defer func() {
		cancel()
		select {
		case <-done:
		case <-time.After(3 * time.Second):
			t.Error("worker did not shut down")
		}
	}()
	f.waitNeurons(t, 1)
	for _, id := range []string{"error", "failure", "effects"} {
		r := req(id)
		r.Artifact, r.Runtime = f.artifact, f.runtime
		result, err := f.scheduler.Execute(ctx, r)
		if err != nil {
			t.Fatal(id, err)
		}
		switch id {
		case "error":
			if result.Status != execution.Failed || result.Error == nil || result.Error.Code != "runtime_error" || result.Error.Message != "" {
				t.Fatal(result)
			}
		case "failure":
			if result.Status != execution.Failed || result.Error == nil || result.Error.Code != "denied" {
				t.Fatal(result)
			}
		default:
			if result.Status != execution.Succeeded || len(result.Effects) != 1 || string(result.Effects[0]) != "{}" {
				t.Fatal(result)
			}
		}
	}
	r := req("cancel")
	r.Artifact, r.Runtime = f.artifact, f.runtime
	callCtx, stopCall := context.WithCancel(ctx)
	defer stopCall()
	result := submit(f.scheduler, callCtx, r)
	select {
	case <-started:
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}
	stopCall()
	select {
	case <-stopped:
	case <-ctx.Done():
		t.Fatal("runtime did not observe cancellation")
	}
	if v := receive(t, result); !errors.Is(v.err, context.Canceled) {
		t.Fatal(v)
	}
	// The same session must still be usable after a cancellation acknowledgement.
	r = req("after-cancel")
	r.Artifact, r.Runtime = f.artifact, f.runtime
	if v, err := f.scheduler.Execute(ctx, r); err != nil || v.Status != execution.Succeeded {
		t.Fatal(v, err)
	}
}
