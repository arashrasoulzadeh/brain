package plugin_test

import (
	"encoding/json"
	"errors"
	"testing"

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
