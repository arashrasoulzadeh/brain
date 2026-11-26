package app_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"brain/internal/admin"
	"brain/internal/app"
	"brain/internal/operator"
	"brain/internal/plugin"
)

func TestAdminRoutesRequireBearerScheme(t *testing.T) {
	operators := &operator.Store{}
	digest := func(value string) string { return value }
	if _, err := operators.Bootstrap("admin", "secret", digest); err != nil {
		t.Fatal(err)
	}
	if err := operators.Authenticate("admin", "secret", digest); err != nil {
		t.Fatal(err)
	}
	token, err := operators.IssueSession(func() string { return "test-operator-token" })
	if err != nil {
		t.Fatal(err)
	}

	registry := plugin.NewRegistry(1)
	manifest := plugin.Manifest{ID: "functions", Title: "Functions", Version: "1.0.0", HostAPI: 1, Capabilities: []string{"functions.invoke"}}
	if err := registry.Register(manifest); err != nil {
		t.Fatal(err)
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
	service := admin.NewLifecycleService(operators, registry)
	handler := app.NewServerWithOperatorAdmin(service, operators, digest).Handler()

	routes := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/admin/projects"},
		{http.MethodGet, "/admin/plugins"},
		{http.MethodGet, "/admin/audit"},
		{http.MethodGet, "/admin/system"},
		{http.MethodGet, "/admin/functions?project_id=dev"},
		{http.MethodPost, "/admin/functions"},
		{http.MethodGet, "/admin/function-logs?project_id=dev"},
		{http.MethodPost, "/admin/actions"},
		{http.MethodDelete, "/admin/session"},
	}
	for _, route := range routes {
		t.Run(route.method+" "+route.path, func(t *testing.T) {
			for _, header := range []string{"", token, "Basic " + token, "Bearer ", "Bearer invalid"} {
				request := httptest.NewRequest(route.method, route.path, nil)
				request.Header.Set("Authorization", header)
				recorder := httptest.NewRecorder()
				handler.ServeHTTP(recorder, request)
				if recorder.Code != http.StatusForbidden {
					t.Errorf("authorization %q: status = %d, want %d", header, recorder.Code, http.StatusForbidden)
				}
			}
		})
	}

	request := httptest.NewRequest(http.MethodGet, "/admin/function-logs?project_id=dev", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("valid bearer authorization: status = %d, body = %s", recorder.Code, recorder.Body.String())
	}
}
