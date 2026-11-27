package plugin

import (
	"fmt"
	"slices"
	"strings"
)

// EnableWithDependencies applies the exact dependency changes confirmed by an
// operator. Only preinstalled, ready, already-granted versions may be enabled.
// Recompute the transitive plan under lifecycle locks; stale consent fails closed.
func (r *Registry) EnableWithDependencies(projectID, id, version string, approved []string) error {
	r.lifecycleMu.Lock()
	defer r.lifecycleMu.Unlock()
	r.grantMu.Lock()
	defer r.grantMu.Unlock()
	r.readinessMu.RLock()
	defer r.readinessMu.RUnlock()
	r.installationMu.Lock()
	defer r.installationMu.Unlock()

	if !isValidProjectID(projectID) {
		return ErrProjectInvalid
	}
	if r.projectResolver != nil {
		p, err := r.projectResolver(projectID)
		if err != nil || p.ID != projectID || !p.Active {
			return ErrProjectInvalid
		}
	}
	if err := r.LifecycleStorageHealth(); err != nil {
		return err
	}
	root := id + "@" + version
	visiting, visited := map[string]bool{}, map[string]bool{}
	selected := map[string]string{}
	plan := []string{}
	var visit func(string) error
	visit = func(key string) error {
		if visiting[key] {
			return fmt.Errorf("%w: dependency cycle at %s", ErrDependencyMissing, key)
		}
		if visited[key] {
			return nil
		}
		pluginID, pluginVersion, valid := strings.Cut(key, "@")
		reg, registered := r.registered[key]
		installation, installed := r.installations[projectID+":"+key]
		if !valid || !registered || !installed || !r.readiness[key] || installation.Draining {
			return fmt.Errorf("%w: %s must be installed, ready and not draining", ErrDependencyMissing, key)
		}
		if other, ok := selected[pluginID]; ok && other != pluginVersion {
			return fmt.Errorf("%w: conflicting exact versions of %s", ErrVersionReferenced, pluginID)
		}
		selected[pluginID] = pluginVersion
		active := r.active[projectID+":"+pluginID]
		if key != root && active != "" && active != pluginVersion {
			return fmt.Errorf("%w: another version of %s is promoted", ErrVersionReferenced, pluginID)
		}
		for _, other := range r.installations {
			if other.ProjectID == projectID && other.PluginID == pluginID && other.Version != pluginVersion && other.Enabled && (key != root || active == "") {
				return fmt.Errorf("%w: another version of %s is enabled", ErrVersionReferenced, pluginID)
			}
		}
		for _, capability := range reg.Manifest.Capabilities {
			if !installation.Grants[capability] {
				return fmt.Errorf("%w: %s needs %s", ErrGrantDenied, key, capability)
			}
		}
		visiting[key] = true
		for _, dependency := range reg.Manifest.Dependencies {
			if err := visit(dependency); err != nil {
				return err
			}
		}
		delete(visiting, key)
		visited[key] = true
		if !installation.Enabled {
			plan = append(plan, key)
		}
		return nil
	}
	if err := visit(root); err != nil {
		return err
	}
	expected := []string{}
	for _, key := range plan {
		if key != root {
			expected = append(expected, key)
		}
	}
	if !slices.Equal(expected, approved) {
		return fmt.Errorf("%w: dependency activation plan changed; review and confirm again", ErrDependencyMissing)
	}

	r.workerMu.Lock()
	defer r.workerMu.Unlock()
	started := map[string]ContractWorker{}
	rollbackWorkers := func() {
		for _, worker := range started {
			_ = worker.Stop()
		}
	}
	for _, key := range plan {
		installKey := projectID + ":" + key
		if current := r.workers[installKey]; current != nil && current.Ready() {
			continue
		}
		if r.workerFactory == nil {
			rollbackWorkers()
			return ErrWorkerUnavailable
		}
		pluginID, pluginVersion, _ := strings.Cut(key, "@")
		worker, err := r.workerFactory(projectID, pluginID, pluginVersion)
		if err != nil || worker == nil {
			rollbackWorkers()
			if err != nil {
				return err
			}
			return ErrWorkerUnavailable
		}
		started[installKey] = worker
		if err := worker.Start(); err != nil {
			rollbackWorkers()
			return err
		}
		if !worker.Ready() {
			rollbackWorkers()
			return ErrWorkerNotReady
		}
	}
	previous := map[string]Installation{}
	for _, key := range plan {
		installKey := projectID + ":" + key
		i := r.installations[installKey]
		previous[installKey] = i
		i.Enabled = true
		r.installations[installKey] = i
	}
	if len(plan) > 0 {
		if err := r.persistLifecycleState(); err != nil {
			for key, i := range previous {
				r.installations[key] = i
			}
			rollbackWorkers()
			return err
		}
		r.lifecycleRevision++
	}
	for key, worker := range started {
		r.workers[key] = worker
	}
	for _, key := range plan {
		pluginID, pluginVersion, _ := strings.Cut(key, "@")
		r.appendAudit(AuditEvent{Action: "enable", ProjectID: projectID, PluginID: pluginID, Version: pluginVersion, Detail: "confirmed dependency enable for " + root})
	}
	return nil
}
