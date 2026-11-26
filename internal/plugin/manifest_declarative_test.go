package plugin

import "testing"

func TestDeclarativeManifestRejectsExecutableSources(t *testing.T) {
	m := Manifest{
		ID: "example", Version: "1", HostAPI: 1, Declarative: true,
		Capabilities: []string{"read"},
		Routes: []Route{{Operation: "read", Function: "read", Capability: "read", Helper: "sdk.response", Response: &ResponseSpec{Output: map[string]any{"project": "$project.id"}}}},
	}
	if err := ValidateManifestContract(m); err != nil {
		t.Fatal(err)
	}
	if m.EntryPoint() != "" {
		t.Fatal("declarative manifest has an entry point")
	}
	m.SourceFiles = []string{"index.ts"}
	if err := ValidateManifestContract(m); err == nil {
		t.Fatal("declarative manifest accepted executable source")
	}
	m.SourceFiles = nil
	m.Routes[0].Helper = "sdk.dispatch"
	m.Routes[0].Response = nil
	if err := ValidateManifestContract(m); err == nil {
		t.Fatal("declarative manifest accepted executable dispatch")
	}
}
