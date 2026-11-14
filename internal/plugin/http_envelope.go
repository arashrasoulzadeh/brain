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
	if e.Protocol != protocol || e.RequestID == "" || !isValidProjectID(e.ProjectID) || e.PluginID == "" || e.Version == "" || e.Operation == "" || len(e.Payload) > 1<<20 {
		return ErrInvalidHTTPEnvelope
	}
	return nil
}

func (e HTTPRequestEnvelope) ValidateForProject(protocol int, projectID string) error {
	if err := e.Validate(protocol); err != nil {
		return err
	}
	if !isValidProjectID(projectID) || e.ProjectID != projectID {
		return ErrInvalidHTTPEnvelope
	}
	return nil
}
