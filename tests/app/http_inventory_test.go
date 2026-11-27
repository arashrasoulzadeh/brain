package app_test

import (
	"encoding/json"
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

func TestOperatorHTTPInventoryAndLifecycle(t *testing.T) {
	operators := &operator.Store{}
	digest := func(value string) string { return "test:" + value }
	if _, err := operators.Bootstrap("operator", "secret", digest); err != nil {
		t.Fatal(err)
	}
	registry := plugin.NewRegistry(1)
	if err := registry.Register(plugin.Manifest{ID: "sample", Version: "1.0.0", HostAPI: 1, Entry: "index.ts", SourceFiles: []string{"index.ts"}, Capabilities: []string{"sample.read"}}); err != nil {
		t.Fatal(err)
	}
	service := admin.NewLifecycleServiceWithProjects(operators, registry, project.NewStore())
	handler := app.NewServerWithOperatorAdmin(service, operators, digest).Handler()
	token := ""
	request := func(method, path, body string, status int) *httptest.ResponseRecorder {
		t.Helper()
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != status {
			t.Fatalf("%s %s: status=%d, want %d, body=%s", method, path, rec.Code, status, rec.Body.String())
		}
		return rec
	}
	request("POST", "/admin/session", "{", 400)
	login := request("POST", "/admin/session", `{"username":"operator","secret":"secret"}`, 200)
	var session struct{ Session string }
	if err := json.Unmarshal(login.Body.Bytes(), &session); err != nil || session.Session == "" {
		t.Fatalf("session = %+v, %v", session, err)
	}
	token = session.Session
	for _, body := range []string{
		`{"operation":"project.create","project_id":"dev"}`,
		`{"operation":"plugin.install","project_id":"dev","plugin_id":"sample","version":"1.0.0"}`,
		`{"operation":"plugin.grant","project_id":"dev","plugin_id":"sample","version":"1.0.0","capability":"sample.read"}`,
		`{"operation":"plugin.readiness","plugin_id":"sample","version":"1.0.0","ready":true}`,
	} {
		rec := request("POST", "/admin/actions", body, 200)
		if !strings.Contains(rec.Body.String(), `"status":"success"`) {
			t.Fatalf("missing success envelope: %s", rec.Body.String())
		}
	}
	for _, endpoint := range []struct{ path, expected string }{
		{"/admin/projects", "dev"},
		{"/admin/plugins?project_id=dev", "sample"},
		{"/admin/system", "sample"},
		{"/admin/audit", "plugin.grant"},
	} {
		rec := request("GET", endpoint.path, "", 200)
		if !strings.Contains(rec.Body.String(), endpoint.expected) {
			t.Fatalf("%s response lacks %s: %s", endpoint.path, endpoint.expected, rec.Body.String())
		}
	}
	request("POST", "/admin/actions", "{", 400)
	request("POST", "/admin/actions", `{"operation":"not.declared"}`, 400)
	request("POST", "/admin/actions", `{"operation":"plugin.configure","project_id":"dev","plugin_id":"sample","version":"1.0.0","values":{"count":1.5}}`, 400)
	request("POST", "/admin/actions", `{"operation":"plugin.configure","project_id":"dev","plugin_id":"sample","version":"1.0.0","values":{"unknown":1}}`, 403)
	request("POST", "/admin/actions", `{"operation":"plugin.admin.inventory","project_id":"dev","plugin_id":"sample","version":"1.0.0","page_id":"not-declared"}`, 403)
	request("POST", "/admin/actions", `{"operation":"plugin.admin.revision","project_id":"dev","plugin_id":"sample","version":"1.0.0","page_id":"not-declared"}`, 403)
	request("POST", "/admin/actions", `{"operation":"plugin.admin.query","project_id":"dev","plugin_id":"sample","version":"1.0.0","page_id":"not-declared"}`, 403)
	request("POST", "/admin/actions", `{"operation":"plugin.worker.start","project_id":"dev","plugin_id":"sample","version":"1.0.0"}`, 403)
	request("GET", "/admin/functions?project_id=dev", "", 403)
	request("POST", "/admin/functions", "{", 400)
	request("POST", "/admin/functions", `{"operation":"functions.create","project_id":"dev"}`, 403)
	request("POST", "/public/dev/missing", "{", 400)
	request("DELETE", "/admin/session", "", 200)
	request("GET", "/admin/projects", "", 403)
}

func TestAdminHandlerMethodAndDependencyBoundaries(t *testing.T) {
	for name, handler := range map[string]http.Handler{
		"projects": app.AuthenticatedProjectsHandler(nil, nil),
		"plugins": app.AuthenticatedPluginsHandler(nil, nil),
		"audit": app.AuthenticatedAuditHandler(nil, nil),
		"system": app.AuthenticatedSystemHandler(nil, nil),
		"actions": app.AdminHandler(nil),
		"session": app.OperatorSessionHandler(nil, nil),
		"logout": app.OperatorLogoutHandler(nil),
	} {
		t.Run(name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
			if rec.Code != http.StatusNotFound {
				t.Fatalf("status = %d, want 404", rec.Code)
			}
		})
	}
}
