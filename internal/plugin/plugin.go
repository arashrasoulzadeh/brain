package plugin

import (
	"errors"
	"fmt"
)

var (
	ErrDuplicate         = errors.New("plugin version already registered")
	ErrIncompatible      = errors.New("plugin host API incompatible")
	ErrNotRegistered     = errors.New("plugin version not registered")
	ErrNotInstalled      = errors.New("plugin not installed for project")
	ErrDisabled          = errors.New("plugin disabled")
	ErrGrantDenied       = errors.New("plugin grant denied")
	ErrProjectInvalid    = errors.New("project identity invalid")
	ErrDependencyMissing = errors.New("plugin dependency missing")
	ErrInvalidManifest   = errors.New("invalid plugin manifest")
	ErrInvalidConfig     = errors.New("invalid plugin configuration")
	ErrNotReady          = errors.New("plugin worker not ready")
)

type Manifest struct {
	ID              string
	Version         string
	HostAPI         int
	Dependencies    []string
	Capabilities    []string
	ExtensionPoints []string
	ConfigSchema    map[string]string
}

type Registered struct{ Manifest Manifest }

type Installation struct {
	ProjectID string
	PluginID  string
	Version   string
	Enabled   bool
	Grants    map[string]bool
	Config    map[string]any
	Resources map[string]map[string]bool
}

type Registry struct {
	HostAPI       int
	registered    map[string]Registered
	installations map[string]Installation
	readiness     map[string]bool
	audit         []AuditEvent
}

type AuditEvent struct {
	Action, ProjectID, PluginID, Version, Detail string
}

type CapabilityRequest struct {
	Capability string
	Required   bool
}

func (m Manifest) RequestedCapabilities() []CapabilityRequest {
	requests := make([]CapabilityRequest, 0, len(m.Capabilities))
	for _, capability := range m.Capabilities {
		requests = append(requests, CapabilityRequest{Capability: capability, Required: true})
	}
	return requests
}

func NewRegistry(hostAPI int) *Registry {
	return &Registry{HostAPI: hostAPI, registered: map[string]Registered{}, installations: map[string]Installation{}, readiness: map[string]bool{}}
}

func (r *Registry) Register(m Manifest) error {
	if m.ID == "" || m.Version == "" || m.HostAPI != r.HostAPI {
		if m.ID == "" || m.Version == "" {
			return ErrInvalidManifest
		}
		return ErrIncompatible
	}
	for _, dependency := range m.Dependencies {
		if _, ok := r.registered[dependency]; !ok {
			return ErrDependencyMissing
		}
	}
	key := m.ID + "@" + m.Version
	if _, ok := r.registered[key]; ok {
		return ErrDuplicate
	}
	r.registered[key] = Registered{Manifest: m}
	r.readiness[key] = false
	r.audit = append(r.audit, AuditEvent{Action: "register", PluginID: m.ID, Version: m.Version})
	return nil
}

func (r *Registry) SetReadiness(id, version string, ready bool) error {
	key := id + "@" + version
	if _, ok := r.registered[key]; !ok {
		return ErrNotRegistered
	}
	r.readiness[key] = ready
	r.audit = append(r.audit, AuditEvent{Action: "readiness", PluginID: id, Version: version, Detail: fmt.Sprintf("ready=%t", ready)})
	return nil
}

func (r *Registry) Install(projectID, id, version string) error {
	key := id + "@" + version
	if _, ok := r.registered[key]; !ok {
		return ErrNotRegistered
	}
	installKey := projectID + ":" + key
	r.installations[installKey] = Installation{ProjectID: projectID, PluginID: id, Version: version, Grants: map[string]bool{}, Resources: map[string]map[string]bool{}, Config: map[string]any{}}
	r.audit = append(r.audit, AuditEvent{Action: "install", ProjectID: projectID, PluginID: id, Version: version})
	return nil
}

