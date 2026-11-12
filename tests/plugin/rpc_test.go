package plugin_test

import (
	"errors"
	"testing"

	"brain/internal/plugin"
)

func TestRPCRequestUsesSharedWorkerContract(t *testing.T) {
	message := plugin.RPCRequest{JSONRPC: "2.0", ID: "req-1", Method: "invoke", Params: plugin.WorkerRequest{Protocol: 1, ProjectID: "dev", PluginID: "auth", Version: "1", Operation: "signIn"}}
	if err := plugin.ValidateRPCRequest(message, 1); err != nil {
		t.Fatal(err)
	}
	message.ID = ""
	if !errors.Is(plugin.ValidateRPCRequest(message, 1), plugin.ErrInvalidRPCMessage) {
		t.Fatal("invalid rpc message accepted")
	}
}
