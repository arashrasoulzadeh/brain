package plugin_test

import (
	"errors"
	"testing"

	"brain/internal/plugin"
)

func TestHTTPEnvelopeRequiresVersionedProjectScopedFields(t *testing.T) {
	envelope := plugin.HTTPRequestEnvelope{Protocol: 1, RequestID: "req-1", ProjectID: "dev", PluginID: "rooms", Version: "1.0.0", Operation: "rooms.create", Payload: []byte(`{}`)}
	if err := envelope.Validate(1); err != nil {
		t.Fatal(err)
	}
	envelope.Protocol = 2
	if !errors.Is(envelope.Validate(1), plugin.ErrInvalidHTTPEnvelope) {
		t.Fatal("wrong protocol accepted")
	}
}
