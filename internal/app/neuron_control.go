package app

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strings"

	"brain/internal/operator"
)

// ManagedNeuron contains only process identity and liveness. The manager owns
// its launch configuration, never the API client.
type ManagedNeuron struct {
	ID      string
	Running bool
}

type ManagedNeuronController interface {
	ManagedNeurons() []ManagedNeuron
	ControlNeuron(context.Context, string, string) error
}

func (s *Server) SetManagedNeuronController(manager ManagedNeuronController) {
	s.neuronManager = manager
}

func (s *Server) neuronControlHandler(sessions *operator.Store) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		actor, ok := operatorActorFromRequest(sessions, r)
		if !ok {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		if s.neuronScheduler == nil {
			http.Error(w, "Neuron scheduler unavailable", http.StatusServiceUnavailable)
			return
		}
		var command struct {
			ID        string `json:"id"`
			Action    string `json:"action"`
			Confirmed bool   `json:"confirmed"`
		}
		decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&command); err != nil || !errors.Is(decoder.Decode(new(any)), io.EOF) ||
			command.ID == "" || len(command.ID) > 1024 || strings.TrimSpace(command.ID) != command.ID || !command.Confirmed {
			http.Error(w, "invalid or unconfirmed Neuron command", http.StatusBadRequest)
			return
		}
		switch command.Action {
		case "drain", "resume", "start", "stop", "restart":
		default:
			http.Error(w, "unsupported Neuron action", http.StatusBadRequest)
			return
		}
		managed := false
		if s.neuronManager != nil {
			for _, neuron := range s.neuronManager.ManagedNeurons() {
				if neuron.ID == command.ID {
					managed = true
					break
				}
			}
		}
		var err error
		if managed {
			err = s.neuronManager.ControlNeuron(r.Context(), command.ID, command.Action)
		} else {
			if _, connected := s.neuronScheduler.Neuron(command.ID); !connected {
				http.Error(w, "Neuron unavailable", http.StatusNotFound)
				return
			}
			if command.Action != "drain" && command.Action != "resume" {
				http.Error(w, "process controls require a Brain-managed local Neuron", http.StatusConflict)
				return
			}
			err = s.neuronScheduler.SetOperatorDrain(command.ID, command.Action == "drain")
		}
		if err != nil {
			log.Printf("Neuron control failed: operator=%q neuron=%q action=%q", actor.ID, command.ID, command.Action)
			http.Error(w, "Neuron action unavailable; inspect current capacity and retry", http.StatusConflict)
			return
		}
		log.Printf("Neuron control completed: operator=%q neuron=%q action=%q", actor.ID, command.ID, command.Action)
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		_ = json.NewEncoder(w).Encode(map[string]any{"status": "success", "id": command.ID, "action": command.Action})
	})
}
