package plugin_test

import (
	"errors"
	"strings"
	"testing"

	"brain/internal/plugin"
)

func TestUserSessionRejectsMalformedIdentityAndAmbiguousAuthVersion(t *testing.T) {
	auth, err := plugin.LoadManifest("../../plugins/auth/plugin.yaml")
	if err != nil {
		t.Fatal(err)
	}
	second := auth
	second.Version = "99.0.0"
	r, _ := accessRegistry(t, auth, second)
	for _, project := range []string{"alpha", "beta"} {
		installAccessTestPlugin(t, r, project, accessManifest(t), &contractWorker{ready: true})
	}
	for _, input := range []struct{ project, token string }{{"", "token"}, {"alpha", ""}, {"alpha", strings.Repeat("x", 1025)}, {"beta", "token"}} {
		if _, err := r.ResolveUserSession(input.project, input.token); !errors.Is(err, plugin.ErrActorDenied) {
			t.Fatalf("invalid session admitted: %v", err)
		}
	}
	if _, err := r.InvokeWithUserSession(plugin.WorkerRequest{}, "token"); !errors.Is(err, plugin.ErrActorDenied) {
		t.Fatalf("malformed request admitted: %v", err)
	}
	var response plugin.WorkerResponse
	worker := &storageCommitWorker{handle: func(req plugin.WorkerRequest) plugin.WorkerResponse {
		return response
	}}
	installAccessTestPlugin(t, r, "alpha", auth, worker)
	identity, _ := auth.Route("auth.identity")
	createAccessRule(t, r, "alpha", "auth-host", "host", auth, identity, "", true)
	for _, payload := range []string{`{`, `{"userId":""}`, `{"userId":"beta:alice"}`, `{"userId":"alpha:"}`, `{"userId":123}`} {
		response = plugin.WorkerResponse{Protocol: 1, Success: true, Payload: []byte(payload)}
		if _, err := r.ResolveUserSession("alpha", "token"); !errors.Is(err, plugin.ErrActorDenied) {
			t.Fatalf("malformed identity %s admitted: %v", payload, err)
		}
	}
	response = plugin.WorkerResponse{Protocol: 1, Success: true, Payload: []byte(`{"userId":"alpha:alice"}`)}
	if actor, err := r.ResolveUserSession("alpha", "token"); err != nil || actor.ID != "alpha:alice" {
		t.Fatalf("valid identity denied: %+v %v", actor, err)
	}
	installAccessTestPlugin(t, r, "alpha", second, worker)
	if _, err := r.ResolveUserSession("alpha", "token"); !errors.Is(err, plugin.ErrActorDenied) {
		t.Fatalf("ambiguous auth version admitted: %v", err)
	}
}
