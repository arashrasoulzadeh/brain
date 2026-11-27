package security_test

import (
	"crypto/rand"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/iotest"

	"brain/internal/security"
)

func TestCSRFEntropyFailureDoesNotReachHandler(t *testing.T) {
	original := rand.Reader
	defer func() { rand.Reader = original }()
	rand.Reader = iotest.ErrReader(errors.New("entropy unavailable"))
	called := false
	handler := security.CSRF(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if called || rec.Code != http.StatusInternalServerError || len(rec.Result().Cookies()) != 0 {
		t.Fatalf("unsafe entropy failure: called=%v status=%d", called, rec.Code)
	}
}
