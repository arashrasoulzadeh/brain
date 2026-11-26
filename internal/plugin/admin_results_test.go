package plugin

import "testing"

func TestAdminResultComponentsAreBounded(t *testing.T) {
	valid := AdminResultSpec{Type: "grid", Children: []AdminResultSpec{{Type: "table", Path: "data.rows", PageSize: 25}, {Type: "metric", Path: "count"}}}
	if !valid.valid(0) {
		t.Fatal("valid components rejected")
	}
	for _, spec := range []AdminResultSpec{{Type: "script"}, {Type: "table", Path: "__proto__"}, {Type: "table", PageSize: 101}, {Type: "chart"}, {Type: "tabs"}} {
		if spec.valid(0) {
			t.Fatalf("invalid result component accepted: %#v", spec)
		}
	}
}
