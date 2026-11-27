package app_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"brain/internal/admin"
	"brain/internal/app"
	"brain/internal/operator"
	"brain/internal/plugin"
	"brain/internal/project"
)

func TestAdminCommandNamespacesRejectCrossScopeAndAllowReactivation(t *testing.T) {
	operators := &operator.Store{}
	digest := func(value string) string { return value }
	if _, err := operators.Bootstrap("admin", "secret", digest); err != nil {
		t.Fatal(err)
	}
	if err := operators.Authenticate("admin", "secret", digest); err != nil {
		t.Fatal(err)
	}
	token, err := operators.IssueSession(func() string { return "scope-token" })
	if err != nil {
		t.Fatal(err)
	}
	projects := project.NewStore(project.Project{ID: "alpha", Active: true}, project.Project{ID: "beta", Active: true})
	service := admin.NewLifecycleServiceWithProjects(operators, plugin.NewRegistry(1), projects)
	handler := app.NewServerWithOperatorAdmin(service, operators, digest).Handler()
	call := func(path, body string, status int) {
		t.Helper()
		req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != status {
			t.Fatalf("%s = %d, want %d: %s", path, rec.Code, status, rec.Body.String())
		}
	}
	call("/_system/admin/actions", `{"operation":"plugin.invoke","project_id":"alpha"}`, http.StatusForbidden)
	call("/alpha/admin/actions", `{"operation":"plugin.readiness"}`, http.StatusForbidden)
	call("/alpha/admin/actions", `{"operation":"project.create","project_id":"alpha"}`, http.StatusForbidden)
	call("/alpha/admin/actions", `{"operation":"plugin.enable","project_id":"beta"}`, http.StatusBadRequest)
	call("/_system/admin/actions", `{"operation":"project.deactivate","project_id":"alpha","confirmed":true}`, http.StatusOK)
	call("/alpha/admin/actions", `{"operation":"plugin.enable"}`, http.StatusForbidden)
	call("/_system/admin/actions", `{"operation":"project.activate","project_id":"alpha"}`, http.StatusOK)
	if _, err := projects.Resolve("alpha"); err != nil {
		t.Fatal(err)
	}
}
