package synapse

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/hex"
	"errors"
	"io"
	"sync"
	"time"

	"brain/internal/cortex"
	"brain/internal/execution"
	pb "brain/proto/synapse/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

// RuntimeDigestHeader identifies the enrolled runtime source distribution.
const RuntimeDigestHeader = "brain-runtime-digest"

// NewGRPCServer exposes only authenticated, bounded Synapse v1 streams.
func NewGRPCServer(tlsConfig *tls.Config, scheduler *cortex.Scheduler, catalog *execution.Catalog, heartbeat time.Duration) (*grpc.Server, error) {
	return newGRPCServer(tlsConfig, scheduler, catalog, heartbeat, "")
}

// NewEnrolledGRPCServer additionally requires the runtime-source compatibility
// digest on each authenticated stream. Use EnrolledServerTLS for leaf pinning.
func NewEnrolledGRPCServer(tlsConfig *tls.Config, scheduler *cortex.Scheduler, catalog *execution.Catalog, heartbeat time.Duration, runtimeDigest string) (*grpc.Server, error) {
	if !validDigest(runtimeDigest) {
		return nil, ErrProtocol
	}
	return newGRPCServer(tlsConfig, scheduler, catalog, heartbeat, runtimeDigest)
}

func newGRPCServer(tlsConfig *tls.Config, scheduler *cortex.Scheduler, catalog *execution.Catalog, heartbeat time.Duration, runtimeDigest string) (*grpc.Server, error) {
	if tlsConfig == nil || tlsConfig.ClientAuth != tls.RequireAndVerifyClientCert || tlsConfig.MinVersion < tls.VersionTLS13 || tlsConfig.ClientCAs == nil || scheduler == nil || catalog == nil || heartbeat < time.Second || heartbeat > 30*time.Second {
		return nil, ErrProtocol
	}
	server := grpc.NewServer(grpc.Creds(credentials.NewTLS(tlsConfig)), grpc.MaxRecvMsgSize(MaxMessageBytes), grpc.MaxSendMsgSize(MaxMessageBytes), grpc.MaxConcurrentStreams(32))
	pb.RegisterSynapseServer(server, &grpcHost{scheduler: scheduler, catalog: catalog, heartbeat: heartbeat, runtimeDigest: runtimeDigest})
	return server, nil
}

type grpcHost struct {
	pb.UnimplementedSynapseServer
	scheduler     *cortex.Scheduler
	catalog       *execution.Catalog
	heartbeat     time.Duration
	runtimeDigest string
}

