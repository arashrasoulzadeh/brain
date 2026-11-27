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

func coverageOperator(t *testing.T) (*operator.Store, string) {
	t.Helper()
	store := &operator.Store{}
	digest := func(value string) string { return "digest:" + value }
	if _, err := store.Bootstrap("admin", "fixture", digest); err != nil {
		t.Fatal(err)
	}
	token, err := store.AuthenticateSession("admin", "fixture", digest, func() string { return "coverage-token" })
	if err != nil {
		t.Fatal(err)
	}
	return store, token
}

func coverageHTTPRequest(t *testing.T, handler http.Handler, method, path, token, body string, status int) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != status {
		t.Fatalf("%s %s: %d want %d; %s", method, path, rec.Code, status, rec.Body.String())
	}
	return rec
}

func TestFunctionHTTPDraftLifecycleAndValidation(t *testing.T) {
	operators, token := coverageOperator(t)
	registry := plugin.NewRegistry(1)
	m := plugin.Manifest{ID: "functions", Title: "Functions", Version: "1.0.0", HostAPI: 1, Capabilities: []string{"functions.invoke"}}
	if err := registry.Register(m); err != nil {
		t.Fatal(err)
	}
	const projectID = "http-functions-coverage"
	for _, id := range []string{projectID, "http-inactive-project"} {
		if err := registry.Install(id, m.ID, m.Version); err != nil {
			t.Fatal(err)
		}
		if err := registry.Grant(id, m.ID, m.Version, "functions.invoke"); err != nil {
			t.Fatal(err)
		}
		if err := registry.SetReadiness(m.ID, m.Version, true); err != nil {
			t.Fatal(err)
		}
		if err := registry.Enable(id, m.ID, m.Version); err != nil {
			t.Fatal(err)
		}
	}
	const projectSlug = "function-demo"
	projects := project.NewStore(project.Project{ID: projectID, Slug: projectSlug, Active: true}, project.Project{ID: "other-project", Active: true})
	service := admin.NewLifecycleServiceWithProjects(operators, registry, projects)
	handler := app.AuthenticatedFunctionsHandler(service, operators)
	path := "https://admin.example.test/admin/functions"
	coverageHTTPRequest(t, app.AuthenticatedFunctionsHandler(nil, operators), http.MethodGet, path, token, "", http.StatusForbidden)
	coverageHTTPRequest(t, handler, http.MethodGet, path, token, "", http.StatusForbidden)
	coverageHTTPRequest(t, handler, http.MethodGet, path+"?project_id=http-inactive-project", token, "", http.StatusBadRequest)
	coverageHTTPRequest(t, handler, http.MethodGet, path+"?project_id="+projectID, token, "", http.StatusOK)
	coverageHTTPRequest(t, handler, http.MethodPut, path, token, "{}", http.StatusNotFound)
	coverageHTTPRequest(t, handler, http.MethodPost, path, token, "{", http.StatusBadRequest)
	coverageHTTPRequest(t, handler, http.MethodPost, path, token, "{}", http.StatusForbidden)
	source := "export const config = { method: 'POST' }; export default (input) => ({echo: input})"
	command := func(operation, name, src string, version int, status int) *httptest.ResponseRecorder {
		t.Helper()
		body, err := json.Marshal(map[string]any{
			"operation": operation, "project_id": projectID, "name": name, "source": src,
			"function_version": version, "input": map[string]string{"value": "one"}, "permissions": []string{"functions.invoke"},
		})
		if err != nil {
			t.Fatal(err)
		}
		return coverageHTTPRequest(t, handler, http.MethodPost, path, token, string(body), status)
	}
	const name = "coverage_echo"
	command("functions.create", "", source, 1, http.StatusBadRequest)
	command("functions.create", name, "export default () => null", 1, http.StatusBadRequest)
	command("functions.create", name, source, -1, http.StatusBadRequest)
	command("functions.create", name, source, 0, http.StatusOK)
	command("functions.test", name, "export default () => null", 1, http.StatusBadRequest)
	command("functions.test", name, source, 1, http.StatusOK)
	command("functions.run", name, "", 1, http.StatusBadRequest)
	result := command("functions.run", name, source, 1, http.StatusOK)
	if !strings.Contains(result.Body.String(), "echo") {
		t.Fatal("function output missing")
	}
	command("functions.permissions", "missing", source, 1, http.StatusBadRequest)
	command("functions.permissions", name, source, 1, http.StatusNoContent)
	command("functions.update", name, "", 2, http.StatusBadRequest)
	command("functions.update", name, "export default () => null", 2, http.StatusBadRequest)
	command("functions.update", "missing", source, 2, http.StatusBadRequest)
	command("functions.update", name, source, 2, http.StatusOK)
	command("functions.publish", "missing", source, 2, http.StatusBadRequest)
	published := command("functions.publish", name, source, 2, http.StatusOK)
	if !strings.Contains(published.Body.String(), "https://admin.example.test/"+projectSlug+"/functions/"+name) {
		t.Fatalf("published URL missing: %s", published.Body.String())
	}
	command("unsupported", name, source, 1, http.StatusBadRequest)
	coverageHTTPRequest(t, handler, http.MethodGet, path+"?project_id="+projectID, token, "", http.StatusOK)
	logs := app.AuthenticatedFunctionLogsHandler(service, operators)
	coverageHTTPRequest(t, logs, http.MethodGet, "/admin/function-logs", token, "", http.StatusForbidden)
	response := coverageHTTPRequest(t, logs, http.MethodGet, "/admin/function-logs?project_id="+projectID+"&name="+name, token, "", http.StatusOK)
	if !strings.Contains(response.Body.String(), "coverage_echo") {
		t.Fatal("function log missing")
	}

	server := app.NewServerWithOperatorAdmin(service, operators, func(value string) string { return "digest:" + value }).Handler()
	publicPath := "/" + projectSlug + "/functions/" + name
	coverageHTTPRequest(t, server, http.MethodPost, publicPath, "", "{", http.StatusBadRequest)
	wrongMethod := coverageHTTPRequest(t, server, http.MethodGet, publicPath, "", "", http.StatusMethodNotAllowed)
	if wrongMethod.Header().Get("Allow") == "" {
		t.Fatal("missing method allowlist")
	}
	coverageHTTPRequest(t, server, http.MethodPost, publicPath, "invalid", "{}", http.StatusNotFound)
	req := httptest.NewRequest(http.MethodPost, publicPath, strings.NewReader("{}"))
	req.Header.Set("Authorization", "Basic invalid")
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("malformed authorization: %d", rec.Code)
	}
	coverageHTTPRequest(t, server, http.MethodPost, "/"+projectID+"/missing", "", "{}", http.StatusNotFound)
	coverageHTTPRequest(t, server, http.MethodPost, publicPath, "", "{}", http.StatusOK)
	coverageHTTPRequest(t, server, http.MethodPost, "/"+projectID+"/functions/"+name, "", "{}", http.StatusOK)
	coverageHTTPRequest(t, server, http.MethodPost, "/other-project/functions/"+name, "", "{}", http.StatusNotFound)
	coverageHTTPRequest(t, server, http.MethodPost, "/functions/"+name, "", "{}", http.StatusNotFound)
	coverageHTTPRequest(t, server, http.MethodPost, "/public/"+projectID+"/"+name, "", "{}", http.StatusNotFound)
	coverageHTTPRequest(t, server, http.MethodPost, publicPath, "", `{"project_id":"other-project"}`, http.StatusBadRequest)
}

