package plugin

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os/exec"
	"strconv"
	"sync"
	"time"
)

var ErrTypeScriptWorkerStarted = errors.New("typescript worker already started")
var ErrTypeScriptWorkerStopped = errors.New("typescript worker stopped")
var ErrTypeScriptWorkerTimedOut = errors.New("typescript worker invocation timed out")

type TypeScriptWorkerOptions struct {
	InvocationTimeout time.Duration
}

type TypeScriptWorker struct {
	command *exec.Cmd
	env     []string
	stdin   io.WriteCloser
	reader  *bufio.Reader
	options TypeScriptWorkerOptions
	mutex   sync.Mutex
	ready   bool
}

func NewTypeScriptWorker(command string, arguments ...string) *TypeScriptWorker {
	return NewTypeScriptWorkerWithOptions(TypeScriptWorkerOptions{}, command, arguments...)
}

func NewTypeScriptWorkerWithOptions(options TypeScriptWorkerOptions, command string, arguments ...string) *TypeScriptWorker {
	return &TypeScriptWorker{command: exec.Command(command, arguments...), options: options}
}

func NewBubblewrapTypeScriptWorker(launcher, workerScript, descriptor, pluginRoot string, policy IsolationPolicy) *TypeScriptWorker {
	return &TypeScriptWorker{
		command: exec.Command(launcher),
		env: []string{
			"PATH=/usr/local/bin:/usr/bin:/bin",
			"WORKER_SCRIPT=" + workerScript,
			"EXECUTION_DESCRIPTOR=" + descriptor,
			"PLUGIN_ROOT=" + pluginRoot,
			"CPU_MILLIS=" + strconv.Itoa(policy.CPUMillis),
			"MEMORY_BYTES=" + strconv.FormatInt(policy.MemoryBytes, 10),
			"PROCESS_LIMIT=" + strconv.Itoa(policy.ProcessLimit),
		},
		options: TypeScriptWorkerOptions{InvocationTimeout: policy.WallTime},
	}
}

func (w *TypeScriptWorker) Start() error {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	if w.ready {
		return ErrTypeScriptWorkerStarted
	}
	if w.env != nil {
		w.command.Env = w.env
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
	responseLine := make(chan []byte, 1)
	readError := make(chan error, 1)
	go func() {
		line, err := w.reader.ReadBytes('\n')
		if err != nil {
			readError <- err
			return
		}
		responseLine <- line
	}()
	var line []byte
	var readErr error
	if w.options.InvocationTimeout > 0 {
		ctx, cancel := context.WithTimeout(context.Background(), w.options.InvocationTimeout)
		defer cancel()
		select {
		case line = <-responseLine:
		case readErr = <-readError:
		case <-ctx.Done():
			w.ready = false
			_ = w.command.Process.Kill()
			return WorkerResponse{Protocol: request.Protocol, ErrorCode: ErrTypeScriptWorkerTimedOut.Error()}
		}
	} else {
		select {
		case line = <-responseLine:
		case readErr = <-readError:
		}
	}
	if readErr != nil {
		w.ready = false
		return WorkerResponse{Protocol: request.Protocol, ErrorCode: readErr.Error()}
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
