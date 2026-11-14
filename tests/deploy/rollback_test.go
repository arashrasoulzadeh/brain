package deploy_test

import (
	"errors"
	"testing"

	"brain/internal/deploy"
)

func TestSelectRollbackManifestRequiresCompatibleSchema(t *testing.T) {
	manifest := deploy.ReleaseManifest{
		CommitSHA:                  "0123456789abcdef0123456789abcdef01234567",
		Backend:                    "backend@sha256:healthy",
		Compiler:                   "compiler@sha256:healthy",
		Runner:                     "runner@sha256:healthy",
		PackagedIsolationVerified:  true,
		PackagedRevocationVerified: true,
	}
	selected, err := deploy.SelectRollbackManifest(manifest, true)
	if err != nil || selected != manifest {
		t.Fatalf("compatible rollback = %#v, %v", selected, err)
	}
	if _, err := deploy.SelectRollbackManifest(manifest, false); !errors.Is(err, deploy.ErrIncompatibleRollbackSchema) {
		t.Fatal("incompatible rollback was accepted")
	}
}
