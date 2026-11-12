package migrate_test

import (
	"strings"
	"testing"

	"brain/internal/migrate"
)

func TestPlatformSchemaCreatesMigrationTable(t *testing.T) {
	for _, fragment := range []string{"CREATE TABLE", "schema_migrations", "version", "applied_at"} {
		if !strings.Contains(migrate.PlatformSchemaSQL, fragment) {
			t.Fatalf("PlatformSchemaSQL lacks %q", fragment)
		}
	}
}
