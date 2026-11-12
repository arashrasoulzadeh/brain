package project_test

import (
	"errors"
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
