package synapse

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"brain/internal/execution"
	"brain/internal/neuron"
	pb "brain/proto/synapse/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/protobuf/proto"
)

// RunNeuron connects an independent Neuron engine to one authenticated Synapse
// session. On disconnect all in-flight runtimes are cancelled before returning;
// callers may reconnect with a new session, but never replay unfinished work.
func RunNeuron(ctx context.Context, address string, tlsConfig *tls.Config, cfg neuron.Config) error {
	if !validID(address) || tlsConfig == nil || tlsConfig.InsecureSkipVerify || tlsConfig.MinVersion < tls.VersionTLS13 || tlsConfig.RootCAs == nil || len(tlsConfig.Certificates) == 0 || cfg.Source != nil {
		return ErrProtocol
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)), grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(MaxMessageBytes), grpc.MaxCallSendMsgSize(MaxMessageBytes)))
	if err != nil {
		return err
	}
	defer conn.Close()
	stream, err := pb.NewSynapseClient(conn).Connect(ctx)
	if err != nil {
		return err
	}
	c := &remoteNeuron{ctx: ctx, cancel: cancel, stream: stream, jobs: make(map[uint64]*remoteJob), maxArtifact: cfg.MaxArtifactBytes}
	cfg.Source = c
	engine, err := neuron.New(cfg)
	if err != nil {
		return err
	}
	c.engine = engine
	defer func() {
		cancel()
		engine.Close()
		c.wg.Wait()
	}()
	reg := &pb.Register{NeuronId: cfg.ID, Slots: uint32(cfg.Slots), SupportedVersions: []uint32{Version}}
	for _, capability := range engine.Snapshot().Runtimes {
		reg.Runtimes = append(reg.Runtimes, runtimeWire(capability))
	}
	// Bound the handshake even if the address is unreachable.
	handshakeTimeout := time.AfterFunc(5*time.Second, cancel)
	if err = stream.Send(&pb.NeuronMessage{Body: &pb.NeuronMessage_Register{Register: reg}}); err != nil {
		handshakeTimeout.Stop()
		return err
	}
	first, err := stream.Recv()
	if !handshakeTimeout.Stop() || err != nil {
		if err != nil {
			return err
		}
		return context.DeadlineExceeded
	}
	registered := first.GetRegistered()
	if registered == nil || first.Version != 0 || first.SessionId != "" || first.Sequence != 0 || registered.Version != Version || !validID(registered.SessionId) || registered.MaxMessageBytes != MaxMessageBytes || registered.HeartbeatIntervalMs < 1000 || registered.HeartbeatIntervalMs > 30000 {
		return ErrProtocol
	}
	c.session = registered.SessionId
	receiver, err := NewReceiver(c.session)
	if err != nil {
		return err
	}
	ticker := time.NewTicker(time.Duration(registered.HeartbeatIntervalMs) * time.Millisecond)
	defer ticker.Stop()
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				state := engine.Snapshot()
				slots := state.Slots
				if state.Draining {
					slots = 0
				}
				if c.send(&pb.NeuronMessage{Body: &pb.NeuronMessage_Pulse{Pulse: &pb.Pulse{Slots: uint32(slots)}}}) != nil {
					cancel()
					return
				}
			}
		}
	}()
	for {
		message, e := stream.Recv()
		if e != nil {
			return e
		}
		if e = receiver.Check(message.Version, message.SessionId, message.Sequence, proto.Size(message)); e != nil {
			return e
		}
		if e = c.receive(message); e != nil {
			return e
		}
	}
}

type remoteJobKey struct{}

type remoteJob struct {
	lease     Lease
	request   execution.Request
	cancel    context.CancelFunc
	assembler *Assembler
	bundle    chan []byte
	fetching  bool
	hostCallID uint64
	hostReply chan hostReply
}

type remoteNeuron struct {
	ctx         context.Context
	cancel      context.CancelFunc
	stream      pb.Synapse_ConnectClient
	engine      *neuron.Engine
	session     string
	maxArtifact int64
	sendMu      sync.Mutex
	sequence    uint64
	mu          sync.Mutex
	jobs        map[uint64]*remoteJob
	lastAttempt uint64
	wg          sync.WaitGroup
}

func (c *remoteNeuron) send(m *pb.NeuronMessage) error {
	c.sendMu.Lock()
	defer c.sendMu.Unlock()
	if err := c.ctx.Err(); err != nil {
		return err
	}
	c.sequence++
	m.Version, m.SessionId, m.Sequence = Version, c.session, c.sequence
	return c.stream.Send(m)
}

