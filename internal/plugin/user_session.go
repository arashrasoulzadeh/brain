package plugin

import (
	"encoding/json"
	"strings"
)

// InvokeWithUserSession resolves the authenticated project user through the
// installed auth plugin before admitting the requested operation. The token
// is never attached to the target request or its invocation context.
// Session validity is checked at admission; policy is rechecked at commit.
func (r *Registry) InvokeWithUserSession(request WorkerRequest, token string) (WorkerResponse, error) {
	if ValidateWorkerRequest(request, r.HostAPI) != nil {
		return WorkerResponse{}, ErrActorDenied
	}
	actor, err := r.ResolveUserSession(request.ProjectID, token)
	if err != nil {
		return WorkerResponse{}, err
	}
	request.accessPrincipal = "user:" + actor.ID
	request.originPrincipal = ""
	request.requireAccessPolicy = true
	return r.Invoke(request)
}

// ResolveUserSession is for trusted adapters of existing host operations.
// It never exposes a token to the operation being authorized.
func (r *Registry) ResolveUserSession(projectID, token string) (InvocationActor, error) {
	if !isValidProjectID(projectID) || token == "" || len(token) > 1024 {
		return InvocationActor{}, ErrActorDenied
	}
	version, err := r.sessionAuthVersion(projectID)
	if err != nil {
		return InvocationActor{}, err
	}
	payload, _ := json.Marshal(map[string]string{"token": token})
	authRequest := WorkerRequest{Protocol: r.HostAPI, ProjectID: projectID, PluginID: "auth", Version: version, Operation: "auth.identity", Payload: payload, accessPrincipal: "host", requireAccessPolicy: true}
	identity, err := r.invokeOnce(authRequest, false)
	if err != nil || !identity.Success || identity.ErrorCode != "" || len(identity.Observations) != 0 {
		return InvocationActor{}, ErrActorDenied
	}
	var result struct {
		UserID string `json:"userId"`
	}
	if json.Unmarshal(identity.Payload, &result) != nil || !validAccessIdentity(result.UserID) || !strings.HasPrefix(result.UserID, projectID+":") || result.UserID == projectID+":" {
		return InvocationActor{}, ErrActorDenied
	}
	return invocationActor("user:" + result.UserID), nil
}

func (r *Registry) sessionAuthVersion(projectID string) (string, error) {
	r.lifecycleMu.RLock()
	defer r.lifecycleMu.RUnlock()
	r.readinessMu.RLock()
	defer r.readinessMu.RUnlock()
	r.installationMu.RLock()
	defer r.installationMu.RUnlock()
	if !r.accessInstalledLocked(projectID) {
		return "", ErrActorDenied
	}
	version := r.active[projectID+":auth"]
	if version == "" {
		for _, installation := range r.installations {
			if installation.ProjectID != projectID || installation.PluginID != "auth" || !installation.Enabled || installation.Draining {
				continue
			}
			if version != "" {
				return "", ErrActorDenied
			}
			version = installation.Version
		}
	}
	manifest, ok := r.registered["auth@"+version]
	route, declared := manifest.Manifest.Route("auth.identity")
	if !ok || !declared || route.Capability != "auth.session" || route.Helper != "sdk.request" || route.ResourceField != "" {
		return "", ErrActorDenied
	}
	return version, nil
}
