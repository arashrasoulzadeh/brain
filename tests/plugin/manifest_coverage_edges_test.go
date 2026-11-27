package plugin_test

import (
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"brain/internal/plugin"
)

func coverageManifest() plugin.Manifest {
	return plugin.Manifest{
		ID: "coverage", Title: "Coverage fixture", Version: "1.0.0", HostAPI: 1,
		Entry: "index.ts", SourceFiles: []string{"index.ts"},
		Capabilities: []string{"read"},
		Routes: []plugin.Route{{Operation: "coverage.read", Function: "read", Capability: "read"}},
	}
}

func TestManifestRejectsInvalidContractsAtEveryDeclarationBoundary(t *testing.T) {
	for name, mutate := range map[string]func(*plugin.Manifest){
		"declarative source": func(m *plugin.Manifest) { m.Declarative = true },
		"duplicate source": func(m *plugin.Manifest) { m.SourceFiles = append(m.SourceFiles, m.SourceFiles[0]) },
		"traversal source": func(m *plugin.Manifest) { m.SourceFiles = append(m.SourceFiles, "../other.ts") },
		"missing entry": func(m *plugin.Manifest) { m.Entry = "missing.ts" },
		"duplicate import": func(m *plugin.Manifest) { m.Imports = []string{"node:crypto", "node:crypto"} },
		"invalid import": func(m *plugin.Manifest) { m.Imports = []string{"../../other.ts"} },
		"duplicate plugin": func(m *plugin.Manifest) { m.PluginImports = []string{"other", "other"} },
		"duplicate capability": func(m *plugin.Manifest) { m.Capabilities = append(m.Capabilities, "read") },
		"duplicate operation": func(m *plugin.Manifest) { m.Routes = append(m.Routes, m.Routes[0]) },
		"unknown capability": func(m *plugin.Manifest) { m.Routes[0].Capability = "other" },
		"unknown helper": func(m *plugin.Manifest) { m.Routes[0].Helper = "arbitrary" },
		"non-dashboard admin helper": func(m *plugin.Manifest) { m.Routes[0].Helper = "host.admin" },
		"response without helper": func(m *plugin.Manifest) { m.Routes[0].Response = &plugin.ResponseSpec{} },
		"pipeline without helper": func(m *plugin.Manifest) { m.Routes[0].Pipeline = &plugin.PipelineSpec{} },
		"model without helper": func(m *plugin.Manifest) { m.Routes[0].Model = &plugin.ModelRoute{} },
		"call without helper": func(m *plugin.Manifest) { m.Routes[0].Call = &plugin.PluginCall{} },
		"resource field": func(m *plugin.Manifest) { m.Routes[0].ResourceField = "__proto__" },
		"invalid admin version": func(m *plugin.Manifest) { m.Admin = &plugin.AdminSpec{Version: 2} },
	} {
		t.Run(name, func(t *testing.T) {
			m := coverageManifest()
			mutate(&m)
			if err := plugin.ValidateManifestContract(m); !errors.Is(err, plugin.ErrInvalidManifest) {
				t.Fatalf("invalid contract accepted: %v", err)
			}
		})
	}
	dir := t.TempDir()
	if _, err := plugin.LoadManifest(filepath.Join(dir, "missing.yaml")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("missing manifest: %v", err)
	}
	for i, raw := range []string{"invalid: [", "id: invalid\n"} {
		path := filepath.Join(dir, fmt.Sprintf("manifest-%d.yaml", i))
		if err := os.WriteFile(path, []byte(raw), 0o600); err != nil {
			t.Fatal(err)
		}
		if _, err := plugin.LoadManifest(path); err == nil {
			t.Fatal("invalid manifest loaded")
		}
	}
	m := coverageManifest()
	if err := plugin.ValidatePluginSources(dir, m); !errors.Is(err, plugin.ErrInvalidManifest) {
		t.Fatalf("missing source: %v", err)
	}
	if err := os.Mkdir(filepath.Join(dir, "index.ts"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := plugin.ValidatePluginSources(dir, m); !errors.Is(err, plugin.ErrInvalidManifest) {
		t.Fatalf("directory source: %v", err)
	}
	if err := os.Remove(filepath.Join(dir, "index.ts")); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "real.ts"), []byte("export const read = 1;"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("real.ts", filepath.Join(dir, "index.ts")); err != nil {
		t.Fatal(err)
	}
	if err := plugin.ValidatePluginSources(dir, m); !errors.Is(err, plugin.ErrInvalidManifest) {
		t.Fatalf("symlink source: %v", err)
	}
}

func TestModelDeclarationsAndGeneratedContracts(t *testing.T) {
	model := plugin.ModelSpec{Version: 1, Fields: map[string]string{"name": "string", "count": "integer", "active": "boolean"}, Initial: "waiting", Transitions: map[string]plugin.ModelTransition{"start": {From: "waiting", To: "running"}}}
	m := coverageManifest()
	m.Models = map[string]plugin.ModelSpec{"Match": model}
	m.Routes = nil
	for _, action := range []string{"create", "read", "list", "update", "delete", "transition"} {
		route := plugin.Route{Operation: "model." + action, Function: action, Capability: "read", Helper: "host.model", Model: &plugin.ModelRoute{Name: "Match", Action: action}}
		if action == "transition" {
			route.Model.Transition = "start"
		}
		m.Routes = append(m.Routes, route)
	}
	if err := plugin.ValidateManifestContract(m); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"Models", "number", "boolean", "revision: string", "after?: string", "model.delete"} {
		if !strings.Contains(plugin.TypeScriptContract(m), want) {
			t.Fatalf("generated type missing %s", want)
		}
	}
	if docs := plugin.Documentation(m); !strings.Contains(docs, "Transition") || !strings.Contains(docs, "schema 1") {
		t.Fatalf("incomplete model documentation: %s", docs)
	}
	admin := m.AdminForOperator()
	if admin == nil || len(admin.Pages) != 1 || len(admin.Pages[0].Actions) != 6 {
		t.Fatalf("generated model actions: %#v", admin)
	}
	id := admin.Pages[0].ID
	if !m.AdminOperation(id, "model.delete") || m.AdminReadOnly(id, "model.delete") || !m.AdminReadOnly(id, "model.list") {
		t.Fatal("incorrect generated admin operations")
	}
	if m.AdminOperation(id, "missing") || m.AdminReadOnly("missing", "model.list") {
		t.Fatal("unknown admin action accepted")
	}
	for name, invalid := range map[string]plugin.ModelSpec{
		"bad version": {Version: 0, Fields: model.Fields},
		"bad fields": {Version: 1, Fields: map[string]string{"bad": "float"}},
		"bad initial": {Version: 1, Fields: model.Fields, Initial: "bad-state"},
		"missing initial": {Version: 1, Fields: model.Fields, Transitions: model.Transitions},
		"bad transition": {Version: 1, Fields: model.Fields, Initial: "waiting", Transitions: map[string]plugin.ModelTransition{"bad": {From: "", To: "running"}}},
	} {
		t.Run(name, func(t *testing.T) {
			candidate := coverageManifest()
			candidate.Models = map[string]plugin.ModelSpec{"Match": invalid}
			if err := plugin.ValidateManifestContract(candidate); !errors.Is(err, plugin.ErrInvalidManifest) {
				t.Fatalf("invalid model accepted: %v", err)
			}
		})
	}
	tooMany := coverageManifest()
	tooMany.Models = map[string]plugin.ModelSpec{}
	for i := 0; i < 33; i++ {
		tooMany.Models[fmt.Sprintf("Model%d", i)] = model
	}
	if err := plugin.ValidateManifestContract(tooMany); !errors.Is(err, plugin.ErrInvalidManifest) {
		t.Fatalf("oversized model set accepted: %v", err)
	}
	for _, route := range []*plugin.ModelRoute{nil, {Name: "missing", Action: "read"}, {Name: "Match", Action: "arbitrary"}} {
		candidate := coverageManifest()
		candidate.Models = m.Models
		candidate.Routes[0].Helper = "host.model"
		candidate.Routes[0].Model = route
		if err := plugin.ValidateManifestContract(candidate); !errors.Is(err, plugin.ErrInvalidManifest) {
			t.Fatalf("invalid model route accepted: %v", err)
		}
	}
	// The generator has a defensive fallback for an unknown schema kind.
	invalid := coverageManifest()
	invalid.Models = map[string]plugin.ModelSpec{"Fixture": {Fields: map[string]string{"field": "unsupported"}}}
	if !strings.Contains(plugin.TypeScriptContract(invalid), "unknown") {
		t.Fatal("generator emitted an unsupported type")
	}
}

func TestManifestResponseAndPipelineReferenceValidation(t *testing.T) {
	for _, spec := range []*plugin.ResponseSpec{
		nil, {Output: map[string]any{}},
		{Input: []plugin.ResponseField{{Name: "bad-key", Type: "string"}}, Output: map[string]any{"value": "ok"}},
		{Input: []plugin.ResponseField{{Name: "name", Type: "value", Trim: true}}, Output: map[string]any{"value": "ok"}},
		{Output: map[string]any{"value": math.Inf(1)}},
		{Output: map[string]any{"value": map[string]string{"unsafe": "value"}}},
	} {
		m := coverageManifest()
		m.Routes[0].Helper, m.Routes[0].Response = "sdk.response", spec
		if err := plugin.ValidateManifestContract(m); !errors.Is(err, plugin.ErrInvalidManifest) {
			t.Fatalf("invalid response accepted: %v", err)
		}
	}
	for _, spec := range []*plugin.PipelineSpec{
		nil, {Steps: nil, Return: "$input"},
		{Steps: []plugin.PipelineStep{{Helper: "untrusted", Input: "$input", Save: "value"}}, Return: "$steps.value"},
		{Steps: []plugin.PipelineStep{{Helper: "value.identity", Input: "$steps.missing", Save: "value"}}, Return: "$steps.value"},
		{Steps: []plugin.PipelineStep{{Helper: "value.identity", Input: "$input.a.b.c.d.e.f.g.h", Save: "value"}}, Return: "$steps.value"},
		{Steps: []plugin.PipelineStep{{Helper: "value.identity", Input: "$arbitrary", Save: "value"}}, Return: "$steps.value"},
	} {
		m := coverageManifest()
		m.Routes[0].Helper, m.Routes[0].Pipeline = "sdk.pipeline", spec
		if err := plugin.ValidateManifestContract(m); !errors.Is(err, plugin.ErrInvalidManifest) {
			t.Fatalf("invalid pipeline accepted: %v", err)
		}
	}
	m := coverageManifest()
	m.Routes[0].Helper = "sdk.pipeline"
	m.Routes[0].Pipeline = &plugin.PipelineSpec{Steps: []plugin.PipelineStep{{Helper: "value.identity", Input: "$auth.actor.id", Save: "value"}}, Return: "$steps.value"}
	if err := plugin.ValidateManifestContract(m); err != nil {
		t.Fatal(err)
	}
}
