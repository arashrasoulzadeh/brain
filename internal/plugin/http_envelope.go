package plugin

import "errors"

var ErrInvalidHTTPEnvelope = errors.New("invalid HTTP request envelope")

type HTTPRequestEnvelope struct {
	Protocol  int    `json:"protocol"`
	RequestID string `json:"request_id"`
	ProjectID string `json:"project_id"`
	PluginID  string `json:"plugin_id"`
	Version   string `json:"version"`
	Operation string `json:"operation"`
	Payload   []byte `json:"payload"`
}

func (e HTTPRequestEnvelope) Validate(protocol int) error {
	if e.Protocol != protocol || e.RequestID == "" || e.ProjectID == "" || e.PluginID == "" || e.Version == "" || e.Operation == "" {
		return ErrInvalidHTTPEnvelope
	}
	return nil
}
