package plugin_test

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
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
		{"name": math.Inf(1)},
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

func TestFunctionsDeclarativeRoutesReachWorkerDescriptor(t *testing.T) {
	root := "../../plugins/functions"
	manifest, err := plugin.LoadManifest(filepath.Join(root, "plugin.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	descriptor := filepath.Join(t.TempDir(), "execution.json")
	if _, err := plugin.NewManifestTypeScriptWorker("node", "worker.mjs", descriptor, root, "project", manifest, plugin.IsolationPolicy{}); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(descriptor)
	if err != nil {
		t.Fatal(err)
	}
	var execution struct {
		Routes []struct {
			Operation string               `json:"operation"`
			Helper    string               `json:"helper"`
			Response  *plugin.ResponseSpec `json:"response"`
		} `json:"routes"`
	}
	if err := json.Unmarshal(data, &execution); err != nil {
		t.Fatal(err)
	}
	found := map[string]string{}
	for _, route := range execution.Routes {
		found[route.Operation] = route.Helper
		if route.Helper == "sdk.response" && (route.Response == nil || route.Response.Output["name"] != "$input.name") {
			t.Fatalf("missing response mapping: %s", data)
		}
	}
	if found["functions.invoke"] != "sdk.response" || found["functions.publish"] != "sdk.response" || found["functions.call"] != "sdk.dispatch" {
		t.Fatalf("unexpected Functions routes: %v", found)
	}
}
