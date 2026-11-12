package plugin

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strconv"
	"strings"
)

var ErrInvalidHandshake = errors.New("invalid external worker handshake")

type Handshake struct {
	Protocol  int
	ProjectID string
	PluginID  string
	Version   string
	ServiceID string
	Signature string
}

func AuthenticateHandshake(h Handshake, secret string) bool {
	if secret == "" || h.Signature == "" {
		return false
	}
	message := strings.Join([]string{strconv.Itoa(h.Protocol), h.ProjectID, h.PluginID, h.Version, h.ServiceID}, "|")
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(message))
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(h.Signature))
}

func ValidateHandshake(h Handshake, hostProtocol int, expectedService string) error {
	if h.Protocol != hostProtocol || h.ProjectID == "" || h.PluginID == "" || h.Version == "" || h.ServiceID == "" || h.ServiceID != expectedService {
		return ErrInvalidHandshake
	}
	return nil
}
