package db_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"brain/internal/db"
	"brain/internal/project"
)

func TestProjectQueryAndOwnerLifecycle(t *testing.T) {
	owner, err := db.Open("sqlite", ":memory:", sql.Open)
	if err != nil {
		t.Fatal(err)
	}
	defer owner.Close()
	ctx := project.WithProject(context.Background(), project.Project{ID: "a", Active: true})
	if err := owner.Ready(ctx, time.Second); err != nil {
		t.Fatal(err)
	}
	if _, err := owner.ExecProjectContext(ctx, "a", "CREATE TABLE records (id TEXT)"); err != nil {
		t.Fatal(err)
	}
	if _, err := owner.ExecProjectContext(ctx, "a", "INSERT INTO records VALUES (?)", "one"); err != nil {
		t.Fatal(err)
	}
	rows, err := owner.QueryProjectContext(ctx, "a", "SELECT id FROM records")
	if err != nil {
		t.Fatal(err)
	}
	if !rows.Next() {
		t.Fatal("missing record")
	}
	var id string
	if err := rows.Scan(&id); err != nil || id != "one" {
		t.Fatalf("record = %q, %v", id, err)
	}
	rows.Close()
	if _, err := owner.QueryProjectContext(ctx, "b", "SELECT id FROM records"); !errors.Is(err, db.ErrProjectContextMismatch) {
		t.Fatalf("cross-project query: %v", err)
	}
	if _, err := owner.QueryProjectContext(context.Background(), "a", "SELECT id FROM records"); !errors.Is(err, db.ErrProjectContextRequired) {
		t.Fatalf("missing project: %v", err)
	}
	if _, err := (*db.Owner)(nil).QueryProjectContext(ctx, "a", "SELECT 1"); !errors.Is(err, db.ErrNotReady) {
		t.Fatalf("nil owner: %v", err)
	}
	if err := (*db.Owner)(nil).Close(); err != nil {
		t.Fatal(err)
	}
	if err := owner.Close(); err != nil {
		t.Fatal(err)
	}
	if err := owner.Ready(ctx, time.Second); !errors.Is(err, db.ErrNotReady) {
		t.Fatalf("closed owner ready: %v", err)
	}
}
