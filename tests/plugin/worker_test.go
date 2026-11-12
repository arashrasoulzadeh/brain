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
