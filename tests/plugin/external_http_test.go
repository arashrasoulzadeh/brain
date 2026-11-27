package plugin_test

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"brain/internal/plugin"
)

func TestExternalHandshakeEndpointAcceptsValidHandshake(t *testing.T) {
	handler := plugin.HandshakeHandlerForProject(1, "arena-dev", "dev")
	body := `{"protocol":1,"projectID":"dev","pluginID":"arena","version":"1","serviceID":"arena-dev"}`
	req := httptest.NewRequest(http.MethodPost, "/handshake", strings.NewReader(body))
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusOK || !strings.Contains(res.Body.String(), `"accepted":true`) {
		t.Fatalf("response = %d %s", res.Code, res.Body.String())
	}
}

func TestExternalHandshakeEndpointRejectsInvalidHandshake(t *testing.T) {
	handler := plugin.HandshakeHandlerForProject(1, "arena-dev", "dev")
	req := httptest.NewRequest(http.MethodPost, "/handshake", strings.NewReader(`{"projectID":"other"}`))
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d", res.Code)
	}
}

func TestExternalHandshakeEndpointRejectsTrailingJSON(t *testing.T) {
	handler := plugin.HandshakeHandlerForProject(1, "arena-dev", "dev")
	body := `{"protocol":1,"projectID":"dev","pluginID":"arena","version":"1","serviceID":"arena-dev"}{"extra":true}`
	req := httptest.NewRequest(http.MethodPost, "/handshake", strings.NewReader(body))
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusUnauthorized)
	}
}

func TestSecretProtectedExternalHandshake(t *testing.T) {
	handler := plugin.HandshakeHandlerForProjectWithSecret(1, "arena-dev", "dev", "fixture-secret")
	handshake := plugin.Handshake{Protocol: 1, ProjectID: "dev", PluginID: "arena", Version: "1", ServiceID: "arena-dev"}
	mac := hmac.New(sha256.New, []byte("fixture-secret"))
	_, _ = mac.Write([]byte("1|dev|arena|1|arena-dev"))
	signature := hex.EncodeToString(mac.Sum(nil))
	for _, tc := range []struct {
		signature string
		want      int
	}{
		{"", http.StatusUnauthorized},
		{"invalid", http.StatusUnauthorized},
		{signature, http.StatusOK},
	} {
		handshake.Signature = tc.signature
		body, err := json.Marshal(handshake)
		if err != nil {
			t.Fatal(err)
		}
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/handshake", strings.NewReader(string(body))))
		if rec.Code != tc.want {
			t.Fatalf("status = %d, want %d", rec.Code, tc.want)
		}
	}
}
