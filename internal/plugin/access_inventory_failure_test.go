package plugin

import (
	"encoding/json"
	"errors"
	"testing"
)

func coverageRuleFields(resource string, allowed bool) map[string]any {
	return map[string]any{"principal": "operator:alice", "plugin": "rooms", "version": "1", "operation": "rooms.read", "capability": "rooms.read", "resource": resource, "allowed": allowed}
}

func coverageAccessRule(t *testing.T, s *SQLiteStorage, id string, version int, fields map[string]any) {
	t.Helper()
	data, err := json.Marshal(fields)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.db.Exec("INSERT INTO plugin_model_records(project_id,plugin_id,model_name,record_id,schema_version,revision,state,fields_json) VALUES(?,?,?,?,?,?,?,?)", "coverage", accessPluginID, accessRuleModel, id, version, "revision", "active", string(data)); err != nil {
		t.Fatal(err)
	}
}

func TestCoverageAccessInventoryCorruptRules(t *testing.T) {
	route := Route{Operation: "rooms.read", Capability: "rooms.read", ResourceField: "roomId"}
	for _, tc := range []struct {
		name, resource string
		version        int
		want           error
	}{
		{"schema mismatch", "room", 2, ErrModelSchema},
		{"invalid resource", "", 1, ErrActorDenied},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := coverageStorage(t)
			coverageAccessRule(t, s, "rule", tc.version, coverageRuleFields(tc.resource, true))
			if _, err := s.accessInventoryFilter("coverage", "operator:alice", "rooms", "1", route); !errors.Is(err, tc.want) {
				t.Fatal(err)
			}
			if tc.version == 2 {
				if allowed, err := s.accessRuleAllows("coverage", "operator:alice", "rooms", "1", route.Operation, route.Capability, tc.resource); allowed || !errors.Is(err, tc.want) {
					t.Fatalf("corrupt rule allowed: %v %v", allowed, err)
				}
			}
		})
	}
}

func TestCoverageAccessInventoryLimitAndDeny(t *testing.T) {
	s := coverageStorage(t)
	route := Route{Operation: "rooms.read", Capability: "rooms.read", ResourceField: "roomId"}
	coverageAccessRule(t, s, "deny", 1, coverageRuleFields("*", false))
	filter, err := s.accessInventoryFilter("coverage", "operator:alice", "rooms", "1", route)
	if err != nil || !filter.DenyAll {
		t.Fatalf("wildcard deny lost: %+v %v", filter, err)
	}
	fields, err := json.Marshal(coverageRuleFields("room", true))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.db.Exec("WITH RECURSIVE n(i) AS (SELECT 1 UNION ALL SELECT i+1 FROM n WHERE i<4096) INSERT INTO plugin_model_records(project_id,plugin_id,model_name,record_id,schema_version,revision,state,fields_json) SELECT ?,?,?,CAST(i AS TEXT),1,'revision','active',? FROM n", "coverage", accessPluginID, accessRuleModel, string(fields)); err != nil {
		t.Fatal(err)
	}
	if _, err := s.accessInventoryFilter("coverage", "operator:alice", "rooms", "1", route); !errors.Is(err, ErrActorDenied) {
		t.Fatalf("oversized policy accepted: %v", err)
	}
}

func TestCoverageAccessRuleValidation(t *testing.T) {
	r := NewRegistry(1)
	manifest := Manifest{ID: "rooms", Version: "1", Routes: []Route{{Operation: "rooms.read", Capability: "rooms.read", ResourceField: "roomId"}}}
	r.registered["rooms@1"] = Registered{Manifest: manifest}
	if err := r.validateAccessRuleLocked(coverageRequest("create", "{}")); !errors.Is(err, ErrInvalidWorkerRequest) {
		t.Fatal(err)
	}
	for _, override := range []map[string]any{
		{"principal": "invalid"},
		{"plugin": "missing"},
		{"operation": "unknown"},
	} {
		fields := coverageRuleFields("room", true)
		for k, v := range override {
			fields[k] = v
		}
		payload, err := json.Marshal(map[string]any{"fields": fields})
		if err != nil {
			t.Fatal(err)
		}
		if err := r.validateAccessRuleLocked(coverageRequest("create", string(payload))); !errors.Is(err, ErrInvalidWorkerRequest) {
			t.Fatalf("invalid policy accepted: %v", err)
		}
	}
	page := AdminPage{Inventory: &AdminInventorySpec{Capability: "rooms.read", SelectField: "roomId"}, Actions: []AdminAction{{Operation: "rooms.read", ReadOnly: true}, {Operation: "rooms.read", ReadOnly: true}}}
	if _, ok := inventoryReadRoute(manifest, page); ok {
		t.Fatal("ambiguous read route accepted")
	}
	if _, ok := inventoryReadRoute(manifest, AdminPage{}); ok {
		t.Fatal("missing inventory accepted")
	}
}

func TestCoverageHostOperationAmbiguousVersion(t *testing.T) {
	r := NewRegistry(1)
	r.installations["coverage:access-management@1"] = Installation{ProjectID: "coverage", PluginID: accessPluginID, Version: "1"}
	for _, version := range []string{"1", "2"} {
		r.installations["coverage:rooms@"+version] = Installation{ProjectID: "coverage", PluginID: "rooms", Version: version, Enabled: true}
	}
	if err := r.AuthorizeHostOperation("coverage", "operator:alice", "rooms", "rooms.read", "room"); !errors.Is(err, ErrActorDenied) {
		t.Fatal(err)
	}
	if err := r.AuthorizeHostOperation("coverage", "operator:alice", "missing", "read", ""); !errors.Is(err, ErrActorDenied) {
		t.Fatal(err)
	}
	if err := r.authorizeActorAccess(HostCall{ProjectID: "coverage", ActorID: "alice", PluginID: "missing", Version: "1", Operation: "read", Capability: "read"}); !errors.Is(err, ErrActorDenied) {
		t.Fatal(err)
	}
}
