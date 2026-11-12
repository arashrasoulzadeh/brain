package dashboard

import "brain/internal/plugin"

func Manifest() plugin.Manifest {
	return plugin.Manifest{
		ID: "dashboard", Version: "0.1.0", HostAPI: 1,
		Capabilities:    []string{"project.manage", "plugin.install", "plugin.grant", "plugin.enable"},
		ExtensionPoints: []string{"admin.navigation", "admin.page"},
		ConfigSchema:    map[string]string{"title": "string"},
	}
}
