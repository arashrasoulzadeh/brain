package plugin_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"brain/internal/plugin"
	"brain/internal/project"
)

func TestWebSocketEndpointRequiresProjectScopedUpgrade(t *testing.T) {
	h := plugin.WebSocketEndpoint()
	req := httptest.NewRequest(http.MethodGet, "/v1/projects/dev/rooms/room-1/ws", nil)
	req.Header.Set("Connection", "Upgrade")
	req.Header.Set("Upgrade", "websocket")
	req.Header.Set("X-Project-ID", "other")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}

func TestWebSocketEndpointRejectsNonUpgrade(t *testing.T) {
	h := plugin.WebSocketEndpoint()
	req := httptest.NewRequest(http.MethodGet, "/v1/projects/dev/rooms/room-1/ws", nil)
	req.Header.Set("X-Project-ID", "dev")
	req = req.WithContext(project.WithProject(req.Context(), project.Project{ID: "dev", Active: true}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusUpgradeRequired {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUpgradeRequired)
	}
}

func TestWebSocketEndpointRejectsMissingHostProjectContext(t *testing.T) {
	h := plugin.WebSocketEndpoint()
	req := httptest.NewRequest(http.MethodGet, "/v1/projects/dev/rooms/room-1/ws", nil)
	req.Header.Set("Connection", "Upgrade")
	req.Header.Set("Upgrade", "websocket")
	req.Header.Set("X-Project-ID", "dev")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}

func TestWebSocketEndpointRejectsTraversalProjectPath(t *testing.T) {
	h := plugin.WebSocketEndpoint()
	req := httptest.NewRequest(http.MethodGet, "/v1/projects/%2e%2e/rooms/room-1/ws", nil)
	req.Header.Set("Connection", "Upgrade")
	req.Header.Set("Upgrade", "websocket")
	req.Header.Set("X-Project-ID", "..")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}
