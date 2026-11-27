package plugin_test

import (
	"encoding/json"
	"os"
	"os/exec"
	"strings"
	"sync/atomic"
	"testing"

	"brain/internal/admin"
	"brain/internal/operator"
	"brain/internal/plugin"
)

func TestFunctionPluginReadBridgeUsesPersistedMetadataAndActualOperator(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("Node is unavailable")
	}
	functions, err := plugin.LoadManifest("../../plugins/functions/plugin.yaml")
	if err != nil { t.Fatal(err) }
	files, err := plugin.LoadManifest("../../plugins/files/plugin.yaml")
	if err != nil { t.Fatal(err) }
	r, storage := accessRegistry(t, functions, files)
	installAccessTestPlugin(t, r, "alpha", accessManifest(t), &contractWorker{ready: true})
	installAccessTestPlugin(t, r, "alpha", functions, &contractWorker{ready: true})
	installAccessTestPlugin(t, r, "alpha", files, &contractWorker{ready: true})
	const key = "images/avatar.png"
	if err := storage.PutFile("alpha", key, "alpha", false); err != nil { t.Fatal(err) }
	if err := storage.PutFile("beta", key, "foreign data", true); err != nil { t.Fatal(err) }
	if err := storage.PutFile("alpha", "private.png", "private", false); err != nil { t.Fatal(err) }

	var calls atomic.Int32
	if err := r.BindWorker("alpha", files.ID, files.Version, &storageCommitWorker{handle: func(req plugin.WorkerRequest) plugin.WorkerResponse {
		calls.Add(1)
		var payload struct {
			Key string `json:"key"`
			Data string `json:"data"`
			Published bool `json:"published"`
		}
		if err := json.Unmarshal(req.Payload, &payload); err != nil || req.ProjectID != "alpha" || req.Operation != "files.metadata" ||
			req.HostContext == nil || req.HostContext.Auth.Actor.ID != "root" || req.HostContext.Auth.Actor.Kind != "operator" ||
			payload.Key != key || payload.Data != "alpha" {
			t.Errorf("unexpected trusted invocation scope: %s %s", req.ProjectID, req.Operation)
			return plugin.WorkerResponse{Protocol: req.Protocol, ErrorCode: "unexpected scope"}
		}
		result, _ := json.Marshal(map[string]any{"key": payload.Key, "size": len(payload.Data), "published": payload.Published})
		return plugin.WorkerResponse{Protocol: req.Protocol, Success: true, Payload: result}
	}}); err != nil { t.Fatal(err) }

	const name = "bridge_metadata"
	publish, _ := functions.Route("functions.publish")
	invoke, _ := functions.Route("functions.invoke")
	public, _ := functions.Route("functions.call")
	metadata, _ := files.Route("files.metadata")
	createAccessRule(t, r, "alpha", "bridge-publish", "operator:root", functions, publish, name, true)
	createAccessRule(t, r, "alpha", "bridge-invoke", "operator:root", functions, invoke, name, true)
	createAccessRule(t, r, "alpha", "bridge-public", "public", functions, public, name, true)

	cwd, err := os.Getwd()
	if err != nil { t.Fatal(err) }
	if err := os.Chdir(t.TempDir()); err != nil { t.Fatal(err) }
	t.Cleanup(func() { if err := os.Chdir(cwd); err != nil { t.Error(err) } })

	service := admin.NewLifecycleService(nil, r)
	actor := operator.Actor{Kind: operator.ActorOperator, ID: "root"}
	source := "export const config = { method: 'POST' }; export default async function handler(input, context) { return await context.plugins.call('files.metadata', input); }"
	input := json.RawMessage(`{"key":"images/avatar.png"}`)
	if _, err := service.CreateFunctionDraftWithSettings(actor, "alpha", name, source, 1, false, []string{"functions.invoke", "files.metadata"}, "POST"); err != nil { t.Fatal(err) }

	deny := func(label string, code string, data json.RawMessage) {
		t.Helper()
		before := calls.Load()
		if _, err := service.RunFunctionDraft(actor, "alpha", name, code, data); err == nil || !strings.Contains(err.Error(), "Plugin call denied") {
			t.Fatalf("%s: expected denial, got %v", label, err)
		}
		if calls.Load() != before { t.Fatalf("%s dispatched a denied call", label) }
	}
	deny("no target access policy", source, input)
	createAccessRule(t, r, "alpha", "bridge-metadata", "operator:root", files, metadata, key, true)
	result, err := service.RunFunctionDraft(actor, "alpha", name, source, input)
	if err != nil { t.Fatal(err) }
	value, ok := result["result"].(map[string]any)
	if !ok || value["key"] != key || value["size"] != float64(5) || value["published"] != false {
		t.Fatalf("unexpected metadata result: %+v", result)
	}
	deny("cross-project override", source, json.RawMessage(`{"key":"images/avatar.png","project_id":"beta"}`))
	deny("forged snapshot", source, json.RawMessage(`{"key":"images/avatar.png","data":"forged"}`))
	deny("unauthorized resource", source, json.RawMessage(`{"key":"private.png"}`))
	deny("non-opted-in route", strings.Replace(source, "files.metadata", "files.download", 1), input)
	if err := r.GrantResource("alpha", files.ID, files.Version, metadata.Capability, "private.png"); err != nil { t.Fatal(err) }
	deny("installation resource grant", source, input)
	if err := r.GrantResource("alpha", files.ID, files.Version, metadata.Capability, key); err != nil { t.Fatal(err) }
	if err := service.UpdateFunctionPermissions(actor, "alpha", name, []string{"functions.invoke"}, false); err != nil { t.Fatal(err) }
	deny("revoked saved function permission", source, input)
	if err := service.UpdateFunctionPermissions(actor, "alpha", name, []string{"functions.invoke", "files.metadata"}, false); err != nil { t.Fatal(err) }
	if _, err := service.PublishFunction(actor, "alpha", name, "POST"); err != nil { t.Fatal(err) }
	before := calls.Load()
	if _, err := service.PublicInvokeFunction("alpha", name, "POST", input); err == nil || !strings.Contains(err.Error(), "operator sandbox test") {
		t.Fatalf("public delegation was not denied: %v", err)
	}
	if calls.Load() != before { t.Fatal("public call inherited operator privileges") }
}

func TestFunctionReadManifestRejectsUnreviewedWrites(t *testing.T) {
	m, err := plugin.LoadManifest("../../plugins/files/plugin.yaml")
	if err != nil { t.Fatal(err) }
	for i := range m.Routes {
		if m.Routes[i].Operation == "files.upload" {
			m.Routes[i].FunctionRead = true
		}
	}
	if err := plugin.ValidateManifestContract(m); err == nil {
		t.Fatal("write route accepted as a function read")
	}
}
