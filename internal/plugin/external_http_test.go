package plugin_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"brain/internal/plugin"
)

func TestExternalHandshakeEndpointAcceptsValidHandshake(t *testing.T) {
	handler := plugin.HandshakeHandler(1, "arena-dev")
	body := `{"protocol":1,"projectID":"dev","pluginID":"arena","version":"1","serviceID":"arena-dev"}`
	req := httptest.NewRequest(http.MethodPost, "/handshake", strings.NewReader(body))
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusOK || !strings.Contains(res.Body.String(), `"accepted":true`) {
		t.Fatalf("response = %d %s", res.Code, res.Body.String())
	}
}

func TestExternalHandshakeEndpointRejectsInvalidHandshake(t *testing.T) {
	handler := plugin.HandshakeHandler(1, "arena-dev")
	req := httptest.NewRequest(http.MethodPost, "/handshake", strings.NewReader(`{"projectID":"other"}`))
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d", res.Code)
	}
}
