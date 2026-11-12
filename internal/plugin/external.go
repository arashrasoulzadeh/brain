package plugin

import "errors"

var ErrInvalidHandshake = errors.New("invalid external worker handshake")

type Handshake struct {
	Protocol  int
	ProjectID string
	PluginID  string
	Version   string
	ServiceID string
}

func ValidateHandshake(h Handshake, hostProtocol int, expectedService string) error {
	if h.Protocol != hostProtocol || h.ProjectID == "" || h.PluginID == "" || h.Version == "" || h.ServiceID == "" || h.ServiceID != expectedService {
		return ErrInvalidHandshake
	}
	return nil
}
