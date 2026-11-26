package plugin

import (
	"errors"
	"fmt"
	"math"
	"time"
)

func (r *Registry) ApproveCapability(projectID, id, version, capability string) error {
	r.readinessMu.RLock()
	defer r.readinessMu.RUnlock()
	r.approvalMu.Lock()
	defer r.approvalMu.Unlock()
	r.installationMu.RLock()
	defer r.installationMu.RUnlock()
	if !isValidProjectID(projectID) || capability == "" {
		return ErrProjectInvalid
	}
	if _, ok := r.installations[projectID+":"+id+"@"+version]; !ok {
		return ErrNotInstalled
	}
	declared := false
	for _, requested := range r.registered[id+"@"+version].Manifest.Capabilities {
		if requested == capability {
			declared = true
			break
		}
	}
	if !declared {
		return ErrGrantDenied
	}
	r.approvals[projectID+":"+id+"@"+version+":"+capability] = true
	r.appendAudit(AuditEvent{Action: "approve_capability", ProjectID: projectID, PluginID: id, Version: version, Detail: capability})
	return nil
}

func (r *Registry) GrantResource(projectID, id, version, capability, resource string) error {
	r.grantMu.Lock()
	defer r.grantMu.Unlock()
	r.readinessMu.RLock()
	defer r.readinessMu.RUnlock()
	r.installationMu.Lock()
	defer r.installationMu.Unlock()
	if !isValidProjectID(projectID) {
		return ErrProjectInvalid
	}
	if resource == "" {
		return ErrGrantDenied
	}
	i, ok := r.installations[projectID+":"+id+"@"+version]
	if !ok {
		return ErrNotInstalled
	}
	allowed := false
	for _, requested := range r.registered[id+"@"+version].Manifest.Capabilities {
		if requested == capability {
			allowed = true
		}
	}
	if !allowed {
		return ErrGrantDenied
	}
	i = cloneInstallation(i)
	if i.Resources == nil {
		i.Resources = map[string]map[string]bool{}
	}
	if i.Grants == nil {
		i.Grants = map[string]bool{}
	}
	if i.Resources[capability] == nil {
		i.Resources[capability] = map[string]bool{}
	}
	i.Grants[capability] = true
	i.Resources[capability][resource] = true
	if err := r.saveInstallationLocked(projectID+":"+id+"@"+version, i); err != nil {
		return err
	}
	r.appendAudit(AuditEvent{Action: "grant_resource", ProjectID: projectID, PluginID: id, Version: version, Detail: capability + ":" + resource})
	return nil
}

func (r *Registry) AuthorizeResource(projectID, id, version, capability, resource string) error {
	r.grantMu.RLock()
	defer r.grantMu.RUnlock()
	if resource == "" {
		return ErrGrantDenied
	}
	r.installationMu.RLock()
	defer r.installationMu.RUnlock()
	if err := r.authorizeGrantLocked(projectID, id, version, capability); err != nil {
		return err
	}
	i := r.installations[projectID+":"+id+"@"+version]
	if resources := i.Resources[capability]; len(resources) > 0 && !resources[resource] {
		return ErrGrantDenied
	}
	return nil
}

func (r *Registry) Configure(projectID, id, version string, values map[string]any) error {
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
	m := r.registered[id+"@"+version].Manifest
	i = cloneInstallation(i)
	if i.Config == nil {
		i.Config = map[string]any{}
	}
	for name, value := range values {
		want, declared := m.ConfigSchema[name]
		if !declared || !configTypeMatches(want, value) {
			return fmt.Errorf("%w: %s", ErrInvalidConfig, name)
		}
		i.Config[name] = value
	}
	if err := r.saveInstallationLocked(key, i); err != nil {
		return err
	}
	r.appendAudit(AuditEvent{Action: "configure", ProjectID: projectID, PluginID: id, Version: version})
	return nil
}

