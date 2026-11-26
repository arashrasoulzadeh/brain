package plugin

import "testing"

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
