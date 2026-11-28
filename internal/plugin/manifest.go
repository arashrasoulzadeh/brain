package plugin

import (
	"fmt"
	"os"
	"path"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

type Manifest struct {
	Title                string               `yaml:"title"`
	Description          string               `yaml:"description"`
	ID                   string               `yaml:"id"`
	Mount                string               `yaml:"mount,omitempty"`
	Entry                string               `yaml:"entrypoint"`
	Declarative          bool                 `yaml:"declarative,omitempty"`
	Version              string               `yaml:"version"`
	HostAPI              int                  `yaml:"host_api"`
	Dependencies         []string             `yaml:"dependencies"`
	OptionalDependencies []string             `yaml:"optional_dependencies,omitempty"`
	Capabilities         []string             `yaml:"capabilities"`
	ExtensionPoints      []string             `yaml:"extension_points"`
	ConfigSchema         map[string]string    `yaml:"config_schema"`
	PublicConfig         []string             `yaml:"public_config,omitempty" json:",omitempty"`
	Routes               []Route              `yaml:"routes"`
	SourceFiles          []string             `yaml:"source_files"`
	Imports              []string             `yaml:"imports"`
	PluginImports        []string             `yaml:"plugin_imports"`
	Events               []EventSpec          `yaml:"events"`
	Observers            []ObserverSpec       `yaml:"observe"`
	Models               map[string]ModelSpec `yaml:"models,omitempty"`
	Admin                *AdminSpec           `yaml:"admin,omitempty" json:",omitempty"`
}

type Route struct {
	Method     string        `yaml:"method,omitempty"`
	Path       string        `yaml:"path,omitempty"`
	Auth       string        `yaml:"auth,omitempty"`
	Operation  string        `yaml:"operation"`
	Function   string        `yaml:"function"`
	Capability string        `yaml:"capability"`
	Helper     string        `yaml:"helper,omitempty"`
	Response   *ResponseSpec `yaml:"response,omitempty"`
	Pipeline   *PipelineSpec `yaml:"pipeline,omitempty"`
	Model      *ModelRoute   `yaml:"model,omitempty"`
	Call       *PluginCall   `yaml:"call,omitempty"`

	// ResourceField binds resource-scoped policy to a manifest-declared payload field.
	ResourceField      string `yaml:"resource_field,omitempty" json:",omitempty"`
	WebSocketAdmission string `yaml:"websocket_admission,omitempty" json:",omitempty"`
	FunctionRead       bool   `yaml:"function_read,omitempty" json:",omitempty"`
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
	public := map[string]bool{}
	for _, name := range manifest.PublicConfig {
		kind := manifest.ConfigSchema[name]
		if public[name] || (kind != "int" && kind != "bool" && kind != "string") {
			return ErrInvalidManifest
		}
		public[name] = true
	}
	if manifest.Declarative && (manifest.Entry != "" || len(manifest.SourceFiles) != 0 || len(manifest.Imports) != 0 || len(manifest.PluginImports) != 0 || len(manifest.Observers) != 0) {
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
	if !manifest.Declarative && !seenFiles[manifest.EntryPoint()] {
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
	if !validOptionalDependencies(manifest) || !validModels(manifest.Models) || !validHTTPRoutes(manifest) {
		return ErrInvalidManifest
	}
	seenRoutes := map[string]bool{}
	for _, route := range manifest.Routes {
		if manifest.Declarative && route.Helper != "sdk.response" && route.Helper != "sdk.pipeline" && route.Helper != "host.model" && route.Helper != "host.call" {
			return ErrInvalidManifest
		}
		if strings.HasPrefix(route.Operation, observationPrefix) || route.Operation == "" || route.Function == "" || route.Capability == "" || seenRoutes[route.Operation] || !declaredCapabilities[route.Capability] || route.ExecutionBoundary() == BoundaryInvalid {
			return ErrInvalidManifest
		}
		if route.Helper == "host.admin" && manifest.ID != "dashboard" {
			return ErrInvalidManifest
		}
		if (route.Helper == "sdk.response" && !route.Response.valid()) || (route.Helper != "sdk.response" && route.Response != nil) {
			return ErrInvalidManifest
		}
		if (route.Helper == "sdk.pipeline" && !route.Pipeline.valid()) || (route.Helper != "sdk.pipeline" && route.Pipeline != nil) {
			return ErrInvalidManifest
		}
		if (route.Helper == "host.model" && !route.Model.valid(manifest.Models)) || (route.Helper != "host.model" && route.Model != nil) {
			return ErrInvalidManifest
		}
		if (route.Helper == "host.call" && !route.Call.valid(manifest, route.Capability)) || (route.Helper != "host.call" && route.Call != nil) {
			return ErrInvalidManifest
		}
		if route.ResourceField != "" && (!validResponseKey(route.ResourceField) || route.Model != nil || route.Call != nil) {
			return ErrInvalidManifest
		}
		if route.FunctionRead && !validFunctionReadRoute(manifest, route) {
			return ErrInvalidManifest
		}
		seenRoutes[route.Operation] = true
	}
	if err := validateAdminSpec(manifest); err != nil {
		return err
	}
	return validateObservationContract(manifest, declaredCapabilities)
}

func (m Manifest) EntryPoint() string {
	if m.Declarative {
		return ""
	}
	if m.Entry == "" {
		return "index.ts"
	}
	return m.Entry
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
		if route.Method != "" {
			fmt.Fprintf(&b, "  - HTTP: `%s %s%s` (auth: `%s`)\n", route.Method, m.Mount, route.Path, route.Auth)
		}
		if route.WebSocketAdmission != "" {
			fmt.Fprintf(&b, "  - WebSocket admission: `%s`\n", route.WebSocketAdmission)
		}
	}
	if len(m.Models) > 0 {
		b.WriteString("\n## Models\n\n")
		for _, name := range sortedContractKeys(m.Models) {
			model := m.Models[name]
			fmt.Fprintf(&b, "### %s (schema %d)\n\n", name, model.Version)
			for _, field := range sortedContractKeys(model.Fields) {
				fmt.Fprintf(&b, "- `%s`: %s\n", field, model.Fields[field])
			}
			for _, action := range sortedContractKeys(model.Transitions) {
				transition := model.Transitions[action]
				fmt.Fprintf(&b, "- Transition `%s`: %s → %s\n", action, transition.From, transition.To)
			}
		}
	}
	if len(m.Events) > 0 {
		fmt.Fprint(&b, "\n## Published events\n\n")
		for _, event := range m.Events {
			source := event.Source
			if source == "" {
				source = "worker"
			}
			fields := make([]string, 0, len(event.Fields))
			for field, kind := range event.Fields {
				fields = append(fields, field+": "+kind)
			}
			sort.Strings(fields)
			fmt.Fprintf(&b, "- `%s.%s` (%s): %s\n", m.ID, event.Name, source, strings.Join(fields, ", "))
		}
	}
	if len(m.Observers) > 0 {
		fmt.Fprint(&b, "\n## Observations\n\n")
		for _, observer := range m.Observers {
			fmt.Fprintf(&b, "- `%s.%s` → `%s` (capability: `%s`, persist: %t)\n", observer.Plugin, observer.Event, observer.Function, observer.Capability, observer.Persist)
		}
	}
	return b.String()
}

func sortedContractKeys[V any](values map[string]V) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
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
