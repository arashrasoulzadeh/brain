package plugin

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"brain/internal/execution"
)

var ErrTypeScriptWorkerStarted = errors.New("typescript worker already started")
var ErrTypeScriptWorkerStopped = errors.New("typescript worker stopped")
var ErrTypeScriptWorkerTimedOut = errors.New("typescript worker invocation timed out")
var ErrTypeScriptWorkerManifest = errors.New("invalid typescript worker manifest")

type executionDescriptor struct {
	Entrypoint     string           `json:"entrypoint"`
	Declarative    bool             `json:"declarative,omitempty"`
	PluginID       string           `json:"pluginId"`
	Version        string           `json:"version"`
	ProjectID      string           `json:"projectId"`
	RequireProject bool             `json:"requireProject"`
	Routes         []executionRoute `json:"routes"`
	Events         []EventSpec      `json:"events,omitempty"`
}

type executionRoute struct {
	Operation string        `json:"operation"`
	Function  string        `json:"function"`
	Helper    string        `json:"helper,omitempty"`
	Response  *ResponseSpec `json:"response,omitempty"`
	Pipeline  *PipelineSpec `json:"pipeline,omitempty"`
}

type TypeScriptWorkerOptions struct {
	InvocationTimeout time.Duration
	ProjectID         string
}

