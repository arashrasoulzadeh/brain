package plugin_test

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"brain/internal/plugin"
)

func TestTypeScriptWorkerRoundTripsJSONPayload(t *testing.T) {
	node, err := exec.LookPath("node")
	if err != nil {
		t.Skip("Node is required for the cross-language contract test")
	}
	root := t.TempDir()
	entry := filepath.Join(root, "index.mjs")
	if err := os.WriteFile(entry, []byte("export function echo(input) { return { echoed: input }; }\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	script, err := filepath.Abs("../../scripts/typescript-worker.mjs")
	if err != nil {
		t.Fatal(err)
	}
	manifest := plugin.Manifest{
		ID: "sample", Version: "1.0.0", Entry: "index.mjs", SourceFiles: []string{"index.mjs"},
		Capabilities: []string{"echo"}, Routes: []plugin.Route{{Operation: "echo", Function: "echo", Capability: "echo"}},
	}
	worker, err := plugin.NewManifestTypeScriptWorker(node, script, filepath.Join(root, "execution.json"), root, "project", manifest, plugin.IsolationPolicy{})
	if err != nil {
		t.Fatal(err)
	}
	if err := worker.Start(); err != nil {
		t.Fatal(err)
	}
	defer worker.Stop()
	request := plugin.WorkerRequest{
		Protocol: 1, ProjectID: "project", PluginID: "sample", Version: "1.0.0", Operation: "echo",
		Payload: []byte(`{"name":"Brain","count":3}`),
	}
	response, err := plugin.InvokeForProject(worker, request, 1, "project")
	if err != nil || !response.Success || response.Protocol != 1 {
		t.Fatalf("response = %#v, err = %v", response, err)
	}
	var result struct {
		Echoed struct {
			Name  string
			Count int
		}
	}
	if err := json.Unmarshal(response.Payload, &result); err != nil || result.Echoed.Name != "Brain" || result.Echoed.Count != 3 {
		t.Fatalf("payload = %s, err = %v", response.Payload, err)
	}
}
