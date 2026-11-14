package project_test

import (
	"errors"
	"fmt"
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

func TestResolveRejectsInactiveProject(t *testing.T) {
	store := project.NewStore(project.Project{ID: "prod", Active: false})
	_, err := store.Resolve("prod")
	if !errors.Is(err, project.ErrInactive) {
		t.Fatalf("Resolve(prod) = %v, want ErrInactive", err)
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
