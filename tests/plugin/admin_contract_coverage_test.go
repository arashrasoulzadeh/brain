package plugin_test

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"brain/internal/plugin"
)

func adminCoverageManifest() plugin.Manifest {
	m := coverageManifest()
	m.Admin = &plugin.AdminSpec{Version: 1, Pages: []plugin.AdminPage{{ID: "records", Title: "Records", Actions: []plugin.AdminAction{{Title: "Read", Operation: "coverage.read"}}}}}
	return m
}

func TestAdminContractRejectsInvalidPageMetadata(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mutate func(*plugin.AdminPage)
	}{
		{"bad ID", func(p *plugin.AdminPage) { p.ID = "bad/id" }},
		{"bad extension", func(p *plugin.AdminPage) { p.Extends = "coverage/records" }},
		{"unbound slot", func(p *plugin.AdminPage) { p.ExtendsSlot = "tab" }},
		{"unsupported slot", func(p *plugin.AdminPage) { p.ExtensionSlots = []string{"arbitrary"} }},
		{"unknown inventory", func(p *plugin.AdminPage) {
			p.Inventory = &plugin.AdminInventorySpec{Kind: "unknown", Capability: "read", Title: "Records"}
		}},
		{"undeclared inventory capability", func(p *plugin.AdminPage) {
			p.Inventory = &plugin.AdminInventorySpec{Kind: "state_keys", Capability: "other", Title: "Records"}
		}},
		{"unknown inventory select field", func(p *plugin.AdminPage) {
			p.Inventory = &plugin.AdminInventorySpec{Kind: "state_keys", Capability: "read", Title: "Records", SelectField: "missing"}
		}},
		{"reserved alias", func(p *plugin.AdminPage) { p.LegacyRoute = "projects" }},
		{"too many events", func(p *plugin.AdminPage) { p.RefreshOn = make([]string, 17) }},
		{"unknown event", func(p *plugin.AdminPage) { p.RefreshOn = []string{"foreign.event"} }},
		{"too many fields", func(p *plugin.AdminPage) { p.Actions[0].Fields = make([]plugin.AdminField, 33) }},
		{"forged project field", func(p *plugin.AdminPage) {
			p.Actions[0].Fields = []plugin.AdminField{{Name: "projectId", Type: "string"}}
		}},
		{"duplicate columns", func(p *plugin.AdminPage) {
			p.Actions[0].Result = &plugin.AdminResultSpec{Type: "table", Columns: []string{"value", "value"}}
		}},
		{"invalid child", func(p *plugin.AdminPage) {
			p.Actions[0].Result = &plugin.AdminResultSpec{Type: "grid", Children: []plugin.AdminResultSpec{{Type: "script"}}}
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := adminCoverageManifest()
			tc.mutate(&m.Admin.Pages[0])
			if err := plugin.ValidateManifestContract(m); !errors.Is(err, plugin.ErrInvalidManifest) {
				t.Fatalf("invalid admin metadata accepted: %v", err)
			}
		})
	}
	m := adminCoverageManifest()
	m.Models = map[string]plugin.ModelSpec{"item": {Version: 1, Fields: map[string]string{"name": "string"}}}
	m.Routes[0].Helper = "host.model"
	m.Routes[0].Model = &plugin.ModelRoute{Name: "item", Action: "create"}
	m.Admin.Pages[0].Actions[0].ReadOnly = true
	if err := plugin.ValidateManifestContract(m); !errors.Is(err, plugin.ErrInvalidManifest) {
		t.Fatalf("read-only mutation accepted: %v", err)
	}
}

func TestAdminPayloadTypesAndActionAuthorization(t *testing.T) {
	m := adminCoverageManifest()
	if err := plugin.ValidateManifestContract(m); err != nil {
		t.Fatal(err)
	}
	if err := m.ValidateAdminPayload("missing", "coverage.read", json.RawMessage(`{}`)); !errors.Is(err, plugin.ErrGrantDenied) {
		t.Fatalf("missing page: %v", err)
	}
	if err := m.ValidateAdminPayload("records", "missing", json.RawMessage(`{}`)); !errors.Is(err, plugin.ErrGrantDenied) {
		t.Fatalf("missing action: %v", err)
	}
	if err := m.ValidateAdminPayload("records", "coverage.read", json.RawMessage(`{"value":1}`)); err != nil {
		t.Fatal(err)
	}
	if err := m.ValidateAdminPayload("records", "coverage.read", json.RawMessage(`{`)); !errors.Is(err, plugin.ErrInvalidWorkerRequest) {
		t.Fatalf("malformed payload: %v", err)
	}
	action := &m.Admin.Pages[0].Actions[0]
	action.Fields = []plugin.AdminField{{Name: "count", Type: "number", Required: true}, {Name: "active", Type: "boolean", Required: true}, {Name: "data", Type: "json", Required: true}, {Name: "note", Type: "string"}}
	if err := m.ValidateAdminPayload("records", "coverage.read", json.RawMessage(`{"count":1,"active":true,"data":{"value":2}}`)); err != nil {
		t.Fatal(err)
	}
	for _, payload := range []string{`{"count":"1","active":true,"data":{}}`, `{"count":1,"active":"true","data":{}}`, `{"count":1,"active":true,"data":null}`, `{"count":1,"active":true,"data":{},"projectId":"other"}`} {
		if err := m.ValidateAdminPayload("records", "coverage.read", json.RawMessage(payload)); !errors.Is(err, plugin.ErrInvalidWorkerRequest) {
			t.Fatalf("invalid input accepted: %v", err)
		}
	}
	if _, err := plugin.NewRegistry(1).AdminInventory("alpha", "coverage", "1", "records", "", 1); !errors.Is(err, plugin.ErrAdminInventoryUnavailable) {
		t.Fatalf("missing inventory provider: %v", err)
	}
	if _, err := plugin.NewRegistry(1).AdminInventory("alpha", "coverage", "1", "records", strings.Repeat("x", 65537), 1); !errors.Is(err, plugin.ErrInvalidWorkerRequest) {
		t.Fatalf("invalid cursor: %v", err)
	}
}
