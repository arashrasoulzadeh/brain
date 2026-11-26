package plugin

import (
	"encoding/json"
	"testing"
)

func TestGeneratedModelPagesOnlyExposeDeclaredRoutes(t *testing.T) {
	m := Manifest{ID: "catalog", Title: "Catalog", Models: map[string]ModelSpec{"item": {Version: 1, Fields: map[string]string{"name": "string"}}}, Routes: []Route{
		{Operation: "catalog.list", Helper: "host.model", Model: &ModelRoute{Name: "item", Action: "list"}},
		{Operation: "catalog.create", Helper: "host.model", Model: &ModelRoute{Name: "item", Action: "create"}},
		{Operation: "catalog.other", Helper: "sdk.response"},
	}}
	spec := m.AdminForOperator()
	if spec == nil || len(spec.Pages) != 1 || len(spec.Pages[0].Actions) != 2 || m.Admin != nil {
		t.Fatalf("generated pages: %#v", spec)
	}
	page := spec.Pages[0]
	if !adminPageID.MatchString(page.ID) || page.Model == nil || page.Model.Name != "item" {
		t.Fatalf("invalid generated page: %#v", page)
	}
	if !m.AdminReadOnly(page.ID, "catalog.list") || m.AdminReadOnly(page.ID, "catalog.create") || m.AdminOperation(page.ID, "catalog.other") {
		t.Fatal("generated action policy is incorrect")
	}
	if err := m.ValidateAdminPayload(page.ID, "catalog.create", json.RawMessage(`{"id":"one","fields":{"name":"first"}}`)); err != nil {
		t.Fatal(err)
	}
	if err := m.ValidateAdminPayload(page.ID, "catalog.list", json.RawMessage(`{"projectId":"other"}`)); err == nil {
		t.Fatal("generated list accepted an undeclared field")
	}
	m.Admin = &AdminSpec{Version: 1, Pages: []AdminPage{{ID: "custom", Title: "Custom", Actions: []AdminAction{{Title: "Create", Operation: "catalog.create", ReadOnly: true}}}}}
	if err := validateAdminSpec(m); err == nil {
		t.Fatal("model write mislabeled read-only")
	}
}