type TypeScriptWorker struct {
	command          *exec.Cmd
	env              []string
	stdin            io.WriteCloser
	reader           *bufio.Reader
	options          TypeScriptWorkerOptions
	mutex            sync.Mutex
	ready            bool
	startupHandshake bool
	spawnNanos, readinessNanos int64
	done             chan struct{}
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

func NewManifestBubblewrapTypeScriptWorker(launcher, workerScript, descriptor, pluginRoot, projectID string, manifest Manifest, policy IsolationPolicy) (*TypeScriptWorker, error) {
	for _, source := range manifest.SourceFiles {
		if strings.ContainsAny(source, "\r\n") {
			return nil, ErrInvalidManifest
		}
	}
	worker, err := NewManifestTypeScriptWorker("", workerScript, descriptor, pluginRoot, projectID, manifest, policy)
	if err != nil {
		return nil, err
	}
	worker.command = exec.Command(launcher)
	worker.startupHandshake = true
	runtimeRoot := filepath.Dir(filepath.Dir(workerScript))
	worker.env = []string{
		"PATH=/usr/local/bin:/usr/bin:/bin",
		"WORKER_SCRIPT=" + workerScript,
		"EXECUTION_DESCRIPTOR=" + descriptor,
		"PLUGIN_ROOT=" + pluginRoot,
		"CPU_MILLIS=" + strconv.Itoa(policy.CPUMillis),
		"MEMORY_BYTES=" + strconv.FormatInt(policy.MemoryBytes, 10),
		"PROCESS_LIMIT=" + strconv.Itoa(policy.ProcessLimit),
		"WORKER_RUNTIME_ROOT=" + runtimeRoot,
		"WORKER_SOURCE_FILES=" + strings.Join(manifest.SourceFiles, "\n"),
		"REQUIRE_DECLARED_SOURCES=1",
		"DECLARATIVE_PLUGIN=" + strconv.FormatBool(manifest.Declarative),
		"BRAIN_WORKER_READY=1",
	}
	return worker, nil
}

func NewManifestTypeScriptWorker(command, workerScript, descriptor, pluginRoot, projectID string, manifest Manifest, policy IsolationPolicy) (*TypeScriptWorker, error) {
	if strings.TrimSpace(projectID) == "" {
		return nil, ErrTypeScriptWorkerManifest
	}
	if strings.TrimSpace(manifest.ID) == "" || strings.TrimSpace(manifest.Version) == "" {
		return nil, ErrTypeScriptWorkerManifest
	}
	if err := ValidatePluginSources(pluginRoot, manifest); err != nil {
		return nil, err
	}
	entrypoint := ""
	if !manifest.Declarative {
		entrypoint = filepath.Join(pluginRoot, manifest.EntryPoint())
	}
	routes := make([]executionRoute, 0, len(manifest.Routes))
	for _, route := range manifest.Routes {
		if route.Helper == "host.admin" {
			continue
		}
		routes = append(routes, executionRoute{Operation: route.Operation, Function: route.Function, Helper: route.Helper, Response: route.Response, Pipeline: route.Pipeline})
	}
	for _, observer := range manifest.Observers {
		routes = append(routes, executionRoute{Operation: observer.operation(), Function: observer.Function, Helper: "sdk.context"})
	}
	data, err := json.Marshal(executionDescriptor{Entrypoint: entrypoint, Declarative: manifest.Declarative, PluginID: manifest.ID, Version: manifest.Version, ProjectID: projectID, RequireProject: true, Routes: routes, Events: manifest.Events})
	if err != nil {
		return nil, err
	}
	if err := os.WriteFile(descriptor, data, 0o600); err != nil {
		return nil, err
	}
	worker := NewTypeScriptWorkerWithOptions(TypeScriptWorkerOptions{InvocationTimeout: policy.WallTime, ProjectID: projectID}, command, workerScript, descriptor)
	worker.options.ProjectID = projectID
	return worker, nil
}

// SetTrustedStartupDiagnostics is for host-generated probes only. Never attach
// production logs to stderr from an untrusted plugin.
func (w *TypeScriptWorker) SetTrustedStartupDiagnostics(writer io.Writer) {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	if !w.ready {
		w.command.Stderr = writer
	}
}

func (w *TypeScriptWorker) Start() error {
	return w.StartContext(context.Background())
}

// StartContext bounds startup by the host assignment deadline.
func (w *TypeScriptWorker) StartContext(ctx context.Context) error {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	if w.ready {
		return ErrTypeScriptWorkerStarted
	}
	if err := ctx.Err(); err != nil {
		return err
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
	spawnStarted := time.Now()
	if err := w.command.Start(); err != nil {
		_ = stdin.Close()
		return err
	}
	w.spawnNanos = time.Since(spawnStarted).Nanoseconds()
	readyStarted := time.Now()
	w.stdin = stdin
	w.reader = bufio.NewReader(stdout)
	if w.startupHandshake {
		ack := make(chan string, 1)
		go func() {
			line, err := w.reader.ReadSlice('\n')
			if err != nil {
				ack <- ""
				return
			}
			ack <- string(line)
		}()
		timer := time.NewTimer(15 * time.Second)
		defer timer.Stop()
		select {
		case line := <-ack:
			if line != "{\"ready\":true}\n" {
				_ = w.command.Process.Kill()
				w.reapProcess()
				return fmt.Errorf("sandbox worker did not acknowledge startup")
			}
		case <-ctx.Done():
			_ = w.command.Process.Kill()
			_ = w.stdin.Close()
			w.reapProcess()
			return ctx.Err()
		case <-timer.C:
			_ = w.command.Process.Kill()
			w.reapProcess()
			return fmt.Errorf("sandbox worker startup timed out")
		}
		w.done = make(chan struct{})
		go func() {
			_ = w.command.Wait()
			close(w.done)
		}()
	}
	w.readinessNanos = time.Since(readyStarted).Nanoseconds()
	w.ready = true
	return nil
}

// StartupTiming separates OS process creation from launcher/Node readiness.
func (w *TypeScriptWorker) StartupTiming() (int64, int64) {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	return w.spawnNanos, w.readinessNanos
}

func (w *TypeScriptWorker) Stop() error {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	if !w.ready {
		return ErrTypeScriptWorkerStopped
	}
	w.ready = false
	_ = w.stdin.Close()
	killErr := w.command.Process.Kill()
	w.reapProcess()
	if killErr != nil && !errors.Is(killErr, os.ErrProcessDone) {
		return killErr
	}
	return nil
}

func (w *TypeScriptWorker) reapProcess() {
	if w.done != nil {
		<-w.done
	} else if w.command.Process != nil {
		_ = w.command.Wait()
	}
}

func (w *TypeScriptWorker) Ready() bool {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	if w.done != nil {
		select {
		case <-w.done:
			return false
		default:
		}
	}
	return w.ready
}

func (w *TypeScriptWorker) Handle(request WorkerRequest) WorkerResponse {
	return w.HandleContext(context.Background(), request)
}

// HandleContext kills and reaps the isolated worker before returning a cancelled
// assignment. The existing invocation timeout remains an upper bound.
func (w *TypeScriptWorker) HandleContext(ctx context.Context, request WorkerRequest) WorkerResponse {
	return w.handleContext(ctx, request, nil)
}

// HandleContextWithHostCalls is reserved for the trusted operator draft
// runtime. Other plugin workers have no host-call authority on their stdio.
func (w *TypeScriptWorker) HandleContextWithHostCalls(ctx context.Context, request WorkerRequest, handler execution.HostCallFunc) WorkerResponse {
	if request.PluginID != "functions" || request.Operation != "functions.invoke" || handler == nil {
		return WorkerResponse{Protocol: request.Protocol, ErrorCode: ErrInvalidWorkerRequest.Error()}
	}
	return w.handleContext(ctx, request, handler)
}

func (w *TypeScriptWorker) handleContext(ctx context.Context, request WorkerRequest, handler execution.HostCallFunc) WorkerResponse {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	if w.options.ProjectID != "" && request.ProjectID != w.options.ProjectID {
		return WorkerResponse{Protocol: request.Protocol, ErrorCode: ErrInvalidWorkerRequest.Error()}
	}
	if !w.ready {
		return WorkerResponse{Protocol: request.Protocol, ErrorCode: ErrTypeScriptWorkerStopped.Error()}
	}
	// The TypeScript SDK expects JSON values, not Go's base64 encoding of []byte.
	message := struct {
		JSONRPC string `json:"jsonrpc"`
		ID      string `json:"id"`
		Method  string `json:"method"`
		Params  struct {
			WorkerRequest
			Payload     json.RawMessage    `json:"payload"`
			HostContext *InvocationContext `json:"hostContext,omitempty"`
		} `json:"params"`
	}{JSONRPC: "2.0", ID: request.Operation, Method: "invoke"}
	message.Params.WorkerRequest = request
	message.Params.Payload = json.RawMessage(request.Payload)
	message.Params.HostContext = request.HostContext
	data, err := json.Marshal(message)
	if err != nil {
		return WorkerResponse{Protocol: request.Protocol, ErrorCode: err.Error()}
	}
	const maxMessage = 1 << 20
	if len(data) >= maxMessage {
		return WorkerResponse{Protocol: request.Protocol, ErrorCode: ErrInvalidRPCMessage.Error()}
	}
	type exchangeResult struct {
		line []byte
		err  error
	}
	result := make(chan exchangeResult, 1)
	if w.options.InvocationTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, w.options.InvocationTimeout)
		defer cancel()
	}
	go func() {
		if _, err := w.stdin.Write(append(data, '\n')); err != nil {
			result <- exchangeResult{err: err}
			return
		}
		for callID := uint64(1); ; {
			line, err := readWorkerLine(w.reader, maxMessage)
			if err != nil {
				result <- exchangeResult{err: err}
				return
			}
			if handler == nil {
				result <- exchangeResult{line: line}
				return
			}
			var message struct {
				HostCall *workerHostCall `json:"hostCall"`
			}
			if err := json.Unmarshal(line, &message); err != nil {
				result <- exchangeResult{err: ErrInvalidRPCMessage}
				return
			}
			if message.HostCall == nil {
				result <- exchangeResult{line: line}
				return
			}
			call := message.HostCall
			if call.ID != callID || callID > execution.MaxHostCalls || !execution.ValidHostCall(call.Operation, call.Input) {
				result <- exchangeResult{err: ErrInvalidRPCMessage}
				return
			}
			reply := workerHostReply{ID: callID}
			output, err := handler(ctx, call.Operation, call.Input)
			if ctx.Err() != nil {
				result <- exchangeResult{err: ctx.Err()}
				return
			}
			if err != nil || len(output) > execution.MaxHostCallBytes || !json.Valid(output) {
				reply.Error = "Plugin call denied or failed; check the operation, saved function permission, plugin installation and resource policy"
			} else {
				reply.Output = output
			}
			encoded, err := json.Marshal(struct {
				HostResponse workerHostReply `json:"hostResponse"`
			}{reply})
			if err != nil || len(encoded) >= maxMessage {
				result <- exchangeResult{err: ErrInvalidRPCMessage}
				return
			}
			if _, err := w.stdin.Write(append(encoded, '\n')); err != nil {
				result <- exchangeResult{err: err}
				return
			}
			callID++
		}
	}()
	var exchange exchangeResult
	select {
	case exchange = <-result:
	case <-ctx.Done():
		w.ready = false
		_ = w.command.Process.Kill()
		_ = w.stdin.Close()
		w.reapProcess()
		return WorkerResponse{Protocol: request.Protocol, ErrorCode: ErrTypeScriptWorkerTimedOut.Error()}
	}
	if exchange.err != nil {
		w.ready = false
		_ = w.command.Process.Kill()
		_ = w.stdin.Close()
		w.reapProcess()
		return WorkerResponse{Protocol: request.Protocol, ErrorCode: exchange.err.Error()}
	}
	line := exchange.line
	var response struct {
		Result *struct {
			Protocol     int                 `json:"protocol"`
			Success      bool                `json:"success"`
			Payload      json.RawMessage     `json:"payload"`
			ErrorCode    string              `json:"errorCode"`
			Observations []ObservationIntent `json:"observations"`
		} `json:"result"`
		ErrorCode string `json:"errorCode"`
	}
	if err := json.Unmarshal(line, &response); err != nil {
		return WorkerResponse{Protocol: request.Protocol, ErrorCode: ErrInvalidRPCMessage.Error()}
	}
	if response.Result == nil {
		return WorkerResponse{Protocol: request.Protocol, ErrorCode: response.ErrorCode}
	}
	if response.Result.Protocol != request.Protocol {
		return WorkerResponse{Protocol: request.Protocol, ErrorCode: ErrInvalidRPCMessage.Error()}
	}
	return WorkerResponse{
		Protocol: response.Result.Protocol, Success: response.Result.Success,
		Payload: response.Result.Payload, ErrorCode: response.Result.ErrorCode, Observations: response.Result.Observations,
	}
}
