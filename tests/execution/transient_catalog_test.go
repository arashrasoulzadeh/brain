package execution_test

import (
	"errors"
	"testing"

	"brain/internal/execution"
)

func TestTransientCatalogScopeAndRelease(t *testing.T) {
	bundle, err := execution.BuildBundle([]execution.BundleFile{{Path: "source.mjs", Data: []byte("export default () => null")}}, 4096)
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := execution.NewCatalog(1, 4096)
	if err != nil {
		t.Fatal(err)
	}
	reg := execution.Registration{Key: execution.ArtifactKey{ProjectID: "one", ID: "invocation", Version: "1"}, Runtime: execution.Runtime{Kind: execution.UserFunctionRuntimeKind, Profile: "test"}, Entrypoints: []string{"functions.call"}, Bundle: bundle}
	release, err := catalog.RegisterTransient(reg)
	if err != nil {
		t.Fatal(err)
	}
	defer release()
	if _, err := catalog.RegisterTransient(reg); !errors.Is(err, execution.ErrArtifactConflict) {
		t.Fatalf("duplicate transient: %v", err)
	}
	if err := catalog.Register(reg); !errors.Is(err, execution.ErrArtifactConflict) {
		t.Fatalf("transient converted to durable: %v", err)
	}
	artifact, _, err := catalog.Resolve(reg.Key, "functions.call")
	if err != nil {
		t.Fatal(err)
	}
	foreign := artifact
	foreign.ProjectID = "two"
	if _, err := catalog.Fetch(foreign); !errors.Is(err, execution.ErrArtifactNotFound) {
		t.Fatalf("foreign fetch: %v", err)
	}
	release()
	release()
	if _, err := catalog.Fetch(artifact); !errors.Is(err, execution.ErrArtifactNotFound) {
		t.Fatalf("released fetch: %v", err)
	}
	second, err := catalog.RegisterTransient(reg)
	if err != nil {
		t.Fatal(err)
	}
	defer second()
	release()
	if _, err := catalog.Fetch(artifact); err != nil {
		t.Fatalf("old release removed new registration: %v", err)
	}
}
