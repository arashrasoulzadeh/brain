package plugin_test

import (
	"errors"
	"reflect"
	"testing"

	"brain/internal/plugin"
)

func TestInvocationContextExposesOnlyReadyAuthorizedDependencies(t *testing.T) {
	r := plugin.NewRegistry(1)
	dependency := plugin.Manifest{ID: "shared", Version: "1", HostAPI: 1, Capabilities: []string{"read"}}
	next := dependency
	next.Version = "2"
	m := coverageManifest()
	m.ConfigSchema = map[string]string{"name": "string", "count": "int", "ready": "bool"}
	m.OptionalDependencies = []string{"shared@1", "shared@2", "missing@1"}
	for _, manifest := range []plugin.Manifest{dependency, next, m} {
		if err := r.Register(manifest); err != nil {
			t.Fatal(err)
		}
		installAccessTestPlugin(t, r, "alpha", manifest, &contractWorker{ready: true})
	}
	if err := r.Configure("alpha", m.ID, m.Version, map[string]any{"name": "fixture", "count": float64(2), "ready": true}); err != nil {
		t.Fatal(err)
	}
	worker := &countingWorker{}
	if err := r.BindWorker("alpha", m.ID, m.Version, worker); err != nil {
		t.Fatal(err)
	}
	request := plugin.WorkerRequest{Protocol: 1, ProjectID: "alpha", PluginID: m.ID, Version: m.Version, Operation: "coverage.read", Payload: []byte(`{}`)}
	check := func(want []string) {
		t.Helper()
		if response, err := r.InvokeAsOperator(request, "root"); err != nil || !response.Success {
			t.Fatalf("invoke: %v", err)
		}
		ctx := worker.last.HostContext
		if ctx == nil || !reflect.DeepEqual(ctx.Available, want) || ctx.Config["name"] != "fixture" || ctx.Config["count"] != float64(2) || ctx.Config["ready"] != true {
			t.Fatalf("incorrect dependency or configuration context: %+v", ctx)
		}
	}
	check([]string{"shared@1", "shared@2"})
	if err := r.Promote("alpha", "shared", "2"); err != nil {
		t.Fatal(err)
	}
	check([]string{"shared@2"})
	if err := r.Revoke("alpha", "shared", "2", "read"); err != nil {
		t.Fatal(err)
	}
	check([]string{})
	if err := r.Grant("alpha", "shared", "2", "read"); err != nil {
		t.Fatal(err)
	}
	if err := r.Drain("alpha", "shared", "2"); err != nil {
		t.Fatal(err)
	}
	check([]string{})
	if _, err := r.InvokeAsOperator(request, "bad\nactor"); !errors.Is(err, plugin.ErrActorDenied) {
		t.Fatalf("invalid actor accepted: %v", err)
	}
	for _, deps := range [][]string{{"coverage@1"}, {"shared@1", "shared@1"}, {"missing"}, make([]string, 33)} {
		invalid := coverageManifest()
		invalid.OptionalDependencies = deps
		if err := plugin.ValidateManifestContract(invalid); !errors.Is(err, plugin.ErrInvalidManifest) {
			t.Fatalf("invalid dependency metadata accepted: %v", err)
		}
	}
}

func TestDelegatedPluginCallRejectsInvalidWorkerResultsAndRevocation(t *testing.T) {
	r := plugin.NewRegistry(1)
	target := plugin.Manifest{ID: "target", Version: "1", HostAPI: 1, Declarative: true, Capabilities: []string{"read"}, Routes: []plugin.Route{{Operation: "read", Function: "read", Capability: "read", Helper: "sdk.response", Response: &plugin.ResponseSpec{Output: map[string]any{"ok": true}}}}}
	source := plugin.Manifest{ID: "source", Version: "1", HostAPI: 1, Declarative: true, OptionalDependencies: []string{"target@1"}, Capabilities: []string{"call.target.read"}, Routes: []plugin.Route{{Operation: "read", Function: "read", Capability: "call.target.read", Helper: "host.call", Call: &plugin.PluginCall{Plugin: "target", Version: "1", Operation: "read"}}}}
	for _, m := range []plugin.Manifest{source, target} {
		if err := r.Register(m); err != nil {
			t.Fatal(err)
		}
		installAccessTestPlugin(t, r, "alpha", m, &contractWorker{ready: true})
	}
	request := plugin.WorkerRequest{Protocol: 1, ProjectID: "alpha", PluginID: "source", Version: "1", Operation: "read", Payload: []byte(`{}`)}
	invalid := request
	invalid.Protocol = 2
	if _, err := r.Invoke(invalid); !errors.Is(err, plugin.ErrInvalidWorkerRequest) {
		t.Fatalf("wrong protocol admitted: %v", err)
	}
	r.UnbindWorker("alpha", "target", "1")
	if _, err := r.Invoke(request); !errors.Is(err, plugin.ErrWorkerNotReady) {
		t.Fatalf("missing target admitted: %v", err)
	}
	for _, response := range []plugin.WorkerResponse{{Protocol: 2, Success: true}, {Protocol: 1, Success: true, Observations: []plugin.ObservationIntent{{Event: "forbidden"}}}} {
		if err := r.BindWorker("alpha", "target", "1", &storageCommitWorker{handle: func(plugin.WorkerRequest) plugin.WorkerResponse { return response }}); err != nil {
			t.Fatal(err)
		}
		if _, err := r.Invoke(request); !errors.Is(err, plugin.ErrInvalidWorkerRequest) {
			t.Fatalf("invalid result admitted: %v", err)
		}
	}
	if err := r.BindWorker("alpha", "target", "1", &storageCommitWorker{handle: func(plugin.WorkerRequest) plugin.WorkerResponse {
		if err := r.Revoke("alpha", "target", "1", "read"); err != nil {
			t.Fatal(err)
		}
		return plugin.WorkerResponse{Protocol: 1, Success: true}
	}}); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Invoke(request); !errors.Is(err, plugin.ErrGrantDenied) {
		t.Fatalf("revocation during delegation ignored: %v", err)
	}
}
