package plugin

import (
	"strings"
	"testing"
)

func TestGeneratedContractUsesManifestModelAndEventTypes(t *testing.T) {
	m := Manifest{
		Models: map[string]ModelSpec{"item": {Version: 1, Fields: map[string]string{"name": "string", "count": "integer"}}},
		Routes: []Route{{Operation: "item.update", Model: &ModelRoute{Name: "item", Action: "update"}}},
		Events: []EventSpec{{Name: "updated", Fields: map[string]string{"ok": "boolean"}}},
	}
	contract := TypeScriptContract(m)
	for _, fragment := range []string{`"count": number;`, `"name": string;`, `"item.update": { id: string; revision: string; fields: Models["item"] };`, `"ok": boolean;`} {
		if !strings.Contains(contract, fragment) {
			t.Fatalf("contract missing %q: %s", fragment, contract)
		}
	}
	if contract != TypeScriptContract(m) {
		t.Fatal("contract generation must be deterministic")
	}
}
