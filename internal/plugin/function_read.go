package plugin

import (
	"encoding/json"
	"strings"
)

// A function-callable read is an explicit manifest opt-in, not a generic
// capability proxy. Require a static operator route with a declared read-only
// payload contract; do not admit model, delegated, admin, or socket operations.
func validFunctionReadRoute(m Manifest, route Route) bool {
	if m.ID == "functions" || m.ID == accessPluginID || route.Auth != "operator" || route.Method == "" ||
		route.Path == "" || strings.ContainsAny(route.Path, "{}") || route.WebSocketAdmission != "" ||
		route.Helper != "sdk.dispatch" || route.ResourceField == "" || m.Admin == nil {
		return false
	}
	for _, page := range m.Admin.Pages {
		for _, action := range page.Actions {
			if action.Operation == route.Operation && action.ReadOnly {
				for _, field := range action.Fields {
					if field.Name == route.ResourceField && field.Required && field.Type == "string" {
						return true
					}
				}
			}
		}
	}
	return false
}

func (r *Registry) resolveFunctionOperation(projectID, operation string) (HTTPRouteMatch, error) {
	var selected HTTPRouteMatch
	count := 0
	for _, route := range r.HTTPRoutes(projectID) {
		if route.Operation == operation && route.Auth == "operator" && !strings.ContainsAny(route.Path, "{}") {
			selected = route
			count++
		}
	}
	if count != 1 {
		return HTTPRouteMatch{}, ErrGrantDenied
	}
	match, ok := r.MatchHTTPRoute(projectID, selected.Method, selected.Path)
	if !ok || match.PluginID != selected.PluginID || match.Version != selected.Version || match.Operation != operation || match.Auth != "operator" {
		return HTTPRouteMatch{}, ErrGrantDenied
	}
	return match, nil
}

// AuthorizeFunctionBridgeSource checks the exact active Functions installation
// and the actual operator's invocation permission, even without Access Management.
func (r *Registry) AuthorizeFunctionBridgeSource(projectID, principal, name string) error {
	if !strings.HasPrefix(principal, "operator:") || !validAccessPrincipal(principal) {
		return ErrActorDenied
	}
	match, err := r.resolveFunctionOperation(projectID, "functions.invoke")
	if err != nil || match.PluginID != "functions" || match.ResourceField != "name" {
		return ErrGrantDenied
	}
	payload, _ := json.Marshal(map[string]string{"name": name})
	request := WorkerRequest{Protocol: r.HostAPI, ProjectID: projectID, PluginID: match.PluginID, Version: match.Version, Operation: match.Operation, Payload: payload, accessPrincipal: principal, httpRoute: &match}
	return r.authorizeInvocation(request, false)
}

// InvokeFunctionRead resolves the current installed version, reuses the normal
// project, worker, grant and resource checks, and admits no storage writes.
// principal and allow are supplied by the trusted function host, not the VM.
func (r *Registry) InvokeFunctionRead(projectID, principal, operation string, payload json.RawMessage, allow func(string) error) (WorkerResponse, error) {
	if allow == nil || !strings.HasPrefix(principal, "operator:") || !validAccessPrincipal(principal) {
		return WorkerResponse{}, ErrActorDenied
	}
	match, err := r.resolveFunctionOperation(projectID, operation)
	if err != nil {
		return WorkerResponse{}, err
	}
	r.readinessMu.RLock()
	manifest := r.registered[match.PluginID+"@"+match.Version].Manifest
	r.readinessMu.RUnlock()
	route, exists := manifest.Route(operation)
	if !exists || !route.FunctionRead || !validFunctionReadRoute(manifest, route) {
		return WorkerResponse{}, ErrGrantDenied
	}
	// The strict admin payload contract rejects project overrides, stored
	// snapshots, extra fields, and missing resource keys before worker dispatch.
	valid := false
	for _, page := range manifest.Admin.Pages {
		if manifest.AdminReadOnly(page.ID, operation) && manifest.ValidateAdminPayload(page.ID, operation, payload) == nil {
			valid = true
			break
		}
	}
	if !valid {
		return WorkerResponse{}, ErrInvalidWorkerRequest
	}
	request := WorkerRequest{Protocol: r.HostAPI, ProjectID: projectID, PluginID: match.PluginID, Version: match.Version, Operation: operation, Payload: payload, accessPrincipal: principal, httpRoute: &match, functionRead: true}
	if _, err := invocationResource(request, route); err != nil {
		return WorkerResponse{}, err
	}
	guard := func() error { return allow(route.Capability) }
	if err := guard(); err != nil {
		return WorkerResponse{}, err
	}
	response, err := r.invokeUsingWorker(request, false, nil, guard)
	if err != nil || !response.Success || response.ErrorCode != "" {
		if err != nil {
			return WorkerResponse{}, err
		}
		return WorkerResponse{}, ErrGrantDenied
	}
	// Recheck the originating function permission before returning its read.
	if err := guard(); err != nil {
		return WorkerResponse{}, err
	}
	return response, nil
}
