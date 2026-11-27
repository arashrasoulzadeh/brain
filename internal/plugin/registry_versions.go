package plugin

import "fmt"

func (r *Registry) Promote(projectID, id, version string) error {
	r.lifecycleMu.Lock()
	defer r.lifecycleMu.Unlock()
	r.readinessMu.RLock()
	defer r.readinessMu.RUnlock()
	r.approvalMu.RLock()
	defer r.approvalMu.RUnlock()
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
	if !i.Enabled {
		return ErrDisabled
	}
	manifest := r.registered[id+"@"+version].Manifest
	for _, capability := range manifest.Capabilities {
		if !i.Grants[capability] {
			return ErrGrantDenied
		}
	}
	if activeVersion, active := r.active[projectID+":"+id]; active && activeVersion != version {
		previousManifest := r.registered[id+"@"+activeVersion].Manifest
		for _, capability := range manifest.Capabilities {
			known := false
			for _, previousCapability := range previousManifest.Capabilities {
				if capability == previousCapability {
					known = true
					break
				}
			}
			if !known && !i.ApprovedCapabilities[capability] {
				return ErrCapabilityApproval
			}
		}
		previous := r.registered[id+"@"+activeVersion].Manifest.ConfigSchema
		for name, previousType := range previous {
			if candidateType, exists := manifest.ConfigSchema[name]; !exists || candidateType != previousType {
				return fmt.Errorf("%w: %s", ErrSchemaIncompatible, name)
			}
		}
	}
	for name, expectedType := range manifest.ConfigSchema {
		value, configured := i.Config[name]
		if !configured || !configTypeMatches(expectedType, value) {
			return fmt.Errorf("%w: %s", ErrInvalidConfig, name)
		}
	}
	previous := make(map[string]Installation)
	for installKey, candidate := range r.installations {
		if candidate.ProjectID == projectID && candidate.PluginID == id {
			previous[installKey] = candidate
			candidate.Promoted = candidate.Version == version
			r.installations[installKey] = candidate
		}
	}
	if err := r.persistLifecycleState(); err != nil {
		for installKey, candidate := range previous {
			r.installations[installKey] = candidate
		}
		return err
	}
	r.active[projectID+":"+id] = version
	r.lifecycleRevision++
	r.appendAudit(AuditEvent{Action: "promote", ProjectID: projectID, PluginID: id, Version: version})
	return nil
}

func (r *Registry) Rollback(projectID, id, version string) error {
	if err := r.Promote(projectID, id, version); err != nil {
		return err
	}
	r.appendAudit(AuditEvent{Action: "rollback", ProjectID: projectID, PluginID: id, Version: version})
	return nil
}

func (r *Registry) ActiveVersion(projectID, id string) (string, bool) {
	r.lifecycleMu.RLock()
	defer r.lifecycleMu.RUnlock()
	if !isValidProjectID(projectID) {
		return "", false
	}
	version, ok := r.active[projectID+":"+id]
	return version, ok
}

func (r *Registry) PinWork(workID, projectID, id, version string) error {
	r.lifecycleMu.Lock()
	defer r.lifecycleMu.Unlock()
	r.installationMu.RLock()
	defer r.installationMu.RUnlock()
	if workID == "" || !isValidProjectID(projectID) {
		return ErrProjectInvalid
	}
	key := projectID + ":" + id + "@" + version
	i, ok := r.installations[key]
	if !ok || !i.Enabled {
		return ErrNotInstalled
	}
	r.pins[workID] = key
	r.appendAudit(AuditEvent{Action: "pin", ProjectID: projectID, PluginID: id, Version: version, Detail: workID})
	return nil
}

func (r *Registry) PinnedVersion(workID string) (string, bool) {
	r.lifecycleMu.RLock()
	defer r.lifecycleMu.RUnlock()
	r.installationMu.RLock()
	defer r.installationMu.RUnlock()
	key, ok := r.pins[workID]
	if !ok {
		return "", false
	}
	for _, installation := range r.installations {
		if installation.ProjectID+":"+installation.PluginID+"@"+installation.Version == key {
			return installation.Version, true
		}
	}
	return "", false
}
