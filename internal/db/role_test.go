package db_test

import (
	"errors"
	"testing"

	"brain/internal/db"
)

func TestRuntimeRoleCannotRunDDL(t *testing.T) {
	if err := db.AllowDDL(db.RuntimeRole); !errors.Is(err, db.ErrDDLForbidden) {
		t.Fatalf("AllowDDL(runtime) = %v", err)
	}
}

func TestMigrationRoleCanRunDDL(t *testing.T) {
	if err := db.AllowDDL(db.MigrationRole); err != nil {
		t.Fatalf("AllowDDL(migration) = %v", err)
	}
}
