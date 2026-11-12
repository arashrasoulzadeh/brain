package app_test

import (
	"testing"

	"brain/internal/app"
)

func TestBootstrapDashboardReturnsPinnedPluginContract(t *testing.T) {
	bootstrap, err := app.BootstrapDashboard()
	if err != nil {
		t.Fatal(err)
	}
	if bootstrap.PluginID != "dashboard" || bootstrap.Version == "" || bootstrap.EntryPoint == "" {
		t.Fatalf("dashboard bootstrap = %#v", bootstrap)
	}
}
