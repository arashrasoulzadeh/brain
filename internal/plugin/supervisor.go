package plugin

import (
	"errors"
	"strings"
	"sync"
)

var ErrWorkerAlreadyRunning = errors.New("worker already running")
var ErrWorkerNotRunning = errors.New("worker not running")
var ErrEnabledWorkerUnavailable = errors.New("enabled worker unavailable")
var ErrWorkerRecoveryUnavailable = errors.New("worker recovery unavailable")
var ErrWorkerRestartLimit = errors.New("worker restart limit exceeded")

type Worker interface {
	Start() error
	Stop() error
	Ready() bool
}

type WorkerFactory func() Worker

type RecoveryPolicy struct{ MaxRestarts int }

type supervisedWorker struct {
	worker       Worker
	factory      WorkerFactory
	policy       RecoveryPolicy
	restartCount int
}

type Supervisor struct {
	mu      sync.RWMutex
	workers map[string]supervisedWorker
}

func NewSupervisor() *Supervisor { return &Supervisor{workers: map[string]supervisedWorker{}} }

func (s *Supervisor) Start(key string, worker Worker) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if worker == nil {
		return ErrWorkerRecoveryUnavailable
	}
	if _, ok := s.workers[key]; ok {
		return ErrWorkerAlreadyRunning
	}
	if err := worker.Start(); err != nil {
		return err
	}
	s.workers[key] = supervisedWorker{worker: worker}
	return nil
}

func (s *Supervisor) StartRecovering(key string, factory WorkerFactory, policy RecoveryPolicy) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if factory == nil || policy.MaxRestarts < 0 {
		return ErrWorkerRecoveryUnavailable
	}
	if _, ok := s.workers[key]; ok {
		return ErrWorkerAlreadyRunning
	}
	worker := factory()
	if worker == nil {
		return ErrWorkerRecoveryUnavailable
	}
	if err := worker.Start(); err != nil {
		return err
	}
	s.workers[key] = supervisedWorker{worker: worker, factory: factory, policy: policy}
	return nil
}

func (s *Supervisor) Recover(key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, ok := s.workers[key]
	if !ok {
		return ErrWorkerNotRunning
	}
	if entry.factory == nil {
		return ErrWorkerRecoveryUnavailable
	}
	if entry.worker.Ready() {
		return nil
	}
	if entry.restartCount >= entry.policy.MaxRestarts {
		return ErrWorkerRestartLimit
	}
	worker := entry.factory()
	if worker == nil {
		return ErrWorkerRecoveryUnavailable
	}
	if err := worker.Start(); err != nil {
		entry.restartCount++
		s.workers[key] = entry
		return err
	}
	entry.worker = worker
	entry.restartCount++
	s.workers[key] = entry
	return nil
}

func (s *Supervisor) Stop(key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, ok := s.workers[key]
	if !ok {
		return ErrWorkerNotRunning
	}
	if err := entry.worker.Stop(); err != nil {
		return err
	}
	delete(s.workers, key)
	return nil
}

func (s *Supervisor) StopPluginVersion(pluginID, version string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if pluginID == "" || version == "" {
		return ErrWorkerNotRunning
	}
	suffix := pluginID + "@" + version
	var firstErr error
	for key, entry := range s.workers {
		if key != suffix && !strings.HasSuffix(key, ":"+suffix) {
			continue
		}
		if err := entry.worker.Stop(); err != nil {
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		delete(s.workers, key)
	}
	if firstErr != nil {
		return firstErr
	}
	return nil
}

func (s *Supervisor) StopProjectPlugin(projectID, pluginID, version string) error {
	if projectID == "" || pluginID == "" || version == "" {
		return ErrWorkerNotRunning
	}
	return s.Stop(projectID + ":" + pluginID + "@" + version)
}

func (s *Supervisor) Ready(key string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	entry, ok := s.workers[key]
	return ok && entry.worker.Ready()
}

func (s *Supervisor) CheckEnabled(keys []string) error {
	for _, key := range keys {
		if !s.Ready(key) {
			return ErrEnabledWorkerUnavailable
		}
	}
	return nil
}
