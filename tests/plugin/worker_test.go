package plugin_test

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"brain/internal/plugin"
)

func TestInvokeWithCredentialBindsProductionWorkerRequest(t *testing.T) {
	r := plugin.NewRegistry(1)
	m := plugin.Manifest{ID: "sample", Version: "1.0.0", HostAPI: 1, Capabilities: []string{"read"}, Routes: []plugin.Route{{Operation: "read", Function: "read", Capability: "read"}}}
	if err := r.Register(m); err != nil {
		t.Fatal(err)
	}
	if err := r.Install("project-a", m.ID, m.Version); err != nil {
		t.Fatal(err)
	}
	if err := r.Grant("project-a", m.ID, m.Version, "read"); err != nil {
		t.Fatal(err)
	}
	if err := r.SetReadiness(m.ID, m.Version, true); err != nil {
		t.Fatal(err)
	}
	if err := r.Enable("project-a", m.ID, m.Version); err != nil {
		t.Fatal(err)
	}
	b := &plugin.Broker{Registry: r, ActorAllows: func(string, string, string) bool { return true }}
	now := time.Unix(100, 0)
	credential, err := b.Issue(plugin.HostCall{ProjectID: "project-a", ActorID: "actor", PluginID: m.ID, Version: m.Version, Capability: "read", Operation: "read"}, now, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	worker := &storageCommitWorker{handle: func(request plugin.WorkerRequest) plugin.WorkerResponse {
		if request.Project == nil || request.Project.ID != "project-a" || !request.Project.Active {
			t.Fatalf("credential did not rebind project model: %+v", request.Project)
		}
		return plugin.WorkerResponse{Protocol: 1, Success: true}
	}}
	request := plugin.WorkerRequest{Protocol: 1, ProjectID: "project-b", PluginID: m.ID, Version: m.Version, Operation: "read"}
	if _, err := plugin.InvokeWithCredential(b, credential, worker, request, 1, now); !errors.Is(err, plugin.ErrInvalidWorkerRequest) {
		t.Fatalf("mismatched request error = %v", err)
	}
	request.ProjectID = "project-a"
	request.Project = &plugin.ProjectSnapshot{ID: "forged", Active: false}
	if response, err := plugin.InvokeWithCredential(b, credential, worker, request, 1, now); err != nil || !response.Success {
		t.Fatalf("valid credential invocation = %#v, %v", response, err)
	}
}

type countingWorker struct{ calls int }

func (w *countingWorker) Start() error { return nil }
func (w *countingWorker) Stop() error  { return nil }
func (w *countingWorker) Ready() bool  { return true }
func (w *countingWorker) Handle(plugin.WorkerRequest) plugin.WorkerResponse {
	w.calls++
	return plugin.WorkerResponse{Protocol: 1, Success: true}
}

func TestWorkerRequestContractRequiresScopeAndProtocol(t *testing.T) {
	request := plugin.WorkerRequest{Protocol: 1, ProjectID: "dev", PluginID: "arena", Version: "1", Operation: "start", Payload: []byte("{}")}
	if err := plugin.ValidateWorkerRequest(request, 1); err != nil {
		t.Fatal(err)
	}
	request.ProjectID = ""
	if !errors.Is(plugin.ValidateWorkerRequest(request, 1), plugin.ErrInvalidWorkerRequest) {
		t.Fatal("unscoped request accepted")
	}
}

func TestWorkerRequestRejectsWrongProtocolAndOversizedPayload(t *testing.T) {
	request := plugin.WorkerRequest{Protocol: 2, ProjectID: "dev", PluginID: "arena", Version: "1", Operation: "start"}
	if !errors.Is(plugin.ValidateWorkerRequest(request, 1), plugin.ErrInvalidWorkerRequest) {
		t.Fatal("wrong protocol accepted")
	}
	request.Protocol = 1
	request.Payload = make([]byte, 1<<20+1)
	if !errors.Is(plugin.ValidateWorkerRequest(request, 1), plugin.ErrInvalidWorkerRequest) {
		t.Fatal("oversized payload accepted")
	}
}

func TestWorkerRequestRejectsHostProjectMismatch(t *testing.T) {
	request := plugin.WorkerRequest{Protocol: 1, ProjectID: "dev", PluginID: "rooms", Version: "1.0.0", Operation: "rooms.create"}
	if err := plugin.ValidateWorkerRequestForProject(request, 1, "prod"); !errors.Is(err, plugin.ErrInvalidWorkerRequest) {
		t.Fatalf("cross-project worker request accepted: %v", err)
	}
	if err := plugin.ValidateWorkerRequestForProject(request, 1, "dev"); err != nil {
		t.Fatal(err)
	}
}

func TestInvokeForProjectBindsExecutionToHostContext(t *testing.T) {
	worker := &readyWorker{}
	request := plugin.WorkerRequest{Protocol: 1, ProjectID: "client-project", PluginID: "sample", Version: "1.0.0", Operation: "health"}
	if _, err := plugin.InvokeForProject(worker, request, 1, "host-project"); !errors.Is(err, plugin.ErrInvalidWorkerRequest) {
		t.Fatalf("client project identity was trusted: %v", err)
	}
}

func TestInvokeForProjectFailsClosedWithoutWorker(t *testing.T) {
	request := plugin.WorkerRequest{Protocol: 1, ProjectID: "dev", PluginID: "sample", Version: "1.0.0", Operation: "health"}
	if _, err := plugin.InvokeForProject(nil, request, 1, "dev"); !errors.Is(err, plugin.ErrWorkerNotReady) {
		t.Fatalf("nil worker error = %v", err)
	}
}

type readyWorker struct{}

func (readyWorker) Start() error { return nil }
func (readyWorker) Stop() error  { return nil }
func (readyWorker) Ready() bool  { return true }
func (readyWorker) Handle(request plugin.WorkerRequest) plugin.WorkerResponse {
	return plugin.WorkerResponse{Protocol: request.Protocol, Success: true}
}

func TestCredentialInvocationSharesObservationAndCommitBoundary(t *testing.T) {
	r := plugin.NewRegistry(1)
	publisher := plugin.Manifest{ID: "external", Version: "1", HostAPI: 1, Capabilities: []string{"external.write"}, Routes: []plugin.Route{{Operation: "external.write", Function: "write", Capability: "external.write"}}, Events: []plugin.EventSpec{{Name: "written", Fields: map[string]string{}}}}
	observer := plugin.Manifest{ID: "observer", Version: "1", HostAPI: 1, Capabilities: []string{"observe.external.written"}, Observers: []plugin.ObserverSpec{{Plugin: "external", Event: "written", Function: "record", Capability: "observe.external.written"}}}
	for _, m := range []plugin.Manifest{publisher, observer} {
		if err := r.Register(m); err != nil {
			t.Fatal(err)
		}
		if err := r.Install("project", m.ID, m.Version); err != nil {
			t.Fatal(err)
		}
		if err := r.SetReadiness(m.ID, m.Version, true); err != nil {
			t.Fatal(err)
		}
		for _, capability := range m.Capabilities {
			if err := r.Grant("project", m.ID, m.Version, capability); err != nil {
				t.Fatal(err)
			}
		}
		if err := r.Enable("project", m.ID, m.Version); err != nil {
			t.Fatal(err)
		}
	}
	delivered := 0
	if err := r.BindWorker("project", "observer", "1", &storageCommitWorker{handle: func(request plugin.WorkerRequest) plugin.WorkerResponse {
		var event plugin.Observation
		if err := json.Unmarshal(request.Payload, &event); err != nil || event.ProjectID != "project" || event.PluginID != "external" {
			t.Fatalf("unexpected observation: %+v %v", event, err)
		}
		delivered++
		return plugin.WorkerResponse{Protocol: 1, Success: true}
	}}); err != nil {
		t.Fatal(err)
	}
	allowed := true
	broker := &plugin.Broker{Registry: r, ActorAllows: func(string, string, string) bool { return allowed }}
	now := time.Now()
	call := plugin.HostCall{ProjectID: "project", ActorID: "actor", PluginID: "external", Version: "1", Capability: "external.write", Operation: "external.write"}
	request := plugin.WorkerRequest{Protocol: 1, ProjectID: "project", PluginID: "external", Version: "1", Operation: "external.write"}
	for _, revoke := range []bool{false, true} {
		allowed = true
		credential, err := broker.Issue(call, now, time.Minute)
		if err != nil {
			t.Fatal(err)
		}
		worker := &storageCommitWorker{handle: func(plugin.WorkerRequest) plugin.WorkerResponse {
			if revoke {
				allowed = false
			}
			return plugin.WorkerResponse{Protocol: 1, Success: true, Observations: []plugin.ObservationIntent{{Event: "written", Payload: json.RawMessage(`{}`)}}}
		}}
		response, err := plugin.InvokeWithCredential(broker, credential, worker, request, 1, now)
		if revoke {
			if !errors.Is(err, plugin.ErrActorDenied) || delivered != 1 {
				t.Fatalf("revoked actor published observation: %v, %d", err, delivered)
			}
		} else if err != nil || !response.Success || delivered != 1 || len(response.Observations) != 0 {
			t.Fatalf("credential invocation did not dispatch: %+v %v %d", response, err, delivered)
		}
	}
}
