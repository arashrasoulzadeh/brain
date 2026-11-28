package synapse_test

import (
	"context"
	"crypto/tls"
	"errors"
	"math"
	"net"
	"testing"
	"time"

	"brain/internal/execution"
	"brain/internal/neuron"
	"brain/internal/synapse"
	pb "brain/proto/synapse/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/protobuf/proto"
)

// A scripted authenticated peer deliberately violates the Brain-side contract.
// No application host or test adapter is exposed to untrusted clients.
type scriptedBrain struct {
	pb.UnimplementedSynapseServer
	connect func(pb.Synapse_ConnectServer) error
}

func (b scriptedBrain) Connect(stream pb.Synapse_ConnectServer) error {
	return b.connect(stream)
}

func clientAgainstBrain(t *testing.T, slots int, script func(pb.Synapse_ConnectServer) error, runner wireRuntime) error {
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
	server := grpc.NewServer(grpc.Creds(credentials.NewTLS(serverTLS)))
	pb.RegisterSynapseServer(server, scriptedBrain{connect: script})
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	go server.Serve(listener)
	defer server.Stop()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cfg := neuron.Config{ID: "worker", Slots: slots, MaxArtifactBytes: 1 << 20, Adapters: []neuron.Adapter{{Capability: execution.Runtime{Kind: "typescript", Profile: "isolated"}, Runtime: runner}}}
	return synapse.RunNeuron(ctx, listener.Addr().String(), clientTLS, cfg)
}

func registeredMessage() *pb.BrainMessage {
	return &pb.BrainMessage{Body: &pb.BrainMessage_Registered{Registered: &pb.Registered{Version: 1, SessionId: "session", HeartbeatIntervalMs: 1000, MaxMessageBytes: synapse.MaxMessageBytes}}}
}

func beginBrain(stream pb.Synapse_ConnectServer) error {
	if _, err := stream.Recv(); err != nil {
		return err
	}
	return stream.Send(registeredMessage())
}

func brainMessage(seq uint64, body *pb.BrainMessage) *pb.BrainMessage {
	body.Version, body.SessionId, body.Sequence = 1, "session", seq
	return body
}

func wireAssignment(bundle []byte, attempt uint64, id string) *pb.Assignment {
	return &pb.Assignment{
		Lease:   &pb.Lease{SessionId: "session", Attempt: attempt, ExecutionId: id},
		TraceId: id, ProjectId: "alpha",
		Artifact:  &pb.Artifact{Id: "plugin", Version: "1.0.0", Sha256Digest: execution.ArtifactDigest(bundle), Entrypoint: "call"},
		Runtime:   &pb.Runtime{Kind: "typescript", IsolationProfile: "isolated"},
		InputJson: []byte("null"), DeadlineUnixNanos: time.Now().Add(8 * time.Second).UnixNano(),
		Limits:        &pb.Limits{CpuNanos: 1000000000, MemoryBytes: 1024, InputBytes: 128, OutputBytes: 128},
		Authorization: &pb.Authorization{OriginalCaller: "user:1", Caller: "user:1", Capability: "call"},
	}
}

func noRuntime(t *testing.T) wireRuntime {
	return func(context.Context, execution.Request, []execution.BundleFile) (execution.Result, error) {
		t.Error("malformed or cancelled work reached the runtime")
		return execution.Result{}, errors.New("unexpected invocation")
	}
}

func TestNeuronRejectsMalformedBrainHandshake(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mutate func(*pb.BrainMessage)
	}{
		{"wrong envelope", func(m *pb.BrainMessage) { m.Sequence = 1 }},
		{"wrong protocol", func(m *pb.BrainMessage) { m.GetRegistered().Version = 2 }},
		{"empty session", func(m *pb.BrainMessage) { m.GetRegistered().SessionId = "" }},
		{"wrong size", func(m *pb.BrainMessage) { m.GetRegistered().MaxMessageBytes = 1 }},
		{"wrong interval", func(m *pb.BrainMessage) { m.GetRegistered().HeartbeatIntervalMs = 0 }},
		{"wrong body", func(m *pb.BrainMessage) { m.Body = nil }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := clientAgainstBrain(t, 1, func(stream pb.Synapse_ConnectServer) error {
				if _, err := stream.Recv(); err != nil {
					return err
				}
				m := registeredMessage()
				tc.mutate(m)
				return stream.Send(m)
			}, noRuntime(t))
			if !errors.Is(err, synapse.ErrProtocol) {
				t.Fatal(err)
			}
		})
	}
	invalid := &tls.Config{InsecureSkipVerify: true}
	if err := synapse.RunNeuron(context.Background(), "127.0.0.1:1", invalid, neuron.Config{}); !errors.Is(err, synapse.ErrProtocol) {
		t.Fatal(err)
	}
}

