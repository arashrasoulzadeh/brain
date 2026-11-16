package admin

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"brain/internal/operator"
)

var ErrFunctionDraftInvalid = errors.New("function draft requires a name and source")
var ErrFunctionSourceInvalid = errors.New("function source must export a handler and may not use host or network access")
var ErrFunctionMethodNotAllowed = errors.New("function method not allowed")

var functionExportPattern = regexp.MustCompile(`(?m)\bexport\s+default\b`)
var functionUnsafePattern = regexp.MustCompile(`(?m)\b(import|require|process|child_process|fs|net|fetch|Deno)\b`)
var functionNamePattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_-]{0,63}$`)
var functionMethodPattern = regexp.MustCompile(`(?s)export\s+const\s+config\s*=\s*\{.*?\bmethod\s*:\s*["'](GET|POST|PUT|PATCH)["']`)

const maxFunctionSourceBytes = 256 * 1024

type FunctionDraft struct {
	ProjectID   string     `json:"project_id"`
	Name        string     `json:"name"`
	Source      string     `json:"source"`
	Version     int        `json:"version"`
	Public      bool       `json:"public"`
	Permissions []string   `json:"permissions"`
	Method      string     `json:"method"`
	PublishedAt *time.Time `json:"published_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}
type FunctionLog struct {
	ProjectID  string            `json:"project_id"`
	Name       string            `json:"name"`
	Status     string            `json:"status"`
	DurationMs int64             `json:"duration_ms"`
	Message    string            `json:"message,omitempty"`
	CreatedAt  time.Time         `json:"created_at"`
	Steps      []FunctionLogStep `json:"steps,omitempty"`
	Error      string            `json:"error,omitempty"`
	RequestID  string            `json:"request_id"`
}
type FunctionLogStep struct {
	Name       string `json:"name"`
	Status     string `json:"status"`
	DurationMs int64  `json:"duration_ms"`
	Reason     string `json:"reason,omitempty"`
	Error      string `json:"error,omitempty"`
}

var functionDrafts = struct {
	sync.RWMutex
	items map[string]FunctionDraft
}{items: make(map[string]FunctionDraft)}

const functionDraftsPath = ".local/function-drafts.json"
const functionLogsPath = ".local/function-logs.json"

var functionLogs struct {
	sync.RWMutex
	items []FunctionLog
}

func appendFunctionLog(logEntry FunctionLog) {
	functionLogs.Lock()
	defer functionLogs.Unlock()
	data, _ := os.ReadFile(functionLogsPath)
	_ = json.Unmarshal(data, &functionLogs.items)
	functionLogs.items = append(functionLogs.items, logEntry)
	data, _ = json.MarshalIndent(functionLogs.items, "", "  ")
	_ = os.MkdirAll(filepath.Dir(functionLogsPath), 0o750)
	_ = os.WriteFile(functionLogsPath, data, 0o640)
}
func (s *LifecycleService) ListFunctionLogs(actor operator.Actor, projectID, name string) ([]FunctionLog, error) {
	if actor.Kind != operator.ActorOperator || strings.TrimSpace(projectID) == "" {
		return nil, ErrFunctionDraftInvalid
	}
	functionLogs.Lock()
	defer functionLogs.Unlock()
	data, _ := os.ReadFile(functionLogsPath)
	_ = json.Unmarshal(data, &functionLogs.items)
	result := make([]FunctionLog, 0)
	for _, item := range functionLogs.items {
		if item.ProjectID == projectID && (name == "" || item.Name == name) {
			result = append(result, item)
		}
	}
	return result, nil
}

func loadFunctionDraftsLocked() {
	data, err := os.ReadFile(functionDraftsPath)
	if err != nil {
		return
	}
	var drafts []FunctionDraft
	if json.Unmarshal(data, &drafts) != nil {
		return
	}
	for _, draft := range drafts {
		functionDrafts.items[draft.ProjectID+":"+draft.Name] = draft
	}
}

func persistFunctionDraftsLocked() error {
	if err := os.MkdirAll(filepath.Dir(functionDraftsPath), 0o750); err != nil {
		return err
	}
	drafts := make([]FunctionDraft, 0, len(functionDrafts.items))
	for _, draft := range functionDrafts.items {
		drafts = append(drafts, draft)
	}
	data, err := json.MarshalIndent(drafts, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(functionDraftsPath, data, 0o640)
}

func (s *LifecycleService) CreateFunctionDraft(actor operator.Actor, projectID, name, source string) (FunctionDraft, error) {
	return s.CreateFunctionDraftWithSettings(actor, projectID, name, source, 1, false, nil, "POST")
}

func normalizeFunctionMethod(method string) (string, error) {
	method = strings.ToUpper(strings.TrimSpace(method))
	if method == "" {
		method = "POST"
	}
	if method != "GET" && method != "POST" && method != "PUT" && method != "PATCH" {
		return "", errors.New("unsupported function method")
	}
	return method, nil
}

func functionMethodFromSource(source string) (string, error) {
	matches := functionMethodPattern.FindStringSubmatch(source)
	if len(matches) != 2 {
		return "", errors.New("function config must export method as GET, POST, PUT, or PATCH")
	}
	return normalizeFunctionMethod(matches[1])
}

func (s *LifecycleService) CreateFunctionDraftWithSettings(actor operator.Actor, projectID, name, source string, version int, public bool, permissions []string, method string) (FunctionDraft, error) {
	if actor.Kind != operator.ActorOperator || strings.TrimSpace(projectID) == "" || !functionNamePattern.MatchString(strings.TrimSpace(name)) || strings.TrimSpace(source) == "" || len(source) > maxFunctionSourceBytes {
		return FunctionDraft{}, ErrFunctionDraftInvalid
	}
	if version == 0 {
		version = 1
	}
	if version < 1 {
		return FunctionDraft{}, ErrFunctionDraftInvalid
	}
	method, err := functionMethodFromSource(source)
	if err != nil {
		return FunctionDraft{}, err
	}
	if s.projects != nil {
		if _, err := s.projects.Resolve(projectID); err != nil {
			return FunctionDraft{}, err
		}
	}
	draft := FunctionDraft{ProjectID: projectID, Name: strings.TrimSpace(name), Source: source, Version: version, Public: public, Permissions: append([]string(nil), permissions...), Method: method, CreatedAt: time.Now().UTC()}
	functionDrafts.Lock()
	loadFunctionDraftsLocked()
	functionDrafts.items[projectID+":"+draft.Name] = draft
	if err := persistFunctionDraftsLocked(); err != nil {
		functionDrafts.Unlock()
		return FunctionDraft{}, err
	}
	functionDrafts.Unlock()
	return draft, nil
}

func (s *LifecycleService) UpdateFunctionDraft(actor operator.Actor, projectID, name, source string, version int) (FunctionDraft, error) {
	if actor.Kind != operator.ActorOperator || strings.TrimSpace(projectID) == "" || !functionNamePattern.MatchString(strings.TrimSpace(name)) || len(source) > maxFunctionSourceBytes || version < 1 {
		return FunctionDraft{}, ErrFunctionDraftInvalid
	}
	if strings.TrimSpace(source) == "" {
		return FunctionDraft{}, errors.New("function source is required")
	}
	functionDrafts.Lock()
	defer functionDrafts.Unlock()
	loadFunctionDraftsLocked()
	key := projectID + ":" + strings.TrimSpace(name)
	draft, ok := functionDrafts.items[key]
	if !ok {
		return FunctionDraft{}, ErrFunctionDraftInvalid
	}
	draft.Source = source
	method, err := functionMethodFromSource(source)
	if err != nil {
		return FunctionDraft{}, err
	}
	draft.Method = method
	draft.Version = version
	functionDrafts.items[key] = draft
	if err := persistFunctionDraftsLocked(); err != nil {
		return FunctionDraft{}, err
	}
	return draft, nil
}

func (s *LifecycleService) ListFunctionDrafts(actor operator.Actor, projectID string) ([]FunctionDraft, error) {
	if actor.Kind != operator.ActorOperator || strings.TrimSpace(projectID) == "" {
		return nil, ErrFunctionDraftInvalid
	}
	if s.projects != nil {
		if _, err := s.projects.Resolve(projectID); err != nil {
			return nil, err
		}
	}
	functionDrafts.Lock()
	defer functionDrafts.Unlock()
	loadFunctionDraftsLocked()
	drafts := make([]FunctionDraft, 0)
	for _, draft := range functionDrafts.items {
		if draft.ProjectID == projectID {
			drafts = append(drafts, draft)
		}
	}
	return drafts, nil
}

func (s *LifecycleService) UpdateFunctionPermissions(actor operator.Actor, projectID, name string, permissions []string, public bool) error {
	if actor.Kind != operator.ActorOperator || strings.TrimSpace(projectID) == "" || !functionNamePattern.MatchString(strings.TrimSpace(name)) {
		return ErrFunctionDraftInvalid
	}
	functionDrafts.Lock()
	defer functionDrafts.Unlock()
	loadFunctionDraftsLocked()
	draft, ok := functionDrafts.items[projectID+":"+strings.TrimSpace(name)]
	if !ok {
		return ErrFunctionDraftInvalid
	}
	draft.Permissions = append([]string(nil), permissions...)
	draft.Public = public
	functionDrafts.items[projectID+":"+draft.Name] = draft
	return persistFunctionDraftsLocked()
}

func (s *LifecycleService) PublishFunction(actor operator.Actor, projectID, name, method string) (FunctionDraft, error) {
	if actor.Kind != operator.ActorOperator || strings.TrimSpace(projectID) == "" || !functionNamePattern.MatchString(strings.TrimSpace(name)) {
		return FunctionDraft{}, ErrFunctionDraftInvalid
	}
	functionDrafts.Lock()
	defer functionDrafts.Unlock()
	loadFunctionDraftsLocked()
	draft, ok := functionDrafts.items[projectID+":"+strings.TrimSpace(name)]
	if !ok {
		return FunctionDraft{}, ErrFunctionDraftInvalid
	}
	method, err := functionMethodFromSource(draft.Source)
	if err != nil {
		return FunctionDraft{}, err
	}
	now := time.Now().UTC()
	draft.Public = true
	draft.Method = method
	draft.PublishedAt = &now
	functionDrafts.items[projectID+":"+draft.Name] = draft
	if err := persistFunctionDraftsLocked(); err != nil {
		return FunctionDraft{}, err
	}
	return draft, nil
}

func (s *LifecycleService) PublicInvokeFunction(projectID, name, method string, input json.RawMessage) (map[string]any, error) {
	edgeStarted := time.Now()
	functionDrafts.Lock()
	loadFunctionDraftsLocked()
	draft, ok := functionDrafts.items[projectID+":"+strings.TrimSpace(name)]
	functionDrafts.Unlock()
	if !ok || !draft.Public {
		return nil, ErrFunctionDraftInvalid
	}
	if draft.Method == "" {
		var err error
		draft.Method, err = functionMethodFromSource(draft.Source)
		if err != nil {
			return nil, err
		}
	}
	if strings.ToUpper(strings.TrimSpace(method)) != draft.Method {
		return nil, ErrFunctionMethodNotAllowed
	}
	return s.runFunctionDraft(operator.Actor{Kind: operator.ActorOperator}, projectID, draft.Name, draft.Source, input, []FunctionLogStep{
		{Name: "edge router", Status: "success", Reason: "public route resolved and function version selected", DurationMs: time.Since(edgeStarted).Milliseconds()},
		{Name: "method authorization", Status: "success", Reason: "request method matched function source config", DurationMs: 0},
	})
}

func (s *LifecycleService) TestFunctionDraft(actor operator.Actor, projectID, name, source string) (map[string]any, error) {
	if actor.Kind != operator.ActorOperator || strings.TrimSpace(projectID) == "" || !functionNamePattern.MatchString(strings.TrimSpace(name)) || strings.TrimSpace(source) == "" || len(source) > maxFunctionSourceBytes {
		return nil, ErrFunctionDraftInvalid
	}
	if s.projects != nil {
		if _, err := s.projects.Resolve(projectID); err != nil {
			return nil, err
		}
	}
	if !functionExportPattern.MatchString(source) || functionUnsafePattern.MatchString(source) {
		return nil, ErrFunctionSourceInvalid
	}
	if _, err := functionMethodFromSource(source); err != nil {
		return nil, err
	}
	return map[string]any{"valid": true, "project_id": projectID, "name": strings.TrimSpace(name), "source_bytes": len(source), "message": "source passed the host preflight and is ready for sandbox compilation"}, nil
}

func (s *LifecycleService) RunFunctionDraft(actor operator.Actor, projectID, name, source string, input json.RawMessage) (map[string]any, error) {
	return s.runFunctionDraft(actor, projectID, name, source, input, nil)
}

func (s *LifecycleService) runFunctionDraft(actor operator.Actor, projectID, name, source string, input json.RawMessage, initialSteps []FunctionLogStep) (map[string]any, error) {
	started := time.Now()
	requestID := createFunctionRequestID()
	steps := append([]FunctionLogStep{}, initialSteps...)
	addStep := func(name, status, reason string, stepStarted time.Time, stepError error) {
		step := FunctionLogStep{Name: name, Status: status, Reason: reason, DurationMs: time.Since(stepStarted).Milliseconds()}
		if stepError != nil {
			step.Error = stepError.Error()
		}
		steps = append(steps, step)
	}
	writeLog := func(status, message string) {
		entry := FunctionLog{ProjectID: projectID, Name: name, Status: status, Message: message, DurationMs: time.Since(started).Milliseconds(), CreatedAt: time.Now().UTC(), Steps: steps, RequestID: requestID}
		if status == "failed" {
			entry.Error = message
		}
		appendFunctionLog(entry)
	}
	validationStarted := time.Now()
	if _, err := s.TestFunctionDraft(actor, projectID, name, source); err != nil {
		addStep("host validation", "failed", "source rejected before execution", validationStarted, err)
		writeLog("failed", err.Error())
		return nil, err
	}
	addStep("host validation", "success", "source and execution policy accepted", validationStarted, nil)
	if len(input) == 0 {
		input = json.RawMessage(`null`)
	}
	inputStarted := time.Now()
	if !json.Valid(input) {
		err := errors.New("function input must be valid JSON")
		addStep("input validation", "failed", "request input is not valid JSON", inputStarted, err)
		writeLog("failed", err.Error())
		return nil, err
	}
	addStep("input validation", "success", "request input is valid JSON", inputStarted, nil)
	const runner = `const vm=await import("node:vm");const clock=()=>Number(process.hrtime.bigint())/1e6;const source=process.env.BRAIN_FUNCTION_SOURCE;const input=JSON.parse(process.env.BRAIN_FUNCTION_INPUT);const context=vm.createContext(Object.create(null));const compileStart=clock();const module=new vm.SourceTextModule(source,{context});await module.link(()=>{throw new Error("imports are not permitted")});const compileMs=clock()-compileStart;const evaluateStart=clock();await module.evaluate({timeout:450});const evaluateMs=clock()-evaluateStart;const handler=module.namespace.default;if(typeof handler!=="function")throw new Error("default export must be a function");const invokeStart=clock();const result=await Promise.race([Promise.resolve(handler(input)),new Promise((_,reject)=>setTimeout(()=>reject(new Error("function execution timed out")),400))]);const invokeMs=clock()-invokeStart;process.stdout.write(JSON.stringify({result:result??null,timings:{compile_ms:compileMs,evaluate_ms:evaluateMs,invoke_ms:invokeMs}}));`
	commandContext, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	command := exec.CommandContext(commandContext, "node", "--experimental-vm-modules", "--input-type=module", "-e", runner)
	command.Env = append(os.Environ(), "BRAIN_FUNCTION_SOURCE="+source, "BRAIN_FUNCTION_INPUT="+string(input))
	executionStarted := time.Now()
	output, err := command.CombinedOutput()
	if err != nil {
		detail := strings.TrimSpace(string(output))
		if len(detail) > 4096 {
			detail = detail[:4096]
		}
		executionError := errors.New("function execution failed")
		if detail != "" {
			executionError = fmt.Errorf("function execution failed: %s", detail)
		}
		addStep("sandbox execution", "failed", "sandbox did not produce a result", executionStarted, executionError)
		writeLog("failed", executionError.Error())
		return nil, executionError
	}
	addStep("sandbox execution", "success", "sandbox returned a JSON payload", executionStarted, nil)
	resultStarted := time.Now()
	var envelope struct {
		Result  json.RawMessage `json:"result"`
		Timings struct {
			CompileMs  float64 `json:"compile_ms"`
			EvaluateMs float64 `json:"evaluate_ms"`
			InvokeMs   float64 `json:"invoke_ms"`
		} `json:"timings"`
	}
	if err := json.Unmarshal(output, &envelope); err != nil {
		resultError := errors.New("function must return a JSON object, array, or null")
		addStep("result validation", "failed", "function returned malformed JSON", resultStarted, resultError)
		writeLog("failed", resultError.Error())
		return nil, resultError
	}
	steps = append(steps,
		FunctionLogStep{Name: "sandbox compile", Status: "success", Reason: "module compiled inside the sandbox", DurationMs: int64(envelope.Timings.CompileMs)},
		FunctionLogStep{Name: "sandbox evaluate", Status: "success", Reason: "module evaluated inside the sandbox", DurationMs: int64(envelope.Timings.EvaluateMs)},
		FunctionLogStep{Name: "handler execution", Status: "success", Reason: "default handler completed inside the sandbox", DurationMs: int64(envelope.Timings.InvokeMs)},
	)
	result := any(nil)
	if err := json.Unmarshal(envelope.Result, &result); err != nil {
		resultError := errors.New("function must return a JSON object, array, or null")
		addStep("result validation", "failed", "function returned malformed JSON", resultStarted, resultError)
		writeLog("failed", resultError.Error())
		return nil, resultError
	}
	switch result.(type) {
	case string, float64, bool:
		resultError := errors.New("function must return a JSON object, array, or null")
		addStep("result validation", "failed", "scalar and string returns are not allowed", resultStarted, resultError)
		writeLog("failed", resultError.Error())
		return nil, resultError
	}
	addStep("result validation", "success", "result conforms to the JSON return contract", resultStarted, nil)
	responseStarted := time.Now()
	if _, err := json.Marshal(result); err != nil {
		addStep("response serialization", "failed", "result could not be encoded for the edge response", responseStarted, err)
		writeLog("failed", "function response serialization failed")
		return nil, errors.New("function response serialization failed")
	}
	addStep("response serialization", "success", "JSON response encoded for the caller", responseStarted, nil)
	writeLog("success", "function completed successfully")
	return map[string]any{"valid": true, "project_id": projectID, "name": name, "result": result}, nil
}

func createFunctionRequestID() string {
	bytes := make([]byte, 12)
	if _, err := rand.Read(bytes); err != nil {
		return fmt.Sprintf("function-%d", time.Now().UnixNano())
	}
	return fmt.Sprintf("function-%x", bytes)
}
