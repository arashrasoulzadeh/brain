package admin_test

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"brain/internal/admin"
	"brain/internal/operator"
	"brain/internal/project"
)

func TestFunctionDraftValidationAndSandboxFailures(t *testing.T) {
	const projectID = "function-failure-coverage"
	service := admin.NewLifecycleServiceWithProjects(nil, nil, project.NewStore(project.Project{ID: projectID, Active: true}))
	actor := operator.Actor{Kind: operator.ActorOperator, ID: "admin"}
	user := operator.Actor{Kind: operator.ActorUser, ID: "user"}
	source := "export const config = { method: 'POST' }; export default () => ({ok:true})"
	if _, err := service.ListFunctionDrafts(user, projectID); !errors.Is(err, admin.ErrFunctionDraftInvalid) {
		t.Fatalf("user drafts: %v", err)
	}
	if _, err := service.ListFunctionDrafts(actor, "missing"); err == nil {
		t.Fatal("missing project accepted")
	}
	if _, err := service.ListFunctionLogs(user, projectID, ""); !errors.Is(err, admin.ErrFunctionDraftInvalid) {
		t.Fatalf("user logs: %v", err)
	}
	if _, err := service.CreateFunctionDraft(actor, "missing", "draft", source); err == nil {
		t.Fatal("draft created in unknown project")
	}
	if _, err := service.CreateFunctionDraft(user, projectID, "draft", source); !errors.Is(err, admin.ErrFunctionDraftInvalid) {
		t.Fatalf("user draft: %v", err)
	}
	if _, err := service.PublishFunction(user, projectID, "draft", "POST"); !errors.Is(err, admin.ErrFunctionDraftInvalid) {
		t.Fatalf("user publication: %v", err)
	}
	if _, err := service.PublishFunction(actor, "missing", "draft", "POST"); err == nil {
		t.Fatal("published in unknown project")
	}
	if err := service.UpdateFunctionPermissions(user, projectID, "draft", nil, true); !errors.Is(err, admin.ErrFunctionDraftInvalid) {
		t.Fatalf("user permissions: %v", err)
	}
	if _, err := service.PublicInvokeFunctionWithSession(projectID, "draft", "POST", nil, "invalid"); err == nil {
		t.Fatal("unverified session admitted")
	}
	for _, tc := range []struct {
		name  string
		src   string
		input json.RawMessage
		want  string
	}{
		{"invalid_source", "invalid", nil, "function source"},
		{"invalid_input", source, json.RawMessage("{"), "valid JSON"},
		{"throwing", "export const config = { method: 'POST' }; export default () => { throw new Error('fixture failure') }", nil, "function execution failed"},
		{"nonfunction", "export const config = { method: 'POST' }; export default 1", nil, "function execution failed"},
		{"long_error", "export const config = { method: 'POST' }; export default () => { throw new Error('x'.repeat(5000)) }", nil, "function execution failed"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := service.RunFunctionDraft(actor, projectID, tc.name, tc.src, tc.input); err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("invalid function result: %v", err)
			}
			logs, err := service.ListFunctionLogs(actor, projectID, tc.name)
			if err != nil || len(logs) == 0 || logs[len(logs)-1].Status != "failed" || logs[len(logs)-1].RequestID == "" {
				t.Fatalf("missing failure log: %#v %v", logs, err)
			}
		})
	}
	if _, err := service.RunFunctionDraft(actor, projectID, "valid", source, nil); err != nil {
		t.Fatal(err)
	}
}
