package plugin

import (
	"fmt"
	"net/url"
	"sort"
	"strings"
)

func (r *Registry) RegisteredManifests() []Manifest {
	r.readinessMu.RLock()
	defer r.readinessMu.RUnlock()
	manifests := make([]Manifest, 0, len(r.registered))
	for _, registered := range r.registered {
		manifests = append(manifests, registered.Manifest)
	}
	sort.Slice(manifests, func(left, right int) bool {
		if manifests[left].ID == manifests[right].ID {
			return manifests[left].Version < manifests[right].Version
		}
		return manifests[left].ID < manifests[right].ID
	})
	return manifests
}

func (r *Registry) InstallationStatuses(projectID string) []InstallationStatus {
	r.installationMu.RLock()
	keys := make([]string, 0, len(r.installations))
	for key, installation := range r.installations {
		if projectID == "" || installation.ProjectID == projectID {
			keys = append(keys, key)
		}
	}
	installations := make([]Installation, 0, len(keys))
	for _, key := range keys {
		installations = append(installations, r.installations[key])
	}
	r.installationMu.RUnlock()
	sort.Slice(installations, func(left, right int) bool {
		if installations[left].PluginID == installations[right].PluginID {
			return installations[left].Version < installations[right].Version
		}
		return installations[left].PluginID < installations[right].PluginID
	})
	statuses := make([]InstallationStatus, 0, len(installations))
	for _, installation := range installations {
		status, err := r.Status(installation.ProjectID, installation.PluginID, installation.Version)
		if err == nil {
			statuses = append(statuses, status)
		}
	}
	return statuses
}

func (r *Registry) Status(projectID, pluginID, version string) (InstallationStatus, error) {
	if !isValidProjectID(projectID) {
		return InstallationStatus{}, ErrProjectInvalid
	}
	r.readinessMu.RLock()
	defer r.readinessMu.RUnlock()
	r.installationMu.RLock()
	defer r.installationMu.RUnlock()
	installation, ok := r.installations[projectID+":"+pluginID+"@"+version]
	if !ok {
		return InstallationStatus{}, ErrNotInstalled
	}
	ready := r.readiness[pluginID+"@"+version]
	r.workerMu.RLock()
	workerReady := false
	if worker := r.workers[projectID+":"+pluginID+"@"+version]; worker != nil {
		workerReady = worker.Ready()
	}
	r.workerMu.RUnlock()
	grants := make([]string, 0, len(installation.Grants))
	for capability, granted := range installation.Grants {
		if granted {
			grants = append(grants, capability)
		}
	}
	sort.Strings(grants)
	return InstallationStatus{ProjectID: projectID, PluginID: pluginID, Version: version, Enabled: installation.Enabled, Ready: ready, WorkerReady: workerReady, Grants: grants, Draining: installation.Draining, DrainUntil: installation.DrainUntil}, nil
}

func (r *Registry) ApproveArtifact(id, version, endpoint string) error {
	r.readinessMu.RLock()
	defer r.readinessMu.RUnlock()
	if _, ok := r.registered[id+"@"+version]; !ok {
		return ErrNotRegistered
	}
	u, err := url.Parse(endpoint)
	if err != nil || u.Scheme != "https" || u.Host == "" {
		return ErrInvalidManifest
	}
	r.artifactMu.Lock()
	defer r.artifactMu.Unlock()
	r.artifacts[id+"@"+version] = endpoint
	r.appendAudit(AuditEvent{Action: "approve_artifact", PluginID: id, Version: version, Detail: endpoint})
	return nil
}

func (r *Registry) ApprovedArtifact(id, version string) (string, bool) {
	r.artifactMu.RLock()
	defer r.artifactMu.RUnlock()
	endpoint, ok := r.artifacts[id+"@"+version]
	return endpoint, ok
}

