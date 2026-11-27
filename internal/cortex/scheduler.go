// Package cortex provides host-owned placement and capacity accounting. It
// accepts only already-authorized requests and authenticated worker sessions.
package cortex

import (
	"bytes"
	"context"
	"errors"
	"slices"
	"strings"
	"sync"
	"time"

	"brain/internal/execution"
)

var (
	ErrInvalidConfig      = errors.New("invalid cortex configuration")
	ErrInvalidNeuron      = errors.New("invalid neuron registration")
	ErrDuplicateNeuron    = errors.New("neuron already connected")
	ErrSessionLost        = errors.New("neuron session lost")
	ErrProjectQuota       = errors.New("project execution quota exceeded")
	ErrDuplicateExecution = errors.New("execution already active")
	ErrQueueFull          = errors.New("execution queue full")
)

const maxSlots = 1024
const maxRuntimes = 32
const maxNeurons = 4096

type Config struct {
	MaxNeurons, MaxProjectRunning, MaxQueued, MaxProjectQueued int
	HeartbeatTTL                                               time.Duration
	// Clock must be safe for concurrent calls and return nondecreasing time.
	// Use the system clock in production.
	Clock func() time.Time
}

type Registration struct {
	ID       string
	Runtimes []execution.Runtime
	Slots    int
	Executor execution.Executor
}

// Snapshot contains only operational metadata, never input, caller or output.
type NeuronSnapshot struct {
	ID             string
	Slots, Running int
	Draining       bool
	LastPulse      time.Time
	Runtimes       []execution.Runtime
}

type Snapshot struct {
	Neurons         []NeuronSnapshot
	Running, Queued int
}

type Scheduler struct {
	mu                            sync.Mutex
	catalog                       *execution.Catalog
	config                        Config
	neurons                       map[string]*Session
	active                        map[string]*assignment
	pending                       map[string]*waiter
	queue                         []*waiter
	projectRunning, projectQueued map[string]int
	changed                       chan struct{}
}

// Session is an opaque connection capability. Only the authenticated transport
// should receive one; its identity must not be reconstructed from a worker ID.
type Session struct {
	scheduler      *Scheduler
	id             string
	runtimes       map[execution.Runtime]bool
	executor       execution.Executor
	slots, running int
	draining       bool
	lastPulse      time.Time
}

type assignment struct {
	session *Session
	project string
	cancel  context.CancelCauseFunc
}

type waiter struct {
	request execution.Request
}

func New(catalog *execution.Catalog, cfg Config) (*Scheduler, error) {
	if catalog == nil || cfg.MaxNeurons <= 0 || cfg.MaxNeurons > maxNeurons || cfg.MaxProjectRunning <= 0 || cfg.HeartbeatTTL <= 0 || cfg.MaxQueued < 0 || cfg.MaxProjectQueued < 0 || cfg.MaxProjectQueued > cfg.MaxQueued || (cfg.MaxQueued > 0 && cfg.MaxProjectQueued == 0) {
		return nil, ErrInvalidConfig
	}
	if cfg.Clock == nil {
		cfg.Clock = time.Now
	}
	return &Scheduler{catalog: catalog, config: cfg, neurons: make(map[string]*Session), active: make(map[string]*assignment), pending: make(map[string]*waiter), projectRunning: make(map[string]int), projectQueued: make(map[string]int), changed: make(chan struct{})}, nil
}

func validID(s string) bool { return strings.TrimSpace(s) != "" && len(s) <= 1024 }

