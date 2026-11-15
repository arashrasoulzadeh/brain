package admin

import (
	"brain/internal/operator"
	"brain/internal/plugin"
)

type LifecycleService struct {
	operators *operator.Store
	registry  *plugin.Registry
}

func NewLifecycleService(operators *operator.Store, registry *plugin.Registry) *LifecycleService {
	return &LifecycleService{operators: operators, registry: registry}
}

func (s *LifecycleService) GrantCapability(actor operator.Actor, projectID, pluginID, version, capability string) error {
	return s.execute(actor, "plugin.grant", func() error {
		return s.registry.Grant(projectID, pluginID, version, capability)
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

func (s *LifecycleService) execute(actor operator.Actor, action string, operation func() error) error {
	if s == nil || s.operators == nil || s.registry == nil {
		return operator.ErrInvalidOperatorAction
	}
	return s.operators.ExecuteAction(actor, action, operation)
}
