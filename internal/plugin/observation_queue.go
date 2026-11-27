package plugin

import (
	"context"
	"encoding/json"
	"errors"
	"time"
)

const maxObservationAttempts = 3
const maxPendingObservations = 10000
const maxPendingProjectObservations = 1000

var ErrObservationCapacity = errors.New("observation queue capacity reached")

type observationDelivery struct {
	ID                  string
	Publisher           WorkerRequest
	Target              WorkerRequest
	Source              string
	Principal           string
	RequireAccessPolicy bool
	Attempts            int
}

func (r *Registry) bindObservationStorage(storage *SQLiteStorage) {
	r.observationMu.Lock()
	defer r.observationMu.Unlock()
	r.observationStore = storage
}

func (r *Registry) observationStorage() *SQLiteStorage {
	r.observationMu.RLock()
	defer r.observationMu.RUnlock()
	return r.observationStore
}

// RunObservationDelivery is the single-process SQLite delivery loop. A failed
// observer never changes the publisher result. Multi-host delivery requires a
// shared queue with leases and is intentionally not enabled by this adapter.
func (r *Registry) RunObservationDelivery(ctx context.Context, interval time.Duration) {
	if interval <= 0 {
		return
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			_, _ = r.DrainObservations(now, maxObservationDeliveries)
		}
	}
}

// DrainObservations is a trusted host operation, not a worker-callable API.
// Queue attempts are serialized; project and grants are checked for every
// delivery, including after process restart.
func (r *Registry) DrainObservations(now time.Time, limit int) (int, error) {
	if limit < 1 || limit > maxObservationDeliveries {
		return 0, ErrInvalidWorkerRequest
	}
	r.observationDrainMu.Lock()
	defer r.observationDrainMu.Unlock()
	store := r.observationStorage()
	if store == nil {
		return 0, nil
	}
	deliveries, err := store.dueObservations(now, limit)
	if err != nil {
		return 0, err
	}
	for i, delivery := range deliveries {
		delivery.Publisher.accessPrincipal = delivery.Principal
		delivery.Publisher.requireAccessPolicy = delivery.RequireAccessPolicy
		delivery.Target.requireAccessPolicy = delivery.RequireAccessPolicy
		delivery.Target.accessPrincipal = pluginAccessPrincipal(delivery.Publisher.PluginID, delivery.Publisher.Version)
		delivery.Target.originPrincipal = delivery.Publisher.originalPrincipal()
		event, persist, err := r.validateQueuedObservation(delivery)
		if err == nil {
			response, invokeErr := r.invokeOnce(delivery.Target, true)
			err = invokeErr
			if err == nil && (!response.Success || response.ErrorCode != "") {
				err = errors.New("observation handler failed")
			}
		}
		detail := delivery.Publisher.PluginID + "." + event.Event
		if err == nil {
			if err := store.finishObservation(delivery, persist, now); err != nil {
				return i, err
			}
			r.recordObservation("observation_delivered", delivery.Target, detail)
			continue
		}
		permanent := errors.Is(err, ErrGrantDenied) || errors.Is(err, ErrDisabled) ||
			errors.Is(err, ErrNotInstalled) || errors.Is(err, ErrNotRegistered) ||
			errors.Is(err, ErrInvalidWorkerRequest) || errors.Is(err, ErrProjectInvalid) || errors.Is(err, ErrActorDenied)
		if permanent || delivery.Attempts+1 >= maxObservationAttempts {
			if err := store.finishObservation(delivery, false, now); err != nil {
				return i, err
			}
			r.recordObservation("observation_failed", delivery.Target, detail)
			continue
		}
		if err := store.retryObservation(delivery, now); err != nil {
			return i, err
		}
		r.recordObservation("observation_retry", delivery.Target, detail)
	}
	return len(deliveries), nil
}

func (r *Registry) validateQueuedObservation(delivery observationDelivery) (Observation, bool, error) {
	var event Observation
	publisher, target := delivery.Publisher, delivery.Target
	if err := json.Unmarshal(target.Payload, &event); err != nil || event.ID != delivery.ID ||
		event.ProjectID != publisher.ProjectID || target.ProjectID != publisher.ProjectID ||
		event.PluginID != publisher.PluginID || event.Version != publisher.Version ||
		target.Operation != observationPrefix+publisher.PluginID+"/"+event.Event ||
		(delivery.Source != "" && delivery.Source != "host") {
		return event, false, ErrInvalidWorkerRequest
	}
	if _, err := r.resolveProject(target.ProjectID); err != nil {
		return event, false, ErrProjectInvalid
	}
	if delivery.Source == "" {
		if err := r.authorizeInvocation(publisher, false); err != nil {
			return event, false, err
		}
	}
	if version, ok := r.ActiveVersion(target.ProjectID, target.PluginID); ok && version != target.Version {
		return event, false, ErrDisabled
	}
	r.readinessMu.RLock()
	registeredPublisher, published := r.registered[publisher.PluginID+"@"+publisher.Version]
	registeredTarget, subscribed := r.registered[target.PluginID+"@"+target.Version]
	r.readinessMu.RUnlock()
	if !published || !subscribed || !registeredPublisher.Manifest.acceptsObservation(ObservationIntent{Event: event.Event, Payload: event.Payload}, delivery.Source) {
		return event, false, ErrInvalidWorkerRequest
	}
	for _, spec := range registeredTarget.Manifest.Observers {
		if spec.operation() == target.Operation {
			if err := r.authorizeInvocation(target, true); err != nil {
				return event, false, err
			}
			return event, spec.Persist, nil
		}
	}
	return event, false, ErrGrantDenied
}
