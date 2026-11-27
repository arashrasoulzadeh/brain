package plugin_test

import (
	"errors"
	"path/filepath"
	"testing"

	"brain/internal/plugin"
)

func TestAdminInventoryRejectsUnresolvedOrInactiveProjects(t *testing.T) {
	storage, err := plugin.OpenSQLiteStorage(filepath.Join(t.TempDir(), "inventory.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()
	registry := plugin.NewRegistry(1)
	if err := registry.BindSQLiteStorage(storage); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name  string
		model plugin.ProjectSnapshot
		err   error
	}{
		{name: "unknown", err: errors.New("unknown project")},
		{name: "inactive", model: plugin.ProjectSnapshot{ID: "one"}},
		{name: "foreign", model: plugin.ProjectSnapshot{ID: "two", Active: true}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			registry.BindProjectResolver(func(string) (plugin.ProjectSnapshot, error) {
				return tc.model, tc.err
			})
			if _, err := registry.AdminInventory("one", "files", "1", "files", "", 10); !errors.Is(err, plugin.ErrProjectInvalid) {
				t.Fatalf("invalid project not rejected: %v", err)
			}
		})
	}
	registry.BindProjectResolver(func(id string) (plugin.ProjectSnapshot, error) {
		return plugin.ProjectSnapshot{ID: id, Active: true}, nil
	})
	if _, err := registry.AdminInventory("one", "files", "1", "files", "", 10); !errors.Is(err, plugin.ErrNotRegistered) {
		t.Fatalf("active project should proceed to plugin validation: %v", err)
	}
}
