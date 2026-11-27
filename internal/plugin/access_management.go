package plugin

import (
	"encoding/json"
	"maps"
	"strings"
	"time"
)

const accessPluginID = "access-management"
const accessRuleModel = "rule"

var accessRuleSpec = ModelSpec{
	Version: 1,
	Fields: map[string]string{
		"principal":  "string",
		"plugin":     "string",
		"version":    "string",
		"operation":  "string",
		"capability": "string",
		"resource":   "string",
		"allowed":    "boolean",
	},
	Initial: "active",
	Transitions: map[string]ModelTransition{
		"revoke": {From: "active", To: "revoked"},
	},
}

func validAccessIdentity(value string) bool {
	return value != "" && len(value) <= 1000 && strings.TrimSpace(value) == value && !strings.ContainsAny(value, "\x00\r\n\uFEFF")
}

func validAccessPrincipal(value string) bool {
	if value == "host" || value == "public" {
		return true
	}
	kind, id, ok := strings.Cut(value, ":")
	return ok && (kind == "operator" || kind == "user" || kind == "plugin") && validAccessIdentity(id)
}

func pluginAccessPrincipal(id, version string) string {
	return "plugin:" + id + "@" + version
}

// InvokeAsOperator is called by the authenticated operator service only.
// The actor is never read from an RPC payload or from a worker result.
func (r *Registry) InvokeAsOperator(request WorkerRequest, actorID string) (WorkerResponse, error) {
	if !validAccessIdentity(actorID) {
		return WorkerResponse{}, ErrActorDenied
	}
	request.accessPrincipal = "operator:" + actorID
	return r.Invoke(request)
}

// CheckAdminReadAsOperator checks the exact admitted read scope without running
// a worker or producing observations. The admin service validates page metadata
// before calling this method. Delegated calls also check the target policy.
func (r *Registry) CheckAdminReadAsOperator(request WorkerRequest, actorID string) error {
	if !validAccessIdentity(actorID) {
		return ErrActorDenied
	}
	if err := ValidateWorkerRequest(request, r.HostAPI); err != nil {
		return err
	}
	if err := validatePayloadProject(request); err != nil {
		return err
	}
	request.accessPrincipal = "operator:" + actorID
	r.readinessMu.RLock()
	route, exists := r.registered[request.PluginID+"@"+request.Version].Manifest.Route(request.Operation)
	r.readinessMu.RUnlock()
	if !exists {
		return ErrGrantDenied
	}
	if route.Helper == "host.call" && route.Call != nil {
		target := WorkerRequest{Protocol: request.Protocol, ProjectID: request.ProjectID, PluginID: route.Call.Plugin, Version: route.Call.Version, Operation: route.Call.Operation, Payload: request.Payload, accessPrincipal: pluginAccessPrincipal(request.PluginID, request.Version), originPrincipal: request.originalPrincipal()}
		return r.checkPluginCall(request, target)
	}
	return r.authorizeInvocation(request, false)
}

// InvokeWithCredential is the registry-backed user invocation boundary. It
// consumes the credential and rechecks the access policy on the registry's
// normal admission and commit paths, including model and delegated operations.
func (r *Registry) InvokeWithCredential(broker *Broker, credential InvocationCredential, request WorkerRequest, now time.Time) (WorkerResponse, error) {
	if broker == nil || broker.Registry != r || credential.Call.ProjectID != request.ProjectID || credential.Call.PluginID != request.PluginID || credential.Call.Version != request.Version || credential.Call.Operation != request.Operation {
		return WorkerResponse{}, ErrInvalidWorkerRequest
	}
	if err := ValidateWorkerRequest(request, r.HostAPI); err != nil {
		return WorkerResponse{}, err
	}
	if err := r.validateCredentialResource(request, credential.Call); err != nil {
		return WorkerResponse{}, err
	}
	if err := broker.CommitCredential(credential, now); err != nil {
		return WorkerResponse{}, err
	}
	request.accessPrincipal = "user:" + credential.Call.ActorID
	return r.Invoke(request)
}

