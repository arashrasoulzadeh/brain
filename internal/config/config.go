package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

type Config struct {
	HTTPAddress  string
	RuntimeDSN   string
	MigrationDSN string
}

// LocalTest returns deterministic, isolated settings for fast non-PostgreSQL tests.
func LocalTest() Config {
	return Config{HTTPAddress: "127.0.0.1:0", RuntimeDSN: "file:local-test?mode=memory&cache=shared", MigrationDSN: "file:local-test-migrations?mode=memory&cache=shared"}
}

func FromEnv() Config {
	return Config{HTTPAddress: os.Getenv("HTTP_ADDRESS"), RuntimeDSN: os.Getenv("RUNTIME_DSN"), MigrationDSN: os.Getenv("MIGRATION_DSN")}
}

func (c Config) Validate() error {
	if strings.TrimSpace(c.HTTPAddress) == "" {
		return errors.New("HTTP_ADDRESS is required")
	}
	if c.RuntimeDSN != "" && c.RuntimeDSN == c.MigrationDSN {
		return errors.New("RUNTIME_DSN and MIGRATION_DSN must differ")
	}
	return nil
}

func (c Config) SafeError() error {
	if err := c.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}
	return nil
}
