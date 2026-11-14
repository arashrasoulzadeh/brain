package plugin

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"os/exec"
	"sync"
)

var ErrTypeScriptWorkerStarted = errors.New("typescript worker already started")
var ErrTypeScriptWorkerStopped = errors.New("typescript worker stopped")

type TypeScriptWorker struct {
	command *exec.Cmd
	stdin   io.WriteCloser
	reader  *bufio.Reader
	mutex   sync.Mutex
	ready   bool
}

func NewTypeScriptWorker(command string, arguments ...string) *TypeScriptWorker {
	return &TypeScriptWorker{command: exec.Command(command, arguments...)}
}

func (w *TypeScriptWorker) Start() error {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	if w.ready {
		return ErrTypeScriptWorkerStarted
	}
	stdin, err := w.command.StdinPipe()
	if err != nil {
		return err
	}
	stdout, err := w.command.StdoutPipe()
	if err != nil {
		_ = stdin.Close()
		return err
	}
	if err := w.command.Start(); err != nil {
		_ = stdin.Close()
		return err
	}
	w.stdin = stdin
	w.reader = bufio.NewReader(stdout)
	w.ready = true
	return nil
}

func (w *TypeScriptWorker) Stop() error {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	if !w.ready {
		return ErrTypeScriptWorkerStopped
	}
	w.ready = false
	_ = w.stdin.Close()
	return w.command.Process.Kill()
}

func (w *TypeScriptWorker) Ready() bool {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	return w.ready
}

func (w *TypeScriptWorker) Handle(request WorkerRequest) WorkerResponse {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	if !w.ready {
		return WorkerResponse{Protocol: request.Protocol, ErrorCode: ErrTypeScriptWorkerStopped.Error()}
	}
	message := RPCRequest{JSONRPC: "2.0", ID: request.Operation, Method: "invoke", Params: request}
	data, err := json.Marshal(message)
	if err != nil {
		return WorkerResponse{Protocol: request.Protocol, ErrorCode: err.Error()}
	}
	if _, err := w.stdin.Write(append(data, '\n')); err != nil {
		w.ready = false
		return WorkerResponse{Protocol: request.Protocol, ErrorCode: err.Error()}
	}
	line, err := w.reader.ReadBytes('\n')
	if err != nil {
		w.ready = false
		return WorkerResponse{Protocol: request.Protocol, ErrorCode: err.Error()}
	}
	var response RPCResponse
	if err := json.Unmarshal(line, &response); err != nil {
		return WorkerResponse{Protocol: request.Protocol, ErrorCode: ErrInvalidRPCMessage.Error()}
	}
	if response.Result == nil {
		return WorkerResponse{Protocol: request.Protocol, ErrorCode: response.ErrorCode}
	}
	return *response.Result
}
