package plugin

import (
	"sort"
	"strings"
)

// InvocationContext is assembled by the host, never decoded from public RPC.
type InvocationContext struct {
	InvocationID    string         `json:"invocationId"`
	ProjectVerified bool           `json:"projectVerified"`
	Config          map[string]any `json:"config"`
	Available       []string       `json:"available"`
}

func validOptionalDependencies(manifest Manifest) bool {
	seen := map[string]bool{}
	for _, dependency := range manifest.Dependencies {
		seen[dependency] = true
	}
	if len(manifest.OptionalDependencies) > 32 {
		return false
	}
	for _, dependency := range manifest.OptionalDependencies {
		id, version, ok := strings.Cut(dependency, "@")
		if !ok || !eventIdentifier.MatchString(id) || version == "" || strings.ContainsAny(version, "@:/ ") || id == manifest.ID || seen[dependency] {
			return false
		}
		seen[dependency] = true
	}
	return true
}

func (r *Registry) invocationContext(request WorkerRequest, observer bool) (*InvocationContext, error) {
	r.lifecycleMu.RLock()
	defer r.lifecycleMu.RUnlock()
	r.readinessMu.RLock()
	defer r.readinessMu.RUnlock()
	r.installationMu.RLock()
	defer r.installationMu.RUnlock()
	if err := r.authorizeInvocationLocked(request, observer); err != nil {
		return nil, err
	}
	key := request.PluginID + "@" + request.Version
	manifest := r.registered[key].Manifest
	installation := r.installations[request.ProjectID+":"+key]
	id, err := newModelRevision()
	if err != nil {
		return nil, err
	}
	context := &InvocationContext{InvocationID: id, ProjectVerified: r.projectResolver != nil, Config: map[string]any{}, Available: []string{}}
	for name, value := range installation.Config {
		if configTypeMatches(manifest.ConfigSchema[name], value) {
			context.Config[name] = value
		}
	}
	dependencies := append(append([]string{}, manifest.Dependencies...), manifest.OptionalDependencies...)
	for _, dependency := range dependencies {
		pluginID, version, ok := strings.Cut(dependency, "@")
		if !ok {
			continue
		}
		candidate, installed := r.installations[request.ProjectID+":"+dependency]
		registered, exists := r.registered[dependency]
		if !installed || !exists || !candidate.Enabled || candidate.Draining || !r.readiness[dependency] {
			continue
		}
		if active, promoted := r.active[request.ProjectID+":"+pluginID]; promoted && active != version {
			continue
		}
		granted := true
		for _, capability := range registered.Manifest.Capabilities {
			if !candidate.Grants[capability] {
				granted = false
				break
			}
		}
		if granted {
			context.Available = append(context.Available, dependency)
		}
	}
	sort.Strings(context.Available)
	return context, nil
}
