package deploy_test

import (
	"errors"
	"strings"
	"testing"

	"brain/internal/deploy"
)

func TestValidateReleaseManifestRequiresAllImages(t *testing.T) {
	manifest := deploy.ReleaseManifest{
		CommitSHA:                  "0123456789abcdef0123456789abcdef01234567",
		Backend:                    "ghcr.io/example/backend@sha256:backend",
		Compiler:                   "ghcr.io/example/compiler@sha256:compiler",
		Runner:                     "ghcr.io/example/runner@sha256:runner",
		PackagedIsolationVerified:  true,
		PackagedRevocationVerified: true,
	}
	if err := deploy.ValidateReleaseManifest(manifest); err != nil {
		t.Fatal(err)
	}
	manifest.Runner = ""
	if !errors.Is(deploy.ValidateReleaseManifest(manifest), deploy.ErrInvalidReleaseManifest) {
		t.Fatal("incomplete release manifest accepted")
	}
}

func TestValidateReleaseManifestRequiresPackagedIsolationVerification(t *testing.T) {
	manifest := deploy.ReleaseManifest{CommitSHA: strings.Repeat("a", 40), Backend: "backend", Compiler: "compiler", Runner: "runner"}
	if !errors.Is(deploy.ValidateReleaseManifest(manifest), deploy.ErrInvalidReleaseManifest) {
		t.Fatal("release manifest accepted without packaged isolation verification")
	}
	manifest.PackagedIsolationVerified = true
	if err := deploy.ValidateReleaseManifest(manifest); !errors.Is(err, deploy.ErrInvalidReleaseManifest) {
		t.Fatal("release manifest accepted without packaged revocation verification")
	}
	manifest.PackagedRevocationVerified = true
	if err := deploy.ValidateReleaseManifest(manifest); err != nil {
		t.Fatal(err)
	}
}
