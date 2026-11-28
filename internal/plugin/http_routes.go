package plugin

import (
	"context"
	"maps"
	"sort"
	"strings"
)

// HTTPRouteMatch is selected by the host from an installed manifest, never from
// client-provided operation, plugin ID, version or capabilities.
type HTTPRouteMatch struct {
	PluginID           string            `json:"plugin_id"`
	Version            string            `json:"version"`
	Method             string            `json:"method"`
	Path               string            `json:"path"`
	Function           string            `json:"function"`
	Capability         string            `json:"capability"`
	Operation          string            `json:"operation"`
	Auth               string            `json:"auth"`
	Params             map[string]string `json:"-"`
	ResourceField      string            `json:"resource_field,omitempty"`
	WebSocketAdmission string            `json:"websocket_admission,omitempty"`

	projectID    string
	relativePath string
}

func mountedHTTPPath(mount, relative string) string {
	if relative == "/" {
		return mount
	}
	return mount + relative
}

func matchHTTPPath(pattern, actual string) (map[string]string, bool) {
	if pattern == "/" && actual == "/" {
		return map[string]string{}, true
	}
	left, right := strings.Split(strings.TrimPrefix(pattern, "/"), "/"), strings.Split(strings.TrimPrefix(actual, "/"), "/")
	if len(left) != len(right) {
		return nil, false
	}
	params := make(map[string]string)
	for i, part := range left {
		if strings.HasPrefix(part, "{") && strings.HasSuffix(part, "}") {
			if right[i] == "" || right[i] == "." || right[i] == ".." {
				return nil, false
			}
			params[part[1:len(part)-1]] = right[i]
		} else if part != right[i] {
			return nil, false
		}
	}
	return params, true
}

// MatchHTTPRoute takes one consistent registry snapshot. The matching version
// must still be enabled, ready and not draining when Invoke is admitted and at
// commit time. A promoted version excludes every other installed version.
func (r *Registry) MatchHTTPRoute(projectID, method, relativePath string) (HTTPRouteMatch, bool) {
	if !isValidProjectID(projectID) {
		return HTTPRouteMatch{}, false
	}
	r.lifecycleMu.RLock()
	defer r.lifecycleMu.RUnlock()
	r.readinessMu.RLock()
	defer r.readinessMu.RUnlock()
	r.installationMu.RLock()
	defer r.installationMu.RUnlock()
	return r.matchHTTPRouteLocked(projectID, method, relativePath)
}

// The caller holds lifecycle, readiness and installation locks in that order.
func (r *Registry) matchHTTPRouteLocked(projectID, method, relativePath string) (HTTPRouteMatch, bool) {
	var result HTTPRouteMatch
	count := 0
	for _, installation := range r.installations {
		if installation.ProjectID != projectID || !installation.Enabled || installation.Draining ||
			!r.readiness[installation.PluginID+"@"+installation.Version] {
			continue
		}
		if version, promoted := r.active[projectID+":"+installation.PluginID]; promoted && version != installation.Version {
			continue
		}
		manifest, exists := r.registered[installation.PluginID+"@"+installation.Version]
		if !exists || manifest.Manifest.Mount == "" {
			continue
		}
		for _, route := range manifest.Manifest.Routes {
			if route.Method != method || route.Path == "" {
				continue
			}
			if !strings.HasPrefix(relativePath, manifest.Manifest.Mount+"/") && relativePath != manifest.Manifest.Mount {
				continue
			}
			suffix := strings.TrimPrefix(relativePath, manifest.Manifest.Mount)
			if suffix == "" {
				suffix = "/"
			}
			params, matches := matchHTTPPath(route.Path, suffix)
			if !matches {
				continue
			}
			count++
			result = HTTPRouteMatch{PluginID: installation.PluginID, Version: installation.Version,
				Method: route.Method, Path: mountedHTTPPath(manifest.Manifest.Mount, route.Path),
				Function: route.Function, Capability: route.Capability,
				Operation: route.Operation, Auth: route.Auth, Params: params, ResourceField: route.ResourceField, WebSocketAdmission: route.WebSocketAdmission, projectID: projectID, relativePath: relativePath}
		}
	}
	// Multiple enabled versions without an active promotion fail closed.
	return result, count == 1
}