func (r *Registry) GrantResource(projectID, id, version, capability, resource string) error {
	if resource == "" {
		return ErrGrantDenied
	}
	i, ok := r.installations[projectID+":"+id+"@"+version]
	if !ok {
		return ErrNotInstalled
	}
	allowed := false
	for _, requested := range r.registered[id+"@"+version].Manifest.Capabilities {
		if requested == capability {
			allowed = true
		}
	}
	if !allowed {
		return ErrGrantDenied
	}
	if i.Resources[capability] == nil {
		i.Resources[capability] = map[string]bool{}
	}
	i.Grants[capability] = true
	i.Resources[capability][resource] = true
	r.installations[projectID+":"+id+"@"+version] = i
	r.audit = append(r.audit, AuditEvent{Action: "grant_resource", ProjectID: projectID, PluginID: id, Version: version, Detail: capability + ":" + resource})
	return nil
}

func (r *Registry) AuthorizeResource(projectID, id, version, capability, resource string) error {
	if err := r.Authorize(projectID, id, version, capability); err != nil {
		return err
	}
	i := r.installations[projectID+":"+id+"@"+version]
	if resources := i.Resources[capability]; len(resources) > 0 && !resources[resource] {
		return ErrGrantDenied
	}
	return nil
}

func (r *Registry) Configure(projectID, id, version string, values map[string]any) error {
	key := projectID + ":" + id + "@" + version
	i, ok := r.installations[key]
	if !ok {
		return ErrNotInstalled
	}
	m := r.registered[id+"@"+version].Manifest
	for name, value := range values {
		want, declared := m.ConfigSchema[name]
		if !declared || !configTypeMatches(want, value) {
			return fmt.Errorf("%w: %s", ErrInvalidConfig, name)
		}
		i.Config[name] = value
	}
	r.installations[key] = i
	r.audit = append(r.audit, AuditEvent{Action: "configure", ProjectID: projectID, PluginID: id, Version: version})
	return nil
}

func configTypeMatches(want string, value any) bool {
	switch want {
	case "string":
		_, ok := value.(string)
		return ok
	case "bool":
		_, ok := value.(bool)
		return ok
	case "int":
		_, ok := value.(int)
		return ok
	default:
		return false
	}
}

func (r *Registry) Grant(projectID, id, version, capability string) error {
	i, ok := r.installations[projectID+":"+id+"@"+version]
	if !ok {
		return ErrNotInstalled
	}
	for _, requested := range r.registered[id+"@"+version].Manifest.Capabilities {
		if requested == capability {
			i.Grants[capability] = true
			r.installations[projectID+":"+id+"@"+version] = i
			r.audit = append(r.audit, AuditEvent{Action: "grant", ProjectID: projectID, PluginID: id, Version: version, Detail: capability})
			return nil
		}
	}
	return ErrGrantDenied
}

func (r *Registry) Enable(projectID, id, version string) error {
	key := projectID + ":" + id + "@" + version
	i, ok := r.installations[key]
	if !ok {
		return ErrNotInstalled
	}
	if !r.readiness[id+"@"+version] {
		return ErrNotReady
	}
	for _, capability := range r.registered[id+"@"+version].Manifest.Capabilities {
		if !i.Grants[capability] {
			return ErrGrantDenied
		}
	}
	i.Enabled = true
	r.installations[key] = i
	r.audit = append(r.audit, AuditEvent{Action: "enable", ProjectID: projectID, PluginID: id, Version: version})
	return nil
}

func (r *Registry) Audit() []AuditEvent { return append([]AuditEvent(nil), r.audit...) }

func (r *Registry) Authorize(projectID, id, version, capability string) error {
	if projectID == "" {
		return ErrProjectInvalid
	}
	i, ok := r.installations[projectID+":"+id+"@"+version]
	if !ok {
		return ErrNotInstalled
	}
	if !i.Enabled {
		return ErrDisabled
	}
	if !i.Grants[capability] {
		return ErrGrantDenied
	}
	return nil
}
