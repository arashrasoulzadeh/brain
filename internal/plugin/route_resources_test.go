package plugin

import "testing"

func TestRouteResourceGrantsFailClosed(t *testing.T) {
	resources := map[string]bool{"allowed": true}
	route := Route{ResourceField: "id"}
	for _, tc := range []struct {
		payload string
		want    bool
	}{
		{`{"id":"allowed"}`, true},
		{`{"id":" allowed "}`, true},
		{`{"id":"other"}`, false},
		{`{"id":7}`, false},
		{`{"resource":"allowed"}`, false},
		{`null`, false},
	} {
		if got := routeResourceAllowed(route, []byte(tc.payload), resources); got != tc.want {
			t.Errorf("%s: got %t, want %t", tc.payload, got, tc.want)
		}
	}
	if routeResourceAllowed(Route{}, []byte(`{"id":"allowed"}`), resources) {
		t.Fatal("undeclared resource binding accepted")
	}
	if !routeResourceAllowed(Route{Model: &ModelRoute{Name: "allowed"}}, nil, resources) {
		t.Fatal("declared model resource rejected")
	}
	if !routeResourceAllowed(Route{}, nil, nil) {
		t.Fatal("unrestricted grant rejected")
	}
}