func (h *grpcHost) Connect(stream pb.Synapse_ConnectServer) error {
	ctx, cancel := context.WithCancel(stream.Context())
	defer cancel()
	p, ok := peer.FromContext(ctx)
	if !ok {
		return status.Error(codes.Unauthenticated, "missing neuron identity")
	}
	info, ok := p.AuthInfo.(credentials.TLSInfo)
	if !ok {
		return status.Error(codes.Unauthenticated, "missing neuron certificate")
	}
	id, err := NeuronIdentity(info.State)
	if err != nil {
		return status.Error(codes.Unauthenticated, "invalid neuron certificate")
	}
	// A TLS connection must not keep an expired leaf enrolled indefinitely.
	ctx, expire := context.WithDeadline(ctx, info.State.PeerCertificates[0].NotAfter)
	defer expire()
	if h.runtimeDigest != "" {
		values := metadata.ValueFromIncomingContext(ctx, RuntimeDigestHeader)
		if len(values) != 1 || values[0] != h.runtimeDigest {
			return status.Error(codes.FailedPrecondition, "incompatible enrolled runtime")
		}
	}
	// Bound unauthenticated application handshakes as well as message sizes.
	type firstMessage struct {
		message *pb.NeuronMessage
		err     error
	}
	first := make(chan firstMessage, 1)
	go func() { m, e := stream.Recv(); first <- firstMessage{m, e} }()
	var message *pb.NeuronMessage
	timer := time.NewTimer(5 * time.Second)
	defer timer.Stop()
	select {
	case v := <-first:
		if v.err != nil {
			return v.err
		}
		message = v.message
	case <-timer.C:
		return status.Error(codes.DeadlineExceeded, "registration timeout")
	case <-ctx.Done():
		return ctx.Err()
	}
	reg := message.GetRegister()
	if reg == nil || message.Version != 0 || message.SessionId != "" || message.Sequence != 0 || reg.NeuronId != id || reg.Slots == 0 || reg.Slots > maxPending || len(reg.Runtimes) == 0 || len(reg.Runtimes) > 32 {
		return status.Error(codes.InvalidArgument, "invalid registration")
	}
	if _, err = Negotiate(reg.SupportedVersions); err != nil {
		return status.Error(codes.FailedPrecondition, "incompatible protocol")
	}
	var runtimes []execution.Runtime
	for _, value := range reg.Runtimes {
		runtime, e := readRuntime(value)
		if e != nil {
			return status.Error(codes.InvalidArgument, "invalid runtime")
		}
		runtimes = append(runtimes, runtime)
	}
	random := make([]byte, 24)
	if _, err = io.ReadFull(rand.Reader, random); err != nil {
		return status.Error(codes.Internal, "session unavailable")
	}
	sessionID := hex.EncodeToString(random)
	session, err := NewSession(Config{WorkerID: id, SessionID: sessionID, MaxPending: int(reg.Slots), QueueSize: int(reg.Slots) * 2})
	if err != nil {
		return status.Error(codes.InvalidArgument, "invalid session")
	}
	defer session.Close()
	worker, err := h.scheduler.Register(cortex.Registration{ID: id, Runtimes: runtimes, Slots: int(reg.Slots), Executor: session})
	if err != nil {
		return status.Error(codes.ResourceExhausted, "registration rejected")
	}
	defer worker.Disconnect()
	if err = stream.Send(&pb.BrainMessage{Body: &pb.BrainMessage_Registered{Registered: &pb.Registered{Version: Version, SessionId: sessionID, HeartbeatIntervalMs: uint32(h.heartbeat.Milliseconds()), MaxMessageBytes: MaxMessageBytes}}}); err != nil {
		return err
	}
	receiver, err := NewReceiver(sessionID)
	if err != nil {
		return err
	}
	var sendMu sync.Mutex
	var sequence uint64
	send := func(m *pb.BrainMessage) error {
		sendMu.Lock()
		defer sendMu.Unlock()
		if err := ctx.Err(); err != nil {
			return err
		}
		sequence++
		m.Version, m.SessionId, m.Sequence = Version, sessionID, sequence
		return stream.Send(m)
	}
	failures := make(chan error, 2)
	pulses := make(chan struct{}, 1)
	go func() {
		for {
			cmd, e := session.Next(ctx)
			if e != nil {
				failures <- e
				return
			}
			var m *pb.BrainMessage
			if cmd.Assignment != nil {
				m = &pb.BrainMessage{Body: &pb.BrainMessage_Assignment{Assignment: assignmentWire(*cmd.Assignment)}}
			} else {
				m = &pb.BrainMessage{Body: &pb.BrainMessage_Cancel{Cancel: &pb.Cancel{Lease: leaseWire(*cmd.Cancel)}}}
			}
			if e = send(m); e != nil {
				failures <- e
				return
			}
		}
	}()
	go func() {
		for {
			m, e := stream.Recv()
			if e != nil {
				failures <- e
				return
			}
			if e = receiver.Check(m.Version, m.SessionId, m.Sequence, proto.Size(m)); e == nil {
				e = h.receive(ctx, session, worker, id, int(reg.Slots), m, send)
			}
			if e != nil {
				failures <- status.Error(codes.InvalidArgument, "invalid synapse message")
				return
			}
			if m.GetPulse() != nil {
				select {
				case pulses <- struct{}{}:
				default:
				}
			}
		}
	}()
	watchdog := time.NewTimer(4 * h.heartbeat)
	defer watchdog.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-session.Done():
			return status.Error(codes.Unavailable, "session closed")
		case e := <-failures:
			if errors.Is(e, io.EOF) {
				return nil
			}
			return e
		case <-pulses:
			watchdog.Reset(4 * h.heartbeat)
		case <-watchdog.C:
			return status.Error(codes.DeadlineExceeded, "heartbeat timeout")
		}
	}
}

func (h *grpcHost) receive(ctx context.Context, session *Session, worker *cortex.Session, id string, maxSlots int, m *pb.NeuronMessage, send func(*pb.BrainMessage) error) error {
	switch body := m.Body.(type) {
	case *pb.NeuronMessage_Pulse:
		if body.Pulse == nil || int(body.Pulse.Slots) > maxSlots {
			return ErrProtocol
		}
		return worker.Pulse(int(body.Pulse.Slots))
	case *pb.NeuronMessage_Accepted:
		l, err := readLease(body.Accepted.GetLease())
		if err != nil {
			return err
		}
		return session.Accept(l)
	case *pb.NeuronMessage_Result:
		l, result, err := readResult(body.Result, id)
		if err != nil {
			return err
		}
		return session.Complete(l, result)
	case *pb.NeuronMessage_ArtifactRequest:
		l, err := readLease(body.ArtifactRequest.GetLease())
		if err != nil {
			return err
		}
		return StreamArtifact(ctx, session, h.catalog, l, func(c Chunk) error {
			return send(&pb.BrainMessage{Body: &pb.BrainMessage_ArtifactChunk{ArtifactChunk: chunkWire(c)}})
		})
	case *pb.NeuronMessage_Drained:
		return worker.Drain()
	default:
		// Nested calls and host service requests are not enabled in this increment.
		return ErrProtocol
	}
}
