package plugin_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"brain/internal/admin"
	"brain/internal/operator"
	"brain/internal/plugin"
)

const accessID = "access-management"
const accessVersion = "0.1.0"

func accessManifest(t *testing.T) plugin.Manifest {
	t.Helper()
	m, err := plugin.LoadManifest("../../plugins/access-management/plugin.yaml")
	if err != nil {
		t.Fatal(err)
	}
	return m
}

func accessRegistry(t *testing.T, manifests ...plugin.Manifest) (*plugin.Registry, *plugin.SQLiteStorage) {
	t.Helper()
	r := plugin.NewRegistry(1)
	for _, m := range append([]plugin.Manifest{accessManifest(t)}, manifests...) {
		if err := r.Register(m); err != nil {
			t.Fatal(err)
		}
	}
	storage, err := plugin.OpenSQLiteStorage(filepath.Join(t.TempDir(), "access.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = storage.Close() })
	if err := r.BindSQLiteStorage(storage); err != nil {
		t.Fatal(err)
	}
	return r, storage
}

func installAccessTestPlugin(t *testing.T, r *plugin.Registry, project string, m plugin.Manifest, worker plugin.ContractWorker) {
	t.Helper()
	if err := r.Install(project, m.ID, m.Version); err != nil {
		t.Fatal(err)
	}
	if err := r.SetReadiness(m.ID, m.Version, true); err != nil {
		t.Fatal(err)
	}
	for _, capability := range m.Capabilities {
		if err := r.Grant(project, m.ID, m.Version, capability); err != nil {
			t.Fatal(err)
		}
	}
	if err := r.Enable(project, m.ID, m.Version); err != nil {
		t.Fatal(err)
	}
	if err := r.BindWorker(project, m.ID, m.Version, worker); err != nil {
		t.Fatal(err)
	}
}

func accessRuleRequest(project, operation string, payload []byte) plugin.WorkerRequest {
	return plugin.WorkerRequest{Protocol: 1, ProjectID: project, PluginID: accessID, Version: accessVersion, Operation: operation, Payload: payload}
}

func createAccessRule(t *testing.T, r *plugin.Registry, project, id, principal string, m plugin.Manifest, route plugin.Route, resource string, allowed bool) string {
	t.Helper()
	payload, err := json.Marshal(map[string]any{"id": id, "fields": map[string]any{"principal": principal, "plugin": m.ID, "version": m.Version, "operation": route.Operation, "capability": route.Capability, "resource": resource, "allowed": allowed}})
	if err != nil {
		t.Fatal(err)
	}
	result, err := r.InvokeAsOperator(accessRuleRequest(project, "access.rules.create", payload), "root")
	if err != nil || !result.Success {
		t.Fatalf("create rule %s: %+v %v", id, result, err)
	}
	var record struct {
		Revision string
	}
	if err := json.Unmarshal(result.Payload, &record); err != nil || record.Revision == "" {
		t.Fatalf("create rule revision: %+v %v", record, err)
	}
	return record.Revision
}

func revokeAccessRule(t *testing.T, r *plugin.Registry, project, id, revision string) {
	t.Helper()
	payload, err := json.Marshal(map[string]string{"id": id, "revision": revision})
	if err != nil {
		t.Fatal(err)
	}
	result, err := r.InvokeAsOperator(accessRuleRequest(project, "access.rules.revoke", payload), "root")
	if err != nil || !result.Success {
		t.Fatalf("revoke rule %s: %+v %v", id, result, err)
	}
}

func TestAccessManagementGatesEveryDiscoveredFeatureRoute(t *testing.T) {
	locations, err := plugin.Discover("../../plugins")
	if err != nil {
		t.Fatal(err)
	}
	manifests := []plugin.Manifest{}
	for _, location := range locations {
		if location.ID == accessID {
			continue
		}
		m, err := plugin.LoadManifest(location.ManifestPath)
		if err != nil {
			t.Fatal(err)
		}
		manifests = append(manifests, m)
	}
	// Required dependencies must be registered and enabled before their dependents.
	ordered := []plugin.Manifest{}
	remaining := append([]plugin.Manifest{}, manifests...)
	ready := map[string]bool{accessID + "@" + accessVersion: true}
	for len(remaining) > 0 {
		pending := []plugin.Manifest{}
		for _, m := range remaining {
			available := true
			for _, dependency := range m.Dependencies {
				if !ready[dependency] {
					available = false
				}
			}
			if !available {
				pending = append(pending, m)
				continue
			}
			ordered = append(ordered, m)
			ready[m.ID+"@"+m.Version] = true
		}
		if len(pending) == len(remaining) {
			t.Fatal("unresolved feature dependencies")
		}
		remaining = pending
	}
	r, _ := accessRegistry(t, ordered...)
	installAccessTestPlugin(t, r, "alpha", accessManifest(t), &contractWorker{ready: true})
	for _, m := range ordered {
		installAccessTestPlugin(t, r, "alpha", m, &contractWorker{ready: true})
	}
	checked := 0
	for _, m := range ordered {
		for _, route := range m.Routes {
			// Delegated and model operations have dedicated integration tests.
			if route.Call != nil || route.Model != nil {
				continue
			}
			input := map[string]string{}
			resource := ""
			if route.ResourceField != "" {
				resource = "resource-a"
				input[route.ResourceField] = resource
			}
			payload, err := json.Marshal(input)
			if err != nil {
				t.Fatal(err)
			}
			req := plugin.WorkerRequest{Protocol: 1, ProjectID: "alpha", PluginID: m.ID, Version: m.Version, Operation: route.Operation, Payload: payload}
			ruleID := fmt.Sprintf("route-%d", checked)
			if route.Helper == "host.admin" {
				// Host-only operations must never become worker-callable,
				// even when a policy rule exists for the principal.
				if _, err := r.InvokeAsOperator(req, "root"); !errors.Is(err, plugin.ErrGrantDenied) {
					t.Fatalf("%s worker call without policy: %v", route.Operation, err)
				}
				createAccessRule(t, r, "alpha", ruleID, "operator:root", m, route, resource, true)
				if _, err := r.InvokeAsOperator(req, "root"); !errors.Is(err, plugin.ErrGrantDenied) {
					t.Fatalf("%s worker call with policy: %v", route.Operation, err)
				}
				checked++
				continue
			}
			if _, err := r.InvokeAsOperator(req, "root"); !errors.Is(err, plugin.ErrActorDenied) {
				t.Fatalf("%s without policy: %v", route.Operation, err)
			}
			createAccessRule(t, r, "alpha", ruleID, "operator:root", m, route, resource, true)
			// This inventory test checks authorization, not product-specific
			// storage effects. Those require valid domain data and have
			// dedicated execution tests.
			if err := r.AuthorizeHostOperation("alpha", "operator:root", m.ID, route.Operation, resource); err != nil {
				t.Fatalf("%s with policy: %v", route.Operation, err)
			}
			if _, err := r.InvokeAsOperator(req, "other"); !errors.Is(err, plugin.ErrActorDenied) {
				t.Fatalf("%s crossed actor boundary: %v", route.Operation, err)
			}
			checked++
		}
	}
	if checked == 0 {
		t.Fatal("no feature routes tested")
	}
}

func TestAccessManagementResourceCredentialAndDenyPrecedence(t *testing.T) {
	route := plugin.Route{Operation: "records.read", Function: "read", Capability: "records.read", ResourceField: "record"}
	m := plugin.Manifest{ID: "records", Version: "1", HostAPI: 1, Capabilities: []string{route.Capability}, Routes: []plugin.Route{route}}
	r, _ := accessRegistry(t, m)
	for _, project := range []string{"alpha", "beta"} {
		installAccessTestPlugin(t, r, project, accessManifest(t), &contractWorker{ready: true})
		installAccessTestPlugin(t, r, project, m, &contractWorker{ready: true})
	}
	now := time.Unix(100, 0)
	b := &plugin.Broker{Registry: r, ActorAllows: func(string, string, string) bool { return true }}
	call := plugin.HostCall{ProjectID: "alpha", ActorID: "alice", PluginID: m.ID, Version: m.Version, Operation: route.Operation, Capability: route.Capability, Resource: "public"}
	req := plugin.WorkerRequest{Protocol: 1, ProjectID: "alpha", PluginID: m.ID, Version: m.Version, Operation: route.Operation, Payload: []byte(`{"record":"public"}`)}
	if _, err := b.Issue(call, now, time.Minute); !errors.Is(err, plugin.ErrActorDenied) {
		t.Fatalf("default allow: %v", err)
	}
	allowRevision := createAccessRule(t, r, "alpha", "allow", "user:alice", m, route, "*", true)
	denyRevision := createAccessRule(t, r, "alpha", "deny", "user:alice", m, route, "private", false)
	privateCall := call
	privateCall.Resource = "private"
	if _, err := b.Issue(privateCall, now, time.Minute); !errors.Is(err, plugin.ErrActorDenied) {
		t.Fatalf("deny did not override wildcard: %v", err)
	}
	otherProject := call
	otherProject.ProjectID = "beta"
	if _, err := b.Issue(otherProject, now, time.Minute); !errors.Is(err, plugin.ErrActorDenied) {
		t.Fatalf("cross-project allow: %v", err)
	}
	otherActor := call
	otherActor.ActorID = "bob"
	if _, err := b.Issue(otherActor, now, time.Minute); !errors.Is(err, plugin.ErrActorDenied) {
		t.Fatalf("cross-actor allow: %v", err)
	}
	if _, err := r.Invoke(req); !errors.Is(err, plugin.ErrActorDenied) {
		t.Fatalf("trusted host inherited user grant: %v", err)
	}
	if _, err := r.InvokeAsOperator(req, "alice"); !errors.Is(err, plugin.ErrActorDenied) {
		t.Fatalf("operator inherited user grant: %v", err)
	}
	credential, err := b.Issue(call, now, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	tampered := req
	tampered.Payload = []byte(`{"record":"private"}`)
	if _, err := r.InvokeWithCredential(b, credential, tampered, now); !errors.Is(err, plugin.ErrInvocationCredentialMismatch) {
		t.Fatalf("credential resource substitution: %v", err)
	}
	if result, err := r.InvokeWithCredential(b, credential, req, now); err != nil || !result.Success {
		t.Fatalf("valid credential: %+v %v", result, err)
	}
	if _, err := r.InvokeWithCredential(b, credential, req, now); !errors.Is(err, plugin.ErrInvocationExpired) {
		t.Fatalf("replayed credential: %v", err)
	}
	revokeAccessRule(t, r, "alpha", "deny", denyRevision)
	if _, err := b.Issue(privateCall, now, time.Minute); err != nil {
		t.Fatalf("revoked deny still applied: %v", err)
	}
	pending, err := b.Issue(call, now, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	revokeAccessRule(t, r, "alpha", "allow", allowRevision)
	if _, err := r.InvokeWithCredential(b, pending, req, now); !errors.Is(err, plugin.ErrActorDenied) {
		t.Fatalf("revoked allow still committed: %v", err)
	}
	management := accessRuleRequest("alpha", "access.rules.list", []byte("{}"))
	if _, err := r.Invoke(management); !errors.Is(err, plugin.ErrActorDenied) {
		t.Fatalf("untrusted management: %v", err)
	}
	managementCall := plugin.HostCall{ProjectID: "alpha", ActorID: "alice", PluginID: accessID, Version: accessVersion, Operation: "access.rules.list", Capability: "access.rules.read", Resource: "rule"}
	if _, err := b.Issue(managementCall, now, time.Minute); !errors.Is(err, plugin.ErrActorDenied) {
		t.Fatalf("user managed policy: %v", err)
	}
}

type accessBlockingWorker struct {
	contractWorker
	started chan struct{}
	release chan struct{}
}

func (w *accessBlockingWorker) Handle(req plugin.WorkerRequest) plugin.WorkerResponse {
	close(w.started)
	<-w.release
	return plugin.WorkerResponse{Protocol: req.Protocol, Success: true, Payload: req.Payload}
}

func TestAccessManagementRechecksRevocationBeforeStorageCommit(t *testing.T) {
	route := plugin.Route{Operation: "tables.write", Function: "write", Capability: "tables.write", ResourceField: "table"}
	m := plugin.Manifest{ID: "tables", Version: "1", HostAPI: 1, Capabilities: []string{route.Capability}, Routes: []plugin.Route{route}}
	r, storage := accessRegistry(t, m)
	installAccessTestPlugin(t, r, "alpha", accessManifest(t), &contractWorker{ready: true})
	worker := &accessBlockingWorker{contractWorker: contractWorker{ready: true}, started: make(chan struct{}), release: make(chan struct{})}
	installAccessTestPlugin(t, r, "alpha", m, worker)
	revision := createAccessRule(t, r, "alpha", "write", "operator:root", m, route, "users", true)
	req := plugin.WorkerRequest{Protocol: 1, ProjectID: "alpha", PluginID: m.ID, Version: m.Version, Operation: route.Operation, Payload: []byte(`{"table":"users","row":{"id":"a"}}`)}
	done := make(chan error, 1)
	go func() {
		_, err := r.InvokeAsOperator(req, "root")
		done <- err
	}()
	select {
	case <-worker.started:
	case err := <-done:
		t.Fatalf("worker not reached: %v", err)
	case <-time.After(5 * time.Second):
		t.Fatal("worker admission timed out")
	}
	revokeAccessRule(t, r, "alpha", "write", revision)
	close(worker.release)
	select {
	case err := <-done:
		if !errors.Is(err, plugin.ErrActorDenied) {
			t.Fatalf("in-flight revoke not enforced: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("commit timed out")
	}
	rows, err := storage.Rows("alpha", "users")
	if err != nil || len(rows) != 0 {
		t.Fatalf("denied write committed: %+v %v", rows, err)
	}
}

func TestAccessManagementPersistenceAndUnavailableProvider(t *testing.T) {
	route := plugin.Route{Operation: "sample.read", Function: "read", Capability: "sample.read"}
	m := plugin.Manifest{ID: "sample", Version: "1", HostAPI: 1, Capabilities: []string{route.Capability}, Routes: []plugin.Route{route}}
	r, storage := accessRegistry(t, m)
	installAccessTestPlugin(t, r, "alpha", m, &contractWorker{ready: true})
	req := plugin.WorkerRequest{Protocol: 1, ProjectID: "alpha", PluginID: m.ID, Version: m.Version, Operation: route.Operation, Payload: []byte("{}")}
	if _, err := r.Invoke(req); err != nil {
		t.Fatalf("legacy project: %v", err)
	}
	installAccessTestPlugin(t, r, "alpha", accessManifest(t), &contractWorker{ready: true})
	createAccessRule(t, r, "alpha", "host", "host", m, route, "", true)
	restored := plugin.NewRegistry(1)
	for _, manifest := range []plugin.Manifest{accessManifest(t), m} {
		if err := restored.Register(manifest); err != nil {
			t.Fatal(err)
		}
	}
	if err := restored.BindSQLiteStorage(storage); err != nil {
		t.Fatal(err)
	}
	if err := restored.BindWorker("alpha", m.ID, m.Version, &contractWorker{ready: true}); err != nil {
		t.Fatal(err)
	}
	if result, err := restored.Invoke(req); err != nil || !result.Success {
		t.Fatalf("persisted policy: %+v %v", result, err)
	}
	if err := restored.SetReadiness(accessID, accessVersion, false); err != nil {
		t.Fatal(err)
	}
	if _, err := restored.Invoke(req); !errors.Is(err, plugin.ErrActorDenied) {
		t.Fatalf("unavailable policy failed open: %v", err)
	}
}

func TestAccessManagementDelegationRequiresSourceAndServiceRules(t *testing.T) {
	targetRoute := plugin.Route{Operation: "read", Function: "read", Capability: "read", Helper: "sdk.response", Response: &plugin.ResponseSpec{Output: map[string]any{"ok": true}}}
	target := plugin.Manifest{ID: "target", Version: "1", HostAPI: 1, Declarative: true, Capabilities: []string{"read"}, Routes: []plugin.Route{targetRoute}}
	sourceRoute := plugin.Route{Operation: "read", Function: "read", Capability: "call.target.read", Helper: "host.call", Call: &plugin.PluginCall{Plugin: "target", Version: "1", Operation: "read"}}
	source := plugin.Manifest{ID: "source", Version: "1", HostAPI: 1, Declarative: true, OptionalDependencies: []string{"target@1"}, Capabilities: []string{sourceRoute.Capability}, Routes: []plugin.Route{sourceRoute}}
	r, _ := accessRegistry(t, source, target)
	for _, m := range []plugin.Manifest{accessManifest(t), source, target} {
		installAccessTestPlugin(t, r, "alpha", m, &contractWorker{ready: true})
	}
	req := plugin.WorkerRequest{Protocol: 1, ProjectID: "alpha", PluginID: source.ID, Version: source.Version, Operation: sourceRoute.Operation, Payload: []byte("{}")}
	createAccessRule(t, r, "alpha", "source", "operator:root", source, sourceRoute, "", true)
	createAccessRule(t, r, "alpha", "operator-target", "operator:root", target, targetRoute, "", true)
	if _, err := r.InvokeAsOperator(req, "root"); !errors.Is(err, plugin.ErrActorDenied) {
		t.Fatalf("delegation inherited operator identity: %v", err)
	}
	createAccessRule(t, r, "alpha", "service-target", "plugin:source@1", target, targetRoute, "", true)
	if result, err := r.InvokeAsOperator(req, "root"); err != nil || !result.Success {
		t.Fatalf("authorized delegation: %+v %v", result, err)
	}
}

type accessObservationWorker struct{ contractWorker }

func (w *accessObservationWorker) Handle(req plugin.WorkerRequest) plugin.WorkerResponse {
	return plugin.WorkerResponse{Protocol: req.Protocol, Success: true, Payload: req.Payload, Observations: []plugin.ObservationIntent{{Event: "connected", Payload: json.RawMessage("{}")}}}
}

func TestAccessManagementObserverUsesPublisherIdentity(t *testing.T) {
	publishRoute := plugin.Route{Operation: "publish", Function: "publish", Capability: "publish"}
	source := plugin.Manifest{ID: "publisher", Version: "1", HostAPI: 1, Capabilities: []string{"publish"}, Routes: []plugin.Route{publishRoute}, Events: []plugin.EventSpec{{Name: "connected", Fields: map[string]string{}}}}
	observationCapability := "observe.publisher.connected"
	target := plugin.Manifest{ID: "observer", Version: "1", HostAPI: 1, Capabilities: []string{observationCapability}, Observers: []plugin.ObserverSpec{{Plugin: source.ID, Event: "connected", Function: "observe", Capability: observationCapability}}}
	r, _ := accessRegistry(t, source, target)
	installAccessTestPlugin(t, r, "alpha", accessManifest(t), &contractWorker{ready: true})
	installAccessTestPlugin(t, r, "alpha", source, &accessObservationWorker{contractWorker{ready: true}})
	worker := &countingWorker{}
	installAccessTestPlugin(t, r, "alpha", target, worker)
	createAccessRule(t, r, "alpha", "publisher", "operator:root", source, publishRoute, "", true)
	req := plugin.WorkerRequest{Protocol: 1, ProjectID: "alpha", PluginID: source.ID, Version: source.Version, Operation: publishRoute.Operation, Payload: []byte("{}")}
	if result, err := r.InvokeAsOperator(req, "root"); err != nil || !result.Success || worker.calls != 0 {
		t.Fatalf("ungranted observation: %+v %v calls=%d", result, err, worker.calls)
	}
	observerRoute := plugin.Route{Operation: "@observe/publisher/connected", Capability: observationCapability}
	createAccessRule(t, r, "alpha", "observer", "plugin:publisher@1", target, observerRoute, "publisher.connected", true)
	if result, err := r.InvokeAsOperator(req, "root"); err != nil || !result.Success || worker.calls != 0 {
		t.Fatalf("queued observation: %+v %v calls=%d", result, err, worker.calls)
	}
	if _, err := r.DrainObservations(time.Now().Add(time.Second), 10); err != nil || worker.calls != 1 {
		t.Fatalf("authorized observation delivery: %v calls=%d", err, worker.calls)
	}
	auth := worker.last.HostContext.Auth
	if auth.Actor.ID != "root" || auth.Actor.Kind != "operator" || auth.Caller.Kind != "plugin" || auth.Caller.ID != "publisher@1" {
		t.Fatalf("observer principal mismatch: %+v", auth)
	}
}

func TestAccessManagementHostAdapterUsesExactPublicPrincipal(t *testing.T) {
	route := plugin.Route{Operation: "functions.call", Function: "call", Capability: "functions.call", ResourceField: "name"}
	m := plugin.Manifest{ID: "functions", Version: "1", HostAPI: 1, Capabilities: []string{route.Capability}, Routes: []plugin.Route{route}}
	r, _ := accessRegistry(t, m)
	installAccessTestPlugin(t, r, "alpha", accessManifest(t), &contractWorker{ready: true})
	installAccessTestPlugin(t, r, "alpha", m, &contractWorker{ready: true})
	createAccessRule(t, r, "alpha", "operator", "operator:root", m, route, "public-function", true)
	if err := r.AuthorizeHostOperation("alpha", "public", m.ID, route.Operation, "public-function"); !errors.Is(err, plugin.ErrActorDenied) {
		t.Fatalf("anonymous inherited operator policy: %v", err)
	}
	revision := createAccessRule(t, r, "alpha", "public", "public", m, route, "public-function", true)
	if err := r.AuthorizeHostOperation("alpha", "public", m.ID, route.Operation, "public-function"); err != nil {
		t.Fatal(err)
	}
	if err := r.AuthorizeHostOperation("alpha", "public", m.ID, route.Operation, "private-function"); !errors.Is(err, plugin.ErrActorDenied) {
		t.Fatalf("anonymous crossed resource boundary: %v", err)
	}
	revokeAccessRule(t, r, "alpha", "public", revision)
	if err := r.AuthorizeHostOperation("alpha", "public", m.ID, route.Operation, "public-function"); !errors.Is(err, plugin.ErrActorDenied) {
		t.Fatalf("revoked public rule admitted: %v", err)
	}
}

func TestAccessManagementFunctionServiceCannotBypassRules(t *testing.T) {
	m, err := plugin.LoadManifest("../../plugins/functions/plugin.yaml")
	if err != nil {
		t.Fatal(err)
	}
	r, _ := accessRegistry(t, m)
	const project = "access-endpoint-test"
	const name = "access_endpoint"
	const source = "export const config = { method: 'POST' }; export default (input) => input"
	installAccessTestPlugin(t, r, project, accessManifest(t), &contractWorker{ready: true})
	installAccessTestPlugin(t, r, project, m, &contractWorker{ready: true})
	service := admin.NewLifecycleService(nil, r)
	actor := operator.Actor{ID: "root", Kind: operator.ActorOperator}
	if _, err := service.CreateFunctionDraft(actor, project, name, source); !errors.Is(err, plugin.ErrActorDenied) {
		t.Fatalf("direct draft creation bypassed access rules: %v", err)
	}
	publish, _ := m.Route("functions.publish")
	createAccessRule(t, r, project, "publish", "operator:root", m, publish, name, true)
	if _, err := service.CreateFunctionDraft(actor, project, name, source); err != nil {
		t.Fatal(err)
	}
	if drafts, err := service.ListFunctionDrafts(actor, project); err != nil || len(drafts) != 0 {
		t.Fatalf("ungranted draft visibility: %v %v", drafts, err)
	}
	inspect, _ := m.Route("functions.inspect")
	createAccessRule(t, r, project, "inspect", "operator:root", m, inspect, name, true)
	if drafts, err := service.ListFunctionDrafts(actor, project); err != nil || len(drafts) != 1 || drafts[0].Name != name {
		t.Fatalf("granted draft visibility: %v %v", drafts, err)
	}
	if _, err := service.TestFunctionDraft(actor, project, name, source); !errors.Is(err, plugin.ErrActorDenied) {
		t.Fatalf("direct preflight bypassed access rules: %v", err)
	}
	invoke, _ := m.Route("functions.invoke")
	createAccessRule(t, r, project, "invoke", "operator:root", m, invoke, name, true)
	if result, err := service.TestFunctionDraft(actor, project, name, source); err != nil || result["valid"] != true {
		t.Fatalf("granted preflight: %v %v", result, err)
	}
	if _, err := service.PublishFunction(actor, project, name, "POST"); err != nil {
		t.Fatal(err)
	}
	if _, err := service.PublicInvokeFunction(project, name, "GET", nil); !errors.Is(err, plugin.ErrActorDenied) {
		t.Fatalf("public endpoint disclosed method before access check: %v", err)
	}
	call, _ := m.Route("functions.call")
	createAccessRule(t, r, project, "public", "public", m, call, name, true)
	if _, err := service.PublicInvokeFunction(project, name, "GET", nil); !errors.Is(err, admin.ErrFunctionMethodNotAllowed) {
		t.Fatalf("public rule bypassed method check: %v", err)
	}
}
