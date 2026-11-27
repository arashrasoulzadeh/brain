package plugin

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"math"
	"regexp"
	"sort"
	"strings"
	"time"
)

const observationPrefix = "@observe/"
const maxObservationEvents = 16
const maxObservationPayload = 16 << 10
const maxObservationDeliveries = 32

var eventIdentifier = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_.-]{0,63}$`)

type EventSpec struct {
	Name   string            `yaml:"name" json:"name"`
	Fields map[string]string `yaml:"fields" json:"fields"`
	Source string            `yaml:"source,omitempty" json:"source,omitempty"`
}

type ObserverSpec struct {
	Plugin     string `yaml:"plugin"`
	Event      string `yaml:"event"`
	Function   string `yaml:"function"`
	Capability string `yaml:"capability"`
	Persist    bool   `yaml:"persist,omitempty"`
}

type ObservationIntent struct {
	Event   string          `json:"event"`
	Payload json.RawMessage `json:"payload"`
}

type Observation struct {
	ID        string          `json:"id"`
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
		if !eventIdentifier.MatchString(event.Name) || events[event.Name] || len(event.Fields) > 32 || (event.Source != "" && event.Source != "host") {
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

func (m Manifest) acceptsObservation(intent ObservationIntent, source string) bool {
	if len(intent.Payload) == 0 || len(intent.Payload) > maxObservationPayload {
		return false
	}
	var payload map[string]any
	if json.Unmarshal(intent.Payload, &payload) != nil || payload == nil {
		return false
	}
	for _, event := range m.Events {
		if event.Name != intent.Event || event.Source != source || len(event.Fields) != len(payload) {
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

// Delivery is bounded and post-commit. With SQLite, the host queues durable
// deliveries instead of blocking the publisher on observer execution. Observer functions cannot
// recursively publish events in this first implementation. No worker-provided
// identity is trusted: the envelope is constructed from the authorized request.
func (r *Registry) deliverObservations(publisher WorkerRequest, intents []ObservationIntent) {
	r.recordCommittedAdminObservations(publisher, intents, "")
	r.executeObservationPlan(r.planWorkerObservations(publisher, intents))
}

func (r *Registry) planWorkerObservations(publisher WorkerRequest, intents []ObservationIntent) []observationDelivery {
	if len(intents) == 0 {
		return nil
	}
	if len(intents) > maxObservationEvents {
		r.recordObservation("observation_rejected", publisher, "event limit")
		return nil
	}
	r.readinessMu.RLock()
	registered, exists := r.registered[publisher.PluginID+"@"+publisher.Version]
	r.readinessMu.RUnlock()
	if !exists || r.authorizeInvocation(publisher, false) != nil {
		return nil
	}
	return r.planObservations(publisher, registered.Manifest, intents, "")
}

// PublishHostObservation is called only by trusted, already-authorized host services
// after their effect succeeds. It cannot be reached through worker RPC. A host
// event is explicitly distinguished from worker-published events in the manifest.
func (r *Registry) PublishHostObservation(projectID, pluginID, event string, payload map[string]any) {
	if !isValidProjectID(projectID) || r.LifecycleStorageHealth() != nil {
		return
	}
	if _, err := r.resolveProject(projectID); err != nil {
		return
	}
	manifests := r.RegisteredManifests()
	var selected Manifest
	for _, manifest := range manifests {
		if manifest.ID != pluginID {
			continue
		}
		for _, spec := range manifest.Events {
			if spec.Name == event && spec.Source == "host" {
				if selected.ID != "" {
					return // Do not guess a publisher version.
				}
				selected = manifest
			}
		}
	}
	if selected.ID == "" {
		return
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return
	}
	publisher := WorkerRequest{Protocol: r.HostAPI, ProjectID: projectID, PluginID: selected.ID, Version: selected.Version}
	intents := []ObservationIntent{{Event: event, Payload: encoded}}
	r.recordCommittedAdminObservations(publisher, intents, "host")
	r.executeObservationPlan(r.planObservations(publisher, selected, intents, "host"))
}

func (r *Registry) planObservations(publisher WorkerRequest, manifest Manifest, intents []ObservationIntent, source string) []observationDelivery {
	deliveries := make([]observationDelivery, 0)
	metadata := publisher
	metadata.Payload, metadata.Project, metadata.HostContext = nil, nil, nil
	// Retain only the authorized resource selector for later admission checks.
	if route, ok := manifest.Route(publisher.Operation); ok && route.ResourceField != "" {
		if resource, err := invocationResource(publisher, route); err == nil {
			metadata.Payload, _ = json.Marshal(map[string]string{route.ResourceField: resource})
		}
	}
	for _, intent := range intents {
		if !manifest.acceptsObservation(intent, source) {
			r.recordObservation("observation_rejected", publisher, "invalid event")
			continue
		}
		var randomID [16]byte
		if _, err := rand.Read(randomID[:]); err != nil {
			r.recordObservation("observation_rejected", publisher, "identity unavailable")
			continue
		}
		id := hex.EncodeToString(randomID[:])
		payload, err := json.Marshal(Observation{ID: id, ProjectID: publisher.ProjectID, PluginID: publisher.PluginID, Version: publisher.Version, Event: intent.Event, Payload: intent.Payload})
		if err != nil {
			continue
		}
		for _, target := range r.observationTargets(publisher, intent.Event) {
			if len(deliveries) == maxObservationDeliveries {
				r.recordObservation("observation_rejected", publisher, "delivery limit")
				return deliveries
			}
			target.Payload = payload
			deliveries = append(deliveries, observationDelivery{ID: id, Publisher: metadata, Target: target, Source: source, Principal: publisher.accessPrincipal})
		}
	}
	return deliveries
}

func (delivery observationDelivery) detail() string {
	return delivery.Publisher.PluginID + "." + strings.TrimPrefix(delivery.Target.Operation, observationPrefix+delivery.Publisher.PluginID+"/")
}

func (r *Registry) executeObservationPlan(deliveries []observationDelivery) {
	for _, delivery := range deliveries {
		if store := r.observationStorage(); store != nil {
			if err := store.enqueueObservation(delivery, time.Now()); err != nil {
				r.recordObservation("observation_enqueue_failed", delivery.Target, delivery.detail())
			} else {
				r.recordObservation("observation_queued", delivery.Target, delivery.detail())
			}
			continue
		}
		result, err := r.invokeOnce(delivery.Target, true)
		if err != nil || !result.Success || result.ErrorCode != "" {
			r.recordObservation("observation_failed", delivery.Target, delivery.detail())
		} else {
			r.recordObservation("observation_delivered", delivery.Target, delivery.detail())
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
				target := WorkerRequest{Protocol: publisher.Protocol, ProjectID: publisher.ProjectID, PluginID: installation.PluginID, Version: installation.Version, Operation: observer.operation(), accessPrincipal: pluginAccessPrincipal(publisher.PluginID, publisher.Version)}
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

func (r *Registry) recordCommittedAdminObservations(publisher WorkerRequest, intents []ObservationIntent, source string) {
	r.readinessMu.RLock()
	defer r.readinessMu.RUnlock()
	r.recordCommittedAdminObservationsLocked(publisher, intents, source)
}

func (r *Registry) recordCommittedAdminObservationsLocked(publisher WorkerRequest, intents []ObservationIntent, source string) {
	if len(intents) > maxObservationEvents {
		return
	}
	registered, ok := r.registered[publisher.PluginID+"@"+publisher.Version]
	if !ok {
		return
	}
	for _, intent := range intents {
		if registered.Manifest.acceptsObservation(intent, source) {
			r.recordAdminObservation(publisher.ProjectID, publisher.PluginID+"."+intent.Event)
		}
	}
}
