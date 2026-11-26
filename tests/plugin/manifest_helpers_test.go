package plugin_test

import (
	"errors"
	"testing"

	"brain/internal/plugin"
)

func TestManifestRouteHelperAllowlist(t *testing.T) {
	manifest := plugin.Manifest{
		SourceFiles:  []string{"index.ts"},
		Capabilities: []string{"sample.read"},
		Routes: []plugin.Route{
			{Operation: "sample.read", Function: "read", Capability: "sample.read", Helper: "sdk.dispatch"},
		},
	}
	if err := plugin.ValidateManifestContract(manifest); err != nil {
		t.Fatalf("declared SDK helper rejected: %v", err)
	}
	manifest.Routes[0].Helper = "arbitrary.module"
	if err := plugin.ValidateManifestContract(manifest); !errors.Is(err, plugin.ErrInvalidManifest) {
		t.Fatalf("unknown helper accepted: %v", err)
	}
}