func (r *Registry) Register(m Manifest) error {
	validationManifest := m
	if len(validationManifest.SourceFiles) == 0 && validationManifest.Entry == "" && len(validationManifest.Imports) == 0 && len(validationManifest.PluginImports) == 0 {
		validationManifest.Entry = "index.ts"
		validationManifest.SourceFiles = []string{"index.ts"}
	}
	if err := ValidateManifestContract(validationManifest); err != nil {
		return err
	}
	r.readinessMu.Lock()
	defer r.readinessMu.Unlock()
	if m.ID == "" || m.Version == "" || m.HostAPI != r.HostAPI {
		if m.ID == "" || m.Version == "" {
			return ErrInvalidManifest
		}
		return ErrIncompatible
	}
	for _, extensionPoint := range m.ExtensionPoints {
		if strings.TrimSpace(extensionPoint) == "" {
			return ErrInvalidManifest
		}
	}
	for _, dependency := range m.Dependencies {
		if _, ok := r.registered[dependency]; !ok {
			return ErrDependencyMissing
		}
	}
	key := m.ID + "@" + m.Version
	if _, ok := r.registered[key]; ok {
		return ErrDuplicate
	}
	r.registered[key] = Registered{Manifest: m}
	r.readiness[key] = false
	r.appendAudit(AuditEvent{Action: "register", PluginID: m.ID, Version: m.Version})
	return nil
}

func (r *Registry) TransitiveDependents(id, version string) []string {
	r.readinessMu.RLock()
	defer r.readinessMu.RUnlock()
	target := id + "@" + version
	seen := map[string]bool{target: true}
	dependents := make([]string, 0)
	for {
		changed := false
		for key, registered := range r.registered {
			if seen[key] {
				continue
			}
			for _, dependency := range registered.Manifest.Dependencies {
				if seen[dependency] {
					seen[key] = true
					dependents = append(dependents, key)
					changed = true
					break
				}
			}
		}
		if !changed {
			break
		}
	}
	sort.Strings(dependents)
	return dependents
}

func (r *Registry) SetReadiness(id, version string, ready bool) error {
	r.readinessMu.Lock()
	defer r.readinessMu.Unlock()
	key := id + "@" + version
	if _, ok := r.registered[key]; !ok {
		return ErrNotRegistered
	}
	r.readiness[key] = ready
	r.persistLifecycleState()
	r.appendAudit(AuditEvent{Action: "readiness", PluginID: id, Version: version, Detail: fmt.Sprintf("ready=%t", ready)})
	return nil
}

func (r *Registry) Install(projectID, id, version string) error {
	r.readinessMu.RLock()
	defer r.readinessMu.RUnlock()
	r.installationMu.Lock()
	defer r.installationMu.Unlock()
	if !isValidProjectID(projectID) {
		return ErrProjectInvalid
	}
	key := id + "@" + version
	if _, ok := r.registered[key]; !ok {
		return ErrNotRegistered
	}
	installKey := projectID + ":" + key
	if _, exists := r.installations[installKey]; exists {
		return ErrDuplicate
	}
	r.installations[installKey] = Installation{ProjectID: projectID, PluginID: id, Version: version, Grants: map[string]bool{}, Resources: map[string]map[string]bool{}, Config: map[string]any{}}
	r.persistLifecycleState()
	r.appendAudit(AuditEvent{Action: "install", ProjectID: projectID, PluginID: id, Version: version})
	return nil
}

func (r *Registry) Unregister(id, version string) error {
	r.lifecycleMu.Lock()
	defer r.lifecycleMu.Unlock()
	r.readinessMu.Lock()
	defer r.readinessMu.Unlock()
	r.artifactMu.Lock()
	defer r.artifactMu.Unlock()
	r.approvalMu.Lock()
	defer r.approvalMu.Unlock()
	r.installationMu.RLock()
	defer r.installationMu.RUnlock()
	key := id + "@" + version
	if _, ok := r.registered[key]; !ok {
		return ErrNotRegistered
	}
	for _, installation := range r.installations {
		if installation.PluginID == id && installation.Version == version {
			return ErrVersionReferenced
		}
	}
	for activeKey, activeVersion := range r.active {
		if strings.HasSuffix(activeKey, ":"+id) && activeVersion == version {
			return ErrVersionReferenced
		}
	}
	for _, pinned := range r.pins {
		if pinned == "" {
			continue
		}
		if strings.HasSuffix(pinned, ":"+key) {
			return ErrVersionReferenced
		}
	}
	for _, registered := range r.registered {
		for _, dependency := range registered.Manifest.Dependencies {
			if dependency == key {
				return ErrVersionReferenced
			}
		}
	}
	delete(r.registered, key)
	delete(r.readiness, key)
	delete(r.artifacts, key)
	for approvalKey := range r.approvals {
		if strings.Contains(approvalKey, ":"+key+":") || strings.HasPrefix(approvalKey, key+":") {
			delete(r.approvals, approvalKey)
		}
	}
	r.appendAudit(AuditEvent{Action: "unregister", PluginID: id, Version: version})
	return nil
}