// AuthorizeHostOperation applies the same policy to trusted host adapters that
// implement a declared plugin operation outside the worker (e.g. HTTP functions).
// The principal and resource must come from the adapter, never request metadata.
func (r *Registry) AuthorizeHostOperation(projectID, principal, id, operation, resource string) error {
	r.lifecycleMu.RLock()
	defer r.lifecycleMu.RUnlock()
	r.readinessMu.RLock()
	defer r.readinessMu.RUnlock()
	r.installationMu.RLock()
	defer r.installationMu.RUnlock()
	if !r.accessInstalledLocked(projectID) {
		if strings.HasPrefix(principal, "user:") {
			return ErrActorDenied
		}
		return nil
	}
	if !validAccessPrincipal(principal) || id == accessPluginID {
		return ErrActorDenied
	}
	version := r.active[projectID+":"+id]
	if version == "" {
		for _, installation := range r.installations {
			if installation.ProjectID != projectID || installation.PluginID != id || !installation.Enabled || installation.Draining {
				continue
			}
			if version != "" {
				return ErrActorDenied
			}
			version = installation.Version
		}
	}
	registered, ok := r.registered[id+"@"+version]
	route, declared := registered.Manifest.Route(operation)
	if !ok || !declared || !validRuleResource(route, resource, false) || route.Model != nil {
		return ErrActorDenied
	}
	payload := []byte("{}")
	if route.ResourceField != "" {
		payload, _ = json.Marshal(map[string]string{route.ResourceField: resource})
	}
	return r.authorizeInvocationLocked(WorkerRequest{Protocol: r.HostAPI, ProjectID: projectID, PluginID: id, Version: version, Operation: operation, Payload: payload, accessPrincipal: principal}, false)
}

func (r *Registry) authorizeActorAccess(call HostCall) error {
	r.readinessMu.RLock()
	defer r.readinessMu.RUnlock()
	r.installationMu.RLock()
	defer r.installationMu.RUnlock()
	var scopeErr error
	if r.accessInstalledLocked(call.ProjectID) {
		registered, ok := r.registered[call.PluginID+"@"+call.Version]
		route, declared := registered.Manifest.Route(call.Operation)
		if !ok || !declared || !validRuleResource(route, call.Resource, false) {
			scopeErr = ErrActorDenied
		}
	}
	return r.authorizeAccessLocked(call.ProjectID, "user:"+call.ActorID, call.PluginID, call.Version, call.Operation, call.Capability, call.Resource, scopeErr)
}

// accessInstalledLocked is called with the installation lock held.
func (r *Registry) accessInstalledLocked(projectID string) bool {
	for _, installation := range r.installations {
		if installation.ProjectID == projectID && installation.PluginID == accessPluginID {
			return true
		}
	}
	return false
}

func (r *Registry) accessInstalled(projectID string) bool {
	r.installationMu.RLock()
	defer r.installationMu.RUnlock()
	return r.accessInstalledLocked(projectID)
}

func validRuleResource(route Route, resource string, rule bool) bool {
	if route.Model != nil {
		return resource == route.Model.Name || (rule && resource == "*")
	}
	if route.ResourceField == "" {
		return resource == ""
	}
	return validAccessIdentity(resource) && (rule || resource != "*")
}

func invocationResource(request WorkerRequest, route Route) (string, error) {
	if route.Model != nil {
		return route.Model.Name, nil
	}
	if route.ResourceField == "" {
		return "", nil
	}
	var fields map[string]json.RawMessage
	var resource string
	if json.Unmarshal(request.Payload, &fields) != nil || json.Unmarshal(fields[route.ResourceField], &resource) != nil || !validRuleResource(route, resource, false) {
		return "", ErrActorDenied
	}
	return resource, nil
}

// Credentials must match the actual resource consumed by the declared route.
// Never authorize a resource supplied independently of the invocation payload.
func (r *Registry) validateCredentialResource(request WorkerRequest, call HostCall) error {
	r.readinessMu.RLock()
	defer r.readinessMu.RUnlock()
	registered, ok := r.registered[request.PluginID+"@"+request.Version]
	route, declared := registered.Manifest.Route(request.Operation)
	if !ok || !declared || route.Capability != call.Capability {
		return ErrInvocationCredentialMismatch
	}
	resource, err := invocationResource(request, route)
	if err != nil || resource != call.Resource {
		return ErrInvocationCredentialMismatch
	}
	return nil
}

