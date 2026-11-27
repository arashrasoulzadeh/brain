package app

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"brain/internal/admin"
	"brain/internal/operator"
	"brain/internal/plugin"
	"brain/internal/project"
)

type httpBoundaryReader struct{}

func (httpBoundaryReader) Read([]byte) (int, error) { return 0, io.ErrUnexpectedEOF }

func httpBoundaryRequest(method, path, body string, identity *project.Project, token string) *http.Request {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if identity != nil {
		req = req.WithContext(project.WithProject(req.Context(), *identity))
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	return req
}

func TestProjectCommandBoundaryRejectsMalformedAndAmbiguousJSON(t *testing.T) {
	identity := project.Project{ID: "immutable", Slug: "alias", Active: true}
	for _, body := range []string{
		"",
		"[]",
		"null",
		`{"value":}`,
		`{"value":1,broken}`,
		`{"value":1,"value":2}`,
		`{"project_id":"immutable","project_id":"other"}`,
		`{"projectId":"other"}`,
		`{"project":42}`,
		`{"project":null}`,
		`{"project_id":""}`,
		`{"value":1} {"value":2}`,
		`{"value":1`,
	} {
		t.Run(body, func(t *testing.T) {
			handler := ProjectCommandHandler(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
				t.Fatal("invalid command forwarded")
			}))
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, httpBoundaryRequest("POST", "/alias/admin/actions", body, &identity, ""))
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("malformed command: %d %s", rec.Code, rec.Body.String())
			}
		})
	}
	handler := ProjectCommandHandler(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		data, err := io.ReadAll(req.Body)
		if err != nil {
			t.Fatal(err)
		}
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(data, &fields); err != nil {
			t.Fatal(err)
		}
		if string(fields["project_id"]) != `"immutable"` || string(fields["value"]) != "9007199254740993" {
			t.Fatalf("identity or numeric value changed: %s", data)
		}
		if req.ContentLength != int64(len(data)) {
			t.Fatalf("incorrect bound length: %d", req.ContentLength)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httpBoundaryRequest("POST", "/alias/admin/actions", `{"project_id":"alias","value":9007199254740993}`, &identity, ""))
	if rec.Code != http.StatusNoContent {
		t.Fatal(rec.Code)
	}
	for _, item := range []struct {
		identity *project.Project
		body     string
		broken   bool
		want     int
	}{
		{nil, "{}", false, http.StatusNotFound},
		{&project.Project{ID: "inactive"}, "{}", false, http.StatusNotFound},
		{&identity, strings.Repeat("x", 1<<20+1), false, http.StatusBadRequest},
		{&identity, "", true, http.StatusBadRequest},
		{&project.Project{ID: "immutable", Active: true}, `{"project":null}`, false, http.StatusBadRequest},
	} {
		req := httpBoundaryRequest("POST", "/alias/admin/actions", item.body, item.identity, "")
		if item.broken {
			req.Body = io.NopCloser(httpBoundaryReader{})
		}
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != item.want {
			t.Fatalf("command body boundary: %d want %d", rec.Code, item.want)
		}
	}
}

