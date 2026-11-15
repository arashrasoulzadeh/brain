package admin

import (
	"brain/internal/deploy"
	"brain/internal/operator"
	"brain/internal/plugin"
	"brain/internal/project"
)

type LifecycleService struct {
	operators *operator.Store
	registry  *plugin.Registry
	projects  *project.Store
}

func NewLifecycleService(operators *operator.Store, registry *plugin.Registry) *LifecycleService {
	return &LifecycleService{operators: operators, registry: registry}
}

func NewLifecycleServiceWithProjects(operators *operator.Store, registry *plugin.Registry, projects *project.Store) *LifecycleService {
	service := NewLifecycleService(operators, registry)
	service.projects = projects
	return service
}

func (s *LifecycleService) CreateProject(actor operator.Actor, projectID string) error {
	return s.execute(actor, "project.create", func() error {
		if s.projects == nil {
			return operator.ErrInvalidOperatorAction
		}
		return s.projects.Create(projectID)
	})
}

func (s *LifecycleService) ListProjects(actor operator.Actor) ([]project.Project, error) {
	if actor.Kind != operator.ActorOperator || s.projects == nil {
		return nil, operator.ErrInvalidOperatorAction
	}
	return s.projects.List(), nil
}

func (s *LifecycleService) ListPlugins(actor operator.Actor, projectID string) ([]plugin.Manifest, []plugin.InstallationStatus, error) {
	if actor.Kind != operator.ActorOperator || s.registry == nil {
		return nil, nil, operator.ErrInvalidOperatorAction
	}
	return s.registry.RegisteredManifests(), s.registry.InstallationStatuses(projectID), nil
}

func (s *LifecycleService) InstallPlugin(actor operator.Actor, projectID, pluginID, version string) error {
	return s.execute(actor, "plugin.install", func() error {
		return s.registry.Install(projectID, pluginID, version)
	})
}

func (s *LifecycleService) GrantCapability(actor operator.Actor, projectID, pluginID, version, capability string) error {
	return s.execute(actor, "plugin.grant", func() error {
		return s.registry.Grant(projectID, pluginID, version, capability)
	})
}

func (s *LifecycleService) GrantResource(actor operator.Actor, projectID, pluginID, version, capability, resource string) error {
	return s.execute(actor, "plugin.grant-resource", func() error {
		return s.registry.GrantResource(projectID, pluginID, version, capability, resource)
	})
}

func (s *LifecycleService) ConfigurePlugin(actor operator.Actor, projectID, pluginID, version string, values map[string]any) error {
	return s.execute(actor, "plugin.configure", func() error {
		return s.registry.Configure(projectID, pluginID, version, values)
	})
}

func (s *LifecycleService) RevokeCapability(actor operator.Actor, projectID, pluginID, version, capability string) error {
	return s.execute(actor, "plugin.revoke", func() error {
		return s.registry.Revoke(projectID, pluginID, version, capability)
	})
}

func (s *LifecycleService) EnablePlugin(actor operator.Actor, projectID, pluginID, version string) error {
	return s.execute(actor, "plugin.enable", func() error {
		return s.registry.Enable(projectID, pluginID, version)
	})
}

func (s *LifecycleService) PromotePlugin(actor operator.Actor, projectID, pluginID, version string) error {
	return s.execute(actor, "plugin.promote", func() error {
		return s.registry.Promote(projectID, pluginID, version)
	})
}

func (s *LifecycleService) RemovePlugin(actor operator.Actor, projectID, pluginID, version string) error {
	return s.execute(actor, "plugin.remove", func() error {
		return s.registry.Remove(projectID, pluginID, version)
	})
}

func (s *LifecycleService) UninstallProject(actor operator.Actor, projectID string) error {
	return s.execute(actor, "project.uninstall", func() error {
		return s.registry.UninstallProject(projectID)
	})
}

func (s *LifecycleService) DeleteDurableData(actor operator.Actor, projectID, pluginID, key string) error {
	return s.execute(actor, "durable-data.delete", func() error {
		return s.registry.DeleteDurableData(projectID, pluginID, key)
	})
}

func (s *LifecycleService) DrainPlugin(actor operator.Actor, projectID, pluginID, version string) error {
	return s.execute(actor, "plugin.drain", func() error {
		return s.registry.Drain(projectID, pluginID, version)
	})
}

func (s *LifecycleService) StopPluginImmediately(actor operator.Actor, projectID, pluginID, version string) error {
	return s.execute(actor, "plugin.stop-immediate", func() error {
		return s.registry.StopImmediately(projectID, pluginID, version)
	})
}

func (s *LifecycleService) RecoverPlugin(actor operator.Actor, projectID, pluginID, version string) error {
	return s.execute(actor, "plugin.recover", func() error {
		return s.registry.Recover(projectID, pluginID, version)
	})
}

func (s *LifecycleService) RollbackPlugin(actor operator.Actor, projectID, pluginID, version string) error {
	return s.execute(actor, "plugin.rollback", func() error {
		return s.registry.Rollback(projectID, pluginID, version)
	})
}

func (s *LifecycleService) RunMigration(actor operator.Actor, version string, operation func() error) error {
	if version == "" {
		return operator.ErrInvalidOperatorAction
	}
	return s.execute(actor, "migration.apply", operation)
}

func (s *LifecycleService) SelectRollbackManifest(actor operator.Actor, manifest deploy.ReleaseManifest, schemaCompatible bool) (deploy.ReleaseManifest, error) {
	var selected deploy.ReleaseManifest
	err := s.execute(actor, "release.rollback", func() error {
		var err error
		selected, err = deploy.SelectRollbackManifest(manifest, schemaCompatible)
		return err
	})
	return selected, err
}

func (s *LifecycleService) execute(actor operator.Actor, action string, operation func() error) error {
	if s == nil || s.operators == nil || s.registry == nil {
		return operator.ErrInvalidOperatorAction
	}
	return s.operators.ExecuteAction(actor, action, operation)
}
