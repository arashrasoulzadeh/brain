package plugin

import (
	"encoding/json"
	"net/http"
)

func HandshakeHandler(hostProtocol int, expectedService string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		defer r.Body.Close()
		var handshake Handshake
		decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096))
		if err := decoder.Decode(&handshake); err != nil || ValidateHandshake(handshake, hostProtocol, expectedService) != nil {
			http.Error(w, "invalid handshake", http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"accepted": true, "protocol": hostProtocol})
	})
}