// Register is a trusted host API; Synapse must authenticate a connection before
// calling it. Reconnection cannot replace an existing live session.
func (s *Scheduler) Register(reg Registration) (*Session, error) {
	if !validID(reg.ID) || reg.Executor == nil || reg.Slots <= 0 || reg.Slots > maxSlots || len(reg.Runtimes) == 0 || len(reg.Runtimes) > maxRuntimes {
		return nil, ErrInvalidNeuron
	}
	runtimes := make(map[execution.Runtime]bool, len(reg.Runtimes))
	for _, rt := range reg.Runtimes {
		if !validID(rt.Kind) || !validID(rt.Profile) || runtimes[rt] {
			return nil, ErrInvalidNeuron
		}
		runtimes[rt] = true
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.expireLocked()
	if s.neurons[reg.ID] != nil {
		return nil, ErrDuplicateNeuron
	}
	if len(s.neurons) >= s.config.MaxNeurons {
		return nil, execution.ErrUnavailable
	}
	w := &Session{scheduler: s, id: reg.ID, runtimes: runtimes, executor: reg.Executor, slots: reg.Slots, lastPulse: s.config.Clock()}
	s.neurons[w.id] = w
	s.signalLocked()
	return w, nil
}

// Pulse updates host-timed heartbeat and advertised capacity. Reducing slots
// below the number running stops new placement without cancelling active work.
func (w *Session) Pulse(slots int) error {
	s := w.scheduler
	s.mu.Lock()
	defer s.mu.Unlock()
	s.expireLocked()
	if s.neurons[w.id] != w {
		return ErrSessionLost
	}
	if slots < 0 || slots > maxSlots {
		return ErrInvalidNeuron
	}
	w.slots, w.lastPulse = slots, s.config.Clock()
	s.signalLocked()
	return nil
}

func (w *Session) Drain() error {
	s := w.scheduler
	s.mu.Lock()
	defer s.mu.Unlock()
	s.expireLocked()
	if s.neurons[w.id] != w {
		return ErrSessionLost
	}
	w.draining = true
	s.signalLocked()
	return nil
}

func (w *Session) Disconnect() {
	s := w.scheduler
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.neurons[w.id] == w {
		s.removeLocked(w)
	}
}

// Expire must also be called periodically by the host so idle connections and
// already-running executions are fenced even with no incoming requests.
func (s *Scheduler) Expire() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.expireLocked()
}

func (s *Scheduler) expireLocked() {
	now := s.config.Clock()
	for _, w := range s.neurons {
		if now.Sub(w.lastPulse) >= s.config.HeartbeatTTL {
			s.removeLocked(w)
		}
	}
}

func (s *Scheduler) removeLocked(w *Session) {
	delete(s.neurons, w.id)
	for _, a := range s.active {
		if a.session == w {
			a.cancel(ErrSessionLost)
		}
	}
	s.signalLocked()
}

func (s *Scheduler) signalLocked() { close(s.changed); s.changed = make(chan struct{}) }

func (s *Scheduler) selectLocked(req execution.Request) *Session {
	if len(s.active) >= s.config.MaxNeurons*maxSlots || s.projectRunning[req.ProjectID] >= s.config.MaxProjectRunning {
		return nil
	}
	var best *Session
	for _, w := range s.neurons {
		if w.draining || w.running >= w.slots || !w.runtimes[req.Runtime] {
			continue
		}
		// Least occupied fraction, deterministic tie break.
		if best == nil || w.running*best.slots < best.running*w.slots || (w.running*best.slots == best.running*w.slots && w.id < best.id) {
			best = w
		}
	}
	return best
}

// Earlier compatible queued work gets first claim on capacity; incompatible
// or quota-blocked head entries do not stall the whole queue.
func (s *Scheduler) firstReadyLocked() *waiter {
	for _, w := range s.queue {
		if s.selectLocked(w.request) != nil {
			return w
		}
	}
	return nil
}

func (s *Scheduler) removeWaiterLocked(w *waiter) {
	delete(s.pending, w.request.ID)
	for i, other := range s.queue {
		if other == w {
			copy(s.queue[i:], s.queue[i+1:])
			s.queue[len(s.queue)-1] = nil
			s.queue = s.queue[:len(s.queue)-1]
			break
		}
	}
	s.projectQueued[w.request.ProjectID]--
	if s.projectQueued[w.request.ProjectID] == 0 {
		delete(s.projectQueued, w.request.ProjectID)
	}
	s.signalLocked()
}

