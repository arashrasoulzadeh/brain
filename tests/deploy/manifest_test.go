package deploy_test

import (
	"errors"
	"testing"

	"brain/internal/deploy"
)

func TestValidateReleaseManifestRequiresAllImages(t *testing.T) {
	manifest := deploy.ReleaseManifest{
		CommitSHA: "0123456789abcdef0123456789abcdef01234567",
		Backend:   "ghcr.io/example/backend@sha256:backend",
		Compiler:  "ghcr.io/example/compiler@sha256:compiler",
		Runner:    "ghcr.io/example/runner@sha256:runner",
	}
	if err := deploy.ValidateReleaseManifest(manifest); err != nil {
		t.Fatal(err)
	}
	manifest.Runner = ""
	if !errors.Is(deploy.ValidateReleaseManifest(manifest), deploy.ErrInvalidReleaseManifest) {
		t.Fatal("incomplete release manifest accepted")
	}
}
