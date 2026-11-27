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

func TestWebSocketEndpointUpgradeHeaderTokens(t *testing.T) {
	h := plugin.WebSocketEndpoint()
	cases := []struct {
		name       string
		connection []string
		upgrade    string
		identity   string
		want       int
	}{
		{"exact", []string{"Upgrade"}, "websocket", "dev", http.StatusUnauthorized},
		{"list", []string{"keep-alive, uPgRaDe"}, "WebSocket", "dev", http.StatusUnauthorized},
		{"multiple headers", []string{"keep-alive", "Upgrade"}, "websocket", "dev", http.StatusUnauthorized},
		{"substring", []string{"not-upgrade"}, "websocket", "dev", http.StatusUpgradeRequired},
		{"missing", nil, "websocket", "dev", http.StatusUpgradeRequired},
		{"wrong protocol", []string{"Upgrade"}, "h2c", "dev", http.StatusUpgradeRequired},
		{"conflicting project", []string{"Upgrade"}, "websocket", "other", http.StatusForbidden},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/dev/ws", nil)
			req = req.WithContext(project.WithProject(req.Context(), project.Project{ID: "dev", Active: true}))
			for _, value := range tc.connection {
				req.Header.Add("Connection", value)
			}
			req.Header.Set("Upgrade", tc.upgrade)
			req.Header.Set("X-Project-ID", tc.identity)
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)
			if rec.Code != tc.want {
				t.Fatalf("status = %d, want %d", rec.Code, tc.want)
			}
		})
	}
}
