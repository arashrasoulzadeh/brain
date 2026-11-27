package app_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"brain/internal/admin"
	"brain/internal/app"
	"brain/internal/operator"
	"brain/internal/plugin"
	"brain/internal/project"
)

type projectHTTPFixtureWorker struct{ ready atomic.Bool }

func (w *projectHTTPFixtureWorker) Start() error { w.ready.Store(true); return nil }
func (w *projectHTTPFixtureWorker) Stop() error  { w.ready.Store(false); return nil }
func (w *projectHTTPFixtureWorker) Ready() bool  { return w.ready.Load() }
func (w *projectHTTPFixtureWorker) Handle(req plugin.WorkerRequest) plugin.WorkerResponse {
	data, err := json.Marshal(map[string]any{"projectID": req.ProjectID, "operation": req.Operation, "input": json.RawMessage(req.Payload)})
	if err != nil {
		return plugin.WorkerResponse{Protocol: req.Protocol, ErrorCode: "BAD_PAYLOAD"}
	}
	return plugin.WorkerResponse{Protocol: req.Protocol, Success: true, Payload: data}
}

func projectHTTPFixture(t *testing.T) (http.Handler, *plugin.Registry, string) {
	t.Helper()
	operators := &operator.Store{}
	if _, err := operators.Bootstrap("admin", "secret", func(value string) string { return value }); err != nil {
		t.Fatal(err)
	}
	if err := operators.Authenticate("admin", "secret", func(value string) string { return value }); err != nil {
		t.Fatal(err)
	}
	token, err := operators.IssueSession(func() string { return "http-project-session" })
	if err != nil {
		t.Fatal(err)
	}
	projects := project.NewStore(project.Project{ID: "immutable-alpha", Slug: "alpha", Active: true},
		project.Project{ID: "immutable-beta", Slug: "beta", Active: true},
		project.Project{ID: "inactive", Active: false})
	registry := plugin.NewRegistry(1)
	m := plugin.Manifest{ID: "rooms", Version: "1.0.0", HostAPI: 1, Mount: "/rooms", Capabilities: []string{"rooms.read"},
		Routes: []plugin.Route{
			{Method: "GET", Path: "/", Auth: "operator", Operation: "rooms.list", Function: "listRooms", Capability: "rooms.read"},
			{Method: "GET", Path: "/{id}", Auth: "operator", Operation: "rooms.get", Function: "getRoom", Capability: "rooms.read"},
		}}
	if err := registry.Register(m); err != nil {
		t.Fatal(err)
	}
	if err := registry.Install("immutable-alpha", "rooms", "1.0.0"); err != nil {
		t.Fatal(err)
	}
	if err := registry.Grant("immutable-alpha", "rooms", "1.0.0", "rooms.read"); err != nil {
		t.Fatal(err)
	}
	if err := registry.SetReadiness("rooms", "1.0.0", true); err != nil {
		t.Fatal(err)
	}
	if err := registry.Enable("immutable-alpha", "rooms", "1.0.0"); err != nil {
		t.Fatal(err)
	}
	w := &projectHTTPFixtureWorker{}
	w.ready.Store(true)
	if err := registry.BindWorker("immutable-alpha", "rooms", "1.0.0", w); err != nil {
		t.Fatal(err)
	}
	service := admin.NewLifecycleServiceWithProjects(operators, registry, projects)
	return app.NewServerWithOperatorAdmin(service, operators, func(value string) string { return value }).Handler(), registry, token
}

func TestProjectHTTPRoutesResolveInstallationsAndNeverTrustClientProject(t *testing.T) {
	handler, _, token := projectHTTPFixture(t)
	testCases := []struct {
		method, path, bearer, body string
		want                       int
		contains                   string
	}{
		{"GET", "/alpha/rooms", token, "", 200, "immutable-alpha"},
		{"GET", "/alpha/rooms/room-1", token, "", 200, "room-1"},
		{"GET", "/alpha/admin/routes", token, "", 200, `"/alpha/rooms/{id}"`},
		{"GET", "/immutable-alpha/admin/routes", token, "", 200, `"/alpha/rooms/{id}"`},
		{"GET", "/immutable-alpha/rooms/room-1", token, "", 200, "immutable-alpha"},
		{"GET", "/beta/admin/routes", token, "", 200, "[]"},
		{"GET", "/inactive/admin/routes", token, "", 403, ""},
		{"GET", "/_system/rooms", token, "", 404, ""},
		{"GET", "/alpha/admin/routes", "", "", 403, ""},
		{"GET", "/beta/rooms", token, "", 404, ""},
		{"GET", "/inactive/rooms", token, "", 403, ""},
		{"GET", "/unknown/rooms", token, "", 404, ""},
		{"GET", "/rooms", token, "", 404, ""},
		{"GET", "/alpha/rooms", "", "", 403, ""},
		{"POST", "/alpha/rooms", token, "{}", 404, ""},
		{"GET", "/alpha/rooms?project_id=immutable-beta", token, "", 400, ""},
		{"GET", "/alpha/rooms?project_id=alpha", token, "", 200, "immutable-alpha"},
		{"GET", "/alpha/rooms?project_id=alpha&project_id=alpha", token, "", 400, ""},
		{"GET", "/alpha/rooms", token, "", 200, "rooms.list"},
	}
	for _, tc := range testCases {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
			if tc.bearer != "" {
				req.Header.Set("Authorization", "Bearer "+tc.bearer)
			}
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)
			if rec.Code != tc.want {
				t.Fatalf("status %d, want %d: %s", rec.Code, tc.want, rec.Body.String())
			}
			if tc.contains != "" && !strings.Contains(rec.Body.String(), tc.contains) {
				t.Fatalf("response missing %q: %s", tc.contains, rec.Body.String())
			}
		})
	}
}

func TestProjectHTTPRequestsDuringDisableCannotEscapeLifecycle(t *testing.T) {
	handler, registry, token := projectHTTPFixture(t)
	const n = 40
	var wg sync.WaitGroup
	start := make(chan struct{})
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			req := httptest.NewRequest(http.MethodGet, "/alpha/rooms", nil)
			req.Header.Set("Authorization", "Bearer "+token)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)
			if rec.Code != http.StatusOK && rec.Code != http.StatusNotFound && rec.Code != http.StatusForbidden {
				t.Errorf("concurrent response %d: %s", rec.Code, rec.Body.String())
			}
		}()
	}
	close(start)
	preview, err := registry.PreviewDisable("immutable-alpha", "rooms", "1.0.0")
	if err != nil {
		t.Fatal(err)
	}
	if err := registry.ApplyDisablePreview(preview); err != nil {
		t.Fatal(err)
	}
	wg.Wait()
	req := httptest.NewRequest(http.MethodGet, "/alpha/rooms", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("disabled route returned %d, want 404", rec.Code)
	}
}
