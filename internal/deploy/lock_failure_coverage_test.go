package deploy

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCoverageLockCloseFailure(t *testing.T) {
	path := filepath.Join(t.TempDir(), "deploy.lock")
	lock := NewLock(path)
	if err := lock.Acquire(); err != nil {
		t.Fatal(err)
	}
	if err := lock.file.Close(); err != nil {
		t.Fatal(err)
	}
	if err := lock.Release(); err == nil {
		t.Fatal("failed close reported as successful release")
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("lock disappeared after close failure: %v", err)
	}
}
