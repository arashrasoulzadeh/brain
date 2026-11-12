package plugin

import (
	"net/http"
	"strings"
)

// WebSocketEndpoint exposes the project-scoped transport boundary. The actual
// room worker and ticket validation are deliberately separate concerns.
func WebSocketEndpoint() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		projectID, roomID, ok := websocketPath(r.URL.Path)
		if !ok || r.Method != http.MethodGet {
			http.NotFound(w, r)
			return
		}
		if projectID == "" || r.Header.Get("X-Project-ID") != projectID {
			http.Error(w, "project boundary denied", http.StatusForbidden)
			return
		}
		if !isWebSocketUpgrade(r) {
			http.Error(w, "websocket upgrade required", http.StatusUpgradeRequired)
			return
		}
		// Ticket authentication and the protocol upgrade are handled by the
		// connection-ticket and room-worker layers added in later steps.
		_ = roomID
		http.Error(w, "connection ticket required", http.StatusUnauthorized)
	})
}

func websocketPath(path string) (string, string, bool) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) != 6 || parts[0] != "v1" || parts[1] != "projects" || parts[3] != "rooms" || parts[5] != "ws" {
		return "", "", false
	}
	return parts[2], parts[4], parts[2] != "" && parts[4] != ""
}

func isWebSocketUpgrade(r *http.Request) bool {
	return strings.EqualFold(r.Header.Get("Upgrade"), "websocket") && strings.Contains(strings.ToLower(r.Header.Get("Connection")), "upgrade")
}
