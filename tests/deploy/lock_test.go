package deploy_test

import (
	"errors"
	"path/filepath"
	"testing"

	"brain/internal/deploy"
)

func TestDeploymentLockSerializesReleases(t *testing.T) {
	path := filepath.Join(t.TempDir(), "deploy.lock")
	first := deploy.NewLock(path)
	second := deploy.NewLock(path)
	if err := first.Acquire(); err != nil {
		t.Fatal(err)
	}
	if !errors.Is(second.Acquire(), deploy.ErrAlreadyLocked) {
		t.Fatal("concurrent deployment was accepted")
	}
	if err := first.Release(); err != nil {
		t.Fatal(err)
	}
	if err := second.Acquire(); err != nil {
		t.Fatal(err)
	}
	if err := second.Release(); err != nil {
		t.Fatal(err)
	}
}
