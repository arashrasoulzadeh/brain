package plugin

import (
	"fmt"
	"strings"
)

// InstallAndActivate is the confirmed, all-or-nothing UI install. The older
// Install/Grant/Enable calls remain available to explicit lifecycle clients.
// Never grant capabilities that were not present in the confirmation.
func (r *Registry) InstallAndActivate(projectID, id, version string, approved []string) error {
	r.lifecycleMu.Lock()
	defer r.lifecycleMu.Unlock()
	r.grantMu.Lock()
	defer r.grantMu.Unlock()
	r.readinessMu.Lock()
	defer r.readinessMu.Unlock()
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
	key := id + "@" + version
	registered, ok := r.registered[key]
	if !ok {
		return ErrNotRegistered
	}
	installKey := projectID + ":" + key
	if _, exists := r.installations[installKey]; exists {
		return ErrDuplicate
	}
	grants := make(map[string]bool, len(approved))
	for _, capability := range approved {
		if capability == "" || grants[capability] {
			return ErrGrantDenied
		}
		grants[capability] = true
	}
	if len(grants) != len(registered.Manifest.Capabilities) {
		return ErrGrantDenied
	}
	for _, capability := range registered.Manifest.Capabilities {
		if !grants[capability] {
			return ErrGrantDenied
		}
	}
	for _, dependency := range registered.Manifest.Dependencies {
		dep, exists := r.installations[projectID+":"+dependency]
		if !exists || !dep.Enabled || dep.Draining || !r.readiness[dependency] {
			return fmt.Errorf("%w: %s must be enabled in this project", ErrDependencyMissing, dependency)
		}
	}
	// Installing another version must not silently switch live HTTP routing or
	// bypass the explicit version-promotion/schema-approval flow.
	for _, existing := range r.installations {
		if existing.ProjectID == projectID && existing.PluginID == id && existing.Enabled {
			return fmt.Errorf("%w: disable the current version or use the staged upgrade flow", ErrVersionReferenced)
		}
	}
	if err := r.LifecycleStorageHealth(); err != nil {
		return err
	}
	r.workerMu.Lock()
	defer r.workerMu.Unlock()
	if r.workerFactory == nil {
		return ErrWorkerUnavailable
	}
	worker, err := r.workerFactory(projectID, id, version)
	if err != nil {
		return err
	}
	if worker == nil {
		return ErrWorkerUnavailable
	}
	if err := worker.Start(); err != nil {
		_ = worker.Stop()
		return err
	}
	if !worker.Ready() {
		_ = worker.Stop()
		return ErrWorkerNotReady
	}
	previousReadiness := r.readiness[key]
	r.readiness[key] = true
	installation := Installation{ProjectID: projectID, PluginID: id, Version: version, Enabled: true, Grants: grants, Resources: map[string]map[string]bool{}, Config: map[string]any{}}
	if err := r.saveInstallationLocked(installKey, installation); err != nil {
		r.readiness[key] = previousReadiness
		_ = worker.Stop()
		return err
	}
	r.workers[installKey] = worker
	r.lifecycleRevision++
	r.appendAudit(AuditEvent{Action: "install", ProjectID: projectID, PluginID: id, Version: version})
	for _, capability := range registered.Manifest.Capabilities {
		r.appendAudit(AuditEvent{Action: "grant", ProjectID: projectID, PluginID: id, Version: version, Detail: capability})
	}
	r.appendAudit(AuditEvent{Action: "enable", ProjectID: projectID, PluginID: id, Version: version, Detail: "confirmed install: " + strings.Join(registered.Manifest.Capabilities, ", ")})
	return nil
}
