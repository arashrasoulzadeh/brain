package db_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"brain/internal/db"
	"brain/internal/project"
	_ "modernc.org/sqlite"
)

func TestOwnerProjectQueriesAndLifecycle(t *testing.T) {
	ctx := project.WithProject(context.Background(), project.Project{ID: "alpha", Active: true})
	owner, err := db.Open("sqlite", ":memory:", sql.Open)
	if err != nil {
		t.Fatal(err)
	}
	defer owner.Close()
	if err := owner.Ready(ctx, time.Second); err != nil {
		t.Fatal(err)
	}
	if _, err := owner.ExecProjectContext(ctx, "alpha", "CREATE TABLE records (id TEXT)"); err != nil {
		t.Fatal(err)
	}
	if _, err := owner.ExecProjectContext(ctx, "alpha", "INSERT INTO records VALUES (?)", "one"); err != nil {
		t.Fatal(err)
	}
	rows, err := owner.QueryProjectContext(ctx, "alpha", "SELECT id FROM records")
	if err != nil {
		t.Fatal(err)
	}
	var id string
	if !rows.Next() || rows.Scan(&id) != nil || id != "one" {
		t.Fatal("expected scoped record")
	}
	if err := rows.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := owner.QueryProjectContext(ctx, "beta", "SELECT id FROM records"); !errors.Is(err, db.ErrProjectContextMismatch) {
		t.Fatalf("cross-project query: %v", err)
	}
	if err := owner.Close(); err != nil {
		t.Fatal(err)
	}
	if err := owner.Ready(ctx, time.Second); !errors.Is(err, db.ErrNotReady) {
		t.Fatalf("closed owner ready: %v", err)
	}
}

func TestOwnerFailsClosedForMissingContextAndDatabase(t *testing.T) {
	var owner *db.Owner
	ctx := project.WithProject(context.Background(), project.Project{ID: "alpha"})
	if _, err := owner.ExecProjectContext(ctx, "alpha", "SELECT 1"); !errors.Is(err, db.ErrNotReady) {
		t.Fatalf("nil exec: %v", err)
	}
	if _, err := owner.QueryProjectContext(ctx, "alpha", "SELECT 1"); !errors.Is(err, db.ErrNotReady) {
		t.Fatalf("nil query: %v", err)
	}
	if _, err := owner.QueryProjectContext(nil, "alpha", "SELECT 1"); !errors.Is(err, db.ErrProjectContextRequired) {
		t.Fatalf("nil context: %v", err)
	}
	empty := project.WithProject(context.Background(), project.Project{})
	if _, err := owner.QueryProjectContext(empty, "alpha", "SELECT 1"); !errors.Is(err, db.ErrProjectContextRequired) {
		t.Fatalf("empty context: %v", err)
	}
	if _, err := owner.QueryProjectContext(ctx, "", "SELECT 1"); !errors.Is(err, db.ErrProjectContextMismatch) {
		t.Fatalf("empty project: %v", err)
	}
	if err := owner.Close(); err != nil {
		t.Fatal(err)
	}
	if err := (&db.Owner{}).Close(); err != nil {
		t.Fatal(err)
	}
	if err := (&db.Owner{}).Ready(ctx, time.Second); !errors.Is(err, db.ErrNotReady) {
		t.Fatalf("empty owner ready: %v", err)
	}
	failure := errors.New("open failed")
	if _, err := db.Open("sqlite", "", func(string, string) (*sql.DB, error) {
		return nil, failure
	}); !errors.Is(err, failure) {
		t.Fatalf("open error: %v", err)
	}
}
