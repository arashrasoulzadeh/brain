package plugin

import (
	"errors"
	"strings"
	"time"
)

type DisablePreview struct {
	ProjectID  string
	PluginID   string
	Version    string
	Revision   uint64
	Dependents []string
}

const DefaultDrainWindow = 5 * time.Minute

func (r *Registry) PreviewDisable(projectID, id, version string) (DisablePreview, error) {
	r.lifecycleMu.Lock()
	defer r.lifecycleMu.Unlock()
	r.readinessMu.RLock()
	defer r.readinessMu.RUnlock()
	r.installationMu.RLock()
	defer r.installationMu.RUnlock()
	if !isValidProjectID(projectID) {
		return DisablePreview{}, ErrProjectInvalid
	}
	if _, ok := r.installations[projectID+":"+id+"@"+version]; !ok {
		return DisablePreview{}, ErrNotInstalled
	}
	return DisablePreview{ProjectID: projectID, PluginID: id, Version: version, Revision: r.lifecycleRevision, Dependents: r.transitiveDependentsLocked(id, version)}, nil
}

func (r *Registry) ApplyDisablePreview(preview DisablePreview) error {
	r.lifecycleMu.Lock()
	defer r.lifecycleMu.Unlock()
	r.installationMu.RLock()
	defer r.installationMu.RUnlock()
	if preview.ProjectID == "" {
		return ErrProjectInvalid
	}
	if preview.Revision != r.lifecycleRevision {
		return ErrStalePreview
	}
	if _, ok := r.installations[preview.ProjectID+":"+preview.PluginID+"@"+preview.Version]; !ok {
		return ErrNotInstalled
	}
	var transitionError error
	if r.supervisor != nil {
		if err := r.supervisor.StopProjectPlugin(preview.ProjectID, preview.PluginID, preview.Version); err != nil && !errors.Is(err, ErrWorkerNotRunning) {
			transitionError = err
		}
	}
	if r.webSockets != nil {
		if err := r.webSockets.CloseProject(preview.ProjectID); err != nil && transitionError == nil {
			transitionError = err
		}
	}
	r.lifecycleRevision++
	return transitionError
}

func (r *Registry) Enable(projectID, id, version string) error {
	r.grantMu.Lock()
	defer r.grantMu.Unlock()
	r.readinessMu.RLock()
	defer r.readinessMu.RUnlock()
	r.installationMu.Lock()
	defer r.installationMu.Unlock()
	if !isValidProjectID(projectID) {
		return ErrProjectInvalid
	}
	key := projectID + ":" + id + "@" + version
	i, ok := r.installations[key]
	if !ok {
		return ErrNotInstalled
	}
	if !r.readiness[id+"@"+version] {
		return ErrNotReady
	}
	for _, capability := range r.registered[id+"@"+version].Manifest.Capabilities {
		if !i.Grants[capability] {
			return ErrGrantDenied
		}
	}
	i.Enabled = true
	r.installations[key] = i
	r.persistLifecycleState()
	if r.workerFactory != nil {
		worker, err := r.workerFactory(projectID, id, version)
		if err != nil {
			i.Enabled = false
			r.installations[key] = i
			return err
		}
		if err := worker.Start(); err != nil {
			i.Enabled = false
			r.installations[key] = i
			return err
		}
		r.workerMu.Lock()
		r.workers[key] = worker
		r.workerMu.Unlock()
	}
	r.appendAudit(AuditEvent{Action: "enable", ProjectID: projectID, PluginID: id, Version: version})
	return nil
}

func (r *Registry) Drain(projectID, id, version string) error {
	return r.DrainUntil(projectID, id, version, time.Now().Add(DefaultDrainWindow))
}

func (r *Registry) DrainUntil(projectID, id, version string, deadline time.Time) error {
	r.lifecycleMu.Lock()
	defer r.lifecycleMu.Unlock()
	r.installationMu.Lock()
	defer r.installationMu.Unlock()
	if !isValidProjectID(projectID) {
		return ErrProjectInvalid
	}
	key := projectID + ":" + id + "@" + version
	i, ok := r.installations[key]
	if !ok {
		return ErrNotInstalled
	}
	i.Draining = true
	i.DrainUntil = deadline
	r.installations[key] = i
	r.appendAudit(AuditEvent{Action: "drain", ProjectID: projectID, PluginID: id, Version: version})
	return nil
}

func (r *Registry) StopImmediately(projectID, id, version string) error {
	r.lifecycleMu.Lock()
	defer r.lifecycleMu.Unlock()
	r.installationMu.Lock()
	defer r.installationMu.Unlock()
	if !isValidProjectID(projectID) {
		return ErrProjectInvalid
	}
	key := projectID + ":" + id + "@" + version
	i, ok := r.installations[key]
	if !ok {
		return ErrNotInstalled
	}
	i.Draining = true
	i.DrainUntil = time.Unix(0, 0)
	r.installations[key] = i
	r.lifecycleRevision++
	r.appendAudit(AuditEvent{Action: "stop_immediate", ProjectID: projectID, PluginID: id, Version: version})
	var stopError error
	if r.supervisor != nil {
		if err := r.supervisor.StopProjectPlugin(projectID, id, version); err != nil && !errors.Is(err, ErrWorkerNotRunning) {
			stopError = err
		}
	}
	if r.webSockets != nil {
		if err := r.webSockets.CloseProject(projectID); err != nil && stopError == nil {
			stopError = err
		}
	}
	return stopError
}

