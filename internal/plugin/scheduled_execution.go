package plugin

import (
	"context"
	"encoding/json"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"

	"brain/internal/execution"

	"gopkg.in/yaml.v3"
)

// ScheduledTypeScriptKind is distinct from the standalone Neuron profile.
// Never advertise it until the plugin-compatible OS sandbox probe passes.
const ScheduledTypeScriptKind = "plugin-typescript-v1"

const (
	ScheduledTypeScriptCPU    = 30 * time.Second
	ScheduledTypeScriptMemory = int64(2 << 30)
	scheduledInvocationLimit  = 20 * time.Second
	scheduledMessageLimit     = 1 << 20
)

// ScheduledArtifact pins a trusted, manifest-selected plugin bundle. It is
// shared across installations; the invocation's project remains mandatory.
type ScheduledArtifact struct {
	catalog *execution.Catalog
	key     execution.ArtifactKey
	runtime execution.Runtime
}

// RegisterScheduledArtifact snapshots only declared source files and the
// manifest selected by the host. The catalog refuses replacement of an exact
// version. Call only with trusted registry metadata and a host-owned source
// directory; the runtime must validate the extracted source again.
func RegisterScheduledArtifact(catalog *execution.Catalog, runtime execution.Runtime, manifest Manifest, directory string) (*ScheduledArtifact, error) {
	if catalog == nil || runtime.Kind != ScheduledTypeScriptKind || ValidateScheduledManifest(manifest) != nil {
		return nil, ErrInvalidWorkerRequest
	}
	if err := ValidatePluginSources(directory, manifest); err != nil {
		return nil, err
	}
	encoded, err := yaml.Marshal(manifest)
	if err != nil {
		return nil, err
	}
	files := []execution.BundleFile{{Path: "plugin.yaml", Data: encoded}}
	for _, source := range manifest.SourceFiles {
		data, err := os.ReadFile(filepath.Join(directory, source))
		if err != nil {
			return nil, err
		}
		files = append(files, execution.BundleFile{Path: source, Data: data})
	}
	bundle, err := execution.BuildBundle(files, execution.MaxBundleBytes)
	if err != nil {
		return nil, err
	}
	entrypoints := make([]string, 0, len(manifest.Routes)+len(manifest.Observers))
	for _, route := range manifest.Routes {
		entrypoints = append(entrypoints, route.Operation)
	}
	for _, observer := range manifest.Observers {
		entrypoints = append(entrypoints, observer.operation())
	}
	key := execution.ArtifactKey{ID: manifest.ID, Version: manifest.Version}
	if err := catalog.Register(execution.Registration{Key: key, Runtime: runtime, Entrypoints: entrypoints, Bundle: bundle}); err != nil {
		return nil, err
	}
	return &ScheduledArtifact{catalog: catalog, key: key, runtime: runtime}, nil
}

// ValidateScheduledManifest permits dependency metadata, not dependency source
// imports. Brain resolves available dependencies and admits any host.call route;
// the Neuron receives only the declared sources and the trusted runtime SDK.
func ValidateScheduledManifest(manifest Manifest) error {
	if manifest.HostAPI != 1 {
		return ErrInvalidWorkerRequest
	}
	if err := ValidateManifestContract(manifest); err != nil {
		return err
	}
	for _, dependency := range manifest.PluginImports {
		if !slices.Contains(manifest.Dependencies, dependency) && !slices.Contains(manifest.OptionalDependencies, dependency) {
			return ErrInvalidWorkerRequest
		}
	}
	for _, imported := range manifest.Imports {
		clean := path.Clean(imported)
		if strings.HasPrefix(clean, "../") && !strings.HasPrefix(clean, "../sdk-typescript/") {
			return ErrInvalidWorkerRequest
		}
	}
	return nil
}

// ScheduledWorker is an installation-scoped lifecycle handle, not a process.
// It has no dedicated slot and dispatches each invocation through the supplied
// trusted executor (Cortex). Capacity loss never invokes a legacy fallback.
type ScheduledWorker struct {
	artifact *ScheduledArtifact
	executor execution.Executor
	project  string
	mu       sync.RWMutex
	ctx      context.Context
	cancel   context.CancelFunc
}

func (a *ScheduledArtifact) Worker(projectID string, executor execution.Executor) (*ScheduledWorker, error) {
	if a == nil || executor == nil || !isValidProjectID(projectID) {
		return nil, ErrInvalidWorkerRequest
	}
	return &ScheduledWorker{artifact: a, executor: executor, project: projectID}, nil
}

func (w *ScheduledWorker) Start() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.ctx != nil {
		return ErrWorkerAlreadyRunning
	}
	w.ctx, w.cancel = context.WithCancel(context.Background())
	return nil
}

