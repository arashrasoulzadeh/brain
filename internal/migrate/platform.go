package migrate

import "context"

// PlatformMigrations is intentionally database-driver agnostic. The SQL executor
// is supplied by the PostgreSQL integration layer, so local unit tests stay fast.
var PlatformMigrations = []Migration{
	{
		Version: 1,
		Name:    "platform_migrations",
		Apply:   func(context.Context) error { return nil },
	},
}
