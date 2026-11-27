package project_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"brain/internal/project"
)

func TestPersistentProjectWriteFailureDoesNotPublishChanges(t *testing.T) {
	for _, fault := range []string{"parent", "temporary", "rename"} {
		t.Run(fault, func(t *testing.T) {
			dir := t.TempDir()
			parent := filepath.Join(dir, "data")
			path := filepath.Join(parent, "projects.json")
			store, err := project.NewPersistentStore(path)
			if err != nil {
				t.Fatal(err)
			}
			if err := store.Create("existing"); err != nil {
				t.Fatal(err)
			}
			if err := store.Create("active"); err != nil {
				t.Fatal(err)
			}
			if err := store.SetActive("existing", false); err != nil {
				t.Fatal(err)
			}
			switch fault {
			case "parent":
				if err := os.RemoveAll(parent); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(parent, []byte("not a directory"), 0o600); err != nil {
					t.Fatal(err)
				}
			case "temporary":
				if err := os.Mkdir(path+".tmp", 0o700); err != nil {
					t.Fatal(err)
				}
			case "rename":
				if err := os.Remove(path); err != nil {
					t.Fatal(err)
				}
				if err := os.Mkdir(path, 0o700); err != nil {
					t.Fatal(err)
				}
			}
			if err := store.Create("new"); err == nil {
				t.Fatal("failed storage accepted new project")
			}
			if _, err := store.Resolve("new"); !errors.Is(err, project.ErrNotFound) {
				t.Fatalf("failed create became visible: %v", err)
			}
			if err := store.SetActive("existing", true); err == nil {
				t.Fatal("failed storage accepted status change")
			}
			if _, err := store.Resolve("existing"); !errors.Is(err, project.ErrInactive) {
				t.Fatalf("failed activation became visible: %v", err)
			}
			if err := store.SetActive("active", false); err == nil {
				t.Fatal("failed storage reported durable deactivation")
			}
			if _, err := store.Resolve("active"); !errors.Is(err, project.ErrInactive) {
				t.Fatalf("failed deactivation restored access: %v", err)
			}
		})
	}
}

func TestSQLiteProjectStoreRejectsUnavailableDatabase(t *testing.T) {
	if _, err := project.NewSQLiteStore(filepath.Join(t.TempDir(), "absent", "db")); err == nil {
		t.Fatal("unavailable database accepted")
	}
	path := filepath.Join(t.TempDir(), "projects.sqlite")
	first, err := project.NewSQLiteStore(path)
	if err != nil {
		t.Fatal(err)
	}
	defer first.Close()
	if err := first.Create("existing"); err != nil {
		t.Fatal(err)
	}
	second, err := project.NewSQLiteStore(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := second.Create("existing"); !errors.Is(err, project.ErrDuplicate) {
		t.Fatalf("persisted duplicate: %v", err)
	}
	if err := second.Close(); err != nil {
		t.Fatal(err)
	}
	if err := second.Create("new"); err == nil {
		t.Fatal("closed database accepted create")
	}
	if _, err := second.Resolve("existing"); err == nil {
		t.Fatal("closed database accepted read")
	}
	if err := second.SetActive("existing", false); err == nil {
		t.Fatal("closed database accepted status change")
	}
	if second.List() != nil || second.CheckHealth(context.Background()) == nil {
		t.Fatal("closed database exposed inventory or health")
	}
}
