package deploy_test

import (
	"errors"
	"testing"

	"brain/internal/deploy"
)

func TestRestorePairRequiresDatabaseAndObjectDataForRelease(t *testing.T) {
	pair := deploy.RestorePair{
		ReleaseSHA:             "0123456789abcdef0123456789abcdef01234567",
		DatabaseRestored:       true,
		ObjectManifestRestored: true,
	}
	if err := pair.Validate(); err != nil {
		t.Fatal(err)
	}
	pair.ObjectManifestRestored = false
	if !errors.Is(pair.Validate(), deploy.ErrIncompleteRestorePair) {
		t.Fatal("restore accepted without object data")
	}
}

func TestRestorePairRejectsUnboundRelease(t *testing.T) {
	pair := deploy.RestorePair{DatabaseRestored: true, ObjectManifestRestored: true}
	if !errors.Is(pair.Validate(), deploy.ErrInvalidRestorePair) {
		t.Fatal("restore accepted without release binding")
	}
}
