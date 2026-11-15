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
	if strings.HasSuffix(message, "\n") {
		message = strings.TrimSpace(message)
	}
	encoded, _ := json.Marshal(APIMessage{Status: "error", Code: http.StatusText(status), Message: message})
	return encoded
}

func withAPIEnvelope(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buffer := newAPIBuffer(w.Header())
		next.ServeHTTP(buffer, r)
		if buffer.status >= http.StatusBadRequest {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(buffer.status)
			_, _ = w.Write(normalizeAPIError(buffer.status, buffer.body.Bytes()))
			return
		}
		w.WriteHeader(buffer.status)
		_, _ = w.Write(buffer.body.Bytes())
	})
}
