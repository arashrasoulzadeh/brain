package plugin

import (
	"path/filepath"
	"testing"
	"time"
)

func TestLifecycleWriteWaitsForObservationTransaction(t *testing.T) {
	path := filepath.Join(t.TempDir(), "shared.db")
	first, err := OpenSQLiteStorage(path)
	if err != nil {
		t.Fatal(err)
	}
	defer first.Close()
	second, err := OpenSQLiteStorage(path)
	if err != nil {
		t.Fatal(err)
	}
	defer second.Close()
	tx, err := first.db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	if err := putSQLiteState(tx, "coverage", "test", "lock", map[string]any{"active": true}); err != nil {
		t.Fatal(err)
	}
	started := make(chan struct{})
	done := make(chan error, 1)
	go func() {
		close(started)
		done <- second.SaveRegistryState(map[string]Installation{}, map[string]bool{})
	}()
	<-started
	select {
	case err := <-done:
		t.Fatalf("write returned before competing transaction committed: %v", err)
	case <-time.After(50 * time.Millisecond):
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(6 * time.Second):
		t.Fatal("write did not complete after lock release")
	}
	if _, _, err := second.LoadRegistryState(); err != nil {
		t.Fatal(err)
	}
}