// Execute implements the execution contract for already-admitted work. No
// retries, fallback, durable commits or authentication are performed here.
func (s *Scheduler) Execute(ctx context.Context, req execution.Request) (execution.Result, error) {
	if err := s.catalog.CheckRequest(req, time.Now()); err != nil {
		return execution.Result{}, err
	}
	req.Input = bytes.Clone(req.Input)
	req.Context = bytes.Clone(req.Context)
	ctx, cancel := context.WithDeadline(ctx, req.Deadline)
	defer cancel()
	var queued *waiter
	s.mu.Lock()
	for {
		s.expireLocked()
		if err := ctx.Err(); err != nil {
			if queued != nil {
				s.removeWaiterLocked(queued)
			}
			s.mu.Unlock()
			return execution.Result{}, err
		}
		if queued == nil && (s.active[req.ID] != nil || s.pending[req.ID] != nil) {
			s.mu.Unlock()
			return execution.Result{}, ErrDuplicateExecution
		}
		w := s.selectLocked(req)
		if first := s.firstReadyLocked(); w != nil && (first == nil || first == queued) {
			if queued != nil {
				s.removeWaiterLocked(queued)
			}
			runCtx, stop := context.WithCancelCause(ctx)
			a := &assignment{session: w, project: req.ProjectID, cancel: stop}
			s.active[req.ID] = a
			w.running++
			s.projectRunning[req.ProjectID]++
			s.mu.Unlock()
			return s.run(runCtx, req, a)
		}
		if queued == nil {
			if s.config.MaxQueued == 0 {
				quota := s.projectRunning[req.ProjectID] >= s.config.MaxProjectRunning
				s.mu.Unlock()
				if quota {
					return execution.Result{}, ErrProjectQuota
				}
				return execution.Result{}, execution.ErrUnavailable
			}
			if s.projectQueued[req.ProjectID] >= s.config.MaxProjectQueued {
				s.mu.Unlock()
				return execution.Result{}, ErrProjectQuota
			}
			if len(s.queue) >= s.config.MaxQueued {
				s.mu.Unlock()
				return execution.Result{}, ErrQueueFull
			}
			queued = &waiter{request: req}
			s.pending[req.ID] = queued
			s.queue = append(s.queue, queued)
			s.projectQueued[req.ProjectID]++
		}
		changed := s.changed
		s.mu.Unlock()
		select {
		case <-ctx.Done():
		case <-changed:
		}
		s.mu.Lock()
	}
}

func (s *Scheduler) run(ctx context.Context, req execution.Request, a *assignment) (execution.Result, error) {
	defer a.cancel(nil)
	result, err := execution.ExecuteAssigned(ctx, s.catalog, a.session.id, a.session.executor, req)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.expireLocked()
	delete(s.active, req.ID)
	a.session.running--
	s.projectRunning[a.project]--
	if s.projectRunning[a.project] == 0 {
		delete(s.projectRunning, a.project)
	}
	s.signalLocked()
	if cause := context.Cause(ctx); cause != nil {
		return execution.Result{}, cause
	}
	return result, err
}

func (s *Scheduler) Snapshot() Snapshot {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.expireLocked()
	out := Snapshot{Running: len(s.active), Queued: len(s.queue)}
	for _, w := range s.neurons {
		v := NeuronSnapshot{ID: w.id, Slots: w.slots, Running: w.running, Draining: w.draining, LastPulse: w.lastPulse}
		for rt := range w.runtimes {
			v.Runtimes = append(v.Runtimes, rt)
		}
		slices.SortFunc(v.Runtimes, func(a, b execution.Runtime) int {
			if kind := strings.Compare(a.Kind, b.Kind); kind != 0 {
				return kind
			}
			return strings.Compare(a.Profile, b.Profile)
		})
		out.Neurons = append(out.Neurons, v)
	}
	slices.SortFunc(out.Neurons, func(a, b NeuronSnapshot) int { return strings.Compare(a.ID, b.ID) })
	return out
}
