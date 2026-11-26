package app

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"

	"brain/internal/operator"
)

func operatorBearerToken(r *http.Request) (string, bool) {
	token, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
	return token, ok && token != ""
}

func operatorActorFromRequest(sessions *operator.Store, r *http.Request) (operator.Actor, bool) {
	token, ok := operatorBearerToken(r)
	if !ok || sessions == nil {
		return operator.Actor{}, false
	}
	return sessions.ValidateSession(token)
}

type operatorLogin struct {
	Username string `json:"username"`
	Secret   string `json:"secret"`
}

func OperatorSessionHandler(sessions *operator.Store, digest func(string) string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if sessions == nil || digest == nil || r.Method != http.MethodPost {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		var login operatorLogin
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 16<<10)).Decode(&login); err != nil {
			http.Error(w, "invalid operator login", http.StatusBadRequest)
			return
		}
		tokenBytes := make([]byte, 32)
		if _, err := rand.Read(tokenBytes); err != nil {
			http.Error(w, "operator session unavailable", http.StatusInternalServerError)
			return
		}
		token, err := sessions.AuthenticateSession(login.Username, login.Secret, digest, func() string { return hex.EncodeToString(tokenBytes) })
		if err != nil {
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"session": token})
	})
}

func OperatorLogoutHandler(sessions *operator.Store) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if sessions == nil || r.Method != http.MethodDelete {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		token, ok := operatorBearerToken(r)
		if !ok || sessions.RevokeSession(token) != nil {
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
}
