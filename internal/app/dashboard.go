package app

// DashboardBootstrap describes the TypeScript dashboard worker the CLI makes available.
// The dashboard remains a plugin; the Go host only exposes its pinned contract.
type DashboardBootstrap struct {
	PluginID   string
	Version    string
	EntryPoint string
}

func BootstrapDashboard() (DashboardBootstrap, error) {
	return DashboardBootstrap{
		PluginID:   "dashboard",
		Version:    "0.1.0",
		EntryPoint: "plugins/dashboard/src/index.ts",
	}, nil
}
