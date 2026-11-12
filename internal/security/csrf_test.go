package security_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"brain/internal/security"
)

func TestCSRFAllowsSafeRequestAndIssuesCookie(t *testing.T) {
	handler := security.CSRF(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNoContent) }))
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/", nil))
	if res.Code != http.StatusNoContent || res.Result().Cookies()[0].Name != "brain_csrf" {
		t.Fatalf("response = %d, cookies = %#v", res.Code, res.Result().Cookies())
	}
}

func TestCSRFRequiresMatchingHeaderForMutation(t *testing.T) {
	handler := security.CSRF(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNoContent) }))
	first := httptest.NewRecorder()
	handler.ServeHTTP(first, httptest.NewRequest(http.MethodGet, "/", nil))
	cookie := first.Result().Cookies()[0]
	for _, provided := range []string{"", "wrong", cookie.Value} {
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		req.AddCookie(cookie)
		req.Header.Set("X-CSRF-Token", provided)
		res := httptest.NewRecorder()
		handler.ServeHTTP(res, req)
		want := http.StatusNoContent
		if provided != cookie.Value {
			want = http.StatusForbidden
		}
		if res.Code != want {
			t.Errorf("token %q status = %d, want %d", provided, res.Code, want)
		}
	}
}
