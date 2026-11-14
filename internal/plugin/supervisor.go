package plugin

import (
	"errors"
	"strings"
)

var ErrWorkerAlreadyRunning = errors.New("worker already running")
var ErrWorkerNotRunning = errors.New("worker not running")
var ErrEnabledWorkerUnavailable = errors.New("enabled worker unavailable")

type Worker interface {
	Start() error
	Stop() error
	Ready() bool
}

type Supervisor struct{ workers map[string]Worker }

func NewSupervisor() *Supervisor { return &Supervisor{workers: map[string]Worker{}} }

func (s *Supervisor) Start(key string, worker Worker) error {
	if _, ok := s.workers[key]; ok {
		return ErrWorkerAlreadyRunning
	}
	if err := worker.Start(); err != nil {
		return err
	}
	s.workers[key] = worker
	return nil
}

func (s *Supervisor) Stop(key string) error {
	worker, ok := s.workers[key]
	if !ok {
		return ErrWorkerNotRunning
	}
	if err := worker.Stop(); err != nil {
		return err
	}
	delete(s.workers, key)
	return nil
}

func (s *Supervisor) StopPluginVersion(pluginID, version string) error {
	if pluginID == "" || version == "" {
		return ErrWorkerNotRunning
	}
	suffix := pluginID + "@" + version
	var firstErr error
	for key, worker := range s.workers {
		if key != suffix && !strings.HasSuffix(key, ":"+suffix) {
			continue
		}
		if err := worker.Stop(); err != nil {
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

func (s *Supervisor) Ready(key string) bool {
	worker, ok := s.workers[key]
	return ok && worker.Ready()
}

func (s *Supervisor) CheckEnabled(keys []string) error {
	for _, key := range keys {
		if !s.Ready(key) {
			return ErrEnabledWorkerUnavailable
		}
	}
	return nil
}
