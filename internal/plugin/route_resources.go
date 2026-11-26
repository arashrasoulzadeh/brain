package plugin

import (
	"encoding/json"
	"strings"
)

// A resource-scoped grant is never treated as an unrestricted grant.
// Model operations use the manifest model name rather than caller input.
func routeResourceAllowed(route Route, payload []byte, resources map[string]bool) bool {
	if len(resources) == 0 {
		return true
	}
	if route.Model != nil {
		return resources[route.Model.Name]
	}
	if route.ResourceField == "" {
		return false
	}
	var input map[string]any
	if json.Unmarshal(payload, &input) != nil {
		return false
	}
	resource, ok := input[route.ResourceField].(string)
	return ok && strings.TrimSpace(resource) != "" && resources[strings.TrimSpace(resource)]
}
