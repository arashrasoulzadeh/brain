package admin_test

import (
	"encoding/json"
	"errors"
	"testing"

	"brain/internal/admin"
	"brain/internal/operator"
	"brain/internal/plugin"
)

func TestRequireEnabledPluginRejectsDirectAPIAccessUntilLifecycleIsComplete(t *testing.T) {
	registry := plugin.NewRegistry(1)
	manifest := plugin.Manifest{ID: "functions", Title: "Functions", Version: "1.0.0", HostAPI: 1, Capabilities: []string{"functions.invoke"}}
	if err := registry.Register(manifest); err != nil {
		t.Fatal(err)
	}
	service := admin.NewLifecycleService(nil, registry)
	actor := operator.Actor{ID: "admin", Kind: operator.ActorOperator}
	if err := service.RequireEnabledPlugin(actor, "dev", "functions"); !errors.Is(err, admin.ErrPluginUnavailable) {
		t.Fatalf("inactive plugin error = %v", err)
	}
	if err := registry.Install("dev", manifest.ID, manifest.Version); err != nil {
		t.Fatal(err)
	}
	if err := registry.SetReadiness(manifest.ID, manifest.Version, true); err != nil {
		t.Fatal(err)
	}
	if err := registry.Grant("dev", manifest.ID, manifest.Version, "functions.invoke"); err != nil {
		t.Fatal(err)
	}
	if err := registry.Enable("dev", manifest.ID, manifest.Version); err != nil {
		t.Fatal(err)
	}
	if err := service.RequireEnabledPlugin(actor, "dev", "functions"); err != nil {
		t.Fatalf("active plugin was rejected: %v", err)
	}
}

func TestFunctionDraftPreservesMethodAndUpdatesSourceBeforePublish(t *testing.T) {
	service := admin.NewLifecycleService(nil, nil)
	actor := operator.Actor{ID: "admin", Kind: operator.ActorOperator}
	source := "export const config = { method: 'GET' };\nexport default (input) => input"
	draft, err := service.CreateFunctionDraftWithSettings(actor, "dev", "method_contract", source, 1, false, []string{"functions.invoke"}, "POST")
	if err != nil || draft.Method != "GET" {
		t.Fatalf("created draft = %#v, err = %v", draft, err)
	}
	updated, err := service.UpdateFunctionDraft(actor, "dev", "method_contract", "export const config = { method: 'GET' };\nexport default (input) => ({updated: input})", 2)
	if err != nil || updated.Version != 2 || updated.Source == draft.Source {
		t.Fatalf("updated draft = %#v, err = %v", updated, err)
	}
	published, err := service.PublishFunction(actor, "dev", "method_contract", "GET")
	if err != nil || !published.Public || published.Method != "GET" || published.PublishedAt == nil {
		t.Fatalf("published draft = %#v, err = %v", published, err)
	}
}

func TestFunctionDraftPreflightRejectsUnsafeSource(t *testing.T) {
	service := admin.NewLifecycleService(nil, nil)
	actor := operator.Actor{ID: "admin", Kind: operator.ActorOperator}
	if _, err := service.TestFunctionDraft(actor, "dev", "unsafe", "export default () => fetch('https://example.test')"); err == nil {
		t.Fatal("unsafe source was accepted")
	}
}

func TestFunctionDraftPreflightAcceptsExportedHandler(t *testing.T) {
	service := admin.NewLifecycleService(nil, nil)
	actor := operator.Actor{ID: "admin", Kind: operator.ActorOperator}
	result, err := service.TestFunctionDraft(actor, "dev", "echo", "export const config = { method: 'POST' };\nexport default (input) => input")
	if err != nil || result["valid"] != true {
		t.Fatalf("preflight result = %#v, err = %v", result, err)
	}
}

func TestPublicFunctionUsesMethodDeclaredInSource(t *testing.T) {
	service := admin.NewLifecycleService(nil, nil)
	actor := operator.Actor{ID: "admin", Kind: operator.ActorOperator}
	if _, err := service.CreateFunctionDraftWithSettings(actor, "dev", "public_method", "export const config = { method: 'POST' };\nexport default (input) => input", 1, false, nil, "GET"); err != nil {
		t.Fatal(err)
	}
	if _, err := service.PublishFunction(actor, "dev", "public_method", "GET"); err != nil {
		t.Fatal(err)
	}
	if _, err := service.PublicInvokeFunction("dev", "public_method", "GET", nil); !errors.Is(err, admin.ErrFunctionMethodNotAllowed) {
		t.Fatalf("GET error = %v, want method denial", err)
	}
}

func TestFunctionDraftReceivesTrustedProjectContext(t *testing.T) {
	service := admin.NewLifecycleService(nil, nil)
	actor := operator.Actor{ID: "admin", Kind: operator.ActorOperator}
	source := "export const config = { method: 'POST' };\nexport default (input, ctx) => ({id: ctx.project.id, active: ctx.project.active, frozen: Object.isFrozen(ctx.project), plugin: ctx.plugin.id, operation: ctx.execution.operation, actor: ctx.auth.actor.id, caller: ctx.auth.caller.kind, allowed: ctx.auth.can('functions.invoke', 'context_contract'), input})"
	result, err := service.RunFunctionDraft(actor, "dev", "context_contract", source, json.RawMessage(`{"projectId":"forged"}`))
	if err != nil {
		t.Fatal(err)
	}
	if result == nil {
		t.Fatal("expected function execution result")
	}
	encoded, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	var output struct {
		Result struct {
			ID      string `json:"id"`
			Active  bool   `json:"active"`
			Frozen  bool   `json:"frozen"`
			Actor   string `json:"actor"`
			Caller  string `json:"caller"`
			Allowed bool   `json:"allowed"`
		} `json:"result"`
	}
	if err := json.Unmarshal(encoded, &output); err != nil || output.Result.ID != "dev" || !output.Result.Active || !output.Result.Frozen || output.Result.Actor != "admin" || output.Result.Caller != "operator" || !output.Result.Allowed {
		t.Fatalf("project context = %s, err = %v", encoded, err)
	}
}

func TestPublicFunctionAuthContextIsAnonymous(t *testing.T) {
	service := admin.NewLifecycleService(nil, nil)
	actor := operator.Actor{ID: "admin", Kind: operator.ActorOperator}
	source := "export const config = { method: 'POST' }; export default (input, ctx) => ({ kind: ctx.auth.actor.kind, authenticated: ctx.auth.actor.authenticated, allowed: ctx.auth.can('functions.call', 'public_auth_context'), admin: ctx.auth.can('functions.invoke', 'public_auth_context') })"
	if _, err := service.CreateFunctionDraftWithSettings(actor, "dev", "public_auth_context", source, 1, true, nil, "POST"); err != nil {
		t.Fatal(err)
	}
	if _, err := service.PublishFunction(actor, "dev", "public_auth_context", "POST"); err != nil {
		t.Fatal(err)
	}
	output, err := service.PublicInvokeFunction("dev", "public_auth_context", "POST", nil)
	if err != nil {
		t.Fatal(err)
	}
	result, ok := output["result"].(map[string]any)
	if !ok || result["kind"] != "public" || result["authenticated"] != false || result["allowed"] != true || result["admin"] != false {
		t.Fatalf("public caller was not anonymous: %+v", output)
	}
}
