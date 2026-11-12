package plugin

import "errors"

var ErrInvalidRPCMessage = errors.New("invalid plugin rpc message")

type RPCRequest struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      string        `json:"id"`
	Method  string        `json:"method"`
	Params  WorkerRequest `json:"params"`
}

type RPCResponse struct {
	JSONRPC   string          `json:"jsonrpc"`
	ID        string          `json:"id"`
	Result    *WorkerResponse `json:"result,omitempty"`
	ErrorCode string          `json:"errorCode,omitempty"`
}

func ValidateRPCRequest(message RPCRequest, protocol int) error {
	if message.JSONRPC != "2.0" || message.ID == "" || message.Method == "" {
		return ErrInvalidRPCMessage
	}
	return ValidateWorkerRequest(message.Params, protocol)
}
