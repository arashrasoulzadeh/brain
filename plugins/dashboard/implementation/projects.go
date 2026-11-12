package dashboard

import (
	"brain/internal/plugin"
	"brain/internal/project"
)

type ProjectAdmin struct {
	Projects *project.Store
	Plugins  *plugin.Registry
}

func (a ProjectAdmin) CreateProject(id string) error { return a.Projects.Create(id) }
func (a ProjectAdmin) InstallPlugin(projectID, pluginID, version string) error {
	return a.Plugins.Install(projectID, pluginID, version)
}
func (a ProjectAdmin) GrantPlugin(projectID, pluginID, version, capability string) error {
	return a.Plugins.Grant(projectID, pluginID, version, capability)
}
func (a ProjectAdmin) EnablePlugin(projectID, pluginID, version string) error {
	return a.Plugins.Enable(projectID, pluginID, version)
}
