package project_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"brain/internal/project"
)

func TestSQLiteProjectStorageHealthAndLegacyImport(t *testing.T) {
	dir := t.TempDir()
	legacy := filepath.Join(dir, "projects.json")
	if err := os.WriteFile(legacy, []byte(`[{"ID":"z","Active":false},{"ID":"a","Active":true},{"ID":"bad id","Active":true}]`), 0600); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "projects.sqlite")
	store, err := project.NewSQLiteStore(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.CheckHealth(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got := store.List(); len(got) != 2 || got[0].ID != "a" || got[1].ID != "z" {
		t.Fatalf("legacy projects = %+v", got)
	}
	if _, err := store.Resolve("z"); !errors.Is(err, project.ErrInactive) {
		t.Fatalf("inactive project: %v", err)
	}
	if _, err := store.Resolve("missing"); !errors.Is(err, project.ErrNotFound) {
		t.Fatalf("missing project: %v", err)
	}
	if err := store.SetActive("missing", true); !errors.Is(err, project.ErrNotFound) {
		t.Fatalf("missing update: %v", err)
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
	if err := store.CheckHealth(context.Background()); err == nil {
		t.Fatal("closed store reported healthy")
	}
	if _, err := store.Resolve("a"); err == nil {
		t.Fatal("closed store resolved a project")
	}
	if err := store.SetActive("a", false); err == nil {
		t.Fatal("closed store changed a project")
	}
	if got := store.List(); got != nil {
		t.Fatalf("closed store returned %+v", got)
	}
	if err := store.Create("new"); err == nil {
		t.Fatal("closed store created a project")
	}
	reopened, err := project.NewSQLiteStore(path)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	if err := reopened.Create("a"); !errors.Is(err, project.ErrDuplicate) {
		t.Fatalf("duplicate persisted project: %v", err)
	}
	if got, err := reopened.Resolve("a"); err != nil || !got.Active {
		t.Fatalf("reopened project = %+v, %v", got, err)
	}
}

func TestProjectStorageFailureBoundaries(t *testing.T) {
	ctx := context.Background()
	for _, store := range []*project.Store{nil, project.NewStore()} {
		if err := store.CheckHealth(ctx); err == nil {
			t.Fatal("non-durable store reported healthy")
		}
		if err := store.Close(); err != nil {
			t.Fatal(err)
		}
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "projects.json")
	if err := os.WriteFile(path, []byte("not json"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := project.NewPersistentStore(path); err == nil {
		t.Fatal("invalid persistence accepted")
	}
	if _, err := project.NewPersistentStore(dir); err == nil {
		t.Fatal("directory accepted as persistent file")
	}
	if _, err := project.NewSQLiteStore(filepath.Join(dir, "missing", "db")); err == nil {
		t.Fatal("unavailable SQLite path accepted")
	}
	store := project.NewStore(project.Project{ID: "bad id"}, project.Project{ID: "z"}, project.Project{ID: "a"})
	if got := store.List(); len(got) != 2 || got[0].ID != "a" {
		t.Fatalf("validated sorted list: %+v", got)
	}
	if err := store.SetActive("bad id", true); !errors.Is(err, project.ErrInvalidID) {
		t.Fatalf("invalid update: %v", err)
	}
	if err := store.Create("a"); !errors.Is(err, project.ErrDuplicate) {
		t.Fatalf("duplicate: %v", err)
	}
}
