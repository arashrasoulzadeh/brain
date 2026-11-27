package plugin

import (
	"net/http"
	"strings"

	"brain/internal/project"
)

// WebSocketEndpoint exposes the project-scoped transport boundary. The actual
// room worker and ticket validation are deliberately separate concerns.
func WebSocketEndpoint() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		namespace, ok := websocketPath(r.URL.Path)
		if !ok || r.Method != http.MethodGet {
			http.NotFound(w, r)
			return
		}
		hostProject, ok := project.FromContext(r.Context())
		if !ok || !hostProject.Active || (hostProject.ID != namespace && hostProject.Slug != namespace) {
			http.Error(w, "project boundary denied", http.StatusForbidden)
			return
		}
		if identity := r.Header.Get("X-Project-ID"); identity != "" && identity != hostProject.ID && identity != hostProject.Slug {
			http.Error(w, "project boundary denied", http.StatusForbidden)
			return
		}
		if !isWebSocketUpgrade(r) {
			http.Error(w, "websocket upgrade required", http.StatusUpgradeRequired)
			return
		}
		// The host must never accept a WebSocket connection before its project
		// authentication and worker-issued connection ticket are verified.
		http.Error(w, "connection ticket required", http.StatusUnauthorized)
	})
}

func websocketPath(path string) (string, bool) {
	parts := strings.Split(strings.TrimPrefix(path, "/"), "/")
	if len(parts) != 2 || parts[1] != "ws" || path != "/"+parts[0]+"/ws" || !isValidProjectID(parts[0]) {
		return "", false
	}
	return parts[0], true
}

func validWebSocketPathSegment(value string) bool {
	return value != "" && value != "." && value != ".." && !strings.ContainsAny(value, "\\\x00")
}

func isWebSocketUpgrade(r *http.Request) bool {
	if !strings.EqualFold(r.Header.Get("Upgrade"), "websocket") {
		return false
	}
	for _, value := range r.Header.Values("Connection") {
		for _, token := range strings.Split(value, ",") {
			if strings.EqualFold(strings.TrimSpace(token), "upgrade") {
				return true
			}
		}
	}
	return false
}
