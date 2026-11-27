package app_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"brain/internal/app"
)

func TestReadinessRequiresHealthyDependenciesWithoutLeakingErrors(t *testing.T) {
	server := app.NewServer()
	check := func(path string, want int) {
		t.Helper()
		recorder := httptest.NewRecorder()
		server.Handler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, path, nil))
		if recorder.Code != want {
			t.Fatalf("%s status = %d, want %d", path, recorder.Code, want)
		}
		if strings.Contains(recorder.Body.String(), "private-database-path") {
			t.Fatal("health response leaked dependency details")
		}
	}
	check("/_system/livez", http.StatusOK)
	check("/_system/readyz", http.StatusServiceUnavailable)
	server.SetReadinessCheck(func(context.Context) error { return errors.New("private-database-path") })
	check("/_system/readyz", http.StatusServiceUnavailable)
	check("/_system/livez", http.StatusOK)
	server.SetReadinessCheck(func(context.Context) error { return nil })
	check("/_system/readyz", http.StatusOK)
}
