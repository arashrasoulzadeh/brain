package db_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"brain/internal/db"
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

func TestNilOwnerIsNotReady(t *testing.T) {
	if err := (*db.Owner)(nil).Ready(context.Background(), time.Millisecond); err != db.ErrNotReady {
		t.Fatalf("Ready() = %v", err)
	}
}