func (r *Registry) Recover(projectID, id, version string) error {
	r.lifecycleMu.Lock()
	defer r.lifecycleMu.Unlock()
	r.installationMu.Lock()
	defer r.installationMu.Unlock()
	if !isValidProjectID(projectID) {
		return ErrProjectInvalid
	}
	key := projectID + ":" + id + "@" + version
	i, ok := r.installations[key]
	if !ok {
		return ErrNotInstalled
	}
	var recoveryError error
	if r.supervisor != nil {
		if err := r.supervisor.Recover(key); err != nil && !errors.Is(err, ErrWorkerNotRunning) {
			recoveryError = err
		}
	}
	if r.webSockets != nil {
		if err := r.webSockets.CloseProject(projectID); err != nil && recoveryError == nil {
			recoveryError = err
		}
	}
	if recoveryError != nil {
		return recoveryError
	}
	i.Draining = false
	i.DrainUntil = time.Time{}
	r.installations[key] = i
	r.lifecycleRevision++
	r.appendAudit(AuditEvent{Action: "recover", ProjectID: projectID, PluginID: id, Version: version})
	return nil
}

func (r *Registry) CanStart(projectID, id, version string, now time.Time) error {
	r.installationMu.RLock()
	defer r.installationMu.RUnlock()
	if !isValidProjectID(projectID) {
		return ErrProjectInvalid
	}
	i, ok := r.installations[projectID+":"+id+"@"+version]
	if !ok {
		return ErrNotInstalled
	}
	if i.Draining && (i.DrainUntil.IsZero() || !now.Before(i.DrainUntil)) {
		return ErrDisabled
	}
	return nil
}

func (r *Registry) CanStartNew(projectID, id, version string) error {
	r.installationMu.RLock()
	defer r.installationMu.RUnlock()
	if !isValidProjectID(projectID) {
		return ErrProjectInvalid
	}
	i, ok := r.installations[projectID+":"+id+"@"+version]
	if !ok {
		return ErrNotInstalled
	}
	if i.Draining {
		return ErrDisabled
	}
	return nil
}

func (r *Registry) CanReconnect(projectID, id, version string, now time.Time) error {
	r.installationMu.RLock()
	defer r.installationMu.RUnlock()
	if !isValidProjectID(projectID) {
		return ErrProjectInvalid
	}
	i, ok := r.installations[projectID+":"+id+"@"+version]
	if !ok {
		return ErrNotInstalled
	}
	if i.Draining && !i.DrainUntil.IsZero() && !now.Before(i.DrainUntil) {
		return ErrDisabled
	}
	return nil
}

func (r *Registry) Remove(projectID, id, version string) error {
	r.lifecycleMu.Lock()
	defer r.lifecycleMu.Unlock()
	r.installationMu.Lock()
	defer r.installationMu.Unlock()
	if !isValidProjectID(projectID) {
		return ErrProjectInvalid
	}
	key := projectID + ":" + id + "@" + version
	if _, ok := r.installations[key]; !ok {
		return ErrNotInstalled
	}
	for _, pinned := range r.pins {
		if pinned == key {
			return ErrVersionReferenced
		}
	}
	if r.active[projectID+":"+id] == version {
		return ErrVersionReferenced
	}
	var removalError error
	if r.supervisor != nil {
		if err := r.supervisor.StopProjectPlugin(projectID, id, version); err != nil && !errors.Is(err, ErrWorkerNotRunning) {
			removalError = err
		}
	}
	delete(r.installations, key)
	r.lifecycleRevision++
	r.appendAudit(AuditEvent{Action: "remove", ProjectID: projectID, PluginID: id, Version: version})
	if r.webSockets != nil {
		if err := r.webSockets.CloseProject(projectID); err != nil && removalError == nil {
			removalError = err
		}
	}
	return removalError
}

func (r *Registry) UninstallProject(projectID string) error {
	r.lifecycleMu.Lock()
	defer r.lifecycleMu.Unlock()
	r.installationMu.Lock()
	defer r.installationMu.Unlock()
	if !isValidProjectID(projectID) {
		return ErrProjectInvalid
	}
	var uninstallError error
	for _, pinnedVersion := range r.pins {
		if strings.HasPrefix(pinnedVersion, projectID+":") {
			return ErrActiveWork
		}
	}
	for key := range r.active {
		if strings.HasPrefix(key, projectID+":") {
			return ErrActiveWork
		}
	}
	for key, installation := range r.installations {
		if installation.ProjectID == projectID {
			if r.supervisor != nil {
				if err := r.supervisor.StopProjectPlugin(projectID, installation.PluginID, installation.Version); err != nil && !errors.Is(err, ErrWorkerNotRunning) && uninstallError == nil {
					uninstallError = err
				}
			}
			for capability := range installation.Grants {
				r.appendAudit(AuditEvent{Action: "revoke_project_grant", ProjectID: projectID, PluginID: installation.PluginID, Version: installation.Version, Detail: capability})
			}
			delete(r.installations, key)
		}
	}
	r.lifecycleRevision++
	r.appendAudit(AuditEvent{Action: "uninstall_project", ProjectID: projectID})
	if r.webSockets != nil {
		if err := r.webSockets.CloseProject(projectID); err != nil && uninstallError == nil {
			uninstallError = err
		}
	}
	return uninstallError
}
