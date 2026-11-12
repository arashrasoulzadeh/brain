package plugin

import "errors"

var ErrInvalidWorkerRequest = errors.New("invalid worker request")

type WorkerRequest struct {
	Protocol  int
	ProjectID string
	PluginID  string
	Version   string
	Operation string
	Payload   []byte
}

type WorkerResponse struct {
	Protocol  int
	Success   bool
	Payload   []byte
	ErrorCode string
}

func ValidateWorkerRequest(request WorkerRequest, hostProtocol int) error {
	if request.Protocol != hostProtocol || request.ProjectID == "" || request.PluginID == "" || request.Version == "" || request.Operation == "" || len(request.Payload) > 1<<20 {
		return ErrInvalidWorkerRequest
	}
	return nil
}
