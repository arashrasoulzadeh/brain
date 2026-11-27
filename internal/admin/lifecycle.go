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
	return s.executeProjectLifecycle(actor, projectID, "", "", "project.create", func() error {
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
	return s.executeProjectLifecycle(actor, projectID, "", "", "project.active", func() error {
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
	manifests := s.registry.RegisteredManifests()
	for index := range manifests {
		manifests[index].Admin = manifests[index].AdminForOperator()
	}
	return manifests, s.registry.InstallationStatuses(projectID), nil
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
	return s.executeProjectLifecycle(actor, projectID, pluginID, version, "plugin.install", func() error {
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
	return s.executeProjectLifecycle(actor, projectID, pluginID, version, "plugin.grant", func() error {
		return s.registry.Grant(projectID, pluginID, version, capability)
	})
}

func (s *LifecycleService) GrantResource(actor operator.Actor, projectID, pluginID, version, capability, resource string) error {
	return s.executeProjectLifecycle(actor, projectID, pluginID, version, "plugin.grant-resource", func() error {
		return s.registry.GrantResource(projectID, pluginID, version, capability, resource)
	})
}

func (s *LifecycleService) ConfigurePlugin(actor operator.Actor, projectID, pluginID, version string, values map[string]any) error {
	return s.executeProjectLifecycle(actor, projectID, pluginID, version, "plugin.configure", func() error {
		return s.registry.Configure(projectID, pluginID, version, values)
	})
}

func (s *LifecycleService) RevokeCapability(actor operator.Actor, projectID, pluginID, version, capability string) error {
	return s.executeProjectLifecycle(actor, projectID, pluginID, version, "plugin.revoke", func() error {
		return s.registry.Revoke(projectID, pluginID, version, capability)
	})
}

func (s *LifecycleService) EnablePlugin(actor operator.Actor, projectID, pluginID, version string) error {
	return s.executeProjectLifecycle(actor, projectID, pluginID, version, "plugin.enable", func() error {
		return s.registry.Enable(projectID, pluginID, version)
	})
}

func (s *LifecycleService) StartPluginWorker(actor operator.Actor, projectID, pluginID, version string) error {
	return s.executeProjectLifecycle(actor, projectID, pluginID, version, "plugin.worker.start", func() error {
		return s.registry.StartWorker(projectID, pluginID, version)
	})
}

func (s *LifecycleService) InvokePlugin(actor operator.Actor, invocation PluginInvocation) (plugin.WorkerResponse, error) {
	var response plugin.WorkerResponse
	err := s.execute(actor, "plugin.invoke", func() error {
		var invokeErr error
		response, invokeErr = s.registry.InvokeAsOperator(plugin.WorkerRequest{Protocol: s.registry.HostAPI, ProjectID: invocation.ProjectID, PluginID: invocation.PluginID, Version: invocation.Version, Operation: invocation.Operation, Payload: invocation.Payload}, actor.ID)
		return invokeErr
	})
	return response, err
}

// adminPageManifest resolves an exact, currently authorized installation.
func (s *LifecycleService) adminPageManifest(actor operator.Actor, pageID string, invocation PluginInvocation) (plugin.Manifest, error) {
	if s == nil || s.registry == nil || s.projects == nil || actor.Kind != operator.ActorOperator {
		return plugin.Manifest{}, operator.ErrInvalidOperatorAction
	}
	if _, err := s.projects.Resolve(invocation.ProjectID); err != nil {
		return plugin.Manifest{}, plugin.ErrProjectInvalid
	}
	status, err := s.registry.Status(invocation.ProjectID, invocation.PluginID, invocation.Version)
	if err != nil || !status.Enabled || !status.Ready || status.Draining {
		return plugin.Manifest{}, ErrPluginUnavailable
	}
	grants := map[string]bool{}
	for _, grant := range status.Grants {
		grants[grant] = true
	}
	for _, manifest := range s.registry.RegisteredManifests() {
		if manifest.ID != invocation.PluginID || manifest.Version != invocation.Version {
			continue
		}
		if _, ok := manifest.AdminPageSpec(pageID); !ok {
			return plugin.Manifest{}, plugin.ErrGrantDenied
		}
		for _, capability := range manifest.Capabilities {
			if !grants[capability] {
				return plugin.Manifest{}, plugin.ErrGrantDenied
			}
		}
		return manifest, nil
	}
	return plugin.Manifest{}, plugin.ErrNotRegistered
}

func (s *LifecycleService) InvokeAdminPage(actor operator.Actor, pageID string, invocation PluginInvocation) (plugin.WorkerResponse, error) {
	return s.invokeAdminPage(actor, pageID, invocation, false)
}

func (s *LifecycleService) QueryAdminPage(actor operator.Actor, pageID string, invocation PluginInvocation) (plugin.WorkerResponse, error) {
	return s.invokeAdminPage(actor, pageID, invocation, true)
}

func (s *LifecycleService) invokeAdminPage(actor operator.Actor, pageID string, invocation PluginInvocation, query bool) (plugin.WorkerResponse, error) {
	var response plugin.WorkerResponse
	auditAction := "plugin.admin.invoke"
	if query {
		auditAction = "plugin.admin.query"
	}
	err := s.execute(actor, auditAction, func() error {
		manifest, err := s.adminPageManifest(actor, pageID, invocation)
		if err != nil {
			return err
		}
		if query && !manifest.AdminReadOnly(pageID, invocation.Operation) {
			return plugin.ErrGrantDenied
		}
		if err := manifest.ValidateAdminPayload(pageID, invocation.Operation, invocation.Payload); err != nil {
			return err
		}
		response, err = s.registry.InvokeAsOperator(plugin.WorkerRequest{Protocol: s.registry.HostAPI, ProjectID: invocation.ProjectID, PluginID: invocation.PluginID, Version: invocation.Version, Operation: invocation.Operation, Payload: invocation.Payload}, actor.ID)
		if err == nil && (!response.Success || response.ErrorCode != "") {
			return errors.New("plugin admin action failed")
		}
		return err
	})
	return response, err
}

// AdminPageRevision is an authenticated read like ListPlugins. No event payload
// is exposed, and every poll rechecks project and plugin authorization.
func (s *LifecycleService) AdminPageRevision(actor operator.Actor, pageID string, invocation PluginInvocation) (uint64, error) {
	if s == nil || s.operators == nil {
		return 0, operator.ErrInvalidOperatorAction
	}
	if err := s.operators.AuthorizeAction(actor, "plugin.admin.revision"); err != nil {
		return 0, err
	}
	manifest, err := s.adminPageManifest(actor, pageID, invocation)
	if err != nil {
		return 0, err
	}
	page, _ := manifest.AdminPageSpec(pageID)
	return s.registry.AdminObservationRevision(invocation.ProjectID, page.RefreshOn), nil
}

func (s *LifecycleService) PromotePlugin(actor operator.Actor, projectID, pluginID, version string) error {
	return s.executeProjectLifecycle(actor, projectID, pluginID, version, "plugin.promote", func() error {
		return s.registry.Promote(projectID, pluginID, version)
	})
}

func (s *LifecycleService) RemovePlugin(actor operator.Actor, projectID, pluginID, version string) error {
	return s.executeProjectLifecycle(actor, projectID, pluginID, version, "plugin.remove", func() error {
		return s.registry.Remove(projectID, pluginID, version)
	})
}

func (s *LifecycleService) UninstallProject(actor operator.Actor, projectID string) error {
	return s.executeProjectLifecycle(actor, projectID, "", "", "project.uninstall", func() error {
		return s.registry.UninstallProject(projectID)
	})
}

func (s *LifecycleService) DeleteDurableData(actor operator.Actor, projectID, pluginID, key string) error {
	return s.execute(actor, "durable-data.delete", func() error {
		return s.registry.DeleteDurableData(projectID, pluginID, key)
	})
}

func (s *LifecycleService) DrainPlugin(actor operator.Actor, projectID, pluginID, version string) error {
	return s.executeProjectLifecycle(actor, projectID, pluginID, version, "plugin.drain", func() error {
		return s.registry.Drain(projectID, pluginID, version)
	})
}

func (s *LifecycleService) StopPluginImmediately(actor operator.Actor, projectID, pluginID, version string) error {
	return s.executeProjectLifecycle(actor, projectID, pluginID, version, "plugin.stop-immediate", func() error {
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
	return s.executeProjectLifecycle(actor, preview.ProjectID, preview.PluginID, preview.Version, "plugin.disable.apply", func() error {
		return s.registry.ApplyDisablePreview(preview)
	})
}

func (s *LifecycleService) RecoverPlugin(actor operator.Actor, projectID, pluginID, version string) error {
	return s.executeProjectLifecycle(actor, projectID, pluginID, version, "plugin.recover", func() error {
		return s.registry.Recover(projectID, pluginID, version)
	})
}

func (s *LifecycleService) ResumePlugin(actor operator.Actor, projectID, pluginID, version string) error {
	return s.executeProjectLifecycle(actor, projectID, pluginID, version, "plugin.resume", func() error {
		return s.registry.Recover(projectID, pluginID, version)
	})
}

func (s *LifecycleService) StopPlugin(actor operator.Actor, projectID, pluginID, version string) error {
	return s.executeProjectLifecycle(actor, projectID, pluginID, version, "plugin.stop", func() error {
		return s.registry.StopImmediately(projectID, pluginID, version)
	})
}

func (s *LifecycleService) RollbackPlugin(actor operator.Actor, projectID, pluginID, version string) error {
	return s.executeProjectLifecycle(actor, projectID, pluginID, version, "plugin.rollback", func() error {
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

func (s *LifecycleService) executeProjectLifecycle(actor operator.Actor, projectID, pluginID, version, action string, operation func() error) error {
	if err := s.execute(actor, action, operation); err != nil {
		return err
	}
	if s.registry == nil {
		return nil
	}
	s.registry.PublishHostObservationAs(projectID, "dashboard", "lifecycleChanged", "operator:"+actor.ID, map[string]any{
		"action": action, "targetPlugin": pluginID, "targetVersion": version,
	})
	return nil
}

func (s *LifecycleService) AdminInventory(actor operator.Actor, pageID string, invocation PluginInvocation, after string, limit int) (plugin.AdminInventoryPage, error) {
	var result plugin.AdminInventoryPage
	err := s.execute(actor, "plugin.admin.inventory", func() error {
		if _, err := s.adminPageManifest(actor, pageID, invocation); err != nil {
			return err
		}
		var err error
		result, err = s.registry.AdminInventory(invocation.ProjectID, invocation.PluginID, invocation.Version, pageID, after, limit)
		return err
	})
	return result, err
}
