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
	req := httptest.NewRequest(http.MethodGet, "/dev/ws", nil)
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
	req := httptest.NewRequest(http.MethodGet, "/dev/ws", nil)
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
	req := httptest.NewRequest(http.MethodGet, "/dev/ws", nil)
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
	req := httptest.NewRequest(http.MethodGet, "/%2e%2e/ws", nil)
	req.Header.Set("Connection", "Upgrade")
	req.Header.Set("Upgrade", "websocket")
	req.Header.Set("X-Project-ID", "..")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestWebSocketEndpointRejectsAmbiguousProjectPath(t *testing.T) {
	h := plugin.WebSocketEndpoint()
	req := httptest.NewRequest(http.MethodGet, "/dev:other/ws", nil)
	req.Header.Set("X-Project-ID", "dev:other")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestWebSocketEndpointUsesResolvedSlugInsteadOfClientHeader(t *testing.T) {
	h := plugin.WebSocketEndpoint()
	req := httptest.NewRequest(http.MethodGet, "/alpha/ws", nil)
	req = req.WithContext(project.WithProject(req.Context(), project.Project{ID: "internal-alpha", Slug: "alpha", Active: true}))
	req.Header.Set("Connection", "Upgrade")
	req.Header.Set("Upgrade", "websocket")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want ticket-required 401", rec.Code)
	}
}
