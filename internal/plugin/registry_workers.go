package plugin

import (
	"encoding/json"
	"errors"
	"time"
)

func (r *Registry) BindSupervisor(supervisor *Supervisor) {
	r.supervisor = supervisor
}

func (r *Registry) BindWebSocketRegistry(webSockets *WebSocketRegistry) {
	r.webSockets = webSockets
}

func (r *Registry) BindWorkerFactory(factory func(projectID, pluginID, version string) (ContractWorker, error)) {
	r.workerMu.Lock()
	defer r.workerMu.Unlock()
	r.workerFactory = factory
}

// BindProjectResolver installs the authoritative project model resolver at host startup.
func (r *Registry) BindProjectResolver(resolve func(string) (ProjectSnapshot, error)) {
	r.workerMu.Lock()
	defer r.workerMu.Unlock()
	r.projectResolver = resolve
}

func (r *Registry) resolveProject(id string) (ProjectSnapshot, error) {
	r.workerMu.RLock()
	resolve := r.projectResolver
	r.workerMu.RUnlock()
	if resolve == nil {
		// Standalone registry fixtures have no project store.
		return ProjectSnapshot{ID: id, Active: true}, nil
	}
	model, err := resolve(id)
	if err != nil {
		return ProjectSnapshot{}, err
	}
	if model.ID != id || !model.Active {
		return ProjectSnapshot{}, ErrProjectInvalid
	}
	return model, nil
}

// BindWorker attaches the host-created runtime worker for one project plugin.
// Lifecycle state remains authoritative; a bound worker alone can never execute.
func (r *Registry) BindWorker(projectID, id, version string, worker ContractWorker) error {
	if !isValidProjectID(projectID) || id == "" || version == "" || worker == nil {
		return ErrInvalidWorkerRequest
	}
	r.workerMu.Lock()
	defer r.workerMu.Unlock()
	r.workers[projectID+":"+id+"@"+version] = worker
	return nil
}

func (r *Registry) UnbindWorker(projectID, id, version string) {
	r.workerMu.Lock()
	defer r.workerMu.Unlock()
	delete(r.workers, projectID+":"+id+"@"+version)
}

// Invoke executes only an installed, ready, enabled and fully granted plugin.
func (r *Registry) Invoke(request WorkerRequest) (WorkerResponse, error) {
	response, err := r.invokeOnce(request, false)
	return r.finishWorkerInvocation(request, response, err)
}

func (r *Registry) finishWorkerInvocation(request WorkerRequest, response WorkerResponse, err error) (WorkerResponse, error) {
	intents := response.Observations
	response.Observations = nil
	// invokeOnce releases storage, installation and readiness locks before delivery.
	if err == nil && response.Success && response.ErrorCode == "" {
		r.deliverObservations(request, intents)
	}
	return response, err
}

func (r *Registry) invokeOnce(request WorkerRequest, observer bool) (WorkerResponse, error) {
	return r.invokeUsingWorker(request, observer, nil, nil)
}

// invokeUsingWorker is shared by installed workers and credential-bound external
// workers. Only the trusted broker boundary can supply a worker override.
func (r *Registry) invokeUsingWorker(request WorkerRequest, observer bool, override ContractWorker, commitGuard func() error) (WorkerResponse, error) {
	if err := ValidateWorkerRequest(request, r.HostAPI); err != nil {
		return WorkerResponse{}, err
	}
	var payload map[string]any
	if json.Unmarshal(request.Payload, &payload) == nil {
		for _, key := range []string{"projectId", "project_id"} {
			if value, exists := payload[key]; exists && value != request.ProjectID {
				return WorkerResponse{}, ErrInvalidWorkerRequest
			}
		}
	}
	model, err := r.resolveProject(request.ProjectID)
	if err != nil {
		return WorkerResponse{}, err
	}
	request.Project = &model
	// Serialize read/modify/write plans for the same project and plugin.
	planner, unlock := r.lockStorageInvocation(request)
	defer unlock()
	if err := r.authorizeInvocation(request, observer); err != nil {
		return WorkerResponse{}, err
	}
	key := request.ProjectID + ":" + request.PluginID + "@" + request.Version
	worker := override
	if worker == nil {
		r.workerMu.RLock()
		worker = r.workers[key]
		r.workerMu.RUnlock()
	}
	if worker == nil || !worker.Ready() {
		return WorkerResponse{}, ErrWorkerNotReady
	}

	// Prepare reads and proposed writes, but do not mutate durable state until
	// the worker has succeeded and lifecycle and grants have been rechecked.
	prepared := request
	var commit func() error
	var atomicCommit func([]observationDelivery) error
	if planner != nil {
		var err error
		if atomic, ok := planner.(interface {
			prepareWithObservations(WorkerRequest) (WorkerRequest, func([]observationDelivery) error, error)
		}); ok {
			prepared, atomicCommit, err = atomic.prepareWithObservations(request)
		} else {
			prepared, commit, err = planner.Prepare(request)
		}
		if err != nil {
			return WorkerResponse{}, err
		}
		if prepared.ProjectID != request.ProjectID || prepared.PluginID != request.PluginID || prepared.Version != request.Version || prepared.Operation != request.Operation || prepared.Protocol != request.Protocol {
			return WorkerResponse{}, ErrInvalidWorkerRequest
		}
	}
	prepared.Project = &model
	response, err := InvokeForProject(worker, prepared, r.HostAPI, request.ProjectID)
	if err != nil || !response.Success || response.ErrorCode != "" {
		return response, err
	}
	if response.Protocol != request.Protocol {
		return WorkerResponse{}, ErrInvalidWorkerRequest
	}
	if _, err := r.resolveProject(request.ProjectID); err != nil {
		return WorkerResponse{}, err
	}
	if commitGuard != nil {
		if err := commitGuard(); err != nil {
			return WorkerResponse{}, err
		}
	}
	var deliveries []observationDelivery
	if atomicCommit != nil && !observer {
		deliveries = r.planWorkerObservations(request, response.Observations)
	}
	r.readinessMu.RLock()
	defer r.readinessMu.RUnlock()
	r.installationMu.RLock()
	defer r.installationMu.RUnlock()
	if err := r.authorizeInvocationLocked(request, observer); err != nil {
		return WorkerResponse{}, err
	}
	if atomicCommit != nil {
		if err := atomicCommit(deliveries); err != nil {
			return WorkerResponse{}, err
		}
		for _, delivery := range deliveries {
			r.recordObservation("observation_queued", delivery.Target, delivery.detail())
		}
		if !observer {
			r.recordCommittedAdminObservationsLocked(request, response.Observations, "")
		}
		response.Observations = nil
	} else if commit != nil {
		if err := commit(); err != nil {
			return WorkerResponse{}, err
		}
	}
	return response, nil
}

