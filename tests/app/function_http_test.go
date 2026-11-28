package app_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"brain/internal/admin"
	"brain/internal/app"
	"brain/internal/execution"
	"brain/internal/operator"
	"brain/internal/plugin"
	"brain/internal/project"
)

type publicFunctionExecutor func(context.Context, execution.Request) (execution.Result, error)

func (f publicFunctionExecutor) Execute(ctx context.Context, req execution.Request) (execution.Result, error) {
	return f(ctx, req)
}

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
		return request(http.MethodPost, "/"+projectID+"/admin/functions", auth, body, status)
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
	list := request(http.MethodGet, "/"+projectID+"/admin/functions", auth, nil, 200)
	if !strings.Contains(list.Body.String(), "functions.invoke") || !strings.Contains(list.Body.String(), "echo") {
		t.Fatalf("draft list: %s", list.Body.String())
	}
	publicPath := "/" + projectID + "/functions/echo"
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
	// Reproduce the editor's save-and-publish flow with a reused version.
	changed := "export const config = { method: 'GET' }; export default () => ({updated: true})"
	command("functions.update", map[string]any{"source": changed, "function_version": 2}, 200)
	request(http.MethodGet, publicPath, "", nil, 405)
	next := command("functions.publish", map[string]any{"function_version": 2}, 200)
	var publication struct {
		Version int    `json:"version"`
		Method  string `json:"method"`
	}
	if err := json.Unmarshal(next.Body.Bytes(), &publication); err != nil || publication.Version != 3 || publication.Method != "GET" {
		t.Fatalf("automatic publication version: %+v, %v", publication, err)
	}
	live := request(http.MethodGet, publicPath, "", nil, 200)
	if !strings.Contains(live.Body.String(), `"updated":true`) {
		t.Fatalf("new publication not live: %s", live.Body.String())
	}
	logs := request(http.MethodGet, "/"+projectID+"/admin/function-logs?name=echo", auth, nil, 200)
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
	// The same published URL must distinguish capacity and runtime failures
	// from denied or nonexistent functions without exposing raw worker errors.
	catalog, err := execution.NewCatalog(8, 1<<20)
	if err != nil {
		t.Fatal(err)
	}
	var dispatchErr error = execution.ErrUnavailable
	executor := publicFunctionExecutor(func(_ context.Context, req execution.Request) (execution.Result, error) {
		if dispatchErr != nil {
			return execution.Result{}, dispatchErr
		}
		return execution.Result{ID: req.ID, WorkerID: "fixture", Status: execution.Succeeded, Output: json.RawMessage(`{"result":{"scheduled":true}}`)}, nil
	})
	if err := service.BindFunctionExecution(catalog, executor, execution.Runtime{Kind: execution.UserFunctionRuntimeKind, Profile: "test"}); err != nil {
		t.Fatal(err)
	}
	request(http.MethodGet, publicPath, "", nil, http.StatusServiceUnavailable)
	dispatchErr = context.DeadlineExceeded
	request(http.MethodGet, publicPath, "", nil, http.StatusGatewayTimeout)
	dispatchErr = errors.New("private runtime failure details")
	failedExecution := request(http.MethodGet, publicPath, "", nil, http.StatusInternalServerError)
	if strings.Contains(failedExecution.Body.String(), "private runtime") {
		t.Fatal("private runtime error leaked")
	}
	request(http.MethodGet, publicPath, "Bearer forged", nil, http.StatusNotFound)
	request(http.MethodGet, "/"+projectID+"/functions/missing", "", nil, http.StatusNotFound)
	dispatchErr = nil
	scheduled := request(http.MethodGet, publicPath, "", nil, http.StatusOK)
	if !strings.Contains(scheduled.Body.String(), `"scheduled":true`) {
		t.Fatal(scheduled.Body.String())
	}
	command("unsupported", nil, 400)
	if err := projects.SetActive(projectID, false); err != nil {
		t.Fatal(err)
	}
	request(http.MethodGet, "/"+projectID+"/admin/functions", auth, nil, 403)
	command("functions.create", map[string]any{"source": source}, 403)
	command("functions.publish", nil, 403)
	command("functions.test", map[string]any{"source": source}, 403)
}
