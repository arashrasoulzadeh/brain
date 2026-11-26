package plugin

import (
	"encoding/json"
	"testing"
)

func TestAdminManifest(t *testing.T) {
	m := Manifest{Admin: &AdminSpec{Version: 1, Pages: []AdminPage{{ID: "records", Title: "Records", Actions: []AdminAction{{Title: "List", Operation: "records.list"}}}}}, Routes: []Route{{Operation: "records.list"}}}
	if err := validateAdminSpec(m); err != nil || !m.AdminOperation("records", "records.list") {
		t.Fatalf("valid page rejected: %v", err)
	}
	if m.AdminOperation("other", "records.list") || m.AdminOperation("records", "records.delete") {
		t.Fatal("undeclared operation exposed")
	}
	m.Admin.Pages[0].Actions[0].Operation = "records.delete"
	if validateAdminSpec(m) == nil {
		t.Fatal("unknown route accepted")
	}
}

func TestAdminFormPayload(t *testing.T) {
	m := Manifest{Admin: &AdminSpec{Version: 1, Pages: []AdminPage{{ID: "records", Title: "Records", Actions: []AdminAction{{Title: "List", Operation: "records.list", Fields: []AdminField{{Name: "table", Type: "string", Required: true}}}}}}}, Routes: []Route{{Operation: "records.list"}}}
	if err := m.ValidateAdminPayload("records", "records.list", json.RawMessage(`{"table":"users"}`)); err != nil {
		t.Fatal(err)
	}
	for _, payload := range []string{`{}`, `{"table":1}`, `{"table":"users","projectId":"another"}`, `null`} {
		if m.ValidateAdminPayload("records", "records.list", json.RawMessage(payload)) == nil {
			t.Fatalf("invalid payload accepted: %s", payload)
		}
	}
}

func TestAdminReadOnlyActionDoesNotAcceptUndeclaredInput(t *testing.T) {
	m := Manifest{Admin: &AdminSpec{Version: 1, Pages: []AdminPage{{ID: "list", Title: "List", Actions: []AdminAction{{Title: "List", Operation: "records.list", ReadOnly: true}}}}}, Routes: []Route{{Operation: "records.list"}}}
	if !m.AdminReadOnly("list", "records.list") || m.AdminReadOnly("list", "records.write") {
		t.Fatal("invalid read-only operation classification")
	}
	if m.ValidateAdminPayload("list", "records.list", json.RawMessage(`{"write":true}`)) == nil {
		t.Fatal("undeclared read input accepted")
	}
}
