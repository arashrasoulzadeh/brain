package admin_test

import (
	"testing"

	"brain/internal/admin"
	"brain/internal/operator"
)

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
