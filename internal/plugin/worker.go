package plugin

import (
	"errors"
	"time"
)

var ErrInvalidWorkerRequest = errors.New("invalid worker request")
var ErrWorkerNotReady = errors.New("worker not ready")
var ErrInvocationCredentialMismatch = errors.New("invocation credential does not match worker request")

type WorkerRequest struct {
	Protocol    int                `json:"protocol"`
	ProjectID   string             `json:"projectId"`
	PluginID    string             `json:"pluginId"`
	Version     string             `json:"version"`
	Operation   string             `json:"operation"`
	Payload     []byte             `json:"payload"`
	HostContext *InvocationContext `json:"-"`

	// Set only by trusted host entry points; never serialized to a worker.
	accessPrincipal string
}

type WorkerResponse struct {
	Protocol     int
	Success      bool
	Payload      []byte
	ErrorCode    string
	Observations []ObservationIntent `json:"observations,omitempty"`
}

type ContractWorker interface {
	Worker
	Handle(WorkerRequest) WorkerResponse
}

func InvokeForProject(worker ContractWorker, request WorkerRequest, protocol int, projectID string) (WorkerResponse, error) {
	if err := ValidateWorkerRequestForProject(request, protocol, projectID); err != nil {
		return WorkerResponse{}, err
	}
	if worker == nil {
		return WorkerResponse{}, ErrWorkerNotReady
	}
	if !worker.Ready() {
		return WorkerResponse{}, ErrWorkerNotReady
	}
	return worker.Handle(request), nil
}

// InvokeWithCredential is the production invocation boundary. The host-issued
// credential is consumed before the worker can observe the request.
func InvokeWithCredential(broker *Broker, credential InvocationCredential, worker ContractWorker, request WorkerRequest, protocol int, now time.Time) (WorkerResponse, error) {
	if broker == nil || credential.Call.ProjectID != request.ProjectID || credential.Call.PluginID != request.PluginID || credential.Call.Version != request.Version || credential.Call.Operation != request.Operation {
		return WorkerResponse{}, ErrInvalidWorkerRequest
	}
	if err := ValidateWorkerRequest(request, protocol); err != nil {
		return WorkerResponse{}, err
	}
	if err := broker.CommitCredential(credential, now); err != nil {
		return WorkerResponse{}, err
	}
	return InvokeForProject(worker, request, protocol, credential.Call.ProjectID)
}

func ValidateWorkerRequest(request WorkerRequest, hostProtocol int) error {
	if request.Protocol != hostProtocol || !isValidProjectID(request.ProjectID) || request.PluginID == "" || request.Version == "" || request.Operation == "" || len(request.Payload) > 1<<20 {
		return ErrInvalidWorkerRequest
	}
	return nil
}

func ValidateWorkerRequestForProject(request WorkerRequest, hostProtocol int, projectID string) error {
	if err := ValidateWorkerRequest(request, hostProtocol); err != nil {
		return err
	}
	if !isValidProjectID(projectID) || request.ProjectID != projectID {
		return ErrInvalidWorkerRequest
	}
	return nil
}
