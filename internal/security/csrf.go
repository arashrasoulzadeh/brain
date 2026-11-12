package security

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"net/http"
)

const csrfCookie = "brain_csrf"
const csrfHeader = "X-CSRF-Token"

func CSRF(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := ""
		if cookie, err := r.Cookie(csrfCookie); err == nil {
			token = cookie.Value
		}
		if token == "" {
			var bytes [32]byte
			if _, err := rand.Read(bytes[:]); err != nil {
				http.Error(w, "csrf unavailable", http.StatusInternalServerError)
				return
			}
			token = base64.RawURLEncoding.EncodeToString(bytes[:])
			http.SetCookie(w, &http.Cookie{Name: csrfCookie, Value: token, Path: "/", Secure: true, HttpOnly: false, SameSite: http.SameSiteStrictMode})
		}
		if r.Method != http.MethodGet && r.Method != http.MethodHead && r.Method != http.MethodOptions {
			provided := r.Header.Get(csrfHeader)
			if subtle.ConstantTimeCompare([]byte(token), []byte(provided)) != 1 {
				http.Error(w, "csrf failed", http.StatusForbidden)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}