func TestAdminHTTPRejectsUnavailableServicesAndInvalidCommands(t *testing.T) {
	operators, token := coverageOperator(t)
	service := admin.NewLifecycleService(operators, nil)
	for _, handler := range []http.Handler{
		app.AuthenticatedProjectsHandler(service, operators),
		app.AuthenticatedPluginsHandler(service, operators),
		app.AuthenticatedAuditHandler(service, operators),
		app.AuthenticatedSystemHandler(service, operators),
	} {
		coverageHTTPRequest(t, handler, http.MethodPost, "/admin/fixture", token, "{}", http.StatusNotFound)
		coverageHTTPRequest(t, handler, http.MethodGet, "/admin/fixture", token, "", http.StatusForbidden)
	}
	coverageHTTPRequest(t, app.AuthenticatedAdminHandler(service, nil), http.MethodPost, "/admin/actions", token, "{}", http.StatusForbidden)
	coverageHTTPRequest(t, app.AdminHandler(nil), http.MethodPost, "/admin/actions", token, "{}", http.StatusNotFound)
	coverageHTTPRequest(t, app.OperatorSessionHandler(nil, nil), http.MethodPost, "/admin/session", token, "{}", http.StatusNotFound)
	coverageHTTPRequest(t, app.OperatorSessionHandler(operators, func(value string) string { return "digest:" + value }), http.MethodPost, "/admin/session", "", "{", http.StatusBadRequest)
	coverageHTTPRequest(t, app.OperatorLogoutHandler(nil), http.MethodDelete, "/admin/session", token, "", http.StatusNotFound)

	projects := project.NewStore()
	registry := plugin.NewRegistry(1)
	service = admin.NewLifecycleServiceWithProjects(operators, registry, projects)
	handler := app.AuthenticatedAdminHandler(service, operators)
	coverageHTTPRequest(t, handler, http.MethodPost, "/admin/actions", token, "{", http.StatusBadRequest)
	coverageHTTPRequest(t, handler, http.MethodPost, "/admin/actions", token, `{"operation":"unsupported"}`, http.StatusBadRequest)
	coverageHTTPRequest(t, handler, http.MethodPost, "/admin/actions", token, `{"operation":"plugin.configure","values":{"limit":1.5}}`, http.StatusBadRequest)
	coverageHTTPRequest(t, handler, http.MethodPost, "/admin/actions", token, `{"operation":"plugin.configure","values":{"limit":4}}`, http.StatusForbidden)
	coverageHTTPRequest(t, handler, http.MethodPost, "/admin/actions", token, `{"operation":"plugin.admin.invoke"}`, http.StatusBadRequest)
	for _, operation := range []string{"plugin.worker.start", "plugin.grant-resource", "plugin.admin.inventory", "plugin.admin.revision", "plugin.admin.query", "plugin.invoke", "plugin.disable.preview"} {
		coverageHTTPRequest(t, handler, http.MethodPost, "/admin/actions", token, `{"operation":"`+operation+`","project_id":"missing","plugin_id":"missing","version":"1"}`, http.StatusForbidden)
	}
	for _, handler := range []http.Handler{
		app.AuthenticatedProjectsHandler(service, operators),
		app.AuthenticatedPluginsHandler(service, operators),
		app.AuthenticatedAuditHandler(service, operators),
		app.AuthenticatedSystemHandler(service, operators),
	} {
		coverageHTTPRequest(t, handler, http.MethodGet, "/admin/fixture", token, "", http.StatusOK)
	}
}