func configTypeMatches(want string, value any) bool {
	switch want {
	case "string":
		_, ok := value.(string)
		return ok
	case "bool":
		_, ok := value.(bool)
		return ok
	case "int":
		switch number := value.(type) {
		case int:
			return math.Abs(float64(number)) <= 9007199254740991
		case float64:
			return math.Trunc(number) == number && math.Abs(number) <= 9007199254740991
		default:
			return false
		}
	default:
		return false
	}
}

func (r *Registry) Grant(projectID, id, version, capability string) error {
	r.grantMu.Lock()
	defer r.grantMu.Unlock()
	r.readinessMu.RLock()
	defer r.readinessMu.RUnlock()
	r.installationMu.Lock()
	defer r.installationMu.Unlock()
	if !isValidProjectID(projectID) {
		return ErrProjectInvalid
	}
	i, ok := r.installations[projectID+":"+id+"@"+version]
	if !ok {
		return ErrNotInstalled
	}
	for _, requested := range r.registered[id+"@"+version].Manifest.Capabilities {
		if requested == capability {
			i = cloneInstallation(i)
			if i.Grants == nil {
				i.Grants = map[string]bool{}
			}
			i.Grants[capability] = true
			if err := r.saveInstallationLocked(projectID+":"+id+"@"+version, i); err != nil {
				return err
			}
			r.appendAudit(AuditEvent{Action: "grant", ProjectID: projectID, PluginID: id, Version: version, Detail: capability})
			return nil
		}
	}
	return ErrGrantDenied
}

func (r *Registry) Revoke(projectID, id, version, capability string) error {
	r.grantMu.Lock()
	defer r.grantMu.Unlock()
	r.readinessMu.RLock()
	defer r.readinessMu.RUnlock()
	r.installationMu.Lock()
	defer r.installationMu.Unlock()
	if !isValidProjectID(projectID) {
		return ErrProjectInvalid
	}
	if capability == "" {
		return ErrGrantDenied
	}
	key := projectID + ":" + id + "@" + version
	i, ok := r.installations[key]
	if !ok {
		return ErrNotInstalled
	}
	if !i.Grants[capability] {
		return ErrGrantDenied
	}
	i = cloneInstallation(i)
	delete(i.Grants, capability)
	delete(i.Resources, capability)
	r.installations[key] = i
	// A failed revoke must remain denied in memory. Report failure and block
	// authorization until the revocation has been durably saved.
	transitionError := r.persistLifecycleState()
	if transitionError == nil {
		r.appendAudit(AuditEvent{Action: "revoke", ProjectID: projectID, PluginID: id, Version: version, Detail: capability})
	}
	if err := r.stopBoundWorker(key); err != nil && transitionError == nil {
		transitionError = err
	}
	if r.supervisor != nil {
		if err := r.supervisor.StopProjectPlugin(projectID, id, version); err != nil && !errors.Is(err, ErrWorkerNotRunning) && transitionError == nil {
			transitionError = err
		}
	}
	if r.webSockets != nil {
		if err := r.webSockets.CloseProject(projectID); err != nil && transitionError == nil {
			transitionError = err
		}
	}
	return transitionError
}

func (r *Registry) Authorize(projectID, id, version, capability string) error {
	r.grantMu.RLock()
	defer r.grantMu.RUnlock()
	return r.authorizeGrant(projectID, id, version, capability)
}

func (r *Registry) authorizeGrant(projectID, id, version, capability string) error {
	r.installationMu.RLock()
	defer r.installationMu.RUnlock()
	return r.authorizeGrantLocked(projectID, id, version, capability)
}

func (r *Registry) authorizeGrantLocked(projectID, id, version, capability string) error {
	if err := r.LifecycleStorageHealth(); err != nil {
		return err
	}
	if !isValidProjectID(projectID) {
		return ErrProjectInvalid
	}
	i, ok := r.installations[projectID+":"+id+"@"+version]
	if !ok {
		return ErrNotInstalled
	}
	if !i.Enabled || (i.Draining && (i.DrainUntil.IsZero() || !time.Now().Before(i.DrainUntil))) {
		return ErrDisabled
	}
	if !i.Grants[capability] {
		return ErrGrantDenied
	}
	return nil
}
