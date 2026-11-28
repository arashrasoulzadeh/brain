package admin_test

import (
	"errors"
	"os"
	"testing"

	"brain/internal/admin"
	"brain/internal/operator"
)

func TestPublishedFunctionPinsSourceAndMethodAcrossDraftEdits(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(t.TempDir()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(cwd); err != nil {
			t.Error(err)
		}
	})
	service := admin.NewLifecycleService(nil, nil)
	actor := operator.Actor{Kind: operator.ActorOperator, ID: "operator"}
	const project = "publication-pinning"
	const initial = "export const config = { method: 'POST' }; export default () => ({value:'first'})"
	const updated = "export const config = { method: 'GET' }; export default () => ({value:'second'})"
	if _, err := service.CreateFunctionDraft(actor, project, "sample", initial); err != nil {
		t.Fatal(err)
	}
	published, err := service.PublishFunction(actor, project, "sample", "POST")
	if err != nil || published.ActivePublication != 1 || len(published.PublishedVersions) != 1 {
		t.Fatalf("publication: %+v %v", published, err)
	}
	if _, err := service.UpdateFunctionDraft(actor, project, "sample", updated, 1); err != nil {
		t.Fatal(err)
	}
	if _, err := service.PublishFunction(actor, project, "sample", "GET"); !errors.Is(err, admin.ErrFunctionVersionConflict) {
		t.Fatalf("same version replaced: %v", err)
	}
	// A new service reads the immutable publication, not editable source.
	service = admin.NewLifecycleService(nil, nil)
	assertValue := func(method, value string) {
		t.Helper()
		response, err := service.PublicInvokeFunction(project, "sample", method, nil)
		if err != nil {
			t.Fatal(err)
		}
		result, ok := response["result"].(map[string]any)
		if !ok || result["value"] != value {
			t.Fatalf("unexpected publication output: %+v", response)
		}
	}
	assertValue("POST", "first")
	if _, err := service.PublicInvokeFunction(project, "sample", "GET", nil); !errors.Is(err, admin.ErrFunctionMethodNotAllowed) {
		t.Fatalf("draft method changed live route: %v", err)
	}
	if _, err := service.UpdateFunctionDraft(actor, project, "sample", updated, 2); err != nil {
		t.Fatal(err)
	}
	published, err = service.PublishFunction(actor, project, "sample", "GET")
	if err != nil || published.ActivePublication != 2 || len(published.PublishedVersions) != 2 {
		t.Fatalf("new publication: %+v %v", published, err)
	}
	assertValue("GET", "second")
	if err := service.UpdateFunctionPermissions(actor, project, "sample", nil, false); err != nil {
		t.Fatal(err)
	}
	if _, err := service.PublicInvokeFunction(project, "sample", "GET", nil); !errors.Is(err, admin.ErrFunctionDraftInvalid) {
		t.Fatalf("revoked visibility still callable: %v", err)
	}
	if err := service.UpdateFunctionPermissions(actor, project, "sample", nil, true); err != nil {
		t.Fatal(err)
	}
	assertValue("GET", "second")
	if err := os.WriteFile(".local/function-drafts.json", []byte("corrupt"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := service.PublicInvokeFunction(project, "sample", "GET", nil); !errors.Is(err, admin.ErrFunctionDraftInvalid) {
		t.Fatalf("corrupt durable publication fell back to cached draft: %v", err)
	}
}
