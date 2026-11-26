package admin_test

import (
	"encoding/json"
	"errors"
	"testing"

	"brain/internal/admin"
	"brain/internal/operator"
	"brain/internal/plugin"
	"brain/internal/project"
)

type adminTestWorker struct{ calls int }

func (w *adminTestWorker) Start() error { return nil }
func (w *adminTestWorker) Stop() error  { return nil }
func (w *adminTestWorker) Ready() bool  { return true }
func (w *adminTestWorker) Handle(request plugin.WorkerRequest) plugin.WorkerResponse {
	w.calls++
	return plugin.WorkerResponse{Protocol: request.Protocol, Success: true, Payload: request.Payload}
}

func TestManifestAdminInvocationAuthorization(t *testing.T) {
	operators := &operator.Store{}
	digest := func(string) string { return "digest" }
	if _, err := operators.Bootstrap("admin", "secret", digest); err != nil {
		t.Fatal(err)
	}
	if err := operators.Authenticate("admin", "secret", digest); err != nil {
		t.Fatal(err)
	}
	registry := plugin.NewRegistry(1)
	m := plugin.Manifest{
		ID: "sample", Version: "1", HostAPI: 1,
		Capabilities: []string{"read", "write"},
		Routes: []plugin.Route{
			{Operation: "sample.read", Function: "read", Capability: "read"},
			{Operation: "sample.write", Function: "write", Capability: "write"},
		},
		Admin: &plugin.AdminSpec{Version: 1, Pages: []plugin.AdminPage{{
			ID: "records", Title: "Records", Actions: []plugin.AdminAction{
				{Title: "Read", Operation: "sample.read", ReadOnly: true},
				{Title: "Write", Operation: "sample.write", Fields: []plugin.AdminField{{Name: "value", Type: "string", Required: true}}},
			},
		}}},
	}
	if err := registry.Register(m); err != nil {
		t.Fatal(err)
	}
	if err := registry.Install("one", m.ID, m.Version); err != nil {
		t.Fatal(err)
	}
	for _, capability := range m.Capabilities {
		if err := registry.Grant("one", m.ID, m.Version, capability); err != nil {
			t.Fatal(err)
		}
	}
	if err := registry.SetReadiness(m.ID, m.Version, true); err != nil {
		t.Fatal(err)
	}
	if err := registry.Enable("one", m.ID, m.Version); err != nil {
		t.Fatal(err)
	}
	worker := &adminTestWorker{}
	if err := registry.BindWorker("one", m.ID, m.Version, worker); err != nil {
		t.Fatal(err)
	}
	projects := project.NewStore(project.Project{ID: "one", Active: true}, project.Project{ID: "two", Active: true})
	service := admin.NewLifecycleServiceWithProjects(operators, registry, projects)
	actor := operator.Actor{ID: "admin", Kind: operator.ActorOperator}
	request := admin.PluginInvocation{ProjectID: "one", PluginID: m.ID, Version: m.Version, Operation: "sample.read", Payload: json.RawMessage(`{}`)}
	if response, err := service.QueryAdminPage(actor, "records", request); err != nil || !response.Success {
		t.Fatalf("read query failed: %#v %v", response, err)
	}
	wrongProject := request
	wrongProject.ProjectID = "two"
	if _, err := service.QueryAdminPage(actor, "records", wrongProject); err == nil {
		t.Fatal("wrong-project query accepted")
	}
	if _, err := service.QueryAdminPage(actor, "missing", request); err == nil {
		t.Fatal("unknown page accepted")
	}
	write := request
	write.Operation = "sample.write"
	write.Payload = json.RawMessage(`{"value":"data"}`)
	if _, err := service.QueryAdminPage(actor, "records", write); !errors.Is(err, plugin.ErrGrantDenied) {
		t.Fatalf("write accepted through query: %v", err)
	}
	if _, err := service.InvokeAdminPage(actor, "records", write); err != nil {
		t.Fatal(err)
	}
	if err := registry.Revoke("one", m.ID, m.Version, "read"); err != nil {
		t.Fatal(err)
	}
	if _, err := service.QueryAdminPage(actor, "records", request); err == nil {
		t.Fatal("revoked query accepted")
	}
	if _, err := service.AdminPageRevision(actor, "records", request); err == nil {
		t.Fatal("revoked event feed accepted")
	}
	if worker.calls != 2 {
		t.Fatalf("worker calls = %d, want 2", worker.calls)
	}
}
