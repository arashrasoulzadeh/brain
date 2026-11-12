package project_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"brain/internal/project"
)

func TestMiddlewareResolvesProjectAtEntryPoint(t *testing.T) {
	store := project.NewStore(project.Project{ID: "dev", Active: true})
	handler := project.Resolver{Store: store}.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p, ok := project.FromContext(r.Context())
		if !ok || p.ID != "dev" {
			t.Fatalf("project context = %#v, %v", p, ok)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("X-Project-ID", "dev")
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusNoContent {
		t.Fatalf("status = %d", res.Code)
	}
}

func TestMiddlewareRejectsUnknownAndInactiveProjects(t *testing.T) {
	handler := project.Resolver{Store: project.NewStore(project.Project{ID: "off", Active: false})}.Middleware(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	for _, test := range []struct {
		id   string
		want int
	}{{"missing", http.StatusNotFound}, {"off", http.StatusForbidden}} {
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("X-Project-ID", test.id)
		res := httptest.NewRecorder()
		handler.ServeHTTP(res, req)
		if res.Code != test.want {
			t.Errorf("project %q status = %d, want %d", test.id, res.Code, test.want)
		}
	}
}
