package plugin

import (
	"encoding/json"
	"math"
	"regexp"
	"sort"
)

const observationPrefix = "@observe/"
const maxObservationEvents = 16
const maxObservationPayload = 16 << 10
const maxObservationDeliveries = 32

var eventIdentifier = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_.-]{0,63}$`)

type EventSpec struct {
	Name   string            `yaml:"name" json:"name"`
	Fields map[string]string `yaml:"fields" json:"fields"`
}

type ObserverSpec struct {
	Plugin     string `yaml:"plugin"`
	Event      string `yaml:"event"`
	Function   string `yaml:"function"`
	Capability string `yaml:"capability"`
}

type ObservationIntent struct {
	Event   string          `json:"event"`
	Payload json.RawMessage `json:"payload"`
}

type Observation struct {
	ProjectID string          `json:"projectId"`
	PluginID  string          `json:"pluginId"`
	Version   string          `json:"version"`
	Event     string          `json:"event"`
	Payload   json.RawMessage `json:"payload"`
}

func (s ObserverSpec) operation() string {
	return observationPrefix + s.Plugin + "/" + s.Event
}

func validateObservationContract(m Manifest, capabilities map[string]bool) error {
	events := map[string]bool{}
	for _, event := range m.Events {
		if !eventIdentifier.MatchString(event.Name) || events[event.Name] || len(event.Fields) > 32 {
			return ErrInvalidManifest
		}
		events[event.Name] = true
		for field, kind := range event.Fields {
			if !validResponseKey(field) || (kind != "string" && kind != "number" && kind != "boolean") {
				return ErrInvalidManifest
			}
		}
	}
	observers := map[string]bool{}
	for _, observer := range m.Observers {
		if !eventIdentifier.MatchString(observer.Plugin) || !eventIdentifier.MatchString(observer.Event) ||
			!validResponseKey(observer.Function) || !capabilities[observer.Capability] ||
			observer.Capability != "observe."+observer.Plugin+"."+observer.Event ||
			observers[observer.operation()] {
			return ErrInvalidManifest
		}
		observers[observer.operation()] = true
	}
	return nil
}

func (m Manifest) acceptsObservation(intent ObservationIntent) bool {
	if len(intent.Payload) == 0 || len(intent.Payload) > maxObservationPayload {
		return false
	}
	var payload map[string]any
	if json.Unmarshal(intent.Payload, &payload) != nil || payload == nil {
		return false
	}
	for _, event := range m.Events {
		if event.Name != intent.Event || len(event.Fields) != len(payload) {
			continue
		}
		for name, kind := range event.Fields {
			switch value := payload[name].(type) {
			case string:
				if kind != "string" || len(value) > 1024 {
					return false
				}
			case float64:
				if kind != "number" || math.IsInf(value, 0) || math.IsNaN(value) {
					return false
				}
			case bool:
				if kind != "boolean" {
					return false
				}
			default:
				return false
			}
		}
		return true
	}
	return false
}

// Delivery is best effort, bounded, and post-commit. Observer functions cannot
// recursively publish events in this first implementation. No worker-provided
// identity is trusted: the envelope is constructed from the authorized request.
func (r *Registry) deliverObservations(publisher WorkerRequest, intents []ObservationIntent) {
	if len(intents) == 0 {
		return
	}
	if len(intents) > maxObservationEvents {
		r.recordObservation("observation_rejected", publisher, "event limit")
		return
	}
	r.readinessMu.RLock()
	registered, exists := r.registered[publisher.PluginID+"@"+publisher.Version]
	r.readinessMu.RUnlock()
	if !exists || r.authorizeInvocation(publisher, false) != nil {
		return
	}
	deliveries := 0
	for _, intent := range intents {
		if !registered.Manifest.acceptsObservation(intent) {
			r.recordObservation("observation_rejected", publisher, "invalid event")
			continue
		}
		r.recordAdminObservation(publisher.ProjectID, publisher.PluginID+"."+intent.Event)
		payload, err := json.Marshal(Observation{ProjectID: publisher.ProjectID, PluginID: publisher.PluginID, Version: publisher.Version, Event: intent.Event, Payload: intent.Payload})
		if err != nil {
			continue
		}
		for _, target := range r.observationTargets(publisher, intent.Event) {
			if deliveries == maxObservationDeliveries {
				r.recordObservation("observation_rejected", publisher, "delivery limit")
				return
			}
			deliveries++
			target.Payload = payload
			result, err := r.invokeOnce(target, true)
			if err != nil || !result.Success || result.ErrorCode != "" {
				r.recordObservation("observation_failed", target, publisher.PluginID+"."+intent.Event)
			} else {
				r.recordObservation("observation_delivered", target, publisher.PluginID+"."+intent.Event)
			}
		}
	}
}

func (r *Registry) observationTargets(publisher WorkerRequest, event string) []WorkerRequest {
	r.lifecycleMu.RLock()
	defer r.lifecycleMu.RUnlock()
	r.readinessMu.RLock()
	defer r.readinessMu.RUnlock()
	r.installationMu.RLock()
	defer r.installationMu.RUnlock()
	candidates := map[string][]WorkerRequest{}
	for _, installation := range r.installations {
		if installation.ProjectID != publisher.ProjectID || !installation.Enabled || installation.Draining {
			continue
		}
		if active, ok := r.active[publisher.ProjectID+":"+installation.PluginID]; ok && active != installation.Version {
			continue
		}
		manifest, ok := r.registered[installation.PluginID+"@"+installation.Version]
		if !ok {
			continue
		}
		for _, observer := range manifest.Manifest.Observers {
			if observer.Plugin == publisher.PluginID && observer.Event == event {
				target := WorkerRequest{Protocol: publisher.Protocol, ProjectID: publisher.ProjectID, PluginID: installation.PluginID, Version: installation.Version, Operation: observer.operation(), accessPrincipal: pluginAccessPrincipal(publisher.PluginID, publisher.Version), originPrincipal: publisher.originalPrincipal()}
				if r.authorizeInvocationLocked(target, true) == nil {
					candidates[installation.PluginID] = append(candidates[installation.PluginID], target)
				}
			}
		}
	}
	targets := []WorkerRequest{}
	for _, versions := range candidates {
		// If multiple versions are enabled, require explicit promotion rather
		// than delivering twice or arbitrarily choosing an implementation.
		if len(versions) == 1 {
			targets = append(targets, versions[0])
		}
	}
	sort.Slice(targets, func(i, j int) bool { return targets[i].PluginID < targets[j].PluginID })
	return targets
}

func (r *Registry) recordObservation(action string, request WorkerRequest, detail string) {
	r.appendAudit(AuditEvent{Action: action, ProjectID: request.ProjectID, PluginID: request.PluginID, Version: request.Version, Detail: detail})
}
