package plugin_test

import (
	"errors"
	"path/filepath"
	"sync"
	"testing"

	"brain/internal/execution"
	"brain/internal/plugin"
)

func artifactRegistration(t *testing.T, project, version, source string) execution.Registration {
	t.Helper()
	bundle, err := execution.BuildBundle([]execution.BundleFile{{Path: "index.ts", Data: []byte(source)}}, 4096)
	if err != nil {
		t.Fatal(err)
	}
	return execution.Registration{
		Key:         execution.ArtifactKey{ProjectID: project, ID: "files", Version: version},
		Runtime:     execution.Runtime{Kind: "typescript", Profile: "sandboxed"},
		Entrypoints: []string{"files.list", "files.metadata"},
		Bundle:      bundle,
	}
}

func openPinnedCatalog(t *testing.T, path string) (*plugin.SQLiteStorage, *execution.Catalog) {
	t.Helper()
	storage, err := plugin.OpenSQLiteStorage(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = storage.Close() })
	catalog, err := execution.NewCatalog(16, 1<<20)
	if err != nil {
		t.Fatal(err)
	}
	if err := storage.BindArtifactPins(catalog); err != nil {
		t.Fatal(err)
	}
	return storage, catalog
}

func TestArtifactPinsSurviveRestart(t *testing.T) {
	path := filepath.Join(t.TempDir(), "brain.sqlite")
	storage, catalog := openPinnedCatalog(t, path)
	reg := artifactRegistration(t, "", "1.0.0", "export {}")
	if err := catalog.Register(reg); err != nil {
		t.Fatal(err)
	}
	if err := storage.Close(); err != nil {
		t.Fatal(err)
	}
	_, restored := openPinnedCatalog(t, path)
	changed := artifactRegistration(t, "", "1.0.0", "export const changed = true")
	if err := restored.Register(changed); !errors.Is(err, execution.ErrArtifactConflict) {
		t.Fatalf("changed source accepted after restart: %v", err)
	}
	if _, _, err := restored.Resolve(reg.Key, "files.list"); !errors.Is(err, execution.ErrArtifactNotFound) {
		t.Fatalf("rejected registration was exposed: %v", err)
	}
	changed = reg
	changed.Runtime.Profile = "other"
	if err := restored.Register(changed); !errors.Is(err, execution.ErrArtifactConflict) {
		t.Fatalf("changed runtime accepted: %v", err)
	}
	changed = reg
	changed.Entrypoints = []string{"files.list"}
	if err := restored.Register(changed); !errors.Is(err, execution.ErrArtifactConflict) {
		t.Fatalf("changed entrypoints accepted: %v", err)
	}
	reg.Entrypoints = []string{"files.metadata", "files.list"}
	if err := restored.Register(reg); err != nil {
		t.Fatalf("identical artifact rejected: %v", err)
	}
	if err := restored.Register(artifactRegistration(t, "", "2.0.0", "export const changed = true")); err != nil {
		t.Fatal(err)
	}
	if err := restored.Register(artifactRegistration(t, "project-a", "1.0.0", "export const changed = true")); err != nil {
		t.Fatal(err)
	}
}

func TestArtifactPinsConcurrentConflictAndStorageFailure(t *testing.T) {
	path := filepath.Join(t.TempDir(), "brain.sqlite")
	storage, first := openPinnedCatalog(t, path)
	_, second := openPinnedCatalog(t, path)
	regA := artifactRegistration(t, "", "1.0.0", "export const a = 1")
	regB := artifactRegistration(t, "", "1.0.0", "export const a = 2")
	var wg sync.WaitGroup
	results := make(chan error, 2)
	wg.Add(2)
	go func() { defer wg.Done(); results <- first.Register(regA) }()
	go func() { defer wg.Done(); results <- second.Register(regB) }()
	wg.Wait()
	close(results)
	success, conflict := 0, 0
	for err := range results {
		switch {
		case err == nil:
			success++
		case errors.Is(err, execution.ErrArtifactConflict):
			conflict++
		default:
			t.Fatal(err)
		}
	}
	if success != 1 || conflict != 1 {
		t.Fatalf("expected one durable winner: %d success, %d conflict", success, conflict)
	}
	if err := storage.Close(); err != nil {
		t.Fatal(err)
	}
	newReg := artifactRegistration(t, "", "2.0.0", "export {}")
	if err := first.Register(newReg); err == nil {
		t.Fatal("registration succeeded with unavailable pin storage")
	}
	if _, _, err := first.Resolve(newReg.Key, "files.list"); !errors.Is(err, execution.ErrArtifactNotFound) {
		t.Fatalf("unpersisted artifact exposed: %v", err)
	}
}

func TestArtifactPinBindingMustPrecedeRegistration(t *testing.T) {
	storage, catalog := openPinnedCatalog(t, filepath.Join(t.TempDir(), "brain.sqlite"))
	if err := storage.BindArtifactPins(catalog); err == nil {
		t.Fatal("replaced existing pin store")
	}
	unbound, err := execution.NewCatalog(4, 4096)
	if err != nil {
		t.Fatal(err)
	}
	if err := unbound.BindPinStore(nil); err == nil {
		t.Fatal("accepted nil pin store")
	}
	if err := unbound.Register(artifactRegistration(t, "", "1.0.0", "export {}")); err != nil {
		t.Fatal(err)
	}
	if err := storage.BindArtifactPins(unbound); err == nil {
		t.Fatal("bound a store after unpersisted registrations")
	}
}
