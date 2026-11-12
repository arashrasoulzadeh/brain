package app_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"brain/internal/app"
)

func TestServerLiveness(t *testing.T) {
	server := app.NewServer()
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/livez", nil)

	server.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("GET /livez status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if body := recorder.Body.String(); body != "ok\n" {
		t.Fatalf("GET /livez body = %q, want %q", body, "ok\n")
	}
}
