package project

import (
	"errors"
	"net/http"
	"strings"
)

// URLMiddleware binds the immutable project ID from the first URL segment.
// Headers and query parameters can only agree with the host-resolved project.
func (r Resolver) URLMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		escaped := req.URL.EscapedPath()
		if escaped == "" || escaped == "/" || strings.Contains(escaped, "//") || strings.Contains(escaped, "\\") {
			http.NotFound(w, req)
			return
		}
		for _, segment := range strings.Split(strings.ToLower(escaped), "/") {
			if segment == "." || segment == ".." || strings.Contains(segment, "%2f") ||
				strings.Contains(segment, "%5c") || strings.Contains(segment, "%2e") ||
				strings.Contains(segment, "%25") {
				http.NotFound(w, req)
				return
			}
		}
		namespace, resource, hasResource := strings.Cut(strings.TrimPrefix(req.URL.Path, "/"), "/")
		if namespace == "_system" {
			next.ServeHTTP(w, req)
			return
		}
		if !hasResource || resource == "" || r.Store == nil {
			http.NotFound(w, req)
			return
		}
		model, err := r.Store.ResolveNamespace(namespace)
		if err != nil {
			status := http.StatusNotFound
			if errors.Is(err, ErrInactive) {
				status = http.StatusForbidden
			}
			http.Error(w, http.StatusText(status), status)
			return
		}
		for _, header := range []string{"X-Project-ID", "X-Project-Slug"} {
			values := req.Header.Values(header)
			if len(values) > 1 {
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}
			for _, identity := range values {
				if identity != "" && identity != model.ID && identity != model.Slug {
					http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
					return
				}
			}
		}
		for _, field := range []string{"project_id", "projectId", "project"} {
			values := req.URL.Query()[field]
			if len(values) > 1 || len(values) == 1 && values[0] != "" && values[0] != model.ID && values[0] != model.Slug {
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}
		}
		next.ServeHTTP(w, req.WithContext(WithProject(req.Context(), model)))
	})
}
