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

func TestOpenUsesInjectedOpener(t *testing.T) {
	called := false
	owner, err := db.Open("sqlite", ":memory:", func(driver, dsn string) (*sql.DB, error) {
		called = driver == "sqlite" && dsn == ":memory:"
		return nil, nil
	})
	if err != nil || !called || owner == nil {
		t.Fatalf("Open() = %#v, %v, called=%v", owner, err, called)
	}
}

func TestRuntimeSQLRequiresMatchingProjectContext(t *testing.T) {
	owner := &db.Owner{}
	ctx := project.WithProject(context.Background(), project.Project{ID: "dev"})
	if _, err := owner.ExecProjectContext(ctx, "prod", "SELECT 1"); !errors.Is(err, db.ErrProjectContextMismatch) {
		t.Fatalf("cross-project SQL was accepted: %v", err)
	}
	if _, err := owner.ExecProjectContext(context.Background(), "dev", "SELECT 1"); !errors.Is(err, db.ErrProjectContextRequired) {
		t.Fatalf("unbound SQL was accepted: %v", err)
	}
}

func TestNilOwnerIsNotReady(t *testing.T) {
	if err := (*db.Owner)(nil).Ready(context.Background(), time.Millisecond); err != db.ErrNotReady {
		t.Fatalf("Ready() = %v", err)
	}
}
