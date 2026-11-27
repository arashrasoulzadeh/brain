package plugin

import "slices"

// PluginCall is a fixed, manifest-declared dependency operation, not a
// caller-selected target. Only source-free pure operations may be delegated.
type PluginCall struct {
	Plugin    string `yaml:"plugin" json:"plugin"`
	Version   string `yaml:"version" json:"version"`
	Operation string `yaml:"operation" json:"operation"`
}

func (c *PluginCall) valid(m Manifest, capability string) bool {
	if c == nil || c.Plugin == m.ID || !eventIdentifier.MatchString(c.Plugin) || !eventIdentifier.MatchString(c.Operation) || c.Version == "" {
		return false
	}
	key := c.Plugin + "@" + c.Version
	return capability == "call."+c.Plugin+"."+c.Operation && (slices.Contains(m.Dependencies, key) || slices.Contains(m.OptionalDependencies, key))
}

func (r *Registry) checkPluginCall(source, target WorkerRequest) error {
	r.lifecycleMu.RLock()
	defer r.lifecycleMu.RUnlock()
	r.readinessMu.RLock()
	defer r.readinessMu.RUnlock()
	r.installationMu.RLock()
	defer r.installationMu.RUnlock()
	for _, request := range []WorkerRequest{source, target} {
		if err := r.authorizeInvocationLocked(request, false); err != nil {
			return err
		}
		key := request.PluginID + "@" + request.Version
		installation := r.installations[request.ProjectID+":"+key]
		route, _ := r.registered[key].Manifest.Route(request.Operation)
		if installation.Draining || len(installation.Resources[route.Capability]) > 0 {
			return ErrGrantDenied
		}
		if active, ok := r.active[request.ProjectID+":"+request.PluginID]; ok && active != request.Version {
			return ErrGrantDenied
		}
	}
	manifest := r.registered[target.PluginID+"@"+target.Version].Manifest
	route, ok := manifest.Route(target.Operation)
	if !ok || !manifest.Declarative || (route.Helper != "sdk.response" && route.Helper != "sdk.pipeline") {
		return ErrGrantDenied
	}
	if route.Pipeline != nil {
		for _, step := range route.Pipeline.Steps {
			if step.Helper == "observe.fire" {
				return ErrGrantDenied
			}
		}
	}
	return nil
}

func (r *Registry) invokePluginCall(request WorkerRequest) (WorkerResponse, bool, error) {
	r.readinessMu.RLock()
	manifest := r.registered[request.PluginID+"@"+request.Version].Manifest
	route, exists := manifest.Route(request.Operation)
	r.readinessMu.RUnlock()
	if !exists || route.Helper != "host.call" {
		return WorkerResponse{}, false, nil
	}
	if err := ValidateWorkerRequest(request, r.HostAPI); err != nil {
		return WorkerResponse{}, true, err
	}
	target := WorkerRequest{Protocol: request.Protocol, ProjectID: request.ProjectID, PluginID: route.Call.Plugin, Version: route.Call.Version, Operation: route.Call.Operation, Payload: request.Payload, accessPrincipal: pluginAccessPrincipal(request.PluginID, request.Version), originPrincipal: request.originalPrincipal(), requireAccessPolicy: request.requireAccessPolicy}
	if err := r.checkPluginCall(request, target); err != nil {
		return WorkerResponse{}, true, err
	}
	r.workerMu.RLock()
	sourceWorker := r.workers[request.ProjectID+":"+request.PluginID+"@"+request.Version]
	targetWorker := r.workers[target.ProjectID+":"+target.PluginID+"@"+target.Version]
	r.workerMu.RUnlock()
	if sourceWorker == nil || !sourceWorker.Ready() || targetWorker == nil || !targetWorker.Ready() {
		return WorkerResponse{}, true, ErrWorkerNotReady
	}
	context, err := r.invocationContext(target, false)
	if err != nil {
		return WorkerResponse{}, true, err
	}
	target.HostContext = context
	// No storage adapter, nested call, or observation dispatch is possible on
	// this path. Recheck both grants before returning even a pure result.
	response, err := InvokeForProject(targetWorker, target, r.HostAPI, target.ProjectID)
	if err != nil {
		return WorkerResponse{}, true, err
	}
	if len(response.Observations) != 0 || response.Protocol != target.Protocol {
		return WorkerResponse{}, true, ErrInvalidWorkerRequest
	}
	if err := r.checkPluginCall(request, target); err != nil {
		return WorkerResponse{}, true, err
	}
	return response, true, nil
}
