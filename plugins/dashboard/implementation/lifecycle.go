package dashboard

import (
	"time"

	"brain/internal/plugin"
)

type Lifecycle struct{ Registry *plugin.Registry }

func (l Lifecycle) Promote(projectID, pluginID, version string) error {
	return l.Registry.Promote(projectID, pluginID, version)
}
func (l Lifecycle) Rollback(projectID, pluginID, version string) error {
	return l.Registry.Rollback(projectID, pluginID, version)
}
func (l Lifecycle) Drain(projectID, pluginID, version string, deadline time.Time) error {
	return l.Registry.DrainUntil(projectID, pluginID, version, deadline)
}
func (l Lifecycle) Remove(projectID, pluginID, version string) error {
	return l.Registry.Remove(projectID, pluginID, version)
}
