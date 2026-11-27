package sqliteconfig

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

func TestDSN(t *testing.T) {
	for _, tc := range []struct{ path, want string }{
		{"test.db", "test.db?_pragma=busy_timeout%3D5000"},
		{"file:test.db?mode=ro", "file:test.db?mode=ro&_pragma=busy_timeout%3D5000"},
		{":memory:", ":memory:?_pragma=busy_timeout%3D5000"},
	} {
		if got := DSN(tc.path); got != tc.want {
			t.Fatalf("DSN(%q) = %q; want %q", tc.path, got, tc.want)
		}
	}
}

func TestBusyTimeoutOnEachConnection(t *testing.T) {
	db, err := sql.Open("sqlite", DSN(filepath.Join(t.TempDir(), "test.db")))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	ctx := context.Background()
	first, err := db.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer first.Close()
	second, err := db.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer second.Close()
	for _, conn := range []*sql.Conn{first, second} {
		var timeout int
		if err := conn.QueryRowContext(ctx, "PRAGMA busy_timeout").Scan(&timeout); err != nil || timeout != 5000 {
			t.Fatalf("busy timeout: %d, %v", timeout, err)
		}
	}
}
