package project_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"brain/internal/project"
)

func TestURLNamespaceResolvesHostProjectAndRejectsClientConflicts(t *testing.T) {
	store := project.NewStore(project.Project{ID: "internal-alpha", Slug: "alpha", Active: true}, project.Project{ID: "beta", Active: false})
	handler := project.Resolver{Store: store}.URLMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		identity, ok := project.FromContext(r.Context())
		if !ok || identity.ID != "internal-alpha" {
			t.Fatalf("project context %#v, %v", identity, ok)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	cases := []struct {
		path, header string
		want         int
	}{
		{"/alpha/rooms", "", 204},
		{"/alpha/rooms?project_id=beta", "", 400},
		{"/alpha/rooms", "beta", 400},
		{"/alpha/rooms?project_id=internal-alpha", "", 204},
		{"/missing/rooms", "", 404},
		{"/beta/rooms", "", 403},
		{"/livez", "", 404},
		{"/alpha/%2e%2e/rooms", "", 404},
		{"/alpha/%2frooms", "", 404},
	}
	for _, tc := range cases {
		req := httptest.NewRequest("GET", tc.path, nil)
		if tc.header != "" {
			req.Header.Set("X-Project-ID", tc.header)
		}
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != tc.want {
			t.Errorf("%s: got %d want %d", tc.path, rec.Code, tc.want)
		}
	}
}

func TestProjectSlugUniqueness(t *testing.T) {
	store := project.NewStore()
	if err := store.CreateWithDetails("immutable-one", "Alpha", "alpha", map[string]any{"region": "eu"}); err != nil {
		t.Fatal(err)
	}
	if err := store.CreateWithDetails("immutable-two", "Second", "alpha", nil); err != project.ErrDuplicate {
		t.Fatalf("duplicate slug: %v", err)
	}
	got, err := store.ResolveNamespace("alpha")
	if err != nil || got.ID != "immutable-one" || got.Name != "Alpha" {
		t.Fatalf("resolved: %#v %v", got, err)
	}
	if _, err := store.ResolveNamespace("_system"); err == nil {
		t.Fatal("reserved namespace accepted")
	}
}

func TestProjectNamespaceConfigureAndResolveAcrossStores(t *testing.T) {
	for _, kind := range []string{"memory", "sqlite"} {
		t.Run(kind, func(t *testing.T) {
			store := project.NewStore()
			path := filepath.Join(t.TempDir(), "projects.db")
			if kind == "sqlite" {
				var err error
				store, err = project.NewSQLiteStore(path)
				if err != nil {
					t.Fatal(err)
				}
				defer store.Close()
			}
			if err := store.CreateWithDetails("id-one", "One", "one", map[string]any{"region": "eu"}); err != nil {
				t.Fatal(err)
			}
			if err := store.CreateWithDetails("id-two", "Two", "two", nil); err != nil {
				t.Fatal(err)
			}
			for _, collision := range []struct{ id, slug string }{{"one", "three"}, {"id-three", "id-one"}, {"id-three", "two"}} {
				if err := store.CreateWithDetails(collision.id, "", collision.slug, nil); !errors.Is(err, project.ErrDuplicate) {
					t.Fatalf("create collision %#v: %v", collision, err)
				}
			}
			for _, slug := range []string{"two", "id-two"} {
				if err := store.Configure("id-one", "", slug, nil); !errors.Is(err, project.ErrDuplicate) {
					t.Fatalf("configure collision %q: %v", slug, err)
				}
			}
			for _, slug := range []string{"_system", "a/b", strings.Repeat("x", 64)} {
				if err := store.Configure("id-one", "", slug, nil); !errors.Is(err, project.ErrInvalidID) {
					t.Fatalf("invalid slug %q: %v", slug, err)
				}
			}
			if err := store.Configure("bad/id", "", "", nil); !errors.Is(err, project.ErrInvalidID) {
				t.Fatalf("invalid id: %v", err)
			}
			if err := store.Configure("missing", "", "", nil); !errors.Is(err, project.ErrNotFound) {
				t.Fatalf("missing project: %v", err)
			}
			if err := store.Configure("id-one", " invalid ", "", nil); !errors.Is(err, project.ErrInvalidID) {
				t.Fatalf("invalid name: %v", err)
			}
			if err := store.Configure("id-one", "", "", map[string]any{"bad": make(chan int)}); err == nil {
				t.Fatal("non-JSON settings accepted")
			}
			if err := store.Configure("id-one", "Renamed", "renamed", nil); err != nil {
				t.Fatal(err)
			}
			if _, err := store.ResolveNamespace("one"); !errors.Is(err, project.ErrNotFound) {
				t.Fatalf("old namespace: %v", err)
			}
			for _, namespace := range []string{"id-one", "renamed"} {
				got, err := store.ResolveNamespace(namespace)
				if err != nil || got.ID != "id-one" || got.Name != "Renamed" || got.Slug != "renamed" || got.Settings["region"] != "eu" {
					t.Fatalf("resolved %q: %#v, %v", namespace, got, err)
				}
			}
			if err := store.SetActive("id-one", false); err != nil {
				t.Fatal(err)
			}
			if err := store.Configure("id-one", "", "", map[string]any{"region": "us"}); err != nil {
				t.Fatal(err)
			}
			if _, err := store.ResolveNamespace("renamed"); !errors.Is(err, project.ErrInactive) {
				t.Fatalf("configure reactivated project: %v", err)
			}
			if kind == "sqlite" {
				reopened, err := project.NewSQLiteStore(path)
				if err != nil {
					t.Fatal(err)
				}
				defer reopened.Close()
				store = reopened
			}
			if err := store.SetActive("id-one", true); err != nil {
				t.Fatal(err)
			}
			got, err := store.ResolveNamespace("renamed")
			if err != nil || got.Name != "Renamed" || got.Settings["region"] != "us" {
				t.Fatalf("persisted configuration: %#v, %v", got, err)
			}
		})
	}
}
