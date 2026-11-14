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

func TestHTTPEnvelopeRejectsHostProjectMismatch(t *testing.T) {
	envelope := plugin.HTTPRequestEnvelope{Protocol: 1, RequestID: "req-1", ProjectID: "dev", PluginID: "rooms", Version: "1.0.0", Operation: "rooms.create"}
	if err := envelope.ValidateForProject(1, "prod"); !errors.Is(err, plugin.ErrInvalidHTTPEnvelope) {
		t.Fatalf("cross-project envelope accepted: %v", err)
	}
	if err := envelope.ValidateForProject(1, "dev"); err != nil {
		t.Fatal(err)
	}
}

func TestHTTPEnvelopeRejectsOversizedPayload(t *testing.T) {
	envelope := plugin.HTTPRequestEnvelope{Protocol: 1, RequestID: "req", ProjectID: "dev", PluginID: "files", Version: "1", Operation: "read", Payload: make([]byte, 1<<20+1)}
	if err := envelope.ValidateForProject(1, "dev"); !errors.Is(err, plugin.ErrInvalidHTTPEnvelope) {
		t.Fatalf("oversized HTTP payload accepted: %v", err)
	}
}