func (r *Registry) HTTPRoutes(projectID string) []HTTPRouteMatch {
	r.lifecycleMu.RLock()
	defer r.lifecycleMu.RUnlock()
	r.readinessMu.RLock()
	defer r.readinessMu.RUnlock()
	r.installationMu.RLock()
	defer r.installationMu.RUnlock()
	routes := make([]HTTPRouteMatch, 0)
	for _, installation := range r.installations {
		if installation.ProjectID != projectID || !installation.Enabled || installation.Draining ||
			!r.readiness[installation.PluginID+"@"+installation.Version] {
			continue
		}
		if version, promoted := r.active[projectID+":"+installation.PluginID]; promoted && version != installation.Version {
			continue
		}
		registered, exists := r.registered[installation.PluginID+"@"+installation.Version]
		if !exists || registered.Manifest.Mount == "" {
			continue
		}
		for _, route := range registered.Manifest.Routes {
			if route.Method == "" {
				continue
			}
			routes = append(routes, HTTPRouteMatch{PluginID: installation.PluginID, Version: installation.Version,
				Method: route.Method, Path: mountedHTTPPath(registered.Manifest.Mount, route.Path),
				Function: route.Function, Capability: route.Capability,
				Operation: route.Operation, Auth: route.Auth, ResourceField: route.ResourceField, WebSocketAdmission: route.WebSocketAdmission})
		}
	}
	counts := make(map[string]int)
	for _, route := range routes {
		counts[route.Method+" "+route.Path]++
	}
	unique := make([]HTTPRouteMatch, 0, len(routes))
	for _, route := range routes {
		if counts[route.Method+" "+route.Path] == 1 {
			unique = append(unique, route)
		}
	}
	sort.Slice(unique, func(i, j int) bool {
		if unique[i].Path == unique[j].Path {
			return unique[i].Method < unique[j].Method
		}
		return unique[i].Path < unique[j].Path
	})
	return unique
}

// InvokeAsPublic cannot inherit the privileged implicit "host" principal.
func (r *Registry) InvokeAsPublic(request WorkerRequest) (WorkerResponse, error) {
	request.accessPrincipal = "public"
	return r.Invoke(request)
}

// InvokeHTTPRoute accepts only a route selected by the host for this project.
func (r *Registry) InvokeHTTPRoute(projectID string, match HTTPRouteMatch, payload []byte, credential string) (WorkerResponse, error) {
	return r.InvokeHTTPRouteContext(context.Background(), projectID, match, payload, credential)
}

// InvokeHTTPRouteContext preserves the host request's cancellation boundary.
func (r *Registry) InvokeHTTPRouteContext(ctx context.Context, projectID string, match HTTPRouteMatch, payload []byte, credential string) (WorkerResponse, error) {
	if match.projectID != projectID || match.relativePath == "" {
		return WorkerResponse{}, ErrNotRegistered
	}
	request := WorkerRequest{Protocol: r.HostAPI, ProjectID: projectID, PluginID: match.PluginID, Version: match.Version, Operation: match.Operation, Payload: payload, httpRoute: &match, callContext: ctx}
	switch match.Auth {
	case "public":
		if credential == "" {
			return r.InvokeAsPublic(request)
		}
	case "user":
		if credential != "" {
			return r.InvokeWithUserSession(request, credential)
		}
	case "operator":
		if credential != "" {
			return r.InvokeAsOperator(request, credential)
		}
	}
	return WorkerResponse{}, ErrActorDenied
}

// Recheck the matched route at admission and under the final commit locks.
func (r *Registry) checkHTTPRouteLocked(request WorkerRequest) error {
	expected := request.httpRoute
	current, ok := r.matchHTTPRouteLocked(request.ProjectID, expected.Method, expected.relativePath)
	if !ok || current.PluginID != request.PluginID || current.Version != request.Version ||
		current.Operation != request.Operation || current.Auth != expected.Auth ||
		current.Capability != expected.Capability || current.Path != expected.Path ||
		current.Function != expected.Function || current.ResourceField != expected.ResourceField ||
		current.WebSocketAdmission != expected.WebSocketAdmission || !maps.Equal(current.Params, expected.Params) {
		return ErrNotRegistered
	}
	return nil
}
