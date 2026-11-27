package app

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
)

type APIMessage struct {
	Status  string `json:"status"`
	Code    string `json:"code,omitempty"`
	Message string `json:"message"`
}

type apiBuffer struct {
	header http.Header
	body   bytes.Buffer
	status int
}

func newAPIBuffer(header http.Header) *apiBuffer {
	return &apiBuffer{header: header, status: http.StatusOK}
}

func (b *apiBuffer) Header() http.Header { return b.header }

func (b *apiBuffer) WriteHeader(status int) { b.status = status }

func (b *apiBuffer) Write(data []byte) (int, error) { return b.body.Write(data) }

func normalizeAPIError(status int, body []byte) []byte {
	message := strings.TrimSpace(string(body))
	if message == "" {
		message = http.StatusText(status)
	}
	encoded, _ := json.Marshal(APIMessage{Status: "error", Code: http.StatusText(status), Message: message})
	return encoded
}

func normalizeAPISuccess(path string, status int, body []byte) []byte {
	if !strings.HasPrefix(path, "/_system/admin/") && !strings.Contains(path, "/admin/") && !strings.Contains(path, "/functions/") {
		return body
	}
	if len(bytes.TrimSpace(body)) > 0 {
		return body
	}
	message := "Request completed."
	if status == http.StatusNoContent {
		message = "Request completed without a response body."
	}
	encoded, _ := json.Marshal(APIMessage{Status: "success", Message: message})
	return encoded
}

func withAPIEnvelope(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only the reserved project socket upgrade may bypass the JSON buffer.
		// The WebSocket library needs the original hijack-capable writer.
		if r.Method == http.MethodGet && strings.Count(r.URL.Path, "/") == 2 && strings.HasSuffix(r.URL.Path, "/ws") && strings.EqualFold(r.Header.Get("Upgrade"), "websocket") {
			next.ServeHTTP(w, r)
			return
		}
		// Only this reserved authenticated binary download uses the streaming writer.
		if r.Method == http.MethodGet && strings.Count(r.URL.Path, "/") == 4 && strings.HasSuffix(r.URL.Path, "/admin/files/blob") {
			next.ServeHTTP(w, r)
			return
		}
		buffer := newAPIBuffer(w.Header())
		next.ServeHTTP(buffer, r)
		if buffer.status >= http.StatusBadRequest {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(buffer.status)
			_, _ = w.Write(normalizeAPIError(buffer.status, buffer.body.Bytes()))
			return
		}
		body := normalizeAPISuccess(r.URL.Path, buffer.status, buffer.body.Bytes())
		if buffer.status == http.StatusNoContent && len(body) > 0 {
			buffer.status = http.StatusOK
		}
		if len(body) > 0 {
			w.Header().Set("Content-Type", "application/json")
		}
		w.WriteHeader(buffer.status)
		_, _ = w.Write(body)
	})
}
