package plugin

import "errors"

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
}

type Registry struct {
	HostAPI       int
	registered    map[string]Registered
	installations map[string]Installation
}

func NewRegistry(hostAPI int) *Registry {
	return &Registry{HostAPI: hostAPI, registered: map[string]Registered{}, installations: map[string]Installation{}}
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
	return nil
}

func (r *Registry) Install(projectID, id, version string) error {
	key := id + "@" + version
	if _, ok := r.registered[key]; !ok {
		return ErrNotRegistered
	}
	installKey := projectID + ":" + key
	r.installations[installKey] = Installation{ProjectID: projectID, PluginID: id, Version: version, Grants: map[string]bool{}}
	return nil
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
	for _, capability := range r.registered[id+"@"+version].Manifest.Capabilities {
		if !i.Grants[capability] {
			return ErrGrantDenied
		}
	}
	i.Enabled = true
	r.installations[key] = i
	return nil
}

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
