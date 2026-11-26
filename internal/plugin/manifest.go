package plugin

import (
	"fmt"
	"os"
	"path"
	"strings"

	"gopkg.in/yaml.v3"
)

type Manifest struct {
	Title           string               `yaml:"title"`
	Description     string               `yaml:"description"`
	ID              string               `yaml:"id"`
	Entry           string               `yaml:"entrypoint"`
	Version         string               `yaml:"version"`
	HostAPI         int                  `yaml:"host_api"`
	Dependencies    []string             `yaml:"dependencies"`
	Capabilities    []string             `yaml:"capabilities"`
	ExtensionPoints []string             `yaml:"extension_points"`
	ConfigSchema    map[string]string    `yaml:"config_schema"`
	Routes          []Route              `yaml:"routes"`
	SourceFiles     []string             `yaml:"source_files"`
	Imports         []string             `yaml:"imports"`
	PluginImports   []string             `yaml:"plugin_imports"`
	Events          []EventSpec          `yaml:"events"`
	Observers       []ObserverSpec       `yaml:"observe"`
	Models          map[string]ModelSpec `yaml:"models,omitempty"`
}

type Route struct {
	Operation  string        `yaml:"operation"`
	Function   string        `yaml:"function"`
	Capability string        `yaml:"capability"`
	Helper     string        `yaml:"helper,omitempty"`
	Response   *ResponseSpec `yaml:"response,omitempty"`
	Pipeline   *PipelineSpec `yaml:"pipeline,omitempty"`
	Model      *ModelRoute   `yaml:"model,omitempty"`
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
	if !validModels(manifest.Models) {
		return ErrInvalidManifest
	}
	seenRoutes := map[string]bool{}
	for _, route := range manifest.Routes {
		if strings.HasPrefix(route.Operation, observationPrefix) || route.Operation == "" || route.Function == "" || route.Capability == "" || seenRoutes[route.Operation] || !declaredCapabilities[route.Capability] || (route.Helper != "" && route.Helper != "sdk.dispatch" && route.Helper != "sdk.request" && route.Helper != "sdk.response" && route.Helper != "sdk.context" && route.Helper != "sdk.pipeline" && route.Helper != "host.model") {
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
		seenRoutes[route.Operation] = true
	}
	return validateObservationContract(manifest, declaredCapabilities)
}

func (m Manifest) EntryPoint() string {
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
	}
	return b.String()
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
