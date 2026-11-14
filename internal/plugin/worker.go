package plugin

import "errors"

var ErrInvalidWorkerRequest = errors.New("invalid worker request")
var ErrWorkerNotReady = errors.New("worker not ready")

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

type ContractWorker interface {
	Worker
	Handle(WorkerRequest) WorkerResponse
}

func Invoke(worker ContractWorker, request WorkerRequest, protocol int) (WorkerResponse, error) {
	if err := ValidateWorkerRequest(request, protocol); err != nil {
		return WorkerResponse{}, err
	}
	if !worker.Ready() {
		return WorkerResponse{}, ErrWorkerNotReady
	}
	return worker.Handle(request), nil
}

func InvokeForProject(worker ContractWorker, request WorkerRequest, protocol int, projectID string) (WorkerResponse, error) {
	if err := ValidateWorkerRequestForProject(request, protocol, projectID); err != nil {
		return WorkerResponse{}, err
	}
	if !worker.Ready() {
		return WorkerResponse{}, ErrWorkerNotReady
	}
	return worker.Handle(request), nil
}

func ValidateWorkerRequest(request WorkerRequest, hostProtocol int) error {
	if request.Protocol != hostProtocol || request.ProjectID == "" || request.PluginID == "" || request.Version == "" || request.Operation == "" || len(request.Payload) > 1<<20 {
		return ErrInvalidWorkerRequest
	}
	return nil
}

func ValidateWorkerRequestForProject(request WorkerRequest, hostProtocol int, projectID string) error {
	if err := ValidateWorkerRequest(request, hostProtocol); err != nil {
		return err
	}
	if projectID == "" || request.ProjectID != projectID {
		return ErrInvalidWorkerRequest
	}
	return nil
}
