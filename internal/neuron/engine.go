// Package neuron owns worker-local execution admission and runtime dispatch.
// A Neuron has no durable application state and no implicit runtime fallback.
package neuron

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"slices"
	"strings"
	"sync"
	"time"

	"brain/internal/execution"
)

var (
	ErrInvalidConfig = errors.New("invalid neuron configuration")
	ErrDraining      = errors.New("neuron is draining")
	ErrClosed        = errors.New("neuron is closed")
	ErrDuplicate     = errors.New("neuron execution already active")
)

// ArtifactSource must obtain only the artifact authorized for the current
// assignment. A remote implementation binds this interface to its Synapse lease.
type ArtifactSource interface {
	Fetch(context.Context, execution.Artifact) ([]byte, error)
}

// Runtime must enforce the selected security profile, actual CPU/memory and
// process limits, cancellation and fresh sandbox isolation. The engine does not
// treat a runtime adapter's resource reports as proof of enforcement.
type Runtime interface {
	Run(context.Context, execution.Request, []execution.BundleFile) (execution.Result, error)
}

type Adapter struct {
	Capability execution.Runtime
	Runtime    Runtime
}
type Config struct {
	ID               string
	Slots            int
	MaxArtifactBytes int64
	Adapters         []Adapter
	Source           ArtifactSource
}
type Snapshot struct {
	ID                        string
	Slots, Running, Available int
	Draining, Closed          bool
	Runtimes                  []execution.Runtime
}
type Engine struct {
	mu               sync.Mutex
	id               string
	slots            int
	maxArtifact      int64
	source           ArtifactSource
	adapters         map[execution.Runtime]Runtime
	active           map[string]context.CancelCauseFunc
	idle             chan struct{}
	draining, closed bool
}

func validID(s string) bool { return strings.TrimSpace(s) != "" && len(s) <= 1024 }
func New(cfg Config) (*Engine, error) {
	if !validID(cfg.ID) || cfg.Slots <= 0 || cfg.Slots > 1024 || cfg.MaxArtifactBytes <= 0 || cfg.MaxArtifactBytes > execution.MaxBundleBytes || cfg.Source == nil || len(cfg.Adapters) == 0 || len(cfg.Adapters) > 32 {
		return nil, ErrInvalidConfig
	}
	adapters := make(map[execution.Runtime]Runtime, len(cfg.Adapters))
	for _, a := range cfg.Adapters {
		if !validID(a.Capability.Kind) || !validID(a.Capability.Profile) || a.Runtime == nil || adapters[a.Capability] != nil {
			return nil, ErrInvalidConfig
		}
		adapters[a.Capability] = a.Runtime
	}
	idle := make(chan struct{})
	close(idle)
	return &Engine{id: cfg.ID, slots: cfg.Slots, maxArtifact: cfg.MaxArtifactBytes, source: cfg.Source, adapters: adapters, active: make(map[string]context.CancelCauseFunc), idle: idle}, nil
}

// Execute accepts host-issued requests only, verifies the complete immutable
// bundle and invokes the exact registered adapter. No durable commit is possible
// here. Each call receives independent input and artifact bytes.
func (e *Engine) Execute(ctx context.Context, r execution.Request) (execution.Result, error) {
	if err := r.Validate(time.Now()); err != nil {
		return execution.Result{}, err
	}
	ctx, deadlineCancel := context.WithDeadline(ctx, r.Deadline)
	defer deadlineCancel()
	if err := ctx.Err(); err != nil {
		return execution.Result{}, err
	}
	ctx, stop := context.WithCancelCause(ctx)
	defer stop(nil)
	r.Deadline, _ = ctx.Deadline()
	r.Input = bytes.Clone(r.Input)
	r.Context = bytes.Clone(r.Context)
	e.mu.Lock()
	switch {
	case e.closed:
		e.mu.Unlock()
		return execution.Result{}, ErrClosed
	case e.draining:
		e.mu.Unlock()
		return execution.Result{}, ErrDraining
	case e.adapters[r.Runtime] == nil:
		e.mu.Unlock()
		return execution.Result{}, execution.ErrUnavailable
	case e.active[r.ID] != nil:
		e.mu.Unlock()
		return execution.Result{}, ErrDuplicate
	case len(e.active) >= e.slots:
		e.mu.Unlock()
		return execution.Result{}, execution.ErrUnavailable
	}
	if len(e.active) == 0 {
		e.idle = make(chan struct{})
	}
	e.active[r.ID] = stop
	adapter := e.adapters[r.Runtime]
	e.mu.Unlock()
	defer e.finish(r.ID)
	started := time.Now()
	bundle, err := e.source.Fetch(ctx, r.Artifact)
	if err != nil {
		return execution.Result{}, err
	}
	files, err := execution.OpenBundle(r.Artifact, bundle, e.maxArtifact)
	if err != nil {
		return execution.Result{}, err
	}
	if cause := context.Cause(ctx); cause != nil {
		return execution.Result{}, cause
	}
	result, err := adapter.Run(ctx, r, files)
	if cause := context.Cause(ctx); cause != nil {
		return execution.Result{}, cause
	}
	if err != nil {
		return execution.Result{}, err
	}
	result.ID, result.WorkerID, result.Duration = r.ID, e.id, time.Since(started)
	if err = result.Validate(r, e.id); err != nil {
		return execution.Result{}, err
	}
	result.Output = bytes.Clone(result.Output)
	if result.Error != nil {
		failure := *result.Error
		result.Error = &failure
	}
	if result.Effects != nil {
		effects := make([]json.RawMessage, len(result.Effects))
		for i, raw := range result.Effects {
			effects[i] = bytes.Clone(raw)
		}
		result.Effects = effects
	}
	return result, nil
}
func (e *Engine) finish(id string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	delete(e.active, id)
	if len(e.active) == 0 {
		close(e.idle)
	}
}

// Drain stops admission without interrupting existing invocations.
func (e *Engine) Drain() { e.mu.Lock(); defer e.mu.Unlock(); e.draining = true }

// Close cancels all active invocations. Wait can be used to await adapter cleanup.
func (e *Engine) Close() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.closed = true
	e.draining = true
	for _, stop := range e.active {
		stop(ErrClosed)
	}
}
func (e *Engine) Wait(ctx context.Context) error {
	e.mu.Lock()
	idle := e.idle
	e.mu.Unlock()
	select {
	case <-idle:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
func (e *Engine) Snapshot() Snapshot {
	e.mu.Lock()
	defer e.mu.Unlock()
	v := Snapshot{ID: e.id, Slots: e.slots, Running: len(e.active), Draining: e.draining, Closed: e.closed}
	if !e.draining {
		v.Available = e.slots - len(e.active)
	}
	for rt := range e.adapters {
		v.Runtimes = append(v.Runtimes, rt)
	}
	slices.SortFunc(v.Runtimes, func(a, b execution.Runtime) int {
		if kind := strings.Compare(a.Kind, b.Kind); kind != 0 {
			return kind
		}
		return strings.Compare(a.Profile, b.Profile)
	})
	return v
}
