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
	request = r.prepareStorageRequest(request)
	key := request.ProjectID + ":" + request.PluginID + "@" + request.Version
	r.installationMu.RLock()
	installation, installed := r.installations[key]
	r.installationMu.RUnlock()
	if !installed {
		return WorkerResponse{}, ErrNotInstalled
	}
	if !installation.Enabled {
		return WorkerResponse{}, ErrDisabled
	}
	r.readinessMu.RLock()
	ready := r.readiness[request.PluginID+"@"+request.Version]
	r.readinessMu.RUnlock()
	if !ready {
		return WorkerResponse{}, ErrNotReady
	}
	r.grantMu.RLock()
	registered, exists := r.registered[request.PluginID+"@"+request.Version]
	granted := exists
	if exists {
		for _, capability := range registered.Manifest.Capabilities {
			if !installation.Grants[capability] {
				granted = false
				break
			}
		}
	}
	r.grantMu.RUnlock()
	if !granted {
		return WorkerResponse{}, ErrGrantDenied
	}
	r.workerMu.RLock()
	worker := r.workers[key]
	r.workerMu.RUnlock()
	return InvokeForProject(worker, request, r.HostAPI, request.ProjectID)
}

// StartWorker starts a worker for an already enabled installation that lost its
// runtime process. Lifecycle and grants are rechecked before the worker starts.
func (r *Registry) StartWorker(projectID, id, version string) error {
	r.installationMu.RLock()
	installation, installed := r.installations[projectID+":"+id+"@"+version]
	r.installationMu.RUnlock()
	if !installed {
		return ErrNotInstalled
	}
	if !installation.Enabled {
		return ErrDisabled
	}
	r.readinessMu.RLock()
	ready := r.readiness[id+"@"+version]
	r.readinessMu.RUnlock()
	if !ready {
		return ErrNotReady
	}
	if r.workerFactory == nil {
		return ErrWorkerUnavailable
	}
	key := projectID + ":" + id + "@" + version
	r.workerMu.RLock()
	worker := r.workers[key]
	r.workerMu.RUnlock()
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
	r.workerMu.Lock()
	r.workers[key] = worker
	r.workerMu.Unlock()
	return nil
}
