package plugin_test

import (
	"testing"

	"brain/internal/plugin"
)

func TestDeclarativeResponseManifest(t *testing.T) {
	route := plugin.Route{
		Operation: "sample.read", Function: "read", Capability: "sample.read", Helper: "sdk.response",
		Response: &plugin.ResponseSpec{
			Input:  []plugin.ResponseField{{Name: "name", Type: "string", Required: true, Error: "name required"}},
			Output: map[string]any{"name": "$input.name", "ok": true},
		},
	}
	manifest := plugin.Manifest{SourceFiles: []string{"index.ts"}, Capabilities: []string{"sample.read"}, Routes: []plugin.Route{route}}
	if err := plugin.ValidateManifestContract(manifest); err != nil {
		t.Fatal(err)
	}
	for _, invalid := range []map[string]any{
		{"name": "$input.missing"},
		{"__proto__": true},
		{"name": "$eval.code"},
		{"name": []string{"not", "scalar"}},
	} {
		route.Response.Output = invalid
		if err := plugin.ValidateManifestContract(manifest); err == nil {
			t.Fatalf("invalid response accepted: %v", invalid)
		}
	}
	route.Response.Output = map[string]any{"name": "$input.name"}
	manifest.Routes[0].Helper = "sdk.dispatch"
	if err := plugin.ValidateManifestContract(manifest); err == nil {
		t.Fatal("response attached to wrong helper")
	}
}