func TestProjectQueryAndRoutesRequireAuthoritativeActiveContext(t *testing.T) {
	handler := ProjectQueryHandler(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Query().Get("project_id") != "immutable" || len(req.URL.Query()["project_id"]) != 1 || req.URL.Query().Get("search") != "sample" {
			t.Fatalf("query not bound: %s", req.URL.RawQuery)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	for _, identity := range []*project.Project{nil, {ID: "inactive"}, {ID: "immutable", Active: true}} {
		req := httpBoundaryRequest("GET", "/immutable/admin/plugins?project_id=other&project_id=more&search=sample", "", identity, "")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		want := http.StatusNotFound
		if identity != nil && identity.Active {
			want = http.StatusNoContent
		}
		if rec.Code != want {
			t.Fatalf("query identity: got %d want %d", rec.Code, want)
		}
		if len(req.URL.Query()["project_id"]) != 2 {
			t.Fatal("query binding mutated original request")
		}
	}
}

type projectBoundaryWorker struct {
	request plugin.WorkerRequest
	calls   int
	handle  func(plugin.WorkerRequest) plugin.WorkerResponse
}

func (*projectBoundaryWorker) Start() error { return nil }
func (*projectBoundaryWorker) Stop() error  { return nil }
func (*projectBoundaryWorker) Ready() bool  { return true }
func (w *projectBoundaryWorker) Handle(req plugin.WorkerRequest) plugin.WorkerResponse {
	w.request = req
	w.calls++
	if w.handle != nil {
		return w.handle(req)
	}
	return plugin.WorkerResponse{Protocol: req.Protocol, Success: true, Payload: req.Payload}
}

func httpBoundaryFixture(t *testing.T) (*admin.LifecycleService, *operator.Store, *plugin.Registry, *projectBoundaryWorker, string) {
	t.Helper()
	sessions := &operator.Store{}
	digest := func(value string) string { return value }
	if _, err := sessions.Bootstrap("operator", "secret", digest); err != nil {
		t.Fatal(err)
	}
	if err := sessions.Authenticate("operator", "secret", digest); err != nil {
		t.Fatal(err)
	}
	token, err := sessions.IssueSession(func() string { return "fixture-token" })
	if err != nil {
		t.Fatal(err)
	}
	m := plugin.Manifest{ID: "boundary", Version: "1", HostAPI: 1, Mount: "/boundary", Capabilities: []string{"write"}, Routes: []plugin.Route{
		{Method: "POST", Path: "/write/{id}", Auth: "operator", Operation: "boundary.write", Function: "write", Capability: "write", ResourceField: "id"},
		{Method: "GET", Path: "/public", Auth: "public", Operation: "boundary.public", Function: "read", Capability: "write"},
		{Method: "POST", Path: "/user", Auth: "user", Operation: "boundary.user", Function: "user", Capability: "write"},
	}}
	r := plugin.NewRegistry(1)
	for _, err := range []error{r.Register(m), r.Install("immutable", m.ID, m.Version), r.Grant("immutable", m.ID, m.Version, "write"), r.SetReadiness(m.ID, m.Version, true), r.Enable("immutable", m.ID, m.Version)} {
		if err != nil {
			t.Fatal(err)
		}
	}
	w := &projectBoundaryWorker{}
	if err := r.BindWorker("immutable", m.ID, m.Version, w); err != nil {
		t.Fatal(err)
	}
	return admin.NewLifecycleService(sessions, r), sessions, r, w, token
}

func TestProjectPluginHTTPAdmissionAndResponseBoundaries(t *testing.T) {
	service, sessions, registry, worker, token := httpBoundaryFixture(t)
	handler := PluginHTTPHandler(service, sessions)
	identity := project.Project{ID: "immutable", Slug: "alias", Active: true}
	for _, item := range []struct {
		method, path, body, bearer string
		identity                   *project.Project
		broken                     bool
		want                       int
	}{
		{"GET", "/alias/boundary/public", "", "", nil, false, 404},
		{"GET", "/alias", "", "", &identity, false, 404},
		{"GET", "/alias/boundary/missing", "", "", &identity, false, 404},
		{"POST", "/alias/boundary/user", "{}", "", &identity, false, 403},
		{"POST", "/alias/boundary/user", "{}", "invalid-user-token", &identity, false, 403},
		{"POST", "/alias/boundary/write/r", "{}", "", &identity, false, 403},
		{"POST", "/alias/boundary/write/r?tag=a&tag=b", "{}", token, &identity, false, 400},
		{"POST", "/alias/boundary/write/r", strings.Repeat("x", 1<<20+1), token, &identity, false, 400},
		{"POST", "/alias/boundary/write/r", "", token, &identity, true, 400},
		{"POST", "/alias/boundary/write/r", `{"x":}`, token, &identity, false, 400},
		{"POST", "/alias/boundary/write/r?x=1", `{"x":2}`, token, &identity, false, 400},
		{"POST", "/alias/boundary/write/r", `{"id":"other"}`, token, &identity, false, 400},
		{"POST", "/alias/boundary/write/r", `{"id":42}`, token, &identity, false, 400},
		{"POST", "/alias/boundary/write/r", `{"projectId":"other"}`, token, &identity, false, 400},
	} {
		req := httpBoundaryRequest(item.method, item.path, item.body, item.identity, item.bearer)
		if item.broken {
			req.Body = io.NopCloser(httpBoundaryReader{})
		}
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != item.want {
			t.Fatalf("%s %s: got %d want %d: %s", item.method, item.path, rec.Code, item.want, rec.Body.String())
		}
	}
	if worker.calls != 0 {
		t.Fatalf("invalid requests reached worker: %d", worker.calls)
	}
	req := httpBoundaryRequest("POST", "/alias/boundary/write/r?tag=sample", `{"id":"r","projectId":"alias","project_id":"immutable","project":"alias","value":9007199254740993}`, &identity, token)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != 200 || rec.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("valid POST: %d %s", rec.Code, rec.Body.String())
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(worker.request.Payload, &fields); err != nil {
		t.Fatal(err)
	}
	if worker.request.ProjectID != "immutable" || string(fields["id"]) != `"r"` || string(fields["tag"]) != `"sample"` || string(fields["value"]) != "9007199254740993" {
		t.Fatalf("wrong worker payload: %s", worker.request.Payload)
	}
	for _, key := range []string{"project", "project_id", "projectId"} {
		if _, ok := fields[key]; ok {
			t.Fatal("caller project identity forwarded", key)
		}
	}
	for _, item := range []struct {
		response plugin.WorkerResponse
		want     int
		body     string
	}{
		{plugin.WorkerResponse{Protocol: 1, Success: false, ErrorCode: "internal-detail"}, 400, "plugin invocation rejected"},
		{plugin.WorkerResponse{Protocol: 1, Success: true, Payload: []byte("not-json")}, 502, "invalid plugin response"},
		{plugin.WorkerResponse{Protocol: 1, Success: true}, 200, "{}"},
	} {
		worker.handle = func(plugin.WorkerRequest) plugin.WorkerResponse { return item.response }
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httpBoundaryRequest("POST", "/alias/boundary/write/r", "", &identity, token))
		if rec.Code != item.want || !strings.Contains(rec.Body.String(), item.body) || strings.Contains(rec.Body.String(), "internal-detail") {
			t.Fatalf("response boundary: %d %s", rec.Code, rec.Body.String())
		}
	}
	registry.BindProjectResolver(func(string) (plugin.ProjectSnapshot, error) {
		return plugin.ProjectSnapshot{}, errors.New("project removed")
	})
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, httpBoundaryRequest("POST", "/alias/boundary/write/r", "{}", &identity, token))
	if rec.Code != 404 {
		t.Fatalf("project removed after route selection: %d %s", rec.Code, rec.Body.String())
	}
}

func TestProjectRouteInventoryUsesImmutableIDWhenSlugIsAbsent(t *testing.T) {
	service, sessions, _, _, token := httpBoundaryFixture(t)
	handler := AuthenticatedProjectRoutesHandler(service, sessions)
	for _, item := range []struct {
		identity *project.Project
		token    string
		want     int
	}{
		{nil, token, 404},
		{&project.Project{ID: "immutable", Active: true}, "", 403},
		{&project.Project{ID: "immutable", Active: true}, token, 200},
	} {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httpBoundaryRequest("GET", "/immutable/admin/routes", "", item.identity, item.token))
		if rec.Code != item.want {
			t.Fatalf("route inventory: %d %s", rec.Code, rec.Body.String())
		}
		if item.want == 200 && !strings.Contains(rec.Body.String(), "/immutable/boundary/write/{id}") {
			t.Fatalf("route missing immutable ID: %s", rec.Body.String())
		}
	}
}
