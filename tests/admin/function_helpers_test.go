package admin_test

import (
	"encoding/json"
	"os/exec"
	"testing"

	"brain/internal/admin"
	"brain/internal/operator"
)

func TestUserFunctionReceivesTrustedProjectAndPureHelpers(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("Node is unavailable")
	}
	service := admin.NewLifecycleService(nil, nil)
	actor := operator.Actor{ID: "admin", Kind: operator.ActorOperator}
	source := "export const config = { method: 'POST' }; export default (input, context) => ({ project: context.project.id, normalized: context.helpers.call('string.trim', input.name) })"
	output, err := service.RunFunctionDraft(actor, "dev", "shared_helpers", source, json.RawMessage(`{"name":" room ","projectId":"forged"}`))
	if err != nil {
		t.Fatal(err)
	}
	result, ok := output["result"].(map[string]any)
	if !ok || result["project"] != "dev" || result["normalized"] != "room" {
		t.Fatalf("unexpected result: %#v", output)
	}
}
