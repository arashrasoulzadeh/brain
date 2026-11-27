package project_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"brain/internal/project"
)

func TestProjectBoundaryRejectsMismatchedTrustedCaller(t *testing.T) {
	store := project.NewStore(project.Project{ID: "alpha", Active: true})
	handler := project.Resolver{Store: store}.Middleware(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("mismatched project reached handler")
	}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Project-ID", "alpha")
	req = req.WithContext(project.WithCaller(req.Context(), "beta", "actor"))
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusForbidden {
		t.Fatalf("mismatched caller status: %d", res.Code)
	}
	if _, ok := project.FromContext(context.Background()); ok {
		t.Fatal("empty context contained a project")
	}
}

func TestProjectCallerRequiresActorAndNonemptyLease(t *testing.T) {
	for name, fn := range map[string]func(){
		"context": func() { project.WithCaller(context.Background(), "alpha", "") },
		"lease":   func() { project.NewCallerLease("alpha", "") },
	} {
		t.Run(name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatal("empty actor was accepted")
				}
			}()
			fn()
		})
	}
	var empty project.CallerLease
	if _, ok := empty.Consume(); ok {
		t.Fatal("empty lease consumed")
	}
}

func TestProjectStoreRejectsDuplicatesAndMissingUpdates(t *testing.T) {
	store := project.NewStore(project.Project{ID: "bad/id"}, project.Project{ID: "z", Active: true}, project.Project{ID: "a", Active: true})
	if got := store.List(); len(got) != 2 || got[0].ID != "a" || got[1].ID != "z" {
		t.Fatalf("invalid or unsorted project list: %#v", got)
	}
	if err := store.Create("a"); !errors.Is(err, project.ErrDuplicate) {
		t.Fatalf("duplicate create: %v", err)
	}
	if err := store.SetActive("missing", true); !errors.Is(err, project.ErrNotFound) {
		t.Fatalf("missing update: %v", err)
	}
	if err := store.SetActive("bad/id", true); !errors.Is(err, project.ErrInvalidID) {
		t.Fatalf("malformed update: %v", err)
	}
}

func TestProjectPersistentStoreHandlesCorruptionAndLegacyRecords(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "projects.json")
	if err := os.WriteFile(path, []byte("not json"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := project.NewPersistentStore(path); err == nil {
		t.Fatal("corrupt store accepted")
	}
	if _, err := project.NewPersistentStore(dir); err == nil {
		t.Fatal("directory accepted as store file")
	}
	if err := os.WriteFile(path, []byte(`[{"ID":"z","Active":false},{"ID":"a","Active":true},{"ID":"bad/id","Active":true}]`), 0o600); err != nil {
		t.Fatal(err)
	}
	store, err := project.NewPersistentStore(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := store.List(); len(got) != 2 || got[0].ID != "a" {
		t.Fatalf("unexpected persisted projects: %#v", got)
	}
	if err := store.Create("middle"); err != nil {
		t.Fatal(err)
	}
	db, err := project.NewSQLiteStore(filepath.Join(dir, "projects.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := db.CheckHealth(context.Background()); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Resolve("a"); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Resolve("z"); !errors.Is(err, project.ErrInactive) {
		t.Fatalf("inactive SQLite project: %v", err)
	}
	if _, err := db.Resolve("missing"); !errors.Is(err, project.ErrNotFound) {
		t.Fatalf("missing SQLite project: %v", err)
	}
	if err := db.SetActive("missing", true); !errors.Is(err, project.ErrNotFound) {
		t.Fatalf("missing SQLite update: %v", err)
	}
	if got := db.List(); len(got) != 3 || got[0].ID != "a" {
		t.Fatalf("unexpected SQLite projects: %#v", got)
	}
	if err := store.CheckHealth(context.Background()); err == nil {
		t.Fatal("file store claimed durable SQLite health")
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
}