func (r *Registry) authorizeInvocation(request WorkerRequest, observer bool) error {
	r.readinessMu.RLock()
	defer r.readinessMu.RUnlock()
	r.installationMu.RLock()
	defer r.installationMu.RUnlock()
	return r.authorizeInvocationLocked(request, observer)
}

func (r *Registry) authorizeInvocationLocked(request WorkerRequest, observer bool) error {
	if err := r.LifecycleStorageHealth(); err != nil {
		return err
	}
	manifestKey := request.PluginID + "@" + request.Version
	installation, installed := r.installations[request.ProjectID+":"+manifestKey]
	if !installed {
		return ErrNotInstalled
	}
	if !installation.Enabled || (installation.Draining && (installation.DrainUntil.IsZero() || !time.Now().Before(installation.DrainUntil))) {
		return ErrDisabled
	}
	if !r.readiness[manifestKey] {
		return ErrNotReady
	}
	registered, exists := r.registered[manifestKey]
	if !exists {
		return ErrGrantDenied
	}
	route, declared := registered.Manifest.Route(request.Operation)
	if observer {
		if installation.Draining {
			return ErrDisabled
		}
		for _, subscription := range registered.Manifest.Observers {
			if subscription.operation() == request.Operation {
				route, declared = Route{Capability: subscription.Capability}, true
				if resources := installation.Resources[route.Capability]; len(resources) > 0 && !resources[subscription.Plugin+"."+subscription.Event] {
					return ErrGrantDenied
				}
				break
			}
		}
	}
	if !declared || route.Helper == "host.admin" || !installation.Grants[route.Capability] {
		return ErrGrantDenied
	}
	for _, capability := range registered.Manifest.Capabilities {
		if !installation.Grants[capability] {
			return ErrGrantDenied
		}
	}
	return nil
}

// StartWorker starts a worker for an already enabled installation that lost its
// runtime process. Lifecycle and grants are rechecked before the worker starts.
func (r *Registry) StartWorker(projectID, id, version string) error {
	r.readinessMu.RLock()
	defer r.readinessMu.RUnlock()
	r.installationMu.RLock()
	defer r.installationMu.RUnlock()
	if err := r.LifecycleStorageHealth(); err != nil {
		return err
	}
	installation, installed := r.installations[projectID+":"+id+"@"+version]
	if !installed {
		return ErrNotInstalled
	}
	if !installation.Enabled || installation.Draining {
		return ErrDisabled
	}
	if !r.readiness[id+"@"+version] {
		return ErrNotReady
	}
	registered, ok := r.registered[id+"@"+version]
	if !ok {
		return ErrNotRegistered
	}
	for _, capability := range registered.Manifest.Capabilities {
		if !installation.Grants[capability] {
			return ErrGrantDenied
		}
	}
	r.workerMu.Lock()
	defer r.workerMu.Unlock()
	if r.workerFactory == nil {
		return ErrWorkerUnavailable
	}
	key := projectID + ":" + id + "@" + version
	worker := r.workers[key]
	if worker != nil && worker.Ready() {
		return nil
	}
	worker, err := r.workerFactory(projectID, id, version)
	if err != nil {
		return err
	}
	if err := worker.Start(); err != nil {
		return err
	}
	r.workers[key] = worker
	return nil
}

// stopBoundWorker removes the actual runtime worker, not just supervisor state.
// The caller holds the installation lock so no new invocation can be admitted.
func (r *Registry) stopBoundWorker(key string) error {
	r.workerMu.Lock()
	defer r.workerMu.Unlock()
	worker := r.workers[key]
	delete(r.workers, key)
	if worker == nil {
		return nil
	}
	if err := worker.Stop(); err != nil && !errors.Is(err, ErrTypeScriptWorkerStopped) {
		return err
	}
	return nil
}