func TestNeuronRejectsMalformedBrainMessages(t *testing.T) {
	bundle, err := execution.BuildBundle([]execution.BundleFile{{Path: "index.ts", Data: []byte("fixture")}}, 4096)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name   string
		mutate func(*pb.Assignment)
	}{
		{"missing artifact", func(a *pb.Assignment) { a.Artifact = nil }},
		{"overflow limit", func(a *pb.Assignment) { a.Limits.CpuNanos = math.MaxUint64 }},
		{"invalid lease", func(a *pb.Assignment) { a.Lease.Attempt = 0 }},
		{"invalid runtime", func(a *pb.Assignment) { a.Runtime.Kind = "" }},
		{"invalid execution", func(a *pb.Assignment) { a.TraceId = "" }},
		{"wrong lease session", func(a *pb.Assignment) { a.Lease.SessionId = "other" }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := clientAgainstBrain(t, 1, func(stream pb.Synapse_ConnectServer) error {
				if err := beginBrain(stream); err != nil {
					return err
				}
				a := wireAssignment(bundle, 1, "job")
				tc.mutate(a)
				if err := stream.Send(brainMessage(1, &pb.BrainMessage{Body: &pb.BrainMessage_Assignment{Assignment: a}})); err != nil {
					return err
				}
				// Wait for the client to reject the message and close its session.
				_, err := stream.Recv()
				return err
			}, noRuntime(t))
			if err == nil || errors.Is(err, context.DeadlineExceeded) {
				t.Fatal("malformed assignment not rejected", err)
			}
		})
	}
	for _, tc := range []struct {
		name string
		msg  *pb.BrainMessage
	}{
		{"unknown body", &pb.BrainMessage{}},
		{"missing cancel lease", &pb.BrainMessage{Body: &pb.BrainMessage_Cancel{Cancel: &pb.Cancel{}}}},
		{"future cancel", &pb.BrainMessage{Body: &pb.BrainMessage_Cancel{Cancel: &pb.Cancel{Lease: &pb.Lease{SessionId: "session", Attempt: 1, ExecutionId: "job"}}}}},
		{"missing chunk lease", &pb.BrainMessage{Body: &pb.BrainMessage_ArtifactChunk{ArtifactChunk: &pb.ArtifactChunk{}}}},
		{"overflow offset", &pb.BrainMessage{Body: &pb.BrainMessage_ArtifactChunk{ArtifactChunk: &pb.ArtifactChunk{Offset: math.MaxUint64}}}},
		{"future chunk", &pb.BrainMessage{Body: &pb.BrainMessage_ArtifactChunk{ArtifactChunk: &pb.ArtifactChunk{Lease: &pb.Lease{SessionId: "session", Attempt: 1, ExecutionId: "job"}}}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := clientAgainstBrain(t, 1, func(stream pb.Synapse_ConnectServer) error {
				if err := beginBrain(stream); err != nil {
					return err
				}
				if err := stream.Send(brainMessage(1, tc.msg)); err != nil {
					return err
				}
				_, err := stream.Recv()
				return err
			}, noRuntime(t))
			if !errors.Is(err, synapse.ErrProtocol) {
				t.Fatal(err)
			}
		})
	}
	err = clientAgainstBrain(t, 1, func(stream pb.Synapse_ConnectServer) error {
		if err := beginBrain(stream); err != nil {
			return err
		}
		return stream.Send(brainMessage(2, &pb.BrainMessage{Body: &pb.BrainMessage_Drain{Drain: &pb.Drain{}}}))
	}, noRuntime(t))
	if !errors.Is(err, synapse.ErrProtocol) {
		t.Fatal("sequence gap accepted", err)
	}
}

func receiveJobMessage(stream pb.Synapse_ConnectServer) (*pb.NeuronMessage, error) {
	for {
		m, err := stream.Recv()
		if err != nil || m.GetPulse() == nil {
			return m, err
		}
	}
}

func sendJobAndWaitFetch(stream pb.Synapse_ConnectServer, a *pb.Assignment) error {
	if err := stream.Send(brainMessage(1, &pb.BrainMessage{Body: &pb.BrainMessage_Assignment{Assignment: a}})); err != nil {
		return err
	}
	accepted, err := receiveJobMessage(stream)
	if err != nil {
		return err
	}
	if !proto.Equal(accepted.GetAccepted().GetLease(), a.Lease) {
		return errors.New("assignment not accepted")
	}
	fetch, err := receiveJobMessage(stream)
	if err != nil {
		return err
	}
	if !proto.Equal(fetch.GetArtifactRequest().GetLease(), a.Lease) {
		return errors.New("artifact request not bound to assignment")
	}
	return nil
}

