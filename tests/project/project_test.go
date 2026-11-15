package project_test

import (
	"errors"
	"fmt"
	"path/filepath"
	"sync"
	"testing"

	"brain/internal/project"
)

func TestResolveActiveProject(t *testing.T) {
	store := project.NewStore(project.Project{ID: "dev", Active: true})
	got, err := store.Resolve("dev")
	if err != nil || got.ID != "dev" {
		t.Fatalf("Resolve(dev) = %#v, %v", got, err)
	}
}

func TestResolveRejectsUnknownProject(t *testing.T) {
	_, err := project.NewStore().Resolve("missing")
	if !errors.Is(err, project.ErrNotFound) {
		t.Fatalf("Resolve(missing) = %v, want ErrNotFound", err)
	}
}

func TestStoreRejectsMalformedProjectIDs(t *testing.T) {
	store := project.NewStore()
	if err := store.Create("dev:other"); !errors.Is(err, project.ErrInvalidID) {
		t.Fatalf("Create(malformed) = %v, want ErrInvalidID", err)
	}
	if _, err := store.Resolve("dev:other"); !errors.Is(err, project.ErrInvalidID) {
		t.Fatalf("Resolve(malformed) = %v, want ErrInvalidID", err)
	}
}

func TestResolveRejectsInactiveProject(t *testing.T) {
	store := project.NewStore(project.Project{ID: "prod", Active: false})
	_, err := store.Resolve("prod")
	if !errors.Is(err, project.ErrInactive) {
		t.Fatalf("Resolve(prod) = %v, want ErrInactive", err)
	}
}

func TestSetActiveChangesProjectStatus(t *testing.T) {
	store := project.NewStore(project.Project{ID: "dev", Active: true})
	if err := store.SetActive("dev", false); err != nil {
		t.Fatal(err)
	}
	if got := store.List()[0].Active; got {
		t.Fatal("project remained active")
	}
	if err := store.SetActive("dev", true); err != nil {
		t.Fatal(err)
	}
	if got := store.List()[0].Active; !got {
		t.Fatal("project remained inactive")
	}
}

func TestPersistentStoreSurvivesReopen(t *testing.T) {
	path := filepath.Join(t.TempDir(), "projects.json")
	first, err := project.NewPersistentStore(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := first.Create("dev"); err != nil {
		t.Fatal(err)
	}
	if err := first.SetActive("dev", false); err != nil {
		t.Fatal(err)
	}
	second, err := project.NewPersistentStore(path)
	if err != nil {
		t.Fatal(err)
	}
	projects := second.List()
	if len(projects) != 1 || projects[0].ID != "dev" || projects[0].Active {
		t.Fatalf("reopened projects = %#v", projects)
	}
}

func TestSQLiteStoreSurvivesReopen(t *testing.T) {
	path := filepath.Join(t.TempDir(), "projects.sqlite")
	first, err := project.NewSQLiteStore(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := first.Create("sqlite-dev"); err != nil {
		t.Fatal(err)
	}
	if err := first.SetActive("sqlite-dev", false); err != nil {
		t.Fatal(err)
	}
	second, err := project.NewSQLiteStore(path)
	if err != nil {
		t.Fatal(err)
	}
	projects := second.List()
	if len(projects) != 1 || projects[0].ID != "sqlite-dev" || projects[0].Active {
		t.Fatalf("reopened sqlite projects = %#v", projects)
	}
}

func TestStoreSupportsConcurrentProjectAccess(t *testing.T) {
	store := project.NewStore()
	var group sync.WaitGroup
	for index := 0; index < 32; index++ {
		index := index
		group.Add(1)
		go func() {
			defer group.Done()
			id := fmt.Sprintf("project-%d", index)
			if err := store.Create(id); err != nil {
				t.Errorf("Create(%q) = %v", id, err)
				return
			}
			if _, err := store.Resolve(id); err != nil {
				t.Errorf("Resolve(%q) = %v", id, err)
			}
		}()
	}
	group.Wait()
}
