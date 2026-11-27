package plugin_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"errors"
	"testing"

	"brain/internal/admin"
	"brain/internal/app"
	"brain/internal/operator"
	"brain/internal/plugin"
)

func TestVerifiedUserSessionSuppliesCurrentActor(t *testing.T) {
	auth, err := plugin.LoadManifest("../../plugins/auth/plugin.yaml")
	if err != nil {
		t.Fatal(err)
	}
	identity, _ := auth.Route("auth.identity")
	read := plugin.Route{Operation: "records.read", Function: "read", Capability: "records.read", ResourceField: "record"}
	records := plugin.Manifest{ID: "records", Version: "1", HostAPI: 1, Capabilities: []string{read.Capability}, Routes: []plugin.Route{read}}
	r, _ := accessRegistry(t, auth, records)
	userWorker := &countingWorker{}
	authWorker := &storageCommitWorker{handle: func(req plugin.WorkerRequest) plugin.WorkerResponse {
		var input map[string]string
		_ = json.Unmarshal(req.Payload, &input)
		if input["token"] != "valid-session" || req.ProjectID != "alpha" {
			return plugin.WorkerResponse{Protocol: req.Protocol, ErrorCode: "invalid session"}
		}
		return plugin.WorkerResponse{Protocol: req.Protocol, Success: true, Payload: []byte(`{"userId":"alpha:alice"}`)}
	}}
	for _, project := range []string{"alpha", "beta"} {
		installAccessTestPlugin(t, r, project, auth, authWorker)
		installAccessTestPlugin(t, r, project, records, userWorker)
	}
	req := plugin.WorkerRequest{Protocol: 1, ProjectID: "alpha", PluginID: records.ID, Version: records.Version, Operation: read.Operation, Payload: []byte(`{"record":"one","actor":"operator:root"}`)}
	if _, err := r.InvokeWithUserSession(req, "valid-session"); !errors.Is(err, plugin.ErrActorDenied) {
		t.Fatalf("unenrolled project admitted: %v", err)
	}
	for _, project := range []string{"alpha", "beta"} {
		installAccessTestPlugin(t, r, project, accessManifest(t), &contractWorker{ready: true})
		createAccessRule(t, r, project, "auth-host", "host", auth, identity, "", true)
	}
	if _, err := r.InvokeWithUserSession(req, "valid-session"); !errors.Is(err, plugin.ErrActorDenied) {
		t.Fatalf("missing user grant admitted: %v", err)
	}
	revision := createAccessRule(t, r, "alpha", "alice-read", "user:alpha:alice", records, read, "one", true)
	if _, err := r.InvokeWithUserSession(req, "invalid"); !errors.Is(err, plugin.ErrActorDenied) {
		t.Fatalf("invalid token admitted: %v", err)
	}
	other := req
	other.ProjectID = "beta"
	if _, err := r.InvokeWithUserSession(other, "valid-session"); !errors.Is(err, plugin.ErrActorDenied) {
		t.Fatalf("cross-project session admitted: %v", err)
	}
	if response, err := r.InvokeWithUserSession(req, "valid-session"); err != nil || !response.Success {
		t.Fatalf("verified user denied: %+v %v", response, err)
	}
	ctx := userWorker.last.HostContext
	if ctx == nil || ctx.Auth.Actor.ID != "alpha:alice" || ctx.Auth.Actor.Kind != "user" || ctx.Auth.Caller != ctx.Auth.Actor || len(ctx.Auth.Permissions) != 1 || ctx.Auth.Permissions[0].Resource != "one" || userWorker.calls != 1 {
		t.Fatalf("wrong user context: %+v", ctx)
	}
	revokeAccessRule(t, r, "alpha", "alice-read", revision)
	if _, err := r.InvokeWithUserSession(req, "valid-session"); !errors.Is(err, plugin.ErrActorDenied) {
		t.Fatalf("revoked grant admitted: %v", err)
	}
}

func TestPublishedFunctionReceivesVerifiedUserSession(t *testing.T) {
	auth, err := plugin.LoadManifest("../../plugins/auth/plugin.yaml")
	if err != nil {
		t.Fatal(err)
	}
	functions, err := plugin.LoadManifest("../../plugins/functions/plugin.yaml")
	if err != nil {
		t.Fatal(err)
	}
	r, _ := accessRegistry(t, auth, functions)
	project, name := "public-session-project", "session_context"
	installAccessTestPlugin(t, r, project, auth, &storageCommitWorker{handle: func(req plugin.WorkerRequest) plugin.WorkerResponse {
		var input map[string]string
		_ = json.Unmarshal(req.Payload, &input)
		if input["token"] != "valid-session" {
			return plugin.WorkerResponse{Protocol: req.Protocol, ErrorCode: "invalid session"}
		}
		return plugin.WorkerResponse{Protocol: req.Protocol, Success: true, Payload: []byte(`{"userId":"public-session-project:alice"}`)}
	}})
	installAccessTestPlugin(t, r, project, functions, &contractWorker{ready: true})
	installAccessTestPlugin(t, r, project, accessManifest(t), &contractWorker{ready: true})
	identity, _ := auth.Route("auth.identity")
	createAccessRule(t, r, project, "auth-host", "host", auth, identity, "", true)
	for _, operation := range []string{"functions.create", "functions.publish"} {
		route, exists := functions.Route(operation)
		if !exists {
			t.Fatalf("missing route %s", operation)
		}
		createAccessRule(t, r, project, strings.ReplaceAll(operation, ".", "-"), "operator:root", functions, route, name, true)
	}
	service := admin.NewLifecycleService(nil, r)
	actor := operator.Actor{Kind: operator.ActorOperator, ID: "root"}
	source := "export const config = { method: 'POST' }; export default (input, ctx) => ({ id:ctx.auth.actor.id, kind:ctx.auth.actor.kind, allowed:ctx.auth.can('functions.call','session_context') })"
	if _, err := service.CreateFunctionDraftWithSettings(actor, project, name, source, 1, false, nil, "POST"); err != nil {
		t.Fatal(err)
	}
	if _, err := service.PublishFunction(actor, project, name, "POST"); err != nil {
		t.Fatal(err)
	}
	call, _ := functions.Route("functions.call")
	createAccessRule(t, r, project, "anonymous-call", "public", functions, call, name, true)
	handler := app.PublicFunctionHandler(service)
	invoke := func(token string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, "/public/"+project+"/"+name, strings.NewReader(`{}`))
		req.SetPathValue("project_id", project)
		req.SetPathValue("function_name", name)
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		return rec
	}
	if rec := invoke("invalid"); rec.Code == http.StatusOK {
		t.Fatal("invalid token fell back to anonymous grant")
	}
	if rec := invoke("valid-session"); rec.Code == http.StatusOK {
		t.Fatal("verified user inherited public grant")
	}
	createAccessRule(t, r, project, "user-call", "user:"+project+":alice", functions, call, name, true)
	rec := invoke("valid-session")
	if rec.Code != http.StatusOK {
		t.Fatalf("verified user rejected: %d %s", rec.Code, rec.Body.String())
	}
	var output struct {
		Result struct {
			ID      string `json:"id"`
			Kind    string `json:"kind"`
			Allowed bool   `json:"allowed"`
		} `json:"result"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &output); err != nil || output.Result.ID != project+":alice" || output.Result.Kind != "user" || !output.Result.Allowed {
		t.Fatalf("incorrect function identity: %s %v", rec.Body.String(), err)
	}
}
