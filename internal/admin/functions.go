package admin

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"brain/internal/execution"
	"brain/internal/operator"
	"brain/internal/plugin"
	pluginsdk "brain/plugins/sdk-typescript"
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
	ProjectID         string                      `json:"project_id"`
	Name              string                      `json:"name"`
	Source            string                      `json:"source"`
	Version           int                         `json:"version"`
	Public            bool                        `json:"public"`
	Permissions       []string                    `json:"permissions"`
	Method            string                      `json:"method"`
	PublishedAt       *time.Time                  `json:"published_at,omitempty"`
	ActivePublication int                         `json:"active_publication,omitempty"`
	PublishedVersions map[int]FunctionPublication `json:"published_versions,omitempty"`
	CreatedAt         time.Time                   `json:"created_at"`
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
func (s *LifecycleService) authorizeFunction(actor operator.Actor, projectID, operation, name string) error {
	if s.registry == nil {
		return nil
	}
	return s.registry.AuthorizeHostOperation(projectID, "operator:"+actor.ID, "functions", operation, strings.TrimSpace(name))
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
		if item.ProjectID == projectID && (name == "" || item.Name == name) && s.authorizeFunction(actor, projectID, "functions.logs", item.Name) == nil {
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
		// Preserve the first known legacy publication before editing the draft.
		draft = adoptLegacyPublication(draft)
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
	temporary, err := os.CreateTemp(filepath.Dir(functionDraftsPath), ".function-drafts-*")
	if err != nil {
		return err
	}
	defer os.Remove(temporary.Name())
	if _, err := temporary.Write(data); err != nil {
		_ = temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	return os.Rename(temporary.Name(), functionDraftsPath)
}

// Save under the draft lock and restore the last admitted state on a failed write.
func saveFunctionDraftLocked(draft FunctionDraft) error {
	key := draft.ProjectID + ":" + draft.Name
	previous, exists := functionDrafts.items[key]
	functionDrafts.items[key] = draft
	if err := persistFunctionDraftsLocked(); err != nil {
		if exists {
			functionDrafts.items[key] = previous
		} else {
			delete(functionDrafts.items, key)
		}
		return err
	}
	return nil
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
	if err := s.authorizeFunction(actor, projectID, "functions.publish", name); err != nil {
		return FunctionDraft{}, err
	}
	draft := FunctionDraft{ProjectID: projectID, Name: strings.TrimSpace(name), Source: source, Version: version, Public: public, Permissions: append([]string(nil), permissions...), Method: method, CreatedAt: time.Now().UTC()}
	functionDrafts.Lock()
	defer functionDrafts.Unlock()
	loadFunctionDraftsLocked()
	if previous, ok := functionDrafts.items[projectID+":"+draft.Name]; ok {
		draft.ActivePublication = previous.ActivePublication
		draft.PublishedVersions = previous.PublishedVersions
		draft.PublishedAt = previous.PublishedAt
	}
	if draft.Public && draft.ActivePublication == 0 {
		draft, err = pinFunctionPublication(draft)
		if err != nil {
			return FunctionDraft{}, err
		}
	}
	if err := saveFunctionDraftLocked(draft); err != nil {
		return FunctionDraft{}, err
	}
	return draft, nil
}

func (s *LifecycleService) UpdateFunctionDraft(actor operator.Actor, projectID, name, source string, version int) (FunctionDraft, error) {
	if actor.Kind != operator.ActorOperator || strings.TrimSpace(projectID) == "" || !functionNamePattern.MatchString(strings.TrimSpace(name)) || len(source) > maxFunctionSourceBytes || version < 1 {
		return FunctionDraft{}, ErrFunctionDraftInvalid
	}
	if strings.TrimSpace(source) == "" {
		return FunctionDraft{}, errors.New("function source is required")
	}
	if err := s.authorizeFunction(actor, projectID, "functions.publish", name); err != nil {
		return FunctionDraft{}, err
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
	if err := saveFunctionDraftLocked(draft); err != nil {
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
		if draft.ProjectID == projectID && s.authorizeFunction(actor, projectID, "functions.inspect", draft.Name) == nil {
			drafts = append(drafts, draft)
		}
	}
	return drafts, nil
}

func (s *LifecycleService) UpdateFunctionPermissions(actor operator.Actor, projectID, name string, permissions []string, public bool) error {
	if actor.Kind != operator.ActorOperator || strings.TrimSpace(projectID) == "" || !functionNamePattern.MatchString(strings.TrimSpace(name)) {
		return ErrFunctionDraftInvalid
	}
	if err := s.authorizeFunction(actor, projectID, "functions.publish", name); err != nil {
		return err
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
	if public && draft.ActivePublication == 0 {
		var err error
		draft, err = pinFunctionPublication(draft)
		if err != nil {
			return err
		}
	}
	return saveFunctionDraftLocked(draft)
}

func (s *LifecycleService) PublishFunction(actor operator.Actor, projectID, name, method string) (FunctionDraft, error) {
	if actor.Kind != operator.ActorOperator || strings.TrimSpace(projectID) == "" || !functionNamePattern.MatchString(strings.TrimSpace(name)) {
		return FunctionDraft{}, ErrFunctionDraftInvalid
	}
	if err := s.authorizeFunction(actor, projectID, "functions.publish", name); err != nil {
		return FunctionDraft{}, err
	}
	if s.projects != nil {
		if _, err := s.projects.Resolve(projectID); err != nil {
			return FunctionDraft{}, err
		}
	}
	draft, err := persistPublishedFunction(projectID, name)
	if err != nil {
		return FunctionDraft{}, err
	}
	if s.registry != nil {
		s.registry.PublishHostObservationAs(projectID, "functions", "published", "operator:"+actor.ID, map[string]any{"name": draft.Name, "version": draft.Version})
	}
	return draft, nil
}

func persistPublishedFunction(projectID, name string) (FunctionDraft, error) {
	functionDrafts.Lock()
	defer functionDrafts.Unlock()
	loadFunctionDraftsLocked()
	draft, ok := functionDrafts.items[projectID+":"+strings.TrimSpace(name)]
	if !ok {
		return FunctionDraft{}, ErrFunctionDraftInvalid
	}
	draft, err := pinNextFunctionPublication(draft)
	if err != nil {
		return FunctionDraft{}, err
	}
	now := time.Now().UTC()
	draft.Public = true
	draft.Method = draft.PublishedVersions[draft.ActivePublication].Method
	draft.PublishedAt = &now
	if err := saveFunctionDraftLocked(draft); err != nil {
		return FunctionDraft{}, err
	}
	return draft, nil
}

func (s *LifecycleService) PublicInvokeFunction(projectID, name, method string, input json.RawMessage) (map[string]any, error) {
	return s.PublicInvokeFunctionContext(context.Background(), projectID, name, method, input)
}

func (s *LifecycleService) PublicInvokeFunctionContext(ctx context.Context, projectID, name, method string, input json.RawMessage) (map[string]any, error) {
	return s.publicInvokeFunction(ctx, projectID, name, method, input, plugin.InvocationActor{Kind: "public", Level: "anonymous"}, "public")
}

// PublicInvokeFunctionWithSession authenticates the existing published function
// endpoint without permitting a bearer token to fall back to anonymous access.
func (s *LifecycleService) PublicInvokeFunctionWithSession(projectID, name, method string, input json.RawMessage, token string) (map[string]any, error) {
	return s.PublicInvokeFunctionWithSessionContext(context.Background(), projectID, name, method, input, token)
}

func (s *LifecycleService) PublicInvokeFunctionWithSessionContext(ctx context.Context, projectID, name, method string, input json.RawMessage, token string) (map[string]any, error) {
	if s.registry == nil {
		return nil, plugin.ErrActorDenied
	}
	actor, err := s.registry.ResolveUserSession(projectID, token)
	if err != nil {
		return nil, err
	}
	return s.publicInvokeFunction(ctx, projectID, name, method, input, actor, "user:"+actor.ID)
}

func (s *LifecycleService) publicInvokeFunction(ctx context.Context, projectID, name, method string, input json.RawMessage, authActor plugin.InvocationActor, principal string) (map[string]any, error) {
	edgeStarted := time.Now()
	draft, publication, err := loadPublicFunctionPublication(projectID, name)
	if err != nil {
		return nil, err
	}
	if s.registry != nil {
		if err := s.registry.AuthorizeHostOperation(projectID, principal, "functions", "functions.call", draft.Name); err != nil {
			return nil, err
		}
	}
	if strings.ToUpper(strings.TrimSpace(method)) != publication.Method {
		return nil, ErrFunctionMethodNotAllowed
	}
	guard := func() error {
		_, current, err := loadPublicFunctionPublication(projectID, draft.Name)
		if err != nil || current != publication {
			return ErrFunctionDraftInvalid
		}
		if s.registry != nil {
			return s.registry.AuthorizeHostOperation(projectID, principal, "functions", "functions.call", draft.Name)
		}
		return nil
	}
	result, err := s.runFunctionDraft(ctx, operator.Actor{Kind: operator.ActorOperator}, projectID, draft.Name, publication.Source, input, authActor, principal, "functions.call", []FunctionLogStep{
		{Name: "edge router", Status: "success", Reason: "public route resolved and function version selected", DurationMs: time.Since(edgeStarted).Milliseconds()},
		{Name: "method authorization", Status: "success", Reason: "request method matched function source config", DurationMs: 0},
	}, guard)
	return result, err
}

func (s *LifecycleService) TestFunctionDraft(actor operator.Actor, projectID, name, source string) (map[string]any, error) {
	if err := s.authorizeFunction(actor, projectID, "functions.invoke", name); err != nil {
		return nil, err
	}
	return s.validateFunctionDraft(actor, projectID, name, source)
}

func (s *LifecycleService) validateFunctionDraft(actor operator.Actor, projectID, name, source string) (map[string]any, error) {
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
	return s.RunFunctionDraftContext(context.Background(), actor, projectID, name, source, input)
}

func (s *LifecycleService) RunFunctionDraftContext(ctx context.Context, actor operator.Actor, projectID, name, source string, input json.RawMessage) (map[string]any, error) {
	if err := s.authorizeFunction(actor, projectID, "functions.invoke", name); err != nil {
		return nil, err
	}
	return s.runFunctionDraft(ctx, actor, projectID, name, source, input, plugin.InvocationActor{Kind: "operator", ID: actor.ID, Authenticated: true, Level: "authenticated"}, "operator:"+actor.ID, "functions.invoke", nil, func() error {
		return s.authorizeFunction(actor, projectID, "functions.invoke", name)
	})
}

func (s *LifecycleService) runFunctionDraft(ctx context.Context, actor operator.Actor, projectID, name, source string, input json.RawMessage, authActor plugin.InvocationActor, principal, operation string, initialSteps []FunctionLogStep, guard func() error) (map[string]any, error) {
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
	if _, err := s.validateFunctionDraft(actor, projectID, name, source); err != nil {
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
	model := map[string]any{"id": projectID, "active": true}
	if s.projects != nil {
		resolved, err := s.projects.Resolve(projectID)
		if err != nil {
			writeLog("failed", err.Error())
			return nil, err
		}
		model = map[string]any{"id": resolved.ID, "active": resolved.Active}
	}
	projectJSON, err := json.Marshal(model)
	if err != nil {
		return nil, err
	}
	authJSON, err := json.Marshal(plugin.InvocationAuth{Actor: authActor, Caller: authActor, Permissions: []plugin.InvocationPermission{{Capability: operation, Operation: operation, Resource: name}}})
	if err != nil {
		return nil, err
	}
	const runner = `const vm=await import("node:vm");
const clock=()=>Number(process.hrtime.bigint())/1e6;
const io=await import("node:fs");
const bridgeMax=256*1024;
const readExact=(fd,length)=>{
  const data=Buffer.alloc(length);let offset=0;
  while(offset<length){const n=io.readSync(fd,data,offset,length-offset,null);if(!n)throw new Error("bridge closed");offset+=n;}
  return data;
};
const writeExact=(fd,data)=>{
  let offset=0;
  while(offset<data.length){const n=io.writeSync(fd,data,offset,data.length-offset,null);if(!n)throw new Error("bridge closed");offset+=n;}
};
// The outer-realm callback accepts and returns only strings. No Node objects,
// promises, errors, file handles or functions are exposed to sandbox source.
const hostCall=(message)=>{
  try{
    if(typeof message!=="string")throw new Error("invalid bridge request");
    const payload=Buffer.from(message,"utf8");
    if(!payload.length||payload.length>bridgeMax)throw new Error("invalid bridge request");
    const header=Buffer.alloc(4);header.writeUInt32BE(payload.length);
    writeExact(3,header);writeExact(3,payload);
    const length=readExact(4,4).readUInt32BE();
    if(!length||length>bridgeMax)throw new Error("invalid bridge response");
    return readExact(4,length).toString("utf8");
  }catch{return '{"error":"Plugin bridge unavailable or call limit exceeded"}';}
};
const sandbox=vm.createContext(Object.create(null),{codeGeneration:{strings:false,wasm:false}});
const helpers=new vm.SourceTextModule(process.env.BRAIN_PURE_HELPERS,{context:sandbox});
await helpers.link(()=>{throw new Error("helper imports are not permitted")});
await helpers.evaluate({timeout:450});
sandbox.__brainHelper=helpers.namespace.callHelper;
// Keep the outer-realm callback private in a sandbox-realm closure. Capture
// JSON primitives before evaluating untrusted source; parse responses in the VM.
const makePlugins=vm.runInContext("(hostCall)=>{const stringify=JSON.stringify,parse=JSON.parse,Failure=Error;return Object.freeze({call:async (operation,payload)=>{if(typeof operation!=='string')throw new Failure('Plugin operation must be a string');const result=parse(hostCall(stringify({operation,payload})));if(result.error)throw new Failure(result.error);return result.result;}})}",sandbox);
sandbox.__brainPlugins=makePlugins(hostCall);
sandbox.__brainInput=vm.runInContext("JSON.parse("+JSON.stringify(process.env.BRAIN_FUNCTION_INPUT)+")",sandbox);
sandbox.__brainAuth=vm.runInContext("(()=>{const a=JSON.parse("+JSON.stringify(process.env.BRAIN_FUNCTION_AUTH)+");const permissions=Object.freeze(a.permissions.map(p=>Object.freeze(p)));return Object.freeze({actor:Object.freeze(a.actor),caller:Object.freeze(a.caller),permissions,can:(capability,resource='')=>permissions.some(p=>p.capability===capability&&p.resource===resource)})})()",sandbox);
sandbox.__brainContext=vm.runInContext("Object.freeze({auth:__brainAuth,project:Object.freeze(JSON.parse("+JSON.stringify(process.env.BRAIN_FUNCTION_PROJECT)+")),plugin:Object.freeze({id:'functions'}),execution:Object.freeze({operation:"+JSON.stringify(process.env.BRAIN_FUNCTION_OPERATION)+",id:"+JSON.stringify(process.env.BRAIN_INVOCATION_ID)+",invocationId:"+JSON.stringify(process.env.BRAIN_INVOCATION_ID)+"}),helpers:Object.freeze({call:__brainHelper}),plugins:__brainPlugins})",sandbox);
const compileStart=clock();
const module=new vm.SourceTextModule(process.env.BRAIN_FUNCTION_SOURCE,{context:sandbox});
await module.link(()=>{throw new Error("imports are not permitted")});
const compileMs=clock()-compileStart;
const evaluateStart=clock();
await module.evaluate({timeout:450});
const evaluateMs=clock()-evaluateStart;
const handler=module.namespace.default;
if(typeof handler!=="function")throw new Error("default export must be a function");
sandbox.__brainHandler=handler;
const invokeStart=clock();
let timer;
try {
const result=await Promise.race([Promise.resolve(vm.runInContext("__brainHandler(__brainInput,__brainContext)",sandbox,{timeout:450})),new Promise((_,reject)=>{timer=setTimeout(()=>reject(new Error("function execution timed out")),400)})]);
const invokeMs=clock()-invokeStart;
process.stdout.write(JSON.stringify({result:result??null,timings:{compile_ms:compileMs,evaluate_ms:evaluateMs,invoke_ms:invokeMs}}));
} finally {clearTimeout(timer);}`
	// Allow cold Node startup under load without relaxing the VM execution deadlines.
	timeout := 5 * time.Second
	if s.functionExecution != nil {
		timeout = 15 * time.Second
	}
	commandContext, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	executionStarted := time.Now()
	var output []byte
	if s.functionExecution != nil {
		if operation == "functions.invoke" {
			commandContext = execution.WithHostHandler(commandContext, func(callCtx context.Context, target string, payload json.RawMessage) (json.RawMessage, error) {
				response := s.functionPluginRead(callCtx, projectID, name, principal, operation, functionPluginRequest{Operation: target, Payload: payload})
				if response.Error != "" {
					return nil, execution.ErrHostCallDenied
				}
				return response.Result, nil
			})
		}
		output, err = s.functionExecution.run(commandContext, projectID, name, source, requestID, principal, operation, input)
	} else {
		command := exec.CommandContext(commandContext, "node", "--no-warnings", "--experimental-vm-modules", "--input-type=module", "-e", runner)
		command.Env = []string{"PATH=" + os.Getenv("PATH"), "BRAIN_FUNCTION_SOURCE=" + source, "BRAIN_FUNCTION_INPUT=" + string(input), "BRAIN_FUNCTION_PROJECT=" + string(projectJSON), "BRAIN_FUNCTION_AUTH=" + string(authJSON), "BRAIN_FUNCTION_OPERATION=" + operation, "BRAIN_INVOCATION_ID=" + requestID, "BRAIN_PURE_HELPERS=" + pluginsdk.PureHelperSource}
		output, err = s.runFunctionCommand(commandContext, command, projectID, name, principal, operation)
	}
	if err != nil {
		var exitError *exec.ExitError
		detail := ""
		if commandContext.Err() != nil {
			detail = commandContext.Err().Error()
		}
		if errors.As(err, &exitError) && detail == "" {
			detail = strings.TrimSpace(string(exitError.Stderr))
		}
		if len(detail) > 4096 {
			detail = detail[:4096]
		}
		executionError := ErrFunctionExecutionFailed
		if errors.Is(commandContext.Err(), context.DeadlineExceeded) {
			executionError = ErrFunctionExecutionTimeout
		} else if s.functionExecution != nil {
			executionError = scheduledFunctionError(err)
		} else if detail != "" {
			executionError = fmt.Errorf("%w: %s", ErrFunctionExecutionFailed, detail)
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
		resultError := ErrFunctionResultInvalid
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
		resultError := ErrFunctionResultInvalid
		addStep("result validation", "failed", "function returned malformed JSON", resultStarted, resultError)
		writeLog("failed", resultError.Error())
		return nil, resultError
	}
	switch result.(type) {
	case string, float64, bool:
		resultError := ErrFunctionResultInvalid
		addStep("result validation", "failed", "scalar and string returns are not allowed", resultStarted, resultError)
		writeLog("failed", resultError.Error())
		return nil, resultError
	}
	addStep("result validation", "success", "result conforms to the JSON return contract", resultStarted, nil)
	responseStarted := time.Now()
	if _, err := json.Marshal(result); err != nil {
		addStep("response serialization", "failed", "result could not be encoded for the edge response", responseStarted, err)
		writeLog("failed", "function response serialization failed")
		return nil, ErrFunctionExecutionFailed
	}
	addStep("response serialization", "success", "JSON response encoded for the caller", responseStarted, nil)
	if err := ctx.Err(); err != nil {
		writeLog("failed", err.Error())
		return nil, err
	}
	if guard != nil {
		if err := guard(); err != nil {
			addStep("result authorization", "failed", "publication or caller permission changed during execution", responseStarted, err)
			writeLog("failed", err.Error())
			return nil, err
		}
	}
	writeLog("success", "function completed successfully")
	if s.registry != nil {
		s.registry.PublishHostObservationAs(projectID, "functions", "invoked", principal, map[string]any{"name": name, "executionId": requestID})
	}
	return map[string]any{"valid": true, "project_id": projectID, "name": name, "result": result}, nil
}

func createFunctionRequestID() string {
	bytes := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, bytes); err != nil {
		return fmt.Sprintf("function-%d", time.Now().UnixNano())
	}
	return fmt.Sprintf("function-%x", bytes)
}
