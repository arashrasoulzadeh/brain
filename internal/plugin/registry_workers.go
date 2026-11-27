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

func (r *Registry) resolveProject(id string) (ProjectSnapshot, error) {
	r.readinessMu.RLock()
	resolve := r.projectResolver
	r.readinessMu.RUnlock()
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
	if err := validatePayloadProject(request); err != nil {
		return WorkerResponse{}, err
	}
	if response, handled, err := r.invokePluginCall(request); handled {
		return response, err
	}
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
	if err := validatePayloadProject(request); err != nil {
		return WorkerResponse{}, err
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

	if !observer && override == nil {
		if result, handled, err := r.invokeModel(request); handled {
			return result, err
		}
	}

	// Prepare reads and proposed writes, but do not mutate durable state until
	// the worker has succeeded and lifecycle and grants have been rechecked.
	prepared := request
	var commit func() error
	var atomicCommit func(*WorkerResponse, []observationDelivery) error
	if planner != nil {
		var err error
		if responsePlanner, ok := planner.(interface {
			prepareResponseWithObservations(WorkerRequest) (WorkerRequest, func(*WorkerResponse, []observationDelivery) error, error)
		}); ok {
			prepared, atomicCommit, err = responsePlanner.prepareResponseWithObservations(request)
		} else {
			prepared, commit, err = planner.Prepare(request)
		}
		if err != nil {
			return WorkerResponse{}, err
		}
		if prepared.ProjectID != request.ProjectID || prepared.PluginID != request.PluginID || prepared.Version != request.Version || prepared.Operation != request.Operation || prepared.Protocol != request.Protocol {
			return WorkerResponse{}, ErrInvalidWorkerRequest
		}
		// A trusted storage planner may enrich a payload but must never
		// substitute a different resource after authorization.
		r.readinessMu.RLock()
		route, _ := r.registered[request.PluginID+"@"+request.Version].Manifest.Route(request.Operation)
		r.readinessMu.RUnlock()
		if route.ResourceField != "" {
			before, beforeErr := invocationResource(request, route)
			after, afterErr := invocationResource(prepared, route)
			if beforeErr != nil || afterErr != nil || before != after {
				return WorkerResponse{}, ErrInvalidWorkerRequest
			}
		}
	}
	hostContext, err := r.invocationContext(request, observer)
	if err != nil {
		return WorkerResponse{}, err
	}
	prepared.HostContext = hostContext
	prepared.Project = &model
	response, err := InvokeForProject(worker, prepared, r.HostAPI, request.ProjectID)
	if err != nil || !response.Success || response.ErrorCode != "" {
		// A failed response-aware operation must never expose private state.
		if request.PluginID == "auth" && atomicCommit != nil {
			response.Payload = nil
			if response.ErrorCode != "" {
				response.ErrorCode = "authentication failed"
			}
		}
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
	if request.httpRoute != nil {
		r.lifecycleMu.RLock()
		defer r.lifecycleMu.RUnlock()
	}
	r.readinessMu.RLock()
	defer r.readinessMu.RUnlock()
	r.installationMu.RLock()
	defer r.installationMu.RUnlock()
	if err := r.authorizeInvocationLocked(request, observer); err != nil {
		return WorkerResponse{}, err
	}
	if atomicCommit != nil {
		if err := atomicCommit(&response, deliveries); err != nil {
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
	if request.httpRoute != nil {
		r.lifecycleMu.RLock()
		defer r.lifecycleMu.RUnlock()
	}
	r.readinessMu.RLock()
	defer r.readinessMu.RUnlock()
	r.installationMu.RLock()
	defer r.installationMu.RUnlock()
	return r.authorizeInvocationLocked(request, observer)
}

func (r *Registry) authorizeInvocationLocked(request WorkerRequest, observer bool) error {
	if request.httpRoute != nil {
		if err := r.checkHTTPRouteLocked(request); err != nil {
			return err
		}
	}
	if request.requireAccessPolicy && !r.accessInstalledLocked(request.ProjectID) {
		return ErrActorDenied
	}
	if err := r.LifecycleStorageHealth(); err != nil {
		return err
	}
	if r.projectResolver != nil {
		model, err := r.projectResolver(request.ProjectID)
		if err != nil || model.ID != request.ProjectID || !model.Active {
			return ErrProjectInvalid
		}
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
	observerResource := ""
	if observer {
		if installation.Draining {
			return ErrDisabled
		}
		for _, subscription := range registered.Manifest.Observers {
			if subscription.operation() == request.Operation {
				route, declared = Route{Capability: subscription.Capability}, true
				observerResource = subscription.Plugin + "." + subscription.Event
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
	resource, scopeErr := invocationResource(request, route)
	if observer {
		resource = observerResource
	}
	if resources := installation.Resources[route.Capability]; len(resources) > 0 && (scopeErr != nil || resource == "" || !resources[resource]) {
		return ErrGrantDenied
	}
	return r.authorizeAccessLocked(request.ProjectID, request.accessPrincipal, request.PluginID, request.Version, request.Operation, route.Capability, resource, scopeErr)
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
	if worker == nil {
		return ErrWorkerUnavailable
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

func validatePayloadProject(request WorkerRequest) error {
	var payload map[string]json.RawMessage
	if json.Unmarshal(request.Payload, &payload) == nil {
		// Durable snapshots are reserved for the trusted host storage adapter.
		if _, supplied := payload["_brainArena"]; request.PluginID == "arena" && supplied {
			return ErrInvalidWorkerRequest
		}
		if _, supplied := payload["_brainAuth"]; request.PluginID == "auth" && supplied {
			return ErrInvalidWorkerRequest
		}
		if _, supplied := payload["_brainInvites"]; request.PluginID == "invites" && supplied {
			return ErrInvalidWorkerRequest
		}
		if _, supplied := payload["_brainSubscriptions"]; request.PluginID == "subscriptions" && supplied {
			return ErrInvalidWorkerRequest
		}
		for _, key := range []string{"projectId", "project_id"} {
			if value, exists := payload[key]; exists {
				var id string
				if json.Unmarshal(value, &id) != nil || id != request.ProjectID {
					return ErrInvalidWorkerRequest
				}
			}
		}
	}
	return nil
}
