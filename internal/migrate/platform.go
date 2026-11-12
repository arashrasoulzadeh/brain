package migrate

import "context"

const PlatformSchemaSQL = `CREATE TABLE IF NOT EXISTS schema_migrations (
  version INTEGER PRIMARY KEY,
  applied_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);`

// PlatformMigrations is intentionally database-driver agnostic. The SQL executor
// is supplied by the PostgreSQL integration layer, so local unit tests stay fast.
var PlatformMigrations = []Migration{
	{
		Version: 1,
		Name:    "platform_migrations",
		Apply:   func(context.Context) error { return nil },
	},
}
