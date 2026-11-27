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
	Auth            InvocationAuth `json:"auth"`
}

// InvocationActor is a safe identity projection, not a credential or account record.
type InvocationActor struct {
	Kind          string `json:"kind"`
	ID            string `json:"id"`
	Authenticated bool   `json:"authenticated"`
	Level         string `json:"level"`
}

// InvocationPermission describes only the operation/resource actually admitted
// by the host. It must not be interpreted as a general grant or a reusable token.
type InvocationPermission struct {
	Capability string `json:"capability"`
	Operation  string `json:"operation"`
	Resource   string `json:"resource"`
}

type InvocationAuth struct {
	Actor       InvocationActor        `json:"actor"`
	Caller      InvocationActor        `json:"caller"`
	Permissions []InvocationPermission `json:"permissions"`
}

func invocationActor(principal string) InvocationActor {
	if principal == "" || principal == "host" {
		return InvocationActor{Kind: "host", Level: "system"}
	}
	if principal == "public" {
		return InvocationActor{Kind: "public", Level: "anonymous"}
	}
	kind, id, ok := strings.Cut(principal, ":")
	if !ok || !validAccessPrincipal(principal) {
		return InvocationActor{Kind: "public", Level: "anonymous"}
	}
	if kind == "user" || kind == "operator" {
		return InvocationActor{Kind: kind, ID: id, Authenticated: true, Level: "authenticated"}
	}
	return InvocationActor{Kind: kind, ID: id, Level: "system"}
}

func (request WorkerRequest) originalPrincipal() string {
	if request.originPrincipal != "" {
		return request.originPrincipal
	}
	if request.accessPrincipal != "" {
		return request.accessPrincipal
	}
	return "host"
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
	route, _ := manifest.Route(request.Operation)
	resource, scopeErr := invocationResource(request, route)
	if observer {
		for _, subscription := range manifest.Observers {
			if subscription.operation() == request.Operation {
				route.Capability = subscription.Capability
				resource = subscription.Plugin + "." + subscription.Event
				scopeErr = nil
				break
			}
		}
	}
	permissions := []InvocationPermission{}
	if scopeErr == nil && route.Capability != "" {
		permissions = append(permissions, InvocationPermission{Capability: route.Capability, Operation: request.Operation, Resource: resource})
	}
	context := &InvocationContext{InvocationID: id, ProjectVerified: r.projectResolver != nil, Config: map[string]any{}, Available: []string{}, Auth: InvocationAuth{
		Actor: invocationActor(request.originalPrincipal()), Caller: invocationActor(request.accessPrincipal), Permissions: permissions,
	}}
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
