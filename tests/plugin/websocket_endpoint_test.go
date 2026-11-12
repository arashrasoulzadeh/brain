package plugin_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"brain/internal/plugin"
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
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusUpgradeRequired {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUpgradeRequired)
	}
}
