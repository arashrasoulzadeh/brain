package admin_test

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"brain/internal/admin"
	"brain/internal/operator"
)

func TestFunctionDraftMutationsFailAtomically(t *testing.T) {
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
	const project = "function-persistence-coverage"
	const source = "export const config = { method: 'POST' }; export default () => ({original:true})"
	const updated = "export const config = { method: 'GET' }; export default () => ({changed:true})"
	if _, err := service.CreateFunctionDraft(actor, project, "original", source); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(".local", "function-drafts.json")
	original, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(path, 0700); err != nil {
		t.Fatal(err)
	}
	if _, err := service.CreateFunctionDraft(actor, project, "failed", source); err == nil {
		t.Fatal("failed create succeeded")
	}
	if _, err := service.UpdateFunctionDraft(actor, project, "original", updated, 2); err == nil {
		t.Fatal("failed update succeeded")
	}
	if err := service.UpdateFunctionPermissions(actor, project, "original", []string{"new"}, true); err == nil {
		t.Fatal("failed permissions change succeeded")
	}
	if _, err := service.PublishFunction(actor, project, "original", "POST"); err == nil {
		t.Fatal("failed publish succeeded")
	}
	drafts, err := service.ListFunctionDrafts(actor, project)
	if err != nil || len(drafts) != 1 || drafts[0].Name != "original" || drafts[0].Source != source || drafts[0].Version != 1 || drafts[0].Public || len(drafts[0].Permissions) != 0 || drafts[0].PublishedAt != nil {
		t.Fatalf("failed mutations became visible: %+v %v", drafts, err)
	}
	if _, err := service.PublicInvokeFunction(project, "original", "POST", nil); !errors.Is(err, admin.ErrFunctionDraftInvalid) {
		t.Fatalf("failed publication callable: %v", err)
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, original, 0640); err != nil {
		t.Fatal(err)
	}
	if _, err := service.PublishFunction(actor, project, "original", "POST"); err != nil {
		t.Fatal(err)
	}
	if _, err := service.PublicInvokeFunction(project, "original", "POST", nil); err != nil {
		t.Fatal(err)
	}
}

func TestLegacyFunctionDraftMethodValidation(t *testing.T) {
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
	const project = "function-legacy-coverage"
	if err := os.Mkdir(".local", 0750); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(".local", "function-drafts.json")
	if err := os.WriteFile(path, []byte("invalid"), 0640); err != nil {
		t.Fatal(err)
	}
	if drafts, err := service.ListFunctionDrafts(actor, project); err != nil || len(drafts) != 0 {
		t.Fatalf("corrupt drafts: %+v %v", drafts, err)
	}
	drafts := []admin.FunctionDraft{{ProjectID: project, Name: "valid", Public: true, Source: "export const config = { method: 'POST' }; export default () => null"}, {ProjectID: project, Name: "invalid", Public: true, Source: "export default () => null"}}
	encoded, err := json.Marshal(drafts)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, encoded, 0640); err != nil {
		t.Fatal(err)
	}
	if _, err := service.PublicInvokeFunction(project, "valid", "POST", nil); err != nil {
		t.Fatal(err)
	}
	if _, err := service.PublicInvokeFunction(project, "invalid", "POST", nil); err == nil {
		t.Fatal("invalid legacy method admitted")
	}
	if _, err := service.PublishFunction(actor, project, "invalid", "POST"); err == nil {
		t.Fatal("invalid legacy publication admitted")
	}
}
