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
	"sync"
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
		if !validRelativeManifestPath(file) {
			return ErrInvalidManifest
		}
		if seenFiles[file] {
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
	declaredCapabilities := map[string]bool{}
	for _, capability := range manifest.Capabilities {
		if capability == "" || declaredCapabilities[capability] {
			return ErrInvalidManifest
		}
		declaredCapabilities[capability] = true
	}
	seenRoutes := map[string]bool{}
	for _, route := range manifest.Routes {
		if route.Operation == "" || route.Function == "" || route.Capability == "" || seenRoutes[route.Operation] || !declaredCapabilities[route.Capability] {
			return ErrInvalidManifest
		}
		seenRoutes[route.Operation] = true
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
var sourceDynamicImportPattern = regexp.MustCompile(`(?m)\bimport\s*\(\s*["']([^"']+)["']\s*\)`)
var sourceRequirePattern = regexp.MustCompile(`(?m)\brequire\s*\(\s*["']([^"']+)["']\s*\)`)

func validRelativeManifestPath(candidate string) bool {
	if candidate == "" || strings.HasPrefix(candidate, "/") || strings.ContainsAny(candidate, "\\\x00") {
		return false
	}
	clean := path.Clean(candidate)
	if clean == "." || clean == ".." || strings.HasPrefix(clean, "../") {
		return false
	}
	return clean == candidate
}

func containsSymlink(root, relativePath string) bool {
	current := root
	for _, component := range strings.Split(relativePath, string(filepath.Separator)) {
		current = filepath.Join(current, component)
		info, err := os.Lstat(current)
		if err != nil {
			return false
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return true
		}
	}
	return false
}

func ValidatePluginSources(directory string, manifest Manifest) error {
	if err := ValidateManifestContract(manifest); err != nil {
		return err
	}
	allowedImports := map[string]bool{}
	for _, importPath := range manifest.Imports {
		allowedImports[importPath] = true
	}
	for _, sourceFile := range manifest.SourceFiles {
		if !validRelativeManifestPath(sourceFile) {
			return ErrInvalidManifest
		}
		path := filepath.Join(directory, sourceFile)
		if containsSymlink(directory, sourceFile) {
			return ErrInvalidManifest
		}
		info, err := os.Lstat(path)
		if err != nil || !info.Mode().IsRegular() {
			return ErrInvalidManifest
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return ErrInvalidManifest
		}
		matches := sourceImportPattern.FindAllStringSubmatch(string(data), -1)
		matches = append(matches, sourceExportPattern.FindAllStringSubmatch(string(data), -1)...)
		matches = append(matches, sourceDynamicImportPattern.FindAllStringSubmatch(string(data), -1)...)
		matches = append(matches, sourceRequirePattern.FindAllStringSubmatch(string(data), -1)...)
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

func (m Manifest) SupportsExtensionPoint(point string) bool {
	for _, declared := range m.ExtensionPoints {
		if declared == point {
			return true
		}
	}
	return false
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
	installationMu    sync.RWMutex
	readiness         map[string]bool
	readinessMu       sync.RWMutex
	grantMu           sync.RWMutex
	approvalMu        sync.RWMutex
	audit             []AuditEvent
	artifactMu        sync.RWMutex
	auditMu           sync.RWMutex
	active            map[string]string
	pins              map[string]string
	artifacts         map[string]string
	approvals         map[string]bool
	lifecycleRevision uint64
	lifecycleMu       sync.RWMutex
	durableData       map[string][]byte
	durableMu         sync.RWMutex
	supervisor        *Supervisor
	webSockets        *WebSocketRegistry
}

func (r *Registry) BindSupervisor(supervisor *Supervisor) {
	r.supervisor = supervisor
}

func (r *Registry) BindWebSocketRegistry(webSockets *WebSocketRegistry) {
	r.webSockets = webSockets
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
	return &Registry{HostAPI: hostAPI, registered: map[string]Registered{}, installations: map[string]Installation{}, readiness: map[string]bool{}, active: map[string]string{}, pins: map[string]string{}, artifacts: map[string]string{}, approvals: map[string]bool{}, durableData: map[string][]byte{}}
}

func (r *Registry) RegisteredManifests() []Manifest {
	r.readinessMu.RLock()
	defer r.readinessMu.RUnlock()
	manifests := make([]Manifest, 0, len(r.registered))
	for _, registered := range r.registered {
		manifests = append(manifests, registered.Manifest)
	}
	sort.Slice(manifests, func(left, right int) bool {
		if manifests[left].ID == manifests[right].ID {
			return manifests[left].Version < manifests[right].Version
		}
		return manifests[left].ID < manifests[right].ID
	})
	return manifests
}

func (r *Registry) InstallationStatuses(projectID string) []InstallationStatus {
	r.installationMu.RLock()
	keys := make([]string, 0, len(r.installations))
	for key, installation := range r.installations {
		if projectID == "" || installation.ProjectID == projectID {
			keys = append(keys, key)
		}
	}
	installations := make([]Installation, 0, len(keys))
	for _, key := range keys {
		installations = append(installations, r.installations[key])
	}
	r.installationMu.RUnlock()
	sort.Slice(installations, func(left, right int) bool {
		if installations[left].PluginID == installations[right].PluginID {
			return installations[left].Version < installations[right].Version
		}
		return installations[left].PluginID < installations[right].PluginID
	})
	statuses := make([]InstallationStatus, 0, len(installations))
	for _, installation := range installations {
		status, err := r.Status(installation.ProjectID, installation.PluginID, installation.Version)
		if err == nil {
			statuses = append(statuses, status)
		}
	}
	return statuses
}

func (r *Registry) SaveDurableData(projectID, pluginID, key string, value []byte) error {
	if !isValidProjectID(projectID) || pluginID == "" || key == "" {
		return ErrProjectInvalid
	}
	r.durableMu.Lock()
	defer r.durableMu.Unlock()
	r.durableData[projectID+":"+pluginID+":"+key] = append([]byte(nil), value...)
	return nil
}

func (r *Registry) DurableData(projectID, pluginID, key string) ([]byte, bool) {
	r.durableMu.RLock()
	defer r.durableMu.RUnlock()
	value, ok := r.durableData[projectID+":"+pluginID+":"+key]
	return append([]byte(nil), value...), ok
}

func (r *Registry) DeleteDurableData(projectID, pluginID, key string) error {
	if !isValidProjectID(projectID) || pluginID == "" || key == "" {
		return ErrProjectInvalid
	}
	r.durableMu.Lock()
	defer r.durableMu.Unlock()
	delete(r.durableData, projectID+":"+pluginID+":"+key)
	r.appendAudit(AuditEvent{Action: "delete_durable_data", ProjectID: projectID, PluginID: pluginID, Detail: key})
	return nil
}

func (r *Registry) ApproveCapability(projectID, id, version, capability string) error {
	r.readinessMu.RLock()
	defer r.readinessMu.RUnlock()
	r.approvalMu.Lock()
	defer r.approvalMu.Unlock()
	r.installationMu.RLock()
	defer r.installationMu.RUnlock()
	if !isValidProjectID(projectID) || capability == "" {
		return ErrProjectInvalid
	}
	if _, ok := r.installations[projectID+":"+id+"@"+version]; !ok {
		return ErrNotInstalled
	}
	declared := false
	for _, requested := range r.registered[id+"@"+version].Manifest.Capabilities {
		if requested == capability {
			declared = true
			break
		}
	}
	if !declared {
		return ErrGrantDenied
	}
	r.approvals[projectID+":"+id+"@"+version+":"+capability] = true
	r.appendAudit(AuditEvent{Action: "approve_capability", ProjectID: projectID, PluginID: id, Version: version, Detail: capability})
	return nil
}

func (r *Registry) Status(projectID, pluginID, version string) (InstallationStatus, error) {
	if !isValidProjectID(projectID) {
		return InstallationStatus{}, ErrProjectInvalid
	}
	r.readinessMu.RLock()
	defer r.readinessMu.RUnlock()
	r.installationMu.RLock()
	defer r.installationMu.RUnlock()
	installation, ok := r.installations[projectID+":"+pluginID+"@"+version]
	if !ok {
		return InstallationStatus{}, ErrNotInstalled
	}
	ready := r.readiness[pluginID+"@"+version]
	return InstallationStatus{ProjectID: projectID, PluginID: pluginID, Version: version, Enabled: installation.Enabled, Ready: ready, Draining: installation.Draining, DrainUntil: installation.DrainUntil}, nil
}

func (r *Registry) ApproveArtifact(id, version, endpoint string) error {
	r.readinessMu.RLock()
	defer r.readinessMu.RUnlock()
	if _, ok := r.registered[id+"@"+version]; !ok {
		return ErrNotRegistered
	}
	u, err := url.Parse(endpoint)
	if err != nil || u.Scheme != "https" || u.Host == "" {
		return ErrInvalidManifest
	}
	r.artifactMu.Lock()
	defer r.artifactMu.Unlock()
	r.artifacts[id+"@"+version] = endpoint
	r.appendAudit(AuditEvent{Action: "approve_artifact", PluginID: id, Version: version, Detail: endpoint})
	return nil
}

func (r *Registry) ApprovedArtifact(id, version string) (string, bool) {
	r.artifactMu.RLock()
	defer r.artifactMu.RUnlock()
	endpoint, ok := r.artifacts[id+"@"+version]
	return endpoint, ok
}

func (r *Registry) Register(m Manifest) error {
	validationManifest := m
	if len(validationManifest.SourceFiles) == 0 && validationManifest.Entry == "" && len(validationManifest.Imports) == 0 && len(validationManifest.PluginImports) == 0 {
		validationManifest.Entry = "index.ts"
		validationManifest.SourceFiles = []string{"index.ts"}
	}
	if err := ValidateManifestContract(validationManifest); err != nil {
		return err
	}
	r.readinessMu.Lock()
	defer r.readinessMu.Unlock()
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
	r.appendAudit(AuditEvent{Action: "register", PluginID: m.ID, Version: m.Version})
	return nil
}

func (r *Registry) TransitiveDependents(id, version string) []string {
	r.readinessMu.RLock()
	defer r.readinessMu.RUnlock()
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
	r.lifecycleMu.Lock()
	defer r.lifecycleMu.Unlock()
	r.installationMu.RLock()
	defer r.installationMu.RUnlock()
	if !isValidProjectID(projectID) {
		return DisablePreview{}, ErrProjectInvalid
	}
	if _, ok := r.installations[projectID+":"+id+"@"+version]; !ok {
		return DisablePreview{}, ErrNotInstalled
	}
	return DisablePreview{ProjectID: projectID, PluginID: id, Version: version, Revision: r.lifecycleRevision, Dependents: r.TransitiveDependents(id, version)}, nil
}

func (r *Registry) ApplyDisablePreview(preview DisablePreview) error {
	r.lifecycleMu.Lock()
	defer r.lifecycleMu.Unlock()
	r.installationMu.RLock()
	defer r.installationMu.RUnlock()
	if preview.ProjectID == "" {
		return ErrProjectInvalid
	}
	if preview.Revision != r.lifecycleRevision {
		return ErrStalePreview
	}
	if _, ok := r.installations[preview.ProjectID+":"+preview.PluginID+"@"+preview.Version]; !ok {
		return ErrNotInstalled
	}
	var transitionError error
	if r.supervisor != nil {
		if err := r.supervisor.StopProjectPlugin(preview.ProjectID, preview.PluginID, preview.Version); err != nil && !errors.Is(err, ErrWorkerNotRunning) {
			transitionError = err
		}
	}
	if r.webSockets != nil {
		if err := r.webSockets.CloseProject(preview.ProjectID); err != nil && transitionError == nil {
			transitionError = err
		}
	}
	r.lifecycleRevision++
	return transitionError
}

func (r *Registry) SetReadiness(id, version string, ready bool) error {
	r.readinessMu.Lock()
	defer r.readinessMu.Unlock()
	key := id + "@" + version
	if _, ok := r.registered[key]; !ok {
		return ErrNotRegistered
	}
	r.readiness[key] = ready
	r.appendAudit(AuditEvent{Action: "readiness", PluginID: id, Version: version, Detail: fmt.Sprintf("ready=%t", ready)})
	return nil
}

func (r *Registry) Install(projectID, id, version string) error {
	r.readinessMu.RLock()
	defer r.readinessMu.RUnlock()
	r.installationMu.Lock()
	defer r.installationMu.Unlock()
	if !isValidProjectID(projectID) {
		return ErrProjectInvalid
	}
	key := id + "@" + version
	if _, ok := r.registered[key]; !ok {
		return ErrNotRegistered
	}
	installKey := projectID + ":" + key
	if _, exists := r.installations[installKey]; exists {
		return ErrDuplicate
	}
	r.installations[installKey] = Installation{ProjectID: projectID, PluginID: id, Version: version, Grants: map[string]bool{}, Resources: map[string]map[string]bool{}, Config: map[string]any{}}
	r.appendAudit(AuditEvent{Action: "install", ProjectID: projectID, PluginID: id, Version: version})
	return nil
}

func (r *Registry) GrantResource(projectID, id, version, capability, resource string) error {
	r.grantMu.Lock()
	defer r.grantMu.Unlock()
	r.readinessMu.RLock()
	defer r.readinessMu.RUnlock()
	r.installationMu.Lock()
	defer r.installationMu.Unlock()
	if !isValidProjectID(projectID) {
		return ErrProjectInvalid
	}
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
	r.appendAudit(AuditEvent{Action: "grant_resource", ProjectID: projectID, PluginID: id, Version: version, Detail: capability + ":" + resource})
	return nil
}

func (r *Registry) AuthorizeResource(projectID, id, version, capability, resource string) error {
	r.grantMu.RLock()
	defer r.grantMu.RUnlock()
	if resource == "" {
		return ErrGrantDenied
	}
	r.installationMu.RLock()
	defer r.installationMu.RUnlock()
	if err := r.authorizeGrant(projectID, id, version, capability); err != nil {
		return err
	}
	i := r.installations[projectID+":"+id+"@"+version]
	if resources := i.Resources[capability]; len(resources) > 0 && !resources[resource] {
		return ErrGrantDenied
	}
	return nil
}

func (r *Registry) Configure(projectID, id, version string, values map[string]any) error {
	r.readinessMu.RLock()
	defer r.readinessMu.RUnlock()
	r.installationMu.Lock()
	defer r.installationMu.Unlock()
	if !isValidProjectID(projectID) {
		return ErrProjectInvalid
	}
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
	r.appendAudit(AuditEvent{Action: "configure", ProjectID: projectID, PluginID: id, Version: version})
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
	r.grantMu.Lock()
	defer r.grantMu.Unlock()
	r.readinessMu.RLock()
	defer r.readinessMu.RUnlock()
	r.installationMu.Lock()
	defer r.installationMu.Unlock()
	if !isValidProjectID(projectID) {
		return ErrProjectInvalid
	}
	i, ok := r.installations[projectID+":"+id+"@"+version]
	if !ok {
		return ErrNotInstalled
	}
	for _, requested := range r.registered[id+"@"+version].Manifest.Capabilities {
		if requested == capability {
			i.Grants[capability] = true
			r.installations[projectID+":"+id+"@"+version] = i
			r.appendAudit(AuditEvent{Action: "grant", ProjectID: projectID, PluginID: id, Version: version, Detail: capability})
			return nil
		}
	}
	return ErrGrantDenied
}

func (r *Registry) Revoke(projectID, id, version, capability string) error {
	r.grantMu.Lock()
	defer r.grantMu.Unlock()
	r.installationMu.Lock()
	defer r.installationMu.Unlock()
	if !isValidProjectID(projectID) {
		return ErrProjectInvalid
	}
	if capability == "" {
		return ErrGrantDenied
	}
	key := projectID + ":" + id + "@" + version
	i, ok := r.installations[key]
	if !ok {
		return ErrNotInstalled
	}
	if !i.Grants[capability] {
		return ErrGrantDenied
	}
	delete(i.Grants, capability)
	delete(i.Resources, capability)
	r.installations[key] = i
	r.appendAudit(AuditEvent{Action: "revoke", ProjectID: projectID, PluginID: id, Version: version, Detail: capability})
	var transitionError error
	if r.supervisor != nil {
		if err := r.supervisor.StopProjectPlugin(projectID, id, version); err != nil && !errors.Is(err, ErrWorkerNotRunning) {
			transitionError = err
		}
	}
	if r.webSockets != nil {
		if err := r.webSockets.CloseProject(projectID); err != nil && transitionError == nil {
			transitionError = err
		}
	}
	return transitionError
}

func (r *Registry) Enable(projectID, id, version string) error {
	r.grantMu.Lock()
	defer r.grantMu.Unlock()
	r.readinessMu.RLock()
	defer r.readinessMu.RUnlock()
	r.installationMu.Lock()
	defer r.installationMu.Unlock()
	if !isValidProjectID(projectID) {
		return ErrProjectInvalid
	}
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
	r.appendAudit(AuditEvent{Action: "enable", ProjectID: projectID, PluginID: id, Version: version})
	return nil
}

func (r *Registry) Promote(projectID, id, version string) error {
	r.lifecycleMu.Lock()
	defer r.lifecycleMu.Unlock()
	r.readinessMu.RLock()
	defer r.readinessMu.RUnlock()
	r.approvalMu.RLock()
	defer r.approvalMu.RUnlock()
	r.installationMu.RLock()
	defer r.installationMu.RUnlock()
	if !isValidProjectID(projectID) {
		return ErrProjectInvalid
	}
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
	r.appendAudit(AuditEvent{Action: "promote", ProjectID: projectID, PluginID: id, Version: version})
	return nil
}

func (r *Registry) Drain(projectID, id, version string) error {
	return r.DrainUntil(projectID, id, version, time.Now().Add(DefaultDrainWindow))
}

func (r *Registry) DrainUntil(projectID, id, version string, deadline time.Time) error {
	r.lifecycleMu.Lock()
	defer r.lifecycleMu.Unlock()
	r.installationMu.Lock()
	defer r.installationMu.Unlock()
	if !isValidProjectID(projectID) {
		return ErrProjectInvalid
	}
	key := projectID + ":" + id + "@" + version
	i, ok := r.installations[key]
	if !ok {
		return ErrNotInstalled
	}
	i.Draining = true
	i.DrainUntil = deadline
	r.installations[key] = i
	r.appendAudit(AuditEvent{Action: "drain", ProjectID: projectID, PluginID: id, Version: version})
	return nil
}

func (r *Registry) StopImmediately(projectID, id, version string) error {
	r.lifecycleMu.Lock()
	defer r.lifecycleMu.Unlock()
	r.installationMu.Lock()
	defer r.installationMu.Unlock()
	if !isValidProjectID(projectID) {
		return ErrProjectInvalid
	}
	key := projectID + ":" + id + "@" + version
	i, ok := r.installations[key]
	if !ok {
		return ErrNotInstalled
	}
	i.Draining = true
	i.DrainUntil = time.Unix(0, 0)
	r.installations[key] = i
	r.lifecycleRevision++
	r.appendAudit(AuditEvent{Action: "stop_immediate", ProjectID: projectID, PluginID: id, Version: version})
	var stopError error
	if r.supervisor != nil {
		if err := r.supervisor.StopProjectPlugin(projectID, id, version); err != nil && !errors.Is(err, ErrWorkerNotRunning) {
			stopError = err
		}
	}
	if r.webSockets != nil {
		if err := r.webSockets.CloseProject(projectID); err != nil && stopError == nil {
			stopError = err
		}
	}
	return stopError
}

func (r *Registry) Recover(projectID, id, version string) error {
	r.lifecycleMu.Lock()
	defer r.lifecycleMu.Unlock()
	r.installationMu.Lock()
	defer r.installationMu.Unlock()
	if !isValidProjectID(projectID) {
		return ErrProjectInvalid
	}
	key := projectID + ":" + id + "@" + version
	i, ok := r.installations[key]
	if !ok {
		return ErrNotInstalled
	}
	var recoveryError error
	if r.supervisor != nil {
		if err := r.supervisor.Recover(key); err != nil && !errors.Is(err, ErrWorkerNotRunning) {
			recoveryError = err
		}
	}
	if r.webSockets != nil {
		if err := r.webSockets.CloseProject(projectID); err != nil && recoveryError == nil {
			recoveryError = err
		}
	}
	if recoveryError != nil {
		return recoveryError
	}
	i.Draining = false
	i.DrainUntil = time.Time{}
	r.installations[key] = i
	r.lifecycleRevision++
	r.appendAudit(AuditEvent{Action: "recover", ProjectID: projectID, PluginID: id, Version: version})
	return nil
}

func (r *Registry) CanStart(projectID, id, version string, now time.Time) error {
	r.installationMu.RLock()
	defer r.installationMu.RUnlock()
	if !isValidProjectID(projectID) {
		return ErrProjectInvalid
	}
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
	r.installationMu.RLock()
	defer r.installationMu.RUnlock()
	if !isValidProjectID(projectID) {
		return ErrProjectInvalid
	}
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
	r.installationMu.RLock()
	defer r.installationMu.RUnlock()
	if !isValidProjectID(projectID) {
		return ErrProjectInvalid
	}
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
	r.appendAudit(AuditEvent{Action: "rollback", ProjectID: projectID, PluginID: id, Version: version})
	return nil
}

func (r *Registry) ActiveVersion(projectID, id string) (string, bool) {
	r.lifecycleMu.RLock()
	defer r.lifecycleMu.RUnlock()
	if !isValidProjectID(projectID) {
		return "", false
	}
	version, ok := r.active[projectID+":"+id]
	return version, ok
}

func (r *Registry) PinWork(workID, projectID, id, version string) error {
	r.lifecycleMu.Lock()
	defer r.lifecycleMu.Unlock()
	r.installationMu.RLock()
	defer r.installationMu.RUnlock()
	if workID == "" || !isValidProjectID(projectID) {
		return ErrProjectInvalid
	}
	key := projectID + ":" + id + "@" + version
	i, ok := r.installations[key]
	if !ok || !i.Enabled {
		return ErrNotInstalled
	}
	r.pins[workID] = key
	r.appendAudit(AuditEvent{Action: "pin", ProjectID: projectID, PluginID: id, Version: version, Detail: workID})
	return nil
}

func (r *Registry) PinnedVersion(workID string) (string, bool) {
	r.lifecycleMu.RLock()
	defer r.lifecycleMu.RUnlock()
	r.installationMu.RLock()
	defer r.installationMu.RUnlock()
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
	r.lifecycleMu.Lock()
	defer r.lifecycleMu.Unlock()
	r.installationMu.Lock()
	defer r.installationMu.Unlock()
	if !isValidProjectID(projectID) {
		return ErrProjectInvalid
	}
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
	var removalError error
	if r.supervisor != nil {
		if err := r.supervisor.StopProjectPlugin(projectID, id, version); err != nil && !errors.Is(err, ErrWorkerNotRunning) {
			removalError = err
		}
	}
	delete(r.installations, key)
	r.lifecycleRevision++
	r.appendAudit(AuditEvent{Action: "remove", ProjectID: projectID, PluginID: id, Version: version})
	if r.webSockets != nil {
		if err := r.webSockets.CloseProject(projectID); err != nil && removalError == nil {
			removalError = err
		}
	}
	return removalError
}

func (r *Registry) Unregister(id, version string) error {
	r.lifecycleMu.Lock()
	defer r.lifecycleMu.Unlock()
	r.readinessMu.Lock()
	defer r.readinessMu.Unlock()
	r.artifactMu.Lock()
	defer r.artifactMu.Unlock()
	r.approvalMu.Lock()
	defer r.approvalMu.Unlock()
	r.installationMu.RLock()
	defer r.installationMu.RUnlock()
	key := id + "@" + version
	if _, ok := r.registered[key]; !ok {
		return ErrNotRegistered
	}
	for _, installation := range r.installations {
		if installation.PluginID == id && installation.Version == version {
			return ErrVersionReferenced
		}
	}
	for activeKey, activeVersion := range r.active {
		if strings.HasSuffix(activeKey, ":"+id) && activeVersion == version {
			return ErrVersionReferenced
		}
	}
	for _, pinned := range r.pins {
		if pinned == "" {
			continue
		}
		if strings.HasSuffix(pinned, ":"+key) {
			return ErrVersionReferenced
		}
	}
	for _, registered := range r.registered {
		for _, dependency := range registered.Manifest.Dependencies {
			if dependency == key {
				return ErrVersionReferenced
			}
		}
	}
	delete(r.registered, key)
	delete(r.readiness, key)
	delete(r.artifacts, key)
	for approvalKey := range r.approvals {
		if strings.Contains(approvalKey, ":"+key+":") || strings.HasPrefix(approvalKey, key+":") {
			delete(r.approvals, approvalKey)
		}
	}
	r.appendAudit(AuditEvent{Action: "unregister", PluginID: id, Version: version})
	return nil
}

func (r *Registry) UninstallProject(projectID string) error {
	r.lifecycleMu.Lock()
	defer r.lifecycleMu.Unlock()
	r.installationMu.Lock()
	defer r.installationMu.Unlock()
	if !isValidProjectID(projectID) {
		return ErrProjectInvalid
	}
	var uninstallError error
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
			if r.supervisor != nil {
				if err := r.supervisor.StopProjectPlugin(projectID, installation.PluginID, installation.Version); err != nil && !errors.Is(err, ErrWorkerNotRunning) && uninstallError == nil {
					uninstallError = err
				}
			}
			for capability := range installation.Grants {
				r.appendAudit(AuditEvent{Action: "revoke_project_grant", ProjectID: projectID, PluginID: installation.PluginID, Version: installation.Version, Detail: capability})
			}
			delete(r.installations, key)
		}
	}
	r.lifecycleRevision++
	r.appendAudit(AuditEvent{Action: "uninstall_project", ProjectID: projectID})
	if r.webSockets != nil {
		if err := r.webSockets.CloseProject(projectID); err != nil && uninstallError == nil {
			uninstallError = err
		}
	}
	return uninstallError
}

func (r *Registry) appendAudit(event AuditEvent) {
	r.auditMu.Lock()
	defer r.auditMu.Unlock()
	r.audit = append(r.audit, event)
}

func (r *Registry) Audit() []AuditEvent {
	r.auditMu.RLock()
	defer r.auditMu.RUnlock()
	return append([]AuditEvent(nil), r.audit...)
}

func (r *Registry) Authorize(projectID, id, version, capability string) error {
	r.grantMu.RLock()
	defer r.grantMu.RUnlock()
	return r.authorizeGrant(projectID, id, version, capability)
}

func (r *Registry) authorizeGrant(projectID, id, version, capability string) error {
	r.installationMu.RLock()
	defer r.installationMu.RUnlock()
	if !isValidProjectID(projectID) {
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