func (w *ScheduledWorker) Stop() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.ctx == nil {
		return ErrWorkerNotRunning
	}
	w.cancel()
	w.ctx, w.cancel = nil, nil
	return nil
}

// Ready reports lifecycle activation, not cluster capacity. The production
// readiness surface must separately check compatible Neuron availability.
func (w *ScheduledWorker) Ready() bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.ctx != nil
}

// ExecutionAvailable separates operator-facing availability from installation
// lifecycle. Busy compatible Neurons still count as ready.
func (w *ScheduledWorker) ExecutionAvailable() bool {
	if !w.Ready() {
		return false
	}
	if pool, ok := w.executor.(interface{ HasRuntime(execution.Runtime) bool }); ok {
		return pool.HasRuntime(w.artifact.runtime)
	}
	return false
}

func (w *ScheduledWorker) Handle(request WorkerRequest) WorkerResponse {
	response, err := w.HandleExecution(request)
	if err != nil {
		return WorkerResponse{Protocol: request.Protocol, ErrorCode: err.Error()}
	}
	return response
}

func (w *ScheduledWorker) HandleExecution(request WorkerRequest) (WorkerResponse, error) {
	w.mu.RLock()
	lifecycle := w.ctx
	w.mu.RUnlock()
	if lifecycle == nil || lifecycle.Err() != nil {
		return WorkerResponse{}, ErrWorkerNotReady
	}
	caller := request.executionContext()
	if err := caller.Err(); err != nil {
		return WorkerResponse{}, err
	}
	ctx, cancel := context.WithTimeout(caller, scheduledInvocationLimit)
	defer cancel()
	stop := context.AfterFunc(lifecycle, cancel)
	defer stop()
	if lifecycle.Err() != nil {
		return WorkerResponse{}, ErrWorkerNotReady
	}
	deadline, _ := ctx.Deadline()
	if err := ValidateWorkerRequestForProject(request, 1, w.project); err != nil {
		return WorkerResponse{}, err
	}
	if request.PluginID != w.artifact.key.ID || request.Version != w.artifact.key.Version || request.Project == nil || request.HostContext == nil {
		return WorkerResponse{}, ErrInvalidWorkerRequest
	}
	artifact, runtime, err := w.artifact.catalog.Resolve(w.artifact.key, request.Operation)
	if err != nil {
		return WorkerResponse{}, ErrInvalidWorkerRequest
	}
	host := request.HostContext
	if !host.ProjectVerified || len(host.Auth.Permissions) != 1 {
		return WorkerResponse{}, ErrInvalidWorkerRequest
	}
	permission := host.Auth.Permissions[0]
	if permission.Operation != request.Operation || permission.Capability == "" {
		return WorkerResponse{}, ErrInvalidWorkerRequest
	}
	original, caller := scheduledPrincipal(host.Auth.Actor), scheduledPrincipal(host.Auth.Caller)
	if original == "" || caller == "" {
		return WorkerResponse{}, ErrInvalidWorkerRequest
	}
	encoded, err := json.Marshal(host)
	if err != nil {
		return WorkerResponse{}, ErrInvalidWorkerRequest
	}
	input := json.RawMessage(request.Payload)
	if len(input) == 0 {
		input = json.RawMessage("null")
	}
	assigned := execution.Request{
		ID: host.InvocationID, TraceID: host.InvocationID, ProjectID: w.project,
		Artifact: artifact, Runtime: runtime, Input: input, Context: encoded,
		Deadline: deadline,
		Limits:   execution.Limits{CPUTime: ScheduledTypeScriptCPU, MemoryBytes: ScheduledTypeScriptMemory, InputBytes: scheduledMessageLimit, OutputBytes: scheduledMessageLimit},
		Auth:     execution.Authorization{ProjectID: w.project, OriginalCaller: original, Caller: caller, Capability: permission.Capability, Resource: permission.Resource},
	}
	if err = assigned.Validate(time.Now()); err != nil {
		return WorkerResponse{}, ErrInvalidWorkerRequest
	}
	result, err := w.executor.Execute(ctx, assigned)
	if callerErr := caller.Err(); callerErr != nil {
		return WorkerResponse{}, callerErr
	}
	if err != nil || ctx.Err() != nil || lifecycle.Err() != nil {
		return WorkerResponse{}, ErrWorkerNotReady
	}
	return decodeScheduledResult(assigned, result)
}

func scheduledPrincipal(actor InvocationActor) string {
	principal := actor.Kind
	if actor.ID != "" {
		principal += ":" + actor.ID
	}
	if !validAccessPrincipal(principal) || invocationActor(principal) != actor {
		return ""
	}
	return principal
}

