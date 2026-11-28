package app

import (
	"encoding/json"
	"net/http"

	"brain/internal/cortex"
	"brain/internal/operator"
)

// SetNeuronScheduler binds the real cluster scheduler before serving requests.
func (s *Server) SetNeuronScheduler(scheduler *cortex.Scheduler) {
	s.neuronScheduler = scheduler
}

type neuronStatus struct {
	Configured  bool                       `json:"configured"`
	Snapshot    cortex.Snapshot            `json:"snapshot"`
	Managed     []ManagedNeuron            `json:"managed,omitempty"`
	Performance []cortex.NeuronPerformance `json:"performance"`
}

func (s *Server) neuronStatusHandler(sessions *operator.Store) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := operatorActorFromRequest(sessions, r); !ok {
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}
		status := neuronStatus{}
		if s.neuronScheduler != nil {
			status.Configured = true
			status.Snapshot = s.neuronScheduler.Snapshot()
			status.Performance = status.Snapshot.PerformanceReport()
			if s.neuronManager != nil {
				status.Managed = s.neuronManager.ManagedNeurons()
			}
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		_ = json.NewEncoder(w).Encode(status)
	})
}
