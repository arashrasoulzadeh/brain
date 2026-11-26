package plugin

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
	if err := ValidateWorkerRequest(request, r.HostAPI); err != nil {
		return WorkerResponse{}, err
	}
	// Serialize read/modify/write plans for the same project and plugin.
	planner, unlock := r.lockStorageInvocation(request)
	defer unlock()
	if err := r.authorizeInvocation(request); err != nil {
		return WorkerResponse{}, err
	}
	key := request.ProjectID + ":" + request.PluginID + "@" + request.Version
	r.workerMu.RLock()
	worker := r.workers[key]
	r.workerMu.RUnlock()
	if worker == nil || !worker.Ready() {
		return WorkerResponse{}, ErrWorkerNotReady
	}

	// Prepare reads and proposed writes, but do not mutate durable state until
	// the worker has succeeded and lifecycle and grants have been rechecked.
	prepared := request
	var commit func() error
	if planner != nil {
		var err error
		prepared, commit, err = planner.Prepare(request)
		if err != nil {
			return WorkerResponse{}, err
		}
		if prepared.ProjectID != request.ProjectID || prepared.PluginID != request.PluginID || prepared.Version != request.Version || prepared.Operation != request.Operation || prepared.Protocol != request.Protocol {
			return WorkerResponse{}, ErrInvalidWorkerRequest
		}
	}
	response, err := InvokeForProject(worker, prepared, r.HostAPI, request.ProjectID)
	if err != nil || !response.Success || response.ErrorCode != "" {
		return response, err
	}
	if response.Protocol != request.Protocol {
		return WorkerResponse{}, ErrInvalidWorkerRequest
	}
	r.readinessMu.RLock()
	defer r.readinessMu.RUnlock()
	r.installationMu.RLock()
	defer r.installationMu.RUnlock()
	if err := r.authorizeInvocationLocked(request); err != nil {
		return WorkerResponse{}, err
	}
	if commit != nil {
		if err := commit(); err != nil {
			return WorkerResponse{}, err
		}
	}
	return response, nil
}

func (r *Registry) authorizeInvocation(request WorkerRequest) error {
	r.readinessMu.RLock()
	defer r.readinessMu.RUnlock()
	r.installationMu.RLock()
	defer r.installationMu.RUnlock()
	return r.authorizeInvocationLocked(request)
}

func (r *Registry) authorizeInvocationLocked(request WorkerRequest) error {
	manifestKey := request.PluginID + "@" + request.Version
	installation, installed := r.installations[request.ProjectID+":"+manifestKey]
	if !installed {
		return ErrNotInstalled
	}
	if !installation.Enabled {
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
	if !declared || !installation.Grants[route.Capability] {
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
