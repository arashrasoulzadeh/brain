package plugin

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

var (
	ErrDuplicate          = errors.New("plugin version already registered")
	ErrIncompatible       = errors.New("plugin host API incompatible")
	ErrNotRegistered      = errors.New("plugin version not registered")
	ErrNotInstalled       = errors.New("plugin not installed for project")
	ErrDisabled           = errors.New("plugin disabled")
	ErrGrantDenied        = errors.New("plugin grant denied")
	ErrProjectInvalid     = errors.New("project identity invalid")
	ErrDependencyMissing  = errors.New("plugin dependency missing")
	ErrInvalidManifest    = errors.New("invalid plugin manifest")
	ErrInvalidConfig      = errors.New("invalid plugin configuration")
	ErrSchemaIncompatible = errors.New("plugin configuration schema incompatible")
	ErrNotReady           = errors.New("plugin worker not ready")
	ErrVersionReferenced  = errors.New("plugin version still referenced")
	ErrCapabilityApproval = errors.New("new plugin capability requires approval")
	ErrStalePreview       = errors.New("plugin lifecycle preview is stale")
	ErrActiveWork         = errors.New("project has active plugin work")
)

type Manifest struct {
	Title           string            `yaml:"title"`
	Description     string            `yaml:"description"`
	ID              string            `yaml:"id"`
	Entry           string            `yaml:"entrypoint"`
	Version         string            `yaml:"version"`
	HostAPI         int               `yaml:"host_api"`
	Dependencies    []string          `yaml:"dependencies"`
	Capabilities    []string          `yaml:"capabilities"`
	ExtensionPoints []string          `yaml:"extension_points"`
	ConfigSchema    map[string]string `yaml:"config_schema"`
	Routes          []Route           `yaml:"routes"`
	SourceFiles     []string          `yaml:"source_files"`
	Imports         []string          `yaml:"imports"`
	PluginImports   []string          `yaml:"plugin_imports"`
}

type Route struct {
	Operation  string `yaml:"operation"`
	Function   string `yaml:"function"`
	Capability string `yaml:"capability"`
}

func LoadManifest(path string) (Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Manifest{}, err
	}
	var manifest Manifest
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		return Manifest{}, err
	}
	if err := ValidateManifestContract(manifest); err != nil {
		return Manifest{}, err
	}
	return manifest, nil
}

func ValidateManifestContract(manifest Manifest) error {
	if manifest.EntryPoint() == "" {
		return ErrInvalidManifest
	}
	seenFiles := map[string]bool{}
	for _, file := range manifest.SourceFiles {
		if file == "" || strings.HasPrefix(file, "/") || path.Clean(file) != file || strings.ContainsAny(file, "\\\x00") {
			return ErrInvalidManifest
		}
		seenFiles[file] = true
	}
	if !seenFiles[manifest.EntryPoint()] {
		return ErrInvalidManifest
	}
	seenImports := map[string]bool{}
	for _, importPath := range manifest.Imports {
		cleanImportPath := path.Clean(importPath)
		if importPath == "" || strings.HasPrefix(importPath, "/") || strings.HasPrefix(cleanImportPath, "../../") || cleanImportPath == ".." {
			return ErrInvalidManifest
		}
		if seenImports[importPath] {
			return ErrInvalidManifest
		}
		seenImports[importPath] = true
	}
	seenPlugins := map[string]bool{}
	for _, pluginID := range manifest.PluginImports {
		if pluginID == "" || seenPlugins[pluginID] {
			return ErrInvalidManifest
		}
		seenPlugins[pluginID] = true
	}
	return nil
}

func (m Manifest) EntryPoint() string {
	if m.Entry == "" {
		return "index.ts"
	}
	return m.Entry
}

var sourceImportPattern = regexp.MustCompile(`(?m)^\s*import[^\n"']*(?:from\s*)?["']([^"']+)["']`)
var sourceExportPattern = regexp.MustCompile(`(?m)^\s*export[^\n"']*from\s*["']([^"']+)["']`)

