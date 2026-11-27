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
            if !httpParameter.MatchString(part[1:len(part)-1]) {
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
        // Public exposure is never implicit. A user-level route needs an
        // authenticated host invocation credential and is not supported yet.
        if route.Auth != "public" && route.Auth != "operator" {
            return false
        }
        key := httpRouteKey(route.Method, route.Path)
        if seen[key] {
            return false
        }
        seen[key] = true
    }
    return true
}
