package app_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"brain/internal/admin"
	"brain/internal/app"
	"brain/internal/operator"
	"brain/internal/plugin"
	"brain/internal/project"
)

func TestFunctionHTTPAuthoringPublishingAndExecution(t *testing.T) {
	manifest, err := plugin.LoadManifest("../../plugins/functions/plugin.yaml")
	if err != nil {
		t.Fatal(err)
	}
	// The draft adapter currently writes .local relative to the process.
	// Use a private working directory rather than sharing test data.
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
	const projectID = "http-functions-coverage"
	projects := project.NewStore(project.Project{ID: projectID, Active: true})
	operators := &operator.Store{}
	digest := func(value string) string { return "fixture:" + value }
	if _, err := operators.Bootstrap("operator", "secret", digest); err != nil {
		t.Fatal(err)
	}
	if err := operators.Authenticate("operator", "secret", digest); err != nil {
		t.Fatal(err)
	}
	token, err := operators.IssueSession(func() string { return "function-http-token" })
	if err != nil {
		t.Fatal(err)
	}
	registry := plugin.NewRegistry(1)
	if err := registry.Register(manifest); err != nil {
		t.Fatal(err)
	}
	if err := registry.Install(projectID, manifest.ID, manifest.Version); err != nil {
		t.Fatal(err)
	}
	if err := registry.SetReadiness(manifest.ID, manifest.Version, true); err != nil {
		t.Fatal(err)
	}
	for _, capability := range manifest.Capabilities {
		if err := registry.Grant(projectID, manifest.ID, manifest.Version, capability); err != nil {
			t.Fatal(err)
		}
	}
	if err := registry.Enable(projectID, manifest.ID, manifest.Version); err != nil {
		t.Fatal(err)
	}
	service := admin.NewLifecycleServiceWithProjects(operators, registry, projects)
	handler := app.NewServerWithOperatorAdmin(service, operators, digest).Handler()
	request := func(method, path, auth string, body any, status int) *httptest.ResponseRecorder {
		t.Helper()
		var encoded string
		if body != nil {
			data, err := json.Marshal(body)
			if err != nil {
				t.Fatal(err)
			}
			encoded = string(data)
		}
		req := httptest.NewRequest(method, "https://fixture.example"+path, strings.NewReader(encoded))
		if auth != "" {
			req.Header.Set("Authorization", auth)
		}
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != status {
			t.Fatalf("%s %s: status=%d, want %d, body=%s", method, path, rec.Code, status, rec.Body.String())
		}
		return rec
	}
	auth := "Bearer " + token
	source := "export const config = { method: 'POST' }; export default (input, ctx) => ({input, project: ctx.project.id})"
	command := func(operation string, fields map[string]any, status int) *httptest.ResponseRecorder {
		t.Helper()
		body := map[string]any{"operation": operation, "project_id": projectID, "name": "echo"}
		for key, value := range fields {
			body[key] = value
		}
		return request(http.MethodPost, "/admin/functions", auth, body, status)
	}
	command("functions.create", map[string]any{"source": source, "function_version": -1}, 400)
	command("functions.create", map[string]any{"source": "export default () => null"}, 400)
	command("functions.create", map[string]any{"name": "bad name", "source": source}, 400)
	command("functions.create", map[string]any{"source": source}, 200)
	command("functions.test", map[string]any{"source": "export default () => fetch('https://fixture.example')"}, 400)
	command("functions.test", map[string]any{"source": source}, 200)
	command("functions.update", map[string]any{"source": source, "function_version": 0}, 400)
	command("functions.update", map[string]any{"source": "", "function_version": 2}, 400)
	command("functions.update", map[string]any{"name": "missing", "source": source, "function_version": 2}, 400)
	command("functions.update", map[string]any{"source": "export default () => null", "function_version": 2}, 400)
	updated := command("functions.update", map[string]any{"source": source, "function_version": 2}, 200)
	var draft admin.FunctionDraft
	if err := json.Unmarshal(updated.Body.Bytes(), &draft); err != nil || draft.Version != 2 || draft.Method != "POST" {
		t.Fatalf("updated draft = %+v, %v", draft, err)
	}
	command("functions.permissions", map[string]any{"name": "missing", "permissions": []string{"functions.invoke"}}, 400)
	command("functions.permissions", map[string]any{"name": "invalid name"}, 400)
	command("functions.permissions", map[string]any{"permissions": []string{"functions.invoke"}, "public": false}, 200)
	list := request(http.MethodGet, "/admin/functions?project_id="+projectID, auth, nil, 200)
	if !strings.Contains(list.Body.String(), "functions.invoke") || !strings.Contains(list.Body.String(), "echo") {
		t.Fatalf("draft list: %s", list.Body.String())
	}
	publicPath := "/public/" + projectID + "/echo"
	request(http.MethodPost, publicPath, "", map[string]any{}, 404)
	run := command("functions.run", map[string]any{"source": source, "input": map[string]any{"value": 7}}, 200)
	if !strings.Contains(run.Body.String(), projectID) || !strings.Contains(run.Body.String(), `"value":7`) {
		t.Fatalf("run response: %s", run.Body.String())
	}
	command("functions.run", map[string]any{"source": "export const config = { method: 'POST' }; export default () => 42"}, 400)
	command("functions.run", map[string]any{"source": "export const config = { method: 'POST' }; export default () => { throw new Error('fixture failure') }"}, 400)
	command("functions.run", map[string]any{"source": "invalid"}, 400)
	command("functions.publish", map[string]any{"name": "missing"}, 400)
	published := command("functions.publish", map[string]any{"method": "POST"}, 200)
	if !strings.Contains(published.Body.String(), "https://fixture.example"+publicPath) || !strings.Contains(published.Body.String(), `"published":true`) {
		t.Fatalf("publish response: %s", published.Body.String())
	}
	request(http.MethodGet, publicPath, "", nil, 405)
	request(http.MethodPost, publicPath, "Basic forged", map[string]any{}, 403)
	request(http.MethodPost, publicPath, "Bearer forged", map[string]any{}, 404)
	request(http.MethodPost, publicPath, "", map[string]any{"value": 8}, 200)
	logs := request(http.MethodGet, "/admin/function-logs?project_id="+projectID+"&name=echo", auth, nil, 200)
	var records []admin.FunctionLog
	if err := json.Unmarshal(logs.Body.Bytes(), &records); err != nil {
		t.Fatal(err)
	}
	var succeeded, failed bool
	for _, record := range records {
		succeeded = succeeded || record.Status == "success"
		failed = failed || record.Status == "failed"
		if record.ProjectID != projectID || record.RequestID == "" || len(record.Steps) == 0 {
			t.Fatalf("invalid execution log: %+v", record)
		}
	}
	if !succeeded || !failed {
		t.Fatalf("expected successful and failed logs: %+v", records)
	}
	command("unsupported", nil, 400)
	if err := projects.SetActive(projectID, false); err != nil {
		t.Fatal(err)
	}
	request(http.MethodGet, "/admin/functions?project_id="+projectID, auth, nil, 400)
	command("functions.create", map[string]any{"source": source}, 400)
	command("functions.publish", nil, 400)
	command("functions.test", map[string]any{"source": source}, 400)
}
