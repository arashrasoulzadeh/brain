package app_test

import (
    "net/http"
    "net/http/httptest"
    "testing"

    "brain/internal/app"
)

func TestSystemNamespaceHasNoUnprefixedHealthAliases(t *testing.T) {
    handler := app.NewServer().Handler()
    for _, test := range []struct{ path string; status int }{
        {"/_system/livez", http.StatusOK},
        {"/livez", http.StatusNotFound},
        {"/readyz", http.StatusNotFound},
        {"/admin/session", http.StatusNotFound},
        {"/public/alpha/hello", http.StatusNotFound},
    } {
        response := httptest.NewRecorder()
        handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, test.path, nil))
        if response.Code != test.status {
            t.Errorf("%s: got %d, want %d", test.path, response.Code, test.status)
        }
    }
}
