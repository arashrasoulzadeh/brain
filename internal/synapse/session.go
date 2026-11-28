// Package synapse implements transport-independent assignment correlation for
// the versioned Synapse stream. The transport must authenticate peers first.
package synapse

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"time"

	"brain/internal/execution"
)

const Version uint32 = 1
const MaxMessageBytes = 2 << 20
const maxPending = 1024

var (
	ErrProtocol     = errors.New("invalid synapse protocol message")
	ErrClosed       = errors.New("synapse session closed")
	ErrBackpressure = errors.New("synapse outbound capacity exhausted")
	ErrExhausted    = errors.New("synapse assignment sequence exhausted")
	ErrDuplicate    = errors.New("synapse execution already pending")
)

// Negotiate supports v1 only. Unknown future versions are not silently selected.
func Negotiate(versions []uint32) (uint32, error) {
	if len(versions) == 0 || len(versions) > 16 {
		return 0, ErrProtocol
	}
	for _, v := range versions {
		if v == Version {
			return Version, nil
		}
	}
	return 0, ErrProtocol
}

// Receiver validates strictly contiguous envelopes after negotiation. The first
// post-handshake sequence is one. The gRPC adapter must enforce encoded size
// before decoding; size here is a redundant check, not a decoder limit.
type Receiver struct {
	mu      sync.Mutex
	session string
	last    uint64
}

func NewReceiver(session string) (*Receiver, error) {
	if !validID(session) {
		return nil, ErrProtocol
	}
	return &Receiver{session: session}, nil
}
func (r *Receiver) Check(version uint32, session string, sequence uint64, encodedBytes int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if version != Version || session != r.session || sequence == 0 || sequence != r.last+1 || encodedBytes <= 0 || encodedBytes > MaxMessageBytes {
		return ErrProtocol
	}
	r.last = sequence
	return nil
}

type Lease struct {
	SessionID   string
	Attempt     uint64
	ExecutionID string
}
type Assignment struct {
	Lease   Lease
	Request execution.Request
}

// Command has exactly one of Assignment or Cancel. It is an internal event,
// not an alternative wire format; the gRPC adapter maps it to protobuf v1.
type Command struct {
	Assignment *Assignment
	Cancel     *Lease
}

type Config struct {
	// WorkerID is derived from authenticated transport identity, not Register.
	// SessionID is a fresh, host-generated, unique connection identifier.
	WorkerID, SessionID   string
	MaxPending, QueueSize int
	// MaxAssignments bounds attempts per connection. Zero uses 2^32.
	MaxAssignments uint64
	// CancelGrace bounds how long a delivered cancellation can occupy its slot.
	// On timeout the entire stream is closed; zero uses two seconds.
	CancelGrace time.Duration
}

type Session struct {
	mu       sync.Mutex
	cfg      Config
	next     uint64
	closed   bool
	done     chan struct{}
	outbound chan Command
	pending  map[uint64]*pending
	ids      map[string]bool
}
type pending struct {
	assignment                      Assignment
	delivered, accepted, cancelling bool
	ctx                             context.Context
	result                          chan execution.Result
	hostCalls                       uint64
	hostBusy                        bool
}

func validID(s string) bool { return strings.TrimSpace(s) != "" && len(s) <= 1024 }
func NewSession(cfg Config) (*Session, error) {
	if !validID(cfg.WorkerID) || !validID(cfg.SessionID) || cfg.MaxPending <= 0 || cfg.MaxPending > maxPending || cfg.QueueSize <= 0 || cfg.QueueSize > 2*maxPending || cfg.CancelGrace < 0 || cfg.CancelGrace > time.Minute {
		return nil, ErrProtocol
	}
	if cfg.MaxAssignments == 0 {
		cfg.MaxAssignments = 1 << 32
	}
	if cfg.CancelGrace == 0 {
		cfg.CancelGrace = 2 * time.Second
	}
	return &Session{cfg: cfg, done: make(chan struct{}), outbound: make(chan Command, cfg.QueueSize), pending: make(map[uint64]*pending), ids: make(map[string]bool)}, nil
}

// Done signals transport teardown; the host must disconnect this Neuron from Cortex.
func (s *Session) Done() <-chan struct{} { return s.done }

// Pending reports outstanding assignments, including cancelling work.
func (s *Session) Pending() int { s.mu.Lock(); defer s.mu.Unlock(); return len(s.pending) }

// Close invalidates the entire connection and releases waiting execution calls.
// A replacement connection must get a new SessionID and a new Session.
func (s *Session) Close() { s.mu.Lock(); defer s.mu.Unlock(); s.closeLocked() }
func (s *Session) closeLocked() {
	if !s.closed {
		s.closed = true
		close(s.done)
		clear(s.pending)
		clear(s.ids)
		close(s.outbound)
		for range s.outbound {
		}
	}
}

