package plugin

import (
	"fmt"
	"strings"
)

// TypeScriptContract emits deterministic types from the validated manifest.
// Runtime permissions and validation remain the host's responsibility.
func TypeScriptContract(m Manifest) string {
	var b strings.Builder
	b.WriteString("// Generated from plugin.yaml. Do not edit.\n\nexport interface Models {\n")
	for _, name := range sortedContractKeys(m.Models) {
		model := m.Models[name]
		fmt.Fprintf(&b, "  %q: {\n", name)
		for _, field := range sortedContractKeys(model.Fields) {
			fmt.Fprintf(&b, "    %q: %s;\n", field, contractType(model.Fields[field]))
		}
		b.WriteString("  };\n")
	}
	b.WriteString("}\n\nexport interface Operations {\n")
	for _, route := range m.Routes {
		input := "unknown"
		if route.Model != nil {
			fields := fmt.Sprintf("Models[%q]", route.Model.Name)
			switch route.Model.Action {
			case "create":
				input = "{ id: string; fields: " + fields + " }"
			case "read":
				input = "{ id: string }"
			case "list":
				input = "{ after?: string }"
			case "update":
				input = "{ id: string; revision: string; fields: " + fields + " }"
			case "delete", "transition":
				input = "{ id: string; revision: string }"
			}
		}
		fmt.Fprintf(&b, "  %q: %s;\n", route.Operation, input)
	}
	b.WriteString("}\n\nexport interface Events {\n")
	for _, event := range m.Events {
		fmt.Fprintf(&b, "  %q: {\n", event.Name)
		for _, field := range sortedContractKeys(event.Fields) {
			fmt.Fprintf(&b, "    %q: %s;\n", field, contractType(event.Fields[field]))
		}
		b.WriteString("  };\n")
	}
	b.WriteString("}\n\nexport type Operation = keyof Operations;\n")
	return b.String()
}

func contractType(kind string) string {
	switch kind {
	case "integer", "number":
		return "number"
	case "string", "boolean":
		return kind
	default:
		return "unknown"
	}
}
