package plugin

import (
	"regexp"
	"strings"
)

// HTTP endpoints are an opt-in projection of the authoritative plugin routes.
// A route without method/path remains callable only through existing host RPC.
var httpSegment = regexp.MustCompile(`^[a-z][a-z0-9-]*$`)
var httpFunction = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_]*$`)
var httpParameter = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_]*$`)

func validHTTPPath(value string, mount bool) bool {
	if value == "" || value[0] != '/' || strings.ContainsAny(value, "%\\?#") || strings.Contains(value, "//") {
		return false
	}
	if mount {
		return httpSegment.MatchString(strings.TrimPrefix(value, "/")) &&
			value != "/_system" && value != "/admin" && value != "/functions" && value != "/ws"
	}
	if value == "/" {
		return true
	}
	for _, part := range strings.Split(strings.TrimPrefix(value, "/"), "/") {
		if part == "" || part == "." || part == ".." {
			return false
		}
		if strings.HasPrefix(part, "{") && strings.HasSuffix(part, "}") {
			if !httpParameter.MatchString(part[1 : len(part)-1]) {
				return false
			}
		} else if !httpSegment.MatchString(part) {
			return false
		}
	}
	return true
}

func httpRouteKey(method, path string) string {
	segments := strings.Split(path, "/")
	for i, segment := range segments {
		if strings.HasPrefix(segment, "{") && strings.HasSuffix(segment, "}") {
			segments[i] = "{}"
		}
	}
	return method + " " + strings.Join(segments, "/")
}

func validHTTPRoutes(m Manifest) bool {
	if m.Mount != "" && !validHTTPPath(m.Mount, true) {
		return false
	}
	seen := make(map[string]bool)
	patterns := make(map[string][]string)
	for _, route := range m.Routes {
		if route.Method == "" && route.Path == "" && route.Auth == "" {
			continue
		}
		if m.Mount == "" || !validHTTPPath(route.Path, false) || !httpFunction.MatchString(route.Function) {
			return false
		}
		switch route.Method {
		case "GET", "POST", "PUT", "PATCH", "DELETE":
		default:
			return false
		}
		// Public exposure is never implicit. User authentication is verified
		// through the host's existing project-scoped session resolver.
		if route.Auth != "public" && route.Auth != "operator" && route.Auth != "user" {
			return false
		}
		key := httpRouteKey(route.Method, route.Path)
		if seen[key] || !uniqueHTTPParameters(route.Path) {
			return false
		}
		for _, existing := range patterns[route.Method] {
			if overlappingHTTPPaths(existing, route.Path) {
				return false
			}
		}
		seen[key] = true
		patterns[route.Method] = append(patterns[route.Method], route.Path)
	}
	return true
}

func uniqueHTTPParameters(path string) bool {
	seen := make(map[string]bool)
	for _, segment := range strings.Split(path, "/") {
		if strings.HasPrefix(segment, "{") && strings.HasSuffix(segment, "}") {
			name := segment[1 : len(segment)-1]
			if seen[name] {
				return false
			}
			seen[name] = true
		}
	}
	return true
}

func overlappingHTTPPaths(left, right string) bool {
	first, second := strings.Split(strings.Trim(left, "/"), "/"), strings.Split(strings.Trim(right, "/"), "/")
	if len(first) != len(second) {
		return false
	}
	for i, segment := range first {
		if segment != second[i] && !strings.HasPrefix(segment, "{") && !strings.HasPrefix(second[i], "{") {
			return false
		}
	}
	return true
}