// authorizeAccessLocked is a second, project-specific layer over the existing
// host installation, capability and resource checks. Callers hold readiness
// and installation locks. Installing the policy plugin opts the project in;
// an installed but unavailable policy provider fails closed.
func (r *Registry) authorizeAccessLocked(projectID, principal, id, version, operation, capability, resource string, scopeErr error) error {
	if id == accessPluginID {
		// Management is reserved for an authenticated operator, not for
		// ordinary users, observation handlers or delegated plugin calls.
		if strings.HasPrefix(principal, "operator:") && validAccessPrincipal(principal) {
			return nil
		}
		return ErrActorDenied
	}

	installed, available := r.accessProviderLocked(projectID)
	if !installed {
		return nil
	}
	if scopeErr != nil || !available || r.LifecycleStorageHealth() != nil {
		return ErrActorDenied
	}
	if principal == "" {
		principal = "host"
	}
	if !validAccessPrincipal(principal) || operation == "" || capability == "" {
		return ErrActorDenied
	}
	storage, ok := r.storage.(*SQLiteStorage)
	if !ok || storage == nil {
		return ErrActorDenied
	}
	allowed, err := storage.accessRuleAllows(projectID, principal, id, version, operation, capability, resource)
	if err != nil || !allowed {
		return ErrActorDenied
	}
	return nil
}

// accessProviderLocked shares the exact policy readiness check with host inventory reads.
// Callers hold readiness and installation locks.
func (r *Registry) accessProviderLocked(projectID string) (bool, bool) {
	installed, available := false, 0
	for _, installation := range r.installations {
		if installation.ProjectID != projectID || installation.PluginID != accessPluginID {
			continue
		}
		installed = true
		key := accessPluginID + "@" + installation.Version
		registered, ok := r.registered[key]
		if !ok || !installation.Enabled || installation.Draining || !r.readiness[key] {
			continue
		}
		spec, ok := registered.Manifest.Models[accessRuleModel]
		if !ok || spec.Version != accessRuleSpec.Version || spec.Initial != accessRuleSpec.Initial || !maps.Equal(spec.Fields, accessRuleSpec.Fields) || !maps.Equal(spec.Transitions, accessRuleSpec.Transitions) {
			continue
		}
		if !installation.Grants["access.rules.read"] || !installation.Grants["access.rules.manage"] {
			continue
		}
		available++
	}
	return installed, available == 1
}

// Rules are immutable except for the revision-checked revoke transition.
// Reject invalid target routes instead of persisting unusable grants.
func (r *Registry) validateAccessRuleLocked(req WorkerRequest) error {
	var input modelInput
	if json.Unmarshal(req.Payload, &input) != nil || !modelFieldsValid(accessRuleSpec, input.Fields) {
		return ErrInvalidWorkerRequest
	}
	principal := input.Fields["principal"].(string)
	id := input.Fields["plugin"].(string)
	version := input.Fields["version"].(string)
	operation := input.Fields["operation"].(string)
	capability := input.Fields["capability"].(string)
	resource := input.Fields["resource"].(string)
	if !validAccessPrincipal(principal) || id == accessPluginID {
		return ErrInvalidWorkerRequest
	}
	registered, ok := r.registered[id+"@"+version]
	if !ok || registered.Manifest.ID != id || registered.Manifest.Version != version {
		return ErrInvalidWorkerRequest
	}
	if route, ok := registered.Manifest.Route(operation); ok && route.Capability == capability && validRuleResource(route, resource, true) {
		return nil
	}
	for _, observer := range registered.Manifest.Observers {
		if observer.operation() == operation && observer.Capability == capability && (resource == observer.Plugin+"."+observer.Event || resource == "*") {
			return nil
		}
	}
	return ErrInvalidWorkerRequest
}

const accessRuleScopeQuery = `SELECT record_id,schema_version,revision,state,fields_json
FROM plugin_model_records
WHERE project_id=? AND plugin_id=? AND model_name=? AND state='active'
AND json_extract(fields_json,'$.principal')=?
AND json_extract(fields_json,'$.plugin')=?
AND json_extract(fields_json,'$.version')=?
AND json_extract(fields_json,'$.operation')=?
AND json_extract(fields_json,'$.capability')=?`

func (s *SQLiteStorage) accessRuleAllows(projectID, principal, id, version, operation, capability, resource string) (bool, error) {
	const query = accessRuleScopeQuery + ` AND (json_extract(fields_json,'$.resource')=? OR json_extract(fields_json,'$.resource')=?)`
	wildcard := ""
	if resource != "" {
		wildcard = "*"
	}
	rows, err := s.db.Query(query, projectID, accessPluginID, accessRuleModel, principal, id, version, operation, capability, resource, wildcard)
	if err != nil {
		return false, err
	}
	defer rows.Close()
	allowed := false
	for rows.Next() {
		record, err := scanModel(rows, accessRuleSpec)
		if err != nil {
			return false, err
		}
		if !record.Fields["allowed"].(bool) {
			return false, nil // explicit deny overrides every allow
		}
		allowed = true
	}
	return allowed, rows.Err()
}
