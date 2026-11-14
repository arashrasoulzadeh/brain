package deploy_test

import (
	"errors"
	"os"
	"testing"

	"brain/internal/deploy"
)

func TestDeployedVersionStoreWritesValidatedSHA(t *testing.T) {
	path := t.TempDir() + "/deployed.sha"
	sha := "0123456789abcdef0123456789abcdef01234567"
	if err := deploy.RecordDeployedVersion(path, sha); err != nil {
		t.Fatal(err)
	}
	read, err := deploy.ReadDeployedVersion(path)
	if err != nil {
		t.Fatal(err)
	}
	if read != sha {
		t.Fatalf("version = %q, want %q", read, sha)
	}
	if _, err := os.Stat(path + ".tmp"); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("temporary deployment version file was left behind")
	}
	if err := deploy.RecordDeployedVersion(path, "invalid"); !errors.Is(err, deploy.ErrInvalidReleaseSHA) {
		t.Fatal("invalid release SHA was recorded")
	}
}
