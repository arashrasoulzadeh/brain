package main

import (
	"fmt"

	"brain/internal/plugin"
)

// Bind exact registry-selected artifacts before restoring installations.
// No process or permanent Neuron placement is created during enable.
func bindCortexPluginWorkers(registry *plugin.Registry, locations map[string]plugin.PluginLocation, cluster *devCluster) error {
	artifacts := make(map[string]*plugin.ScheduledArtifact, len(locations))
	for _, manifest := range registry.RegisteredManifests() {
		key := manifest.ID + "@" + manifest.Version
		location, ok := locations[key]
		if !ok {
			return fmt.Errorf("plugin location unavailable: %s", key)
		}
		artifact, err := plugin.RegisterScheduledArtifact(cluster.catalog, cluster.pluginRuntime, manifest, location.Directory)
		if err != nil {
			return fmt.Errorf("plugin artifact %s rejected: %w", key, err)
		}
		artifacts[key] = artifact
	}
	return registry.BindScheduledWorkerFactory(func(projectID, pluginID, version string) (*plugin.ScheduledWorker, error) {
		artifact := artifacts[pluginID+"@"+version]
		if artifact == nil {
			return nil, plugin.ErrWorkerUnavailable
		}
		return artifact.Worker(projectID, cluster.scheduler)
	})
}
