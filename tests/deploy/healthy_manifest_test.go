package deploy_test

import (
	"errors"
	"testing"

	"brain/internal/deploy"
)

func TestHealthyManifestStoreRetainsValidatedImageSet(t *testing.T) {
	path := t.TempDir() + "/healthy.json"
	manifest := deploy.ReleaseManifest{
		CommitSHA: "0123456789abcdef0123456789abcdef01234567",
		Backend:   "ghcr.io/example/backend@sha256:backend",
		Compiler:  "ghcr.io/example/compiler@sha256:compiler",
		Runner:    "ghcr.io/example/runner@sha256:runner",
	}
	if err := deploy.RecordHealthyManifest(path, manifest); err != nil {
		t.Fatal(err)
	}
	read, err := deploy.ReadHealthyManifest(path)
	if err != nil {
		t.Fatal(err)
	}
	if read != manifest {
		t.Fatalf("manifest = %#v, want %#v", read, manifest)
	}
	manifest.Runner = ""
	if err := deploy.RecordHealthyManifest(path, manifest); !errors.Is(err, deploy.ErrInvalidReleaseManifest) {
		t.Fatal("invalid healthy manifest was recorded")
	}
}
