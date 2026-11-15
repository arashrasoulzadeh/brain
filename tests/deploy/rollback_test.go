package deploy_test

import (
	"errors"
	"strings"
	"testing"

	"brain/internal/deploy"
)

func TestSelectRollbackManifestRequiresCompatibleSchema(t *testing.T) {
	manifest := deploy.ReleaseManifest{
		CommitSHA:                  "0123456789abcdef0123456789abcdef01234567",
		Backend:                    "backend@sha256:" + strings.Repeat("b", 64),
		Compiler:                   "compiler@sha256:" + strings.Repeat("c", 64),
		Runner:                     "runner@sha256:" + strings.Repeat("d", 64),
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
