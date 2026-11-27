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
	if ValidateWorkerRequest(request, r.HostAPI) != nil || token == "" || len(token) > 1024 {
		return WorkerResponse{}, ErrActorDenied
	}
	version, err := r.sessionAuthVersion(request.ProjectID)
	if err != nil {
		return WorkerResponse{}, err
	}
	payload, _ := json.Marshal(map[string]string{"token": token})
	authRequest := WorkerRequest{Protocol: r.HostAPI, ProjectID: request.ProjectID, PluginID: "auth", Version: version, Operation: "auth.identity", Payload: payload, accessPrincipal: "host", requireAccessPolicy: true}
	identity, err := r.invokeOnce(authRequest, false)
	if err != nil || !identity.Success || identity.ErrorCode != "" || len(identity.Observations) != 0 {
		return WorkerResponse{}, ErrActorDenied
	}
	var result struct {
		UserID string `json:"userId"`
	}
	if json.Unmarshal(identity.Payload, &result) != nil || !validAccessIdentity(result.UserID) || !strings.HasPrefix(result.UserID, request.ProjectID+":") || result.UserID == request.ProjectID+":" {
		return WorkerResponse{}, ErrActorDenied
	}
	request.accessPrincipal = "user:" + result.UserID
	request.originPrincipal = ""
	request.requireAccessPolicy = true
	return r.Invoke(request)
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