// DecodeScheduledRequest reconstructs only the already-admitted worker input.
// It never reads a project or actor from plugin payload, and does not issue
// credentials or authorize additional operations. Called by a trusted Neuron.
func DecodeScheduledRequest(assigned execution.Request, manifest Manifest) (WorkerRequest, error) {
	if err := assigned.Validate(time.Now()); err != nil {
		return WorkerRequest{}, err
	}
	if assigned.Runtime.Kind != ScheduledTypeScriptKind || assigned.Artifact.ID != manifest.ID || assigned.Artifact.Version != manifest.Version || ValidateScheduledManifest(manifest) != nil {
		return WorkerRequest{}, ErrInvalidWorkerRequest
	}
	operation := assigned.Artifact.Entrypoint
	capability := ""
	if route, ok := manifest.Route(operation); ok && route.Helper != "host.admin" && route.Helper != "host.model" && route.Helper != "host.call" {
		capability = route.Capability
	}
	for _, observer := range manifest.Observers {
		if observer.operation() == operation {
			capability = observer.Capability
			if assigned.Auth.Resource != observer.Plugin+"."+observer.Event {
				return WorkerRequest{}, ErrInvalidWorkerRequest
			}
			break
		}
	}
	if capability == "" || assigned.Auth.Capability != capability {
		return WorkerRequest{}, ErrInvalidWorkerRequest
	}
	var host InvocationContext
	if err := json.Unmarshal(assigned.Context, &host); err != nil || !host.ProjectVerified || host.InvocationID != assigned.ID ||
		scheduledPrincipal(host.Auth.Actor) != assigned.Auth.OriginalCaller || scheduledPrincipal(host.Auth.Caller) != assigned.Auth.Caller ||
		len(host.Auth.Permissions) != 1 || host.Auth.Permissions[0] != (InvocationPermission{Capability: capability, Operation: operation, Resource: assigned.Auth.Resource}) {
		return WorkerRequest{}, ErrInvalidWorkerRequest
	}
	request := WorkerRequest{Protocol: manifest.HostAPI, ProjectID: assigned.ProjectID, PluginID: manifest.ID, Version: manifest.Version, Operation: operation, Payload: assigned.Input, Project: &ProjectSnapshot{ID: assigned.ProjectID, Active: true}, HostContext: &host}
	if err := ValidateWorkerRequestForProject(request, manifest.HostAPI, assigned.ProjectID); err != nil {
		return WorkerRequest{}, err
	}
	return request, nil
}

// EncodeScheduledResult translates the legacy sandbox response. Observation
// intents are not committed here: the registry rechecks grants and validates
// them against the publishing manifest before any durable effect.
func EncodeScheduledResult(request WorkerRequest, response WorkerResponse) (execution.Result, error) {
	if response.Protocol != request.Protocol {
		return execution.Result{}, ErrInvalidWorkerRequest
	}
	if !response.Success || response.ErrorCode != "" {
		message := response.ErrorCode
		if message == "" {
			message = "plugin execution failed"
		}
		return execution.Result{Status: execution.Failed, Error: &execution.Failure{Code: "plugin_error", Message: message}}, nil
	}
	output := json.RawMessage(response.Payload)
	if len(output) == 0 {
		output = json.RawMessage("null")
	}
	if !json.Valid(output) || len(response.Observations) > maxObservationEvents {
		return execution.Result{}, ErrInvalidWorkerRequest
	}
	result := execution.Result{Status: execution.Succeeded, Output: output}
	for _, intent := range response.Observations {
		data, err := json.Marshal(intent)
		if err != nil {
			return execution.Result{}, ErrInvalidWorkerRequest
		}
		result.Effects = append(result.Effects, data)
	}
	return result, nil
}

func decodeScheduledResult(assigned execution.Request, result execution.Result) (WorkerResponse, error) {
	if err := result.Validate(assigned, result.WorkerID); err != nil {
		return WorkerResponse{}, ErrInvalidWorkerRequest
	}
	if result.Status != execution.Succeeded {
		if result.Status == execution.Failed && result.Error.Code == "plugin_error" {
			return WorkerResponse{Protocol: 1, ErrorCode: result.Error.Message}, nil
		}
		return WorkerResponse{}, ErrWorkerNotReady
	}
	if len(result.Effects) > maxObservationEvents {
		return WorkerResponse{}, ErrInvalidWorkerRequest
	}
	response := WorkerResponse{Protocol: 1, Success: true, Payload: result.Output}
	for _, effect := range result.Effects {
		var intent ObservationIntent
		if err := json.Unmarshal(effect, &intent); err != nil || intent.Event == "" || len(intent.Payload) == 0 || len(intent.Payload) > maxObservationPayload {
			return WorkerResponse{}, ErrInvalidWorkerRequest
		}
		response.Observations = append(response.Observations, intent)
	}
	return response, nil
}
