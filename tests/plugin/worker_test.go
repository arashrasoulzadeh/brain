package plugin_test

import (
	"errors"
	"testing"

	"brain/internal/plugin"
)

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

type readyWorker struct{}

func (readyWorker) Start() error { return nil }
func (readyWorker) Stop() error  { return nil }
func (readyWorker) Ready() bool  { return true }
func (readyWorker) Handle(request plugin.WorkerRequest) plugin.WorkerResponse {
	return plugin.WorkerResponse{Protocol: request.Protocol, Success: true}
}