func TestNeuronRejectsLiveLeaseAndArtifactSubstitution(t *testing.T) {
	bundle, err := execution.BuildBundle([]execution.BundleFile{{Path: "index.ts", Data: []byte("fixture")}}, 4096)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name    string
		slots   int
		message func(*pb.Assignment) *pb.BrainMessage
		want    error
	}{
		{"duplicate attempt", 2, func(a *pb.Assignment) *pb.BrainMessage {
			return &pb.BrainMessage{Body: &pb.BrainMessage_Assignment{Assignment: a}}
		}, synapse.ErrProtocol},
		{"duplicate execution", 2, func(a *pb.Assignment) *pb.BrainMessage {
			return &pb.BrainMessage{Body: &pb.BrainMessage_Assignment{Assignment: wireAssignment(bundle, 2, a.Lease.ExecutionId)}}
		}, synapse.ErrProtocol},
		{"slot overflow", 1, func(a *pb.Assignment) *pb.BrainMessage {
			return &pb.BrainMessage{Body: &pb.BrainMessage_Assignment{Assignment: wireAssignment(bundle, 2, "other")}}
		}, synapse.ErrProtocol},
		{"cancel substitution", 1, func(a *pb.Assignment) *pb.BrainMessage {
			return &pb.BrainMessage{Body: &pb.BrainMessage_Cancel{Cancel: &pb.Cancel{Lease: &pb.Lease{SessionId: "session", Attempt: 1, ExecutionId: "other"}}}}
		}, synapse.ErrProtocol},
		{"chunk substitution", 1, func(a *pb.Assignment) *pb.BrainMessage {
			return &pb.BrainMessage{Body: &pb.BrainMessage_ArtifactChunk{ArtifactChunk: &pb.ArtifactChunk{Lease: &pb.Lease{SessionId: "session", Attempt: 1, ExecutionId: "other"}}}}
		}, synapse.ErrProtocol},
		{"invalid offset", 1, func(a *pb.Assignment) *pb.BrainMessage {
			return &pb.BrainMessage{Body: &pb.BrainMessage_ArtifactChunk{ArtifactChunk: &pb.ArtifactChunk{Lease: a.Lease, Offset: 1, TotalBytes: 2, Data: []byte("x"), FinalChunk: true}}}
		}, synapse.ErrInvalidChunk},
		{"tampered bytes", 1, func(a *pb.Assignment) *pb.BrainMessage {
			return &pb.BrainMessage{Body: &pb.BrainMessage_ArtifactChunk{ArtifactChunk: &pb.ArtifactChunk{Lease: a.Lease, TotalBytes: 1, Data: []byte("x"), FinalChunk: true}}}
		}, execution.ErrArtifactMismatch},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := clientAgainstBrain(t, tc.slots, func(stream pb.Synapse_ConnectServer) error {
				if err := beginBrain(stream); err != nil {
					return err
				}
				a := wireAssignment(bundle, 1, "job")
				if err := sendJobAndWaitFetch(stream, a); err != nil {
					return err
				}
				if err := stream.Send(brainMessage(2, tc.message(a))); err != nil {
					return err
				}
				_, err := stream.Recv()
				return err
			}, noRuntime(t))
			if !errors.Is(err, tc.want) {
				t.Fatal(err)
			}
		})
	}
}

func TestNeuronCancelsArtifactFetchAndDrains(t *testing.T) {
	bundle, err := execution.BuildBundle([]execution.BundleFile{{Path: "index.ts", Data: []byte("fixture")}}, 4096)
	if err != nil {
		t.Fatal(err)
	}
	drained := make(chan struct{})
	err = clientAgainstBrain(t, 1, func(stream pb.Synapse_ConnectServer) error {
		if err := beginBrain(stream); err != nil {
			return err
		}
		a := wireAssignment(bundle, 1, "job")
		if err := sendJobAndWaitFetch(stream, a); err != nil {
			return err
		}
		cancel := func(seq uint64) error {
			return stream.Send(brainMessage(seq, &pb.BrainMessage{Body: &pb.BrainMessage_Cancel{Cancel: &pb.Cancel{Lease: a.Lease}}}))
		}
		if err := cancel(2); err != nil {
			return err
		}
		result, err := receiveJobMessage(stream)
		if err != nil || result.GetResult().GetStatus() != pb.Result_CANCELLED {
			return errors.New("cancelled fetch did not report terminal cancellation")
		}
		// In-flight cancellation and chunks can cross a completed job.
		if err := cancel(3); err != nil {
			return err
		}
		chunk := &pb.ArtifactChunk{Lease: a.Lease, TotalBytes: uint64(len(bundle)), Data: bundle, FinalChunk: true}
		if err := stream.Send(brainMessage(4, &pb.BrainMessage{Body: &pb.BrainMessage_ArtifactChunk{ArtifactChunk: chunk}})); err != nil {
			return err
		}
		if err := stream.Send(brainMessage(5, &pb.BrainMessage{Body: &pb.BrainMessage_Drain{Drain: &pb.Drain{}}})); err != nil {
			return err
		}
		gotDrained, gotPulse := false, false
		for !gotDrained || !gotPulse {
			message, err := stream.Recv()
			if err != nil {
				return err
			}
			if message.GetDrained() != nil {
				gotDrained = true
			}
			if p := message.GetPulse(); p != nil && p.Slots == 0 {
				gotPulse = true
			}
		}
		close(drained)
		return nil
	}, noRuntime(t))
	select {
	case <-drained:
	default:
		t.Fatal("draining Neuron did not finish and withdraw capacity", err)
	}
}
