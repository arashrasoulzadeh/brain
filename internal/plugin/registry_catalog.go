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
		if scheduled, ok := worker.(interface{ ExecutionAvailable() bool }); ok {
			workerReady = scheduled.ExecutionAvailable()
		}
	}
	r.workerMu.RUnlock()
	grants := make([]string, 0, len(installation.Grants))
	for capability, granted := range installation.Grants {
		if granted {
			grants = append(grants, capability)
		}
	}
	sort.Strings(grants)
	resourceGrants := make(map[string][]string)
	for _, capability := range grants {
		for resource, granted := range installation.Resources[capability] {
			if granted {
				resourceGrants[capability] = append(resourceGrants[capability], resource)
			}
		}
		sort.Strings(resourceGrants[capability])
	}
	var approved []string
	for capability, granted := range installation.ApprovedCapabilities {
		if granted {
			approved = append(approved, capability)
		}
	}
	sort.Strings(approved)
	var configured []string
	for name := range installation.Config {
		configured = append(configured, name)
	}
	sort.Strings(configured)
	var publicConfig map[string]any
	for _, name := range r.registered[pluginID+"@"+version].Manifest.PublicConfig {
		if value, ok := installation.Config[name]; ok {
			if publicConfig == nil {
				publicConfig = map[string]any{}
			}
			publicConfig[name] = value
		}
	}
	maxUploadBytes := 0
	if pluginID == "files" {
		maxUploadBytes = defaultFileUploadLimit
		if limit, ok := configuredFileUploadLimit(installation.Config); ok {
			maxUploadBytes = limit
		}
	}
	return InstallationStatus{ProjectID: projectID, PluginID: pluginID, Version: version, Enabled: installation.Enabled, Ready: ready, WorkerReady: workerReady, Grants: grants, ResourceGrants: resourceGrants, Promoted: installation.Promoted, ApprovedCapabilities: approved, Draining: installation.Draining, DrainUntil: installation.DrainUntil, ConfiguredFields: configured, PublicConfig: publicConfig, MaxUploadBytes: maxUploadBytes}, nil
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
	if !validationManifest.Declarative && len(validationManifest.SourceFiles) == 0 && validationManifest.Entry == "" && len(validationManifest.Imports) == 0 && len(validationManifest.PluginImports) == 0 {
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
	if m.Mount != "" {
		for _, existing := range r.registered {
			if existing.Manifest.ID != m.ID && existing.Manifest.Mount == m.Mount {
				return ErrInvalidManifest
			}
		}
	}
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
	return r.transitiveDependentsLocked(id, version)
}

func (r *Registry) transitiveDependentsLocked(id, version string) []string {
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
	r.installationMu.RLock()
	defer r.installationMu.RUnlock()
	key := id + "@" + version
	if _, ok := r.registered[key]; !ok {
		return ErrNotRegistered
	}
	previous := r.readiness[key]
	r.readiness[key] = ready
	if err := r.persistLifecycleState(); err != nil {
		// Never restore readiness on a failed attempt to make a worker unready.
		if ready {
			r.readiness[key] = previous
		}
		return err
	}
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
	installation := Installation{ProjectID: projectID, PluginID: id, Version: version, Grants: map[string]bool{}, Resources: map[string]map[string]bool{}, Config: map[string]any{}}
	if err := r.saveInstallationLocked(installKey, installation); err != nil {
		return err
	}
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

// BindProjectResolver installs a trusted active-project lookup for invocations.
func (r *Registry) BindProjectResolver(resolve func(string) (ProjectSnapshot, error)) {
	r.readinessMu.Lock()
	defer r.readinessMu.Unlock()
	r.projectResolver = resolve
}
