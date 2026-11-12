package migrate_test

import (
	"testing"

	"brain/internal/migrate"
)

func TestPlatformMigrationsStartAtVersionOne(t *testing.T) {
	if len(migrate.PlatformMigrations) == 0 || migrate.PlatformMigrations[0].Version != 1 {
		t.Fatalf("platform migrations = %#v", migrate.PlatformMigrations)
	}
}
