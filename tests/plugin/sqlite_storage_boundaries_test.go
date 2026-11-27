package plugin_test

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"testing"

	"brain/internal/plugin"
)

func TestSQLiteStorageIsolationAndFailureBoundaries(t *testing.T) {
	path := filepath.Join(t.TempDir(), "storage.sqlite")
	store, err := plugin.OpenSQLiteStorage(path)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	if err := store.CheckHealth(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := store.Append("a", "messages", map[string]any{"text": "first"}); err != nil {
		t.Fatal(err)
	}
	if err := store.Append("a", "messages", map[string]any{"text": "second"}); err != nil {
		t.Fatal(err)
	}
	if rows, err := store.Rows("a", "messages"); err != nil || len(rows) != 2 || rows[1]["text"] != "second" {
		t.Fatalf("stored rows = %+v, %v", rows, err)
	}
	if rows, err := store.Rows("b", "messages"); err != nil || len(rows) != 0 {
		t.Fatalf("foreign rows = %+v, %v", rows, err)
	}
	if err := store.Append("a", "", nil); !errors.Is(err, plugin.ErrStorageTableRequired) {
		t.Fatalf("empty table: %v", err)
	}
	if err := store.Append("a", "messages", map[string]any{"invalid": make(chan int)}); err == nil {
		t.Fatal("unserializable row accepted")
	}
	if err := store.PutState("a", "test", "key", map[string]any{"invalid": make(chan int)}); err == nil {
		t.Fatal("unserializable state accepted")
	}
	if err := store.PutFile("a", "file", "content", true); err != nil {
		t.Fatal(err)
	}
	if data, published, err := store.File("a", "file"); err != nil || !published || data != "content" {
		t.Fatalf("file = %q, %t, %v", data, published, err)
	}
	if data, published, err := store.File("b", "file"); err != nil || published || data != "" {
		t.Fatalf("foreign file = %q, %t, %v", data, published, err)
	}
	raw, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	defer raw.Close()
	if _, err := raw.Exec("INSERT INTO plugin_state VALUES ('a','test','broken','not-json')"); err != nil {
		t.Fatal(err)
	}
	if _, err := store.State("a", "test", "broken"); err == nil {
		t.Fatal("corrupt state accepted")
	}
	if _, err := raw.Exec("INSERT INTO plugin_table_rows VALUES ('a','broken','not-json')"); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Rows("a", "broken"); err == nil {
		t.Fatal("corrupt rows accepted")
	}
	if err := store.PutState("", "registry", "lifecycle", map[string]any{"installations": "wrong-type"}); err != nil {
		t.Fatal(err)
	}
	if _, _, err := store.LoadRegistryState(); err == nil {
		t.Fatal("corrupt lifecycle accepted")
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
	if err := store.CheckHealth(context.Background()); err == nil {
		t.Fatal("closed store reported healthy")
	}
	if _, err := store.State("a", "test", "key"); err == nil {
		t.Fatal("closed state read succeeded")
	}
	if _, err := store.Rows("a", "messages"); err == nil {
		t.Fatal("closed row read succeeded")
	}
}

func TestSQLiteStorageUnavailable(t *testing.T) {
	for _, store := range []*plugin.SQLiteStorage{nil, {}} {
		if err := store.CheckHealth(context.Background()); err == nil {
			t.Fatal("uninitialized store reported healthy")
		}
		if err := store.Close(); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := plugin.OpenSQLiteStorage(filepath.Join(t.TempDir(), "missing", "storage.sqlite")); err == nil {
		t.Fatal("unavailable storage path accepted")
	}
}
