package plugin

// AllowedWebSocketOrigin applies an exact allowlist comparison. Wildcards and
// implicit subdomains are intentionally unsupported at this trust boundary.
func AllowedWebSocketOrigin(origin string, allowed []string) bool {
	if origin == "" {
		return false
	}
	for _, candidate := range allowed {
		if origin == candidate {
			return true
		}
	}
	return false
}
