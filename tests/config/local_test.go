package config_test

import (
	"testing"

	"brain/internal/config"
)

func TestLocalTestSettingsAreIsolated(t *testing.T) {
	c := config.LocalTest()
	if c.HTTPAddress != "127.0.0.1:0" {
		t.Fatalf("HTTPAddress = %q", c.HTTPAddress)
	}
	if c.RuntimeDSN == c.MigrationDSN || c.RuntimeDSN == "" || c.MigrationDSN == "" {
		t.Fatalf("local database settings are not isolated")
	}
}
