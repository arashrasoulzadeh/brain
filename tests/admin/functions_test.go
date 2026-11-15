package admin_test

import (
	"testing"

	"brain/internal/admin"
	"brain/internal/operator"
)

func TestFunctionDraftPreservesMethodAndUpdatesSourceBeforePublish(t *testing.T) {
	service := admin.NewLifecycleService(nil, nil)
	actor := operator.Actor{ID: "admin", Kind: operator.ActorOperator}
	draft, err := service.CreateFunctionDraftWithSettings(actor, "dev", "method_contract", "export default (input) => input", 1, false, []string{"functions.invoke"}, "GET")
	if err != nil || draft.Method != "GET" {
		t.Fatalf("created draft = %#v, err = %v", draft, err)
	}
	updated, err := service.UpdateFunctionDraft(actor, "dev", "method_contract", "export default (input) => ({updated: input})", 2)
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
	result, err := service.TestFunctionDraft(actor, "dev", "echo", "export default (input) => input")
	if err != nil || result["valid"] != true {
		t.Fatalf("preflight result = %#v, err = %v", result, err)
	}
}
