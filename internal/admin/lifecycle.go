package admin

import (
	"brain/internal/deploy"
	"brain/internal/operator"
	"brain/internal/plugin"
	"brain/internal/project"
	"encoding/json"
	"errors"
	"strings"
)

var ErrPluginUnavailable = errors.New("plugin is not enabled and ready")

type PluginInvocation struct {
	ProjectID string
	PluginID  string
	Version   string
	Operation string
	Payload   json.RawMessage
}

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

func (s *LifecycleService) SetProjectActive(actor operator.Actor, projectID string, active bool) error {
	return s.execute(actor, "project.active", func() error {
		if s.projects == nil {
			return operator.ErrInvalidOperatorAction
		}
		return s.projects.SetActive(projectID, active)
	})
}

func (s *LifecycleService) ListPlugins(actor operator.Actor, projectID string) ([]plugin.Manifest, []plugin.InstallationStatus, error) {
	if actor.Kind != operator.ActorOperator || s.registry == nil {
		return nil, nil, operator.ErrInvalidOperatorAction
	}
	return s.registry.RegisteredManifests(), s.registry.InstallationStatuses(projectID), nil
}

// RequireEnabledPlugin is the server-side gate for plugin-backed admin APIs.
// The browser sidebar is only a convenience; API authorization must recheck lifecycle state.
func (s *LifecycleService) RequireEnabledPlugin(actor operator.Actor, projectID, pluginID string) error {
	if actor.Kind != operator.ActorOperator || s.registry == nil || strings.TrimSpace(projectID) == "" || strings.TrimSpace(pluginID) == "" {
		return operator.ErrInvalidOperatorAction
	}
	for _, installation := range s.registry.InstallationStatuses(projectID) {
		if installation.PluginID == pluginID && installation.Enabled && installation.Ready {
			return nil
		}
	}
	return ErrPluginUnavailable
}

func (s *LifecycleService) SystemCapabilities(actor operator.Actor) ([]plugin.Manifest, error) {
	if actor.Kind != operator.ActorOperator || s.registry == nil {
		return nil, operator.ErrInvalidOperatorAction
	}
	return s.registry.RegisteredManifests(), nil
}

func (s *LifecycleService) InstallPlugin(actor operator.Actor, projectID, pluginID, version string) error {
	return s.execute(actor, "plugin.install", func() error {
		return s.registry.Install(projectID, pluginID, version)
	})
}

func (s *LifecycleService) SetPluginReadiness(actor operator.Actor, pluginID, version string, ready bool) error {
	return s.execute(actor, "plugin.readiness", func() error {
		return s.registry.SetReadiness(pluginID, version, ready)
	})
}

type AuditRecord struct {
	Source    string `json:"source"`
	Action    string `json:"action"`
	ActorID   string `json:"actor_id,omitempty"`
	ProjectID string `json:"project_id,omitempty"`
	PluginID  string `json:"plugin_id,omitempty"`
	Version   string `json:"version,omitempty"`
	Success   bool   `json:"success"`
	Detail    string `json:"detail,omitempty"`
}

func (s *LifecycleService) AuditRecords(actor operator.Actor) ([]AuditRecord, error) {
	if actor.Kind != operator.ActorOperator || s.operators == nil || s.registry == nil {
		return nil, operator.ErrInvalidOperatorAction
	}
	records := make([]AuditRecord, 0)
	for _, event := range s.operators.Audit() {
		records = append(records, AuditRecord{Source: "operator", Action: event.Action, ActorID: event.ActorID, Success: event.Success, Detail: event.Detail})
	}
	for _, event := range s.registry.Audit() {
		records = append(records, AuditRecord{Source: "plugin", Action: event.Action, ProjectID: event.ProjectID, PluginID: event.PluginID, Version: event.Version, Success: true, Detail: event.Detail})
	}
	return records, nil
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

func (s *LifecycleService) StartPluginWorker(actor operator.Actor, projectID, pluginID, version string) error {
	return s.execute(actor, "plugin.worker.start", func() error {
		return s.registry.StartWorker(projectID, pluginID, version)
	})
}

func (s *LifecycleService) InvokePlugin(actor operator.Actor, invocation PluginInvocation) (plugin.WorkerResponse, error) {
	var response plugin.WorkerResponse
	err := s.execute(actor, "plugin.invoke", func() error {
		var invokeErr error
		response, invokeErr = s.registry.Invoke(plugin.WorkerRequest{Protocol: s.registry.HostAPI, ProjectID: invocation.ProjectID, PluginID: invocation.PluginID, Version: invocation.Version, Operation: invocation.Operation, Payload: invocation.Payload})
		return invokeErr
	})
	return response, err
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

func (s *LifecycleService) PreviewDisable(actor operator.Actor, projectID, pluginID, version string) (plugin.DisablePreview, error) {
	var preview plugin.DisablePreview
	err := s.execute(actor, "plugin.disable.preview", func() error {
		var err error
		preview, err = s.registry.PreviewDisable(projectID, pluginID, version)
		return err
	})
	return preview, err
}

func (s *LifecycleService) ApplyDisable(actor operator.Actor, preview plugin.DisablePreview) error {
	return s.execute(actor, "plugin.disable.apply", func() error {
		return s.registry.ApplyDisablePreview(preview)
	})
}

func (s *LifecycleService) RecoverPlugin(actor operator.Actor, projectID, pluginID, version string) error {
	return s.execute(actor, "plugin.recover", func() error {
		return s.registry.Recover(projectID, pluginID, version)
	})
}

func (s *LifecycleService) ResumePlugin(actor operator.Actor, projectID, pluginID, version string) error {
	return s.execute(actor, "plugin.resume", func() error {
		return s.registry.Recover(projectID, pluginID, version)
	})
}

func (s *LifecycleService) StopPlugin(actor operator.Actor, projectID, pluginID, version string) error {
	return s.execute(actor, "plugin.stop", func() error {
		return s.registry.StopImmediately(projectID, pluginID, version)
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
