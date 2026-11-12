package config_test

import (
	"strings"
	"testing"

	"brain/internal/config"
)

func TestConfigRequiresHTTPAddress(t *testing.T) {
	err := (config.Config{}).SafeError()
	if err == nil || !strings.Contains(err.Error(), "HTTP_ADDRESS") {
		t.Fatalf("SafeError() = %v, want a missing HTTP_ADDRESS error", err)
	}
}

func TestConfigValidationAcceptsAddress(t *testing.T) {
	if err := (config.Config{HTTPAddress: ":8080"}).SafeError(); err != nil {
		t.Fatalf("SafeError() = %v, want nil", err)
	}
}

func TestFromEnvReadsHTTPAddress(t *testing.T) {
	t.Setenv("HTTP_ADDRESS", ":9090")
	if got := config.FromEnv().HTTPAddress; got != ":9090" {
		t.Fatalf("HTTPAddress = %q, want %q", got, ":9090")
	}
}

func TestConfigRejectsSharedDatabaseCredentials(t *testing.T) {
	err := (config.Config{HTTPAddress: ":8080", RuntimeDSN: "runtime", MigrationDSN: "runtime"}).SafeError()
	if err == nil || !strings.Contains(err.Error(), "must differ") {
		t.Fatalf("SafeError() = %v", err)
	}
}
