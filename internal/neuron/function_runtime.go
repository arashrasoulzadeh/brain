package neuron

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	"brain/internal/execution"
	"brain/internal/plugin"
)

type UserFunctionRuntime struct {
	*TypeScriptRuntime
}

func NewUserFunctionRuntime(root string, localMacOS bool) (*UserFunctionRuntime, error) {
	runner, err := NewTypeScriptRuntime(root, localMacOS)
	if err != nil {
		return nil, err
	}
	runner.capability.Kind = execution.UserFunctionRuntimeKind
	return &UserFunctionRuntime{runner}, nil
}

func (r *UserFunctionRuntime) Run(ctx context.Context, req execution.Request, files []execution.BundleFile) (out execution.Result, runErr error) {
	if err := req.Validate(time.Now()); err != nil {
		return execution.Result{}, err
	}
	if req.Runtime != r.capability || req.Artifact.ProjectID != req.ProjectID || req.Artifact.Entrypoint != "functions.call" || req.Auth.Capability != "functions.call" || req.Auth.Resource == "" || req.Auth.Caller != req.Auth.OriginalCaller || req.ParentID != "" || req.Limits.CPUTime != TypeScriptCPU || req.Limits.MemoryBytes != TypeScriptMemory || req.Limits.InputBytes > 1<<20 || req.Limits.OutputBytes > 1<<20 || len(files) != 1 || files[0].Path != "source.mjs" || len(files[0].Data) > 256*1024 {
		return execution.Result{}, ErrRuntimeArtifact
	}
	actor := safeActor(req.Auth.Caller)
	if req.Auth.Caller != "public" && (actor.Kind != "user" || !strings.HasPrefix(req.Auth.Caller, "user:")) {
		return execution.Result{}, ErrRuntimeArtifact
	}
	bundle, err := execution.BuildBundle(files, execution.MaxBundleBytes)
	if err != nil {
		return execution.Result{}, err
	}
	if err := req.Artifact.Verify(bundle); err != nil {
		return execution.Result{}, err
	}
	ctx, cancel := context.WithDeadline(ctx, req.Deadline)
	defer cancel()
	dir, err := os.MkdirTemp("", "brain-function-")
	if err != nil {
		return execution.Result{}, err
	}
	defer os.RemoveAll(dir)
	dir, err = filepath.EvalSymlinks(dir)
	if err != nil {
		return execution.Result{}, err
	}
	root := filepath.Join(dir, "source")
	if err = os.Mkdir(root, 0o700); err != nil {
		return execution.Result{}, err
	}
	if err = os.WriteFile(filepath.Join(root, "source.mjs"), files[0].Data, 0o600); err != nil {
		return execution.Result{}, err
	}
	manifest := plugin.Manifest{ID: "functions", Version: req.Artifact.Version, HostAPI: 1, Entry: "source.mjs", SourceFiles: []string{"source.mjs"}, Capabilities: []string{"functions.call"}, Routes: []plugin.Route{{Operation: "functions.call", Function: "handler", Capability: "functions.call"}}}
	policy := plugin.IsolationPolicy{FilesystemReadOnly: true, CPUMillis: int(TypeScriptCPU.Milliseconds()), MemoryBytes: TypeScriptMemory, ProcessLimit: 512, WallTime: 2 * time.Second}
	script := filepath.Join(r.root, "scripts", "function-worker.mjs")
	descriptor := filepath.Join(dir, "execution.json")
	var worker *plugin.TypeScriptWorker
	if r.capability.Profile == MacOSProfile {
		worker, err = plugin.NewManifestMacOSTypeScriptWorker(filepath.Join(r.root, "scripts", "run-macos-typescript-worker.sh"), script, descriptor, root, req.ProjectID, manifest, policy)
	} else {
		worker, err = plugin.NewManifestBubblewrapTypeScriptWorker(filepath.Join(r.root, "scripts", "run-sandboxed-typescript-worker.sh"), script, descriptor, root, req.ProjectID, manifest, policy)
	}
	if err != nil {
		return execution.Result{}, err
	}
	if err := worker.StartContext(ctx); err != nil {
		return execution.Result{}, err
	}
	defer func() {
		_ = worker.Stop()
		if runErr == nil {
			out.Usage = worker.ProcessUsage()
		}
	}()
	invocation := plugin.WorkerRequest{Protocol: 1, ProjectID: req.ProjectID, Project: &plugin.ProjectSnapshot{ID: req.ProjectID, Active: true}, PluginID: "functions", Version: manifest.Version, Operation: "functions.call", Payload: req.Input, HostContext: &plugin.InvocationContext{InvocationID: req.ID, ProjectVerified: true, Auth: plugin.InvocationAuth{Actor: actor, Caller: actor, Permissions: []plugin.InvocationPermission{{Capability: "functions.call", Operation: "functions.call", Resource: req.Auth.Resource}}}}}
	result := worker.HandleContext(ctx, invocation)
	if err := ctx.Err(); err != nil {
		return execution.Result{}, err
	}
	if !result.Success || result.ErrorCode != "" || result.Protocol != 1 || len(result.Observations) != 0 || !json.Valid(result.Payload) {
		return execution.Result{Status: execution.Failed, Error: &execution.Failure{Code: "function_error"}}, nil
	}
	return execution.Result{Status: execution.Succeeded, Output: result.Payload}, nil
}

func (r *UserFunctionRuntime) Probe(ctx context.Context) error {
	files := []execution.BundleFile{{Path: "source.mjs", Data: []byte("export default (input, context) => ({accepted: context.project.id === 'probe' && context.auth.can('functions.call', 'probe') && typeof process === 'undefined'})")}}
	bundle, err := execution.BuildBundle(files, 4096)
	if err != nil {
		return err
	}
	req := execution.Request{ID: "function-probe", TraceID: "function-probe", ProjectID: "probe", Artifact: execution.Artifact{ProjectID: "probe", ID: "function-probe", Version: "1", Digest: execution.ArtifactDigest(bundle), Entrypoint: "functions.call"}, Runtime: r.capability, Input: json.RawMessage("null"), Deadline: time.Now().Add(15 * time.Second), Limits: execution.Limits{CPUTime: TypeScriptCPU, MemoryBytes: TypeScriptMemory, InputBytes: 4096, OutputBytes: 4096}, Auth: execution.Authorization{ProjectID: "probe", OriginalCaller: "public", Caller: "public", Capability: "functions.call", Resource: "probe"}}
	result, err := r.Run(ctx, req, files)
	if err != nil {
		return err
	}
	var envelope struct {
		Result struct {
			Accepted bool `json:"accepted"`
		} `json:"result"`
	}
	if result.Status != execution.Succeeded || json.Unmarshal(result.Output, &envelope) != nil || !envelope.Result.Accepted {
		return errors.New("isolated function runtime probe failed")
	}
	return nil
}
