package plugin

import "testing"

func TestInvocationResourceRejectsAmbiguousIdentity(t *testing.T) {
	route := Route{ResourceField: "id"}
	for _, tc := range []struct {
		payload string
		want    string
	}{
		{`{"id":"allowed"}`, "allowed"},
		{`{"id":" allowed "}`, ""},
		{`{"id":7}`, ""},
		{`{"resource":"allowed"}`, ""},
		{`{"id":"*"}`, ""},
		{`null`, ""},
	} {
		resource, err := invocationResource(WorkerRequest{Payload: []byte(tc.payload)}, route)
		if resource != tc.want || (err != nil) != (tc.want == "") {
			t.Errorf("%s: got %q, %v; want %q", tc.payload, resource, err, tc.want)
		}
	}
	resource, err := invocationResource(WorkerRequest{}, Route{Model: &ModelRoute{Name: "allowed"}})
	if err != nil || resource != "allowed" {
		t.Fatalf("model resource = %q, %v", resource, err)
	}
}