// Execute implements execution.Executor for one authenticated stream. Cortex
// still owns admission, placement, deadlines, cancellation and commit policy.
func (s *Session) Execute(ctx context.Context, req execution.Request) (execution.Result, error) {
	if err := req.Validate(time.Now()); err != nil {
		return execution.Result{}, err
	}
	ctx, cancel := context.WithDeadline(ctx, req.Deadline)
	defer cancel()
	if err := ctx.Err(); err != nil {
		return execution.Result{}, err
	}
	req.Deadline, _ = ctx.Deadline()
	req.Input = bytes.Clone(req.Input)
	req.Context = bytes.Clone(req.Context)
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return execution.Result{}, ErrClosed
	}
	if s.ids[req.ID] {
		s.mu.Unlock()
		return execution.Result{}, ErrDuplicate
	}
	if len(s.pending) >= s.cfg.MaxPending {
		s.mu.Unlock()
		return execution.Result{}, ErrBackpressure
	}
	if s.next == s.cfg.MaxAssignments {
		s.mu.Unlock()
		return execution.Result{}, ErrExhausted
	}
	s.next++
	lease := Lease{SessionID: s.cfg.SessionID, Attempt: s.next, ExecutionID: req.ID}
	p := &pending{ctx: ctx, assignment: Assignment{Lease: lease, Request: req}, result: make(chan execution.Result, 1)}
	s.pending[lease.Attempt] = p
	s.ids[req.ID] = true
	select {
	case s.outbound <- Command{Assignment: &p.assignment}:
	default:
		s.removeLocked(p)
		s.mu.Unlock()
		return execution.Result{}, ErrBackpressure
	}
	s.mu.Unlock()
	var result execution.Result
	select {
	case <-ctx.Done():
	case <-s.done:
		return execution.Result{}, ErrClosed
	case result = <-p.result:
	}
	if err := ctx.Err(); err != nil {
		s.cancelAndWait(lease)
		return execution.Result{}, err
	}
	return result, nil
}

func (s *Session) removeLocked(p *pending) {
	delete(s.pending, p.assignment.Lease.Attempt)
	delete(s.ids, p.assignment.Lease.ExecutionID)
}
func (s *Session) lookupLocked(lease Lease) *pending {
	if s.closed || lease.SessionID != s.cfg.SessionID {
		return nil
	}
	p := s.pending[lease.Attempt]
	if p == nil || p.assignment.Lease != lease {
		return nil
	}
	return p
}

// Keep a delivered slot occupied until the worker reports a terminal result,
// or the stream is closed. Returning early would allow Cortex to overschedule
// a worker that is still terminating the previous sandbox.
func (s *Session) cancelAndWait(lease Lease) {
	s.mu.Lock()
	p := s.lookupLocked(lease)
	if p != nil && p.delivered {
		p.cancelling = true
		select {
		case s.outbound <- Command{Cancel: &lease}:
		default:
			s.closeLocked()
			s.mu.Unlock()
			return
		}
		s.mu.Unlock()
		timer := time.NewTimer(s.cfg.CancelGrace)
		defer timer.Stop()
		select {
		case <-p.result:
		case <-s.done:
		case <-timer.C:
			s.Close()
		}
		return
	}
	if p != nil {
		s.removeLocked(p)
	}
	s.mu.Unlock()
}

// Next is called by the single stream sender. Cancelled, undelivered work is
// skipped. Closing the session prevents a queued command from being dispatched.
func (s *Session) Next(ctx context.Context) (Command, error) {
	for {
		if err := ctx.Err(); err != nil {
			return Command{}, err
		}
		select {
		case <-ctx.Done():
			return Command{}, ctx.Err()
		case cmd := <-s.outbound:
			s.mu.Lock()
			if s.closed {
				s.mu.Unlock()
				return Command{}, ErrClosed
			}
			if cmd.Assignment != nil {
				p := s.lookupLocked(cmd.Assignment.Lease)
				if p == nil || p.ctx.Err() != nil {
					s.mu.Unlock()
					continue
				}
				p.delivered = true
				// The sender receives its own copy; it cannot change the pending identity.
				a := p.assignment
				a.Request.Input = bytes.Clone(a.Request.Input)
				a.Request.Context = bytes.Clone(a.Request.Context)
				cmd.Assignment = &a
			}
			s.mu.Unlock()
			return cmd, nil
		}
	}
}

// Accept and Complete are called only by the authenticated stream receiver.
// Every assignment must be delivered and accepted once before a terminal result,
// except that a cancellation before acceptance may be acknowledged as cancelled.
func (s *Session) Accept(lease Lease) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	p := s.lookupLocked(lease)
	if p == nil || !p.delivered || p.accepted {
		return ErrProtocol
	}
	p.accepted = true
	return nil
}
func (s *Session) Complete(lease Lease, result execution.Result) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	p := s.lookupLocked(lease)
	if p == nil || (!p.accepted && !(p.cancelling && result.Status == execution.Cancelled)) {
		return ErrProtocol
	}
	if p.hostBusy && result.Status == execution.Succeeded {
		return ErrProtocol
	}
	if err := result.Validate(p.assignment.Request, s.cfg.WorkerID); err != nil {
		return err
	}
	result.Output = bytes.Clone(result.Output)
	if result.Error != nil {
		e := *result.Error
		result.Error = &e
	}
	if result.Effects != nil {
		effects := make([]json.RawMessage, len(result.Effects))
		for i, e := range result.Effects {
			effects[i] = bytes.Clone(e)
		}
		result.Effects = effects
	}
	s.removeLocked(p)
	p.result <- result
	return nil
}

// ArtifactFor returns only the artifact pinned to a delivered live assignment.
// The transport must fetch from the trusted catalog and enforce chunk bounds.
func (s *Session) ArtifactFor(lease Lease) (execution.Artifact, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	p := s.lookupLocked(lease)
	if p == nil || !p.delivered || p.ctx.Err() != nil {
		return execution.Artifact{}, ErrProtocol
	}
	return p.assignment.Request.Artifact, nil
}