func ValidatePluginSources(directory string, manifest Manifest) error {
	if err := ValidateManifestContract(manifest); err != nil {
		return err
	}
	allowedImports := map[string]bool{}
	for _, importPath := range manifest.Imports {
		allowedImports[importPath] = true
	}
	for _, sourceFile := range manifest.SourceFiles {
		data, err := os.ReadFile(filepath.Join(directory, sourceFile))
		if err != nil {
			return ErrInvalidManifest
		}
		matches := append(sourceImportPattern.FindAllStringSubmatch(string(data), -1), sourceExportPattern.FindAllStringSubmatch(string(data), -1)...)
		for _, match := range matches {
			if !allowedImports[match[1]] {
				return fmt.Errorf("%w: %s imports %s", ErrInvalidManifest, sourceFile, match[1])
			}
		}
	}
	return nil
}

func (m Manifest) Route(operation string) (Route, bool) {
	for _, route := range m.Routes {
		if route.Operation == operation {
			return route, true
		}
	}
	return Route{}, false
}

func Documentation(m Manifest) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# %s\n\n%s\n\n## Routes\n\n", m.Title, m.Description)
	for _, route := range m.Routes {
		fmt.Fprintf(&b, "- `%s` → `%s` (capability: `%s`)\n", route.Operation, route.Function, route.Capability)
	}
	return b.String()
}

type Registered struct{ Manifest Manifest }

type Installation struct {
	ProjectID  string
	PluginID   string
	Version    string
	Enabled    bool
	Grants     map[string]bool
	Config     map[string]any
	Resources  map[string]map[string]bool
	Draining   bool
	DrainUntil time.Time
}

type InstallationStatus struct {
	ProjectID  string
	PluginID   string
	Version    string
	Enabled    bool
	Ready      bool
	Draining   bool
	DrainUntil time.Time
}

type Registry struct {
	HostAPI           int
	registered        map[string]Registered
	installations     map[string]Installation
	readiness         map[string]bool
	audit             []AuditEvent
	active            map[string]string
	pins              map[string]string
	artifacts         map[string]string
	approvals         map[string]bool
	lifecycleRevision uint64
}

type AuditEvent struct {
	Action, ProjectID, PluginID, Version, Detail string
}

type DisablePreview struct {
	ProjectID  string
	PluginID   string
	Version    string
	Revision   uint64
	Dependents []string
}

const DefaultDrainWindow = 5 * time.Minute

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
	return &Registry{HostAPI: hostAPI, registered: map[string]Registered{}, installations: map[string]Installation{}, readiness: map[string]bool{}, active: map[string]string{}, pins: map[string]string{}, artifacts: map[string]string{}, approvals: map[string]bool{}}
}

func (r *Registry) ApproveCapability(projectID, id, version, capability string) error {
	if projectID == "" || capability == "" {
		return ErrProjectInvalid
	}
	if _, ok := r.installations[projectID+":"+id+"@"+version]; !ok {
		return ErrNotInstalled
	}
	r.approvals[projectID+":"+id+"@"+version+":"+capability] = true
	r.audit = append(r.audit, AuditEvent{Action: "approve_capability", ProjectID: projectID, PluginID: id, Version: version, Detail: capability})
	return nil
}

func (r *Registry) Status(projectID, pluginID, version string) (InstallationStatus, error) {
	installation, ok := r.installations[projectID+":"+pluginID+"@"+version]
	if !ok {
		return InstallationStatus{}, ErrNotInstalled
	}
	return InstallationStatus{ProjectID: projectID, PluginID: pluginID, Version: version, Enabled: installation.Enabled, Ready: r.readiness[pluginID+"@"+version], Draining: installation.Draining, DrainUntil: installation.DrainUntil}, nil
}

func (r *Registry) ApproveArtifact(id, version, endpoint string) error {
	if _, ok := r.registered[id+"@"+version]; !ok {
		return ErrNotRegistered
	}
	u, err := url.Parse(endpoint)
	if err != nil || u.Scheme != "https" || u.Host == "" {
		return ErrInvalidManifest
	}
	r.artifacts[id+"@"+version] = endpoint
	r.audit = append(r.audit, AuditEvent{Action: "approve_artifact", PluginID: id, Version: version, Detail: endpoint})
	return nil
}

func (r *Registry) ApprovedArtifact(id, version string) (string, bool) {
	endpoint, ok := r.artifacts[id+"@"+version]
	return endpoint, ok
}

