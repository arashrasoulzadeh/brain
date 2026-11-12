package plugin_test

import (
	"testing"

	"brain/internal/plugin"
)

func TestExternalWorkerUsesSameContractFixture(t *testing.T) {
	handshake := plugin.Handshake{Protocol: 1, ProjectID: "dev", PluginID: "arena", Version: "1", ServiceID: "arena-dev"}
	if err := plugin.ValidateHandshake(handshake, 1, "arena-dev"); err != nil {
		t.Fatal(err)
	}
	worker := &contractWorker{ready: true}
	request := plugin.WorkerRequest{Protocol: 1, ProjectID: handshake.ProjectID, PluginID: handshake.PluginID, Version: handshake.Version, Operation: "fixture", Payload: []byte(`{"ok":true}`)}
	response, err := plugin.Invoke(worker, request, handshake.Protocol)
	if err != nil || !response.Success {
		t.Fatalf("external fixture response = %#v, %v", response, err)
	}
}