func (c *remoteNeuron) receive(m *pb.BrainMessage) error {
	switch body := m.Body.(type) {
	case *pb.BrainMessage_Assignment:
		a, err := readAssignment(body.Assignment)
		if err != nil {
			return err
		}
		return c.start(a)
	case *pb.BrainMessage_Cancel:
		lease, err := readLease(body.Cancel.GetLease())
		if err != nil {
			return err
		}
		c.mu.Lock()
		defer c.mu.Unlock()
		if lease.SessionID != c.session || lease.Attempt > c.lastAttempt {
			return ErrProtocol
		}
		if job := c.jobs[lease.Attempt]; job != nil {
			if job.lease != lease {
				return ErrProtocol
			}
			job.cancel()
		}
		// A terminal result can cross a host cancellation on the wire.
		return nil
	case *pb.BrainMessage_ArtifactChunk:
		chunk, err := readChunk(body.ArtifactChunk)
		if err != nil {
			return err
		}
		c.mu.Lock()
		defer c.mu.Unlock()
		if chunk.Lease.SessionID != c.session || chunk.Lease.Attempt > c.lastAttempt {
			return ErrProtocol
		}
		job := c.jobs[chunk.Lease.Attempt]
		if job == nil {
			// An in-flight chunk may arrive after cancellation completed.
			return nil
		}
		if job.lease != chunk.Lease || !job.fetching {
			return ErrProtocol
		}
		if err = job.assembler.Push(chunk); err != nil {
			return err
		}
		if chunk.Final {
			data, err := job.assembler.Finish()
			if err != nil {
				return err
			}
			job.fetching = false
			job.bundle <- data
		}
		return nil
	case *pb.BrainMessage_HostResponse:
		return c.receiveHostResponse(body.HostResponse)
	case *pb.BrainMessage_Drain:
		if body.Drain == nil {
			return ErrProtocol
		}
		c.engine.Drain()
		c.wg.Add(1)
		go func() {
			defer c.wg.Done()
			if c.engine.Wait(c.ctx) == nil && c.send(&pb.NeuronMessage{Body: &pb.NeuronMessage_Drained{Drained: &pb.Drained{}}}) != nil {
				c.cancel()
			}
		}()
		return nil
	default:
		return ErrProtocol
	}
}

func (c *remoteNeuron) start(a Assignment) error {
	if err := a.Request.Validate(time.Now()); err != nil && !errors.Is(err, context.DeadlineExceeded) {
		return err
	}
	c.mu.Lock()
	state := c.engine.Snapshot()
	if a.Lease.SessionID != c.session || a.Lease.Attempt <= c.lastAttempt || len(c.jobs) >= state.Slots || state.Draining {
		c.mu.Unlock()
		return ErrProtocol
	}
	for _, job := range c.jobs {
		if job.request.ID == a.Request.ID {
			c.mu.Unlock()
			return ErrProtocol
		}
	}
	assembler, err := NewAssembler(a.Lease, a.Request.Artifact, c.maxArtifact)
	if err != nil {
		c.mu.Unlock()
		return err
	}
	ctx, cancel := context.WithCancel(c.ctx)
	job := &remoteJob{lease: a.Lease, request: a.Request, cancel: cancel, assembler: assembler, bundle: make(chan []byte, 1)}
	c.lastAttempt = a.Lease.Attempt
	c.jobs[a.Lease.Attempt] = job
	c.mu.Unlock()
	if err = c.send(&pb.NeuronMessage{Body: &pb.NeuronMessage_Accepted{Accepted: &pb.Accepted{Lease: leaseWire(a.Lease)}}}); err != nil {
		cancel()
		return err
	}
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		defer cancel()
		jobCtx := context.WithValue(ctx, remoteJobKey{}, job)
		jobCtx = execution.WithHostCaller(jobCtx, func(ctx context.Context, operation string, input json.RawMessage) (json.RawMessage, error) {
			return c.callHost(ctx, job, operation, input)
		})
		r, err := c.engine.Execute(jobCtx, a.Request)
		if err != nil {
			r = execution.Result{Status: execution.Failed, Error: &execution.Failure{Code: "runtime_error"}}
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				r.Status, r.Error.Code = execution.Cancelled, "cancelled"
			}
		}
		c.mu.Lock()
		delete(c.jobs, a.Lease.Attempt)
		c.mu.Unlock()
		if c.send(&pb.NeuronMessage{Body: &pb.NeuronMessage_Result{Result: resultWire(a.Lease, r)}}) != nil {
			c.cancel()
		}
	}()
	return nil
}

// Fetch is reachable only from the engine with its host-issued job context.
func (c *remoteNeuron) Fetch(ctx context.Context, artifact execution.Artifact) ([]byte, error) {
	job, ok := ctx.Value(remoteJobKey{}).(*remoteJob)
	if !ok || job.request.Artifact != artifact {
		return nil, ErrProtocol
	}
	c.mu.Lock()
	if c.jobs[job.lease.Attempt] != job || job.fetching {
		c.mu.Unlock()
		return nil, ErrProtocol
	}
	job.fetching = true
	c.mu.Unlock()
	if err := c.send(&pb.NeuronMessage{Body: &pb.NeuronMessage_ArtifactRequest{ArtifactRequest: &pb.ArtifactRequest{Lease: leaseWire(job.lease)}}}); err != nil {
		return nil, err
	}
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case data := <-job.bundle:
		return data, nil
	}
}