func (r *Registry) Register(m Manifest) error {
	if m.ID == "" || m.Version == "" || m.HostAPI != r.HostAPI {
		if m.ID == "" || m.Version == "" {
			return ErrInvalidManifest
		}
		return ErrIncompatible
	}
	for _, extensionPoint := range m.ExtensionPoints {
		if strings.TrimSpace(extensionPoint) == "" {
			return ErrInvalidManifest
		}
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

func (r *Registry) TransitiveDependents(id, version string) []string {
	target := id + "@" + version
	seen := map[string]bool{target: true}
	dependents := make([]string, 0)
	for {
		changed := false
		for key, registered := range r.registered {
			if seen[key] {
				continue
			}
			for _, dependency := range registered.Manifest.Dependencies {
				if seen[dependency] {
					seen[key] = true
					dependents = append(dependents, key)
					changed = true
					break
				}
			}
		}
		if !changed {
			break
		}
	}
	sort.Strings(dependents)
	return dependents
}

func (r *Registry) PreviewDisable(projectID, id, version string) (DisablePreview, error) {
	if _, ok := r.installations[projectID+":"+id+"@"+version]; !ok {
		return DisablePreview{}, ErrNotInstalled
	}
	return DisablePreview{ProjectID: projectID, PluginID: id, Version: version, Revision: r.lifecycleRevision, Dependents: r.TransitiveDependents(id, version)}, nil
}

func (r *Registry) ApplyDisablePreview(preview DisablePreview) error {
	if preview.Revision != r.lifecycleRevision {
		return ErrStalePreview
	}
	if _, ok := r.installations[preview.ProjectID+":"+preview.PluginID+"@"+preview.Version]; !ok {
		return ErrNotInstalled
	}
	r.lifecycleRevision++
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

func (r *Registry) Revoke(projectID, id, version, capability string) error {
	key := projectID + ":" + id + "@" + version
	i, ok := r.installations[key]
	if !ok {
		return ErrNotInstalled
	}
	delete(i.Grants, capability)
	delete(i.Resources, capability)
	r.installations[key] = i
	r.audit = append(r.audit, AuditEvent{Action: "revoke", ProjectID: projectID, PluginID: id, Version: version, Detail: capability})
	return nil
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

func (r *Registry) Promote(projectID, id, version string) error {
	key := projectID + ":" + id + "@" + version
	i, ok := r.installations[key]
	if !ok {
		return ErrNotInstalled
	}
	if !r.readiness[id+"@"+version] {
		return ErrNotReady
	}
	if !i.Enabled {
		return ErrDisabled
	}
	manifest := r.registered[id+"@"+version].Manifest
	for _, capability := range manifest.Capabilities {
		if !i.Grants[capability] {
			return ErrGrantDenied
		}
	}
	if activeVersion, active := r.active[projectID+":"+id]; active && activeVersion != version {
		previousManifest := r.registered[id+"@"+activeVersion].Manifest
		for _, capability := range manifest.Capabilities {
			known := false
			for _, previousCapability := range previousManifest.Capabilities {
				if capability == previousCapability {
					known = true
					break
				}
			}
			if !known && !r.approvals[projectID+":"+id+"@"+version+":"+capability] {
				return ErrCapabilityApproval
			}
		}
		previous := r.registered[id+"@"+activeVersion].Manifest.ConfigSchema
		for name, previousType := range previous {
			if candidateType, exists := manifest.ConfigSchema[name]; !exists || candidateType != previousType {
				return fmt.Errorf("%w: %s", ErrSchemaIncompatible, name)
			}
		}
	}
	for name, expectedType := range manifest.ConfigSchema {
		value, configured := i.Config[name]
		if !configured || !configTypeMatches(expectedType, value) {
			return fmt.Errorf("%w: %s", ErrInvalidConfig, name)
		}
	}
	r.active[projectID+":"+id] = version
	r.lifecycleRevision++
	r.audit = append(r.audit, AuditEvent{Action: "promote", ProjectID: projectID, PluginID: id, Version: version})
	return nil
}

func (r *Registry) Drain(projectID, id, version string) error {
	return r.DrainUntil(projectID, id, version, time.Now().Add(DefaultDrainWindow))
}

func (r *Registry) DrainUntil(projectID, id, version string, deadline time.Time) error {
	key := projectID + ":" + id + "@" + version
	i, ok := r.installations[key]
	if !ok {
		return ErrNotInstalled
	}
	i.Draining = true
	i.DrainUntil = deadline
	r.installations[key] = i
	r.audit = append(r.audit, AuditEvent{Action: "drain", ProjectID: projectID, PluginID: id, Version: version})
	return nil
}

func (r *Registry) StopImmediately(projectID, id, version string) error {
	key := projectID + ":" + id + "@" + version
	i, ok := r.installations[key]
	if !ok {
		return ErrNotInstalled
	}
	i.Draining = true
	i.DrainUntil = time.Unix(0, 0)
	r.installations[key] = i
	r.lifecycleRevision++
	r.audit = append(r.audit, AuditEvent{Action: "stop_immediate", ProjectID: projectID, PluginID: id, Version: version})
	return nil
}

func (r *Registry) CanStart(projectID, id, version string, now time.Time) error {
	i, ok := r.installations[projectID+":"+id+"@"+version]
	if !ok {
		return ErrNotInstalled
	}
	if i.Draining && (i.DrainUntil.IsZero() || !now.Before(i.DrainUntil)) {
		return ErrDisabled
	}
	return nil
}

func (r *Registry) CanStartNew(projectID, id, version string) error {
	i, ok := r.installations[projectID+":"+id+"@"+version]
	if !ok {
		return ErrNotInstalled
	}
	if i.Draining {
		return ErrDisabled
	}
	return nil
}

func (r *Registry) CanReconnect(projectID, id, version string, now time.Time) error {
	i, ok := r.installations[projectID+":"+id+"@"+version]
	if !ok {
		return ErrNotInstalled
	}
	if i.Draining && !i.DrainUntil.IsZero() && !now.Before(i.DrainUntil) {
		return ErrDisabled
	}
	return nil
}

func (r *Registry) Rollback(projectID, id, version string) error {
	if err := r.Promote(projectID, id, version); err != nil {
		return err
	}
	r.audit = append(r.audit, AuditEvent{Action: "rollback", ProjectID: projectID, PluginID: id, Version: version})
	return nil
}

func (r *Registry) ActiveVersion(projectID, id string) (string, bool) {
	version, ok := r.active[projectID+":"+id]
	return version, ok
}

func (r *Registry) PinWork(workID, projectID, id, version string) error {
	if workID == "" {
		return ErrProjectInvalid
	}
	key := projectID + ":" + id + "@" + version
	i, ok := r.installations[key]
	if !ok || !i.Enabled {
		return ErrNotInstalled
	}
	r.pins[workID] = key
	r.audit = append(r.audit, AuditEvent{Action: "pin", ProjectID: projectID, PluginID: id, Version: version, Detail: workID})
	return nil
}

func (r *Registry) PinnedVersion(workID string) (string, bool) {
	key, ok := r.pins[workID]
	if !ok {
		return "", false
	}
	for _, installation := range r.installations {
		if installation.ProjectID+":"+installation.PluginID+"@"+installation.Version == key {
			return installation.Version, true
		}
	}
	return "", false
}

func (r *Registry) Remove(projectID, id, version string) error {
	key := projectID + ":" + id + "@" + version
	if _, ok := r.installations[key]; !ok {
		return ErrNotInstalled
	}
	for _, pinned := range r.pins {
		if pinned == key {
			return ErrVersionReferenced
		}
	}
	if r.active[projectID+":"+id] == version {
		return ErrVersionReferenced
	}
	delete(r.installations, key)
	r.audit = append(r.audit, AuditEvent{Action: "remove", ProjectID: projectID, PluginID: id, Version: version})
	return nil
}

func (r *Registry) UninstallProject(projectID string) error {
	for _, pinnedVersion := range r.pins {
		if strings.HasPrefix(pinnedVersion, projectID+":") {
			return ErrActiveWork
		}
	}
	for key := range r.active {
		if strings.HasPrefix(key, projectID+":") {
			return ErrActiveWork
		}
	}
	for key, installation := range r.installations {
		if installation.ProjectID == projectID {
			delete(r.installations, key)
		}
	}
	r.audit = append(r.audit, AuditEvent{Action: "uninstall_project", ProjectID: projectID})
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
