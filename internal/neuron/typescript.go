package neuron

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"brain/internal/execution"
	"brain/internal/plugin"
)

const (
	TypeScriptKind = "standalone-typescript"
	LinuxProfile   = "linux-bubblewrap"
	MacOSProfile   = "macos-seatbelt-dev"
	// Node's address space includes V8 reservations. The existing launcher
	// separately caps its JavaScript heap. macOS lacks Linux's address-space cap.
	TypeScriptMemory = int64(2 << 30)
	TypeScriptCPU    = 30 * time.Second
)

var ErrRuntimeArtifact = errors.New("unsupported standalone TypeScript artifact")

// TypeScriptRuntime runs one manifest-declared operation in a fresh existing
// Brain OS sandbox. Standalone and prepared-plugin capabilities have separate
// admission rules; neither exposes host credentials or database connections.
type TypeScriptRuntime struct {
	root       string
	capability execution.Runtime
}

func NewTypeScriptRuntime(root string, localMacOS bool) (*TypeScriptRuntime, error) {
	absolute, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	absolute, err = filepath.EvalSymlinks(absolute)
	if err != nil {
		return nil, err
	}
	profile := LinuxProfile
	switch runtime.GOOS {
	case "linux":
	case "darwin":
		if !localMacOS {
			return nil, plugin.ErrInvalidIsolationPolicy
		}
		profile = MacOSProfile
	default:
		return nil, plugin.ErrInvalidIsolationPolicy
	}
	return &TypeScriptRuntime{root: absolute, capability: execution.Runtime{Kind: TypeScriptKind, Profile: profile}}, nil
}

func (r *TypeScriptRuntime) Capability() execution.Runtime { return r.capability }

func (r *TypeScriptRuntime) Run(ctx context.Context, request execution.Request, files []execution.BundleFile) (out execution.Result, runErr error) {
	if err := request.Validate(time.Now()); err != nil {
		return execution.Result{}, err
	}
	if request.Runtime != r.capability || request.Limits.CPUTime != TypeScriptCPU || request.Limits.MemoryBytes != TypeScriptMemory ||
		request.Limits.InputBytes > 1<<20 || request.Limits.OutputBytes > 1<<20 {
		return execution.Result{}, plugin.ErrInvalidIsolationPolicy
	}
	ctx, cancel := context.WithDeadline(ctx, request.Deadline)
	defer cancel()
	if err := ctx.Err(); err != nil {
		return execution.Result{}, err
	}
	dir, err := os.MkdirTemp("", "brain-neuron-")
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
	// Recheck canonical bytes and digest even for direct callers of the adapter.
	bundle, err := execution.BuildBundle(files, execution.MaxBundleBytes)
	if err != nil {
		return execution.Result{}, err
	}
	if err = request.Artifact.Verify(bundle); err != nil {
		return execution.Result{}, err
	}
	for _, file := range files {
		target := filepath.Join(root, filepath.FromSlash(file.Path))
		if err = os.MkdirAll(filepath.Dir(target), 0o700); err != nil {
			return execution.Result{}, err
		}
		if err = os.WriteFile(target, file.Data, 0o600); err != nil {
			return execution.Result{}, err
		}
	}
	manifest, err := plugin.LoadManifest(filepath.Join(root, "plugin.yaml"))
	if err != nil {
		return execution.Result{}, ErrRuntimeArtifact
	}
	if manifest.ID != request.Artifact.ID || manifest.Version != request.Artifact.Version || manifest.HostAPI != 1 || len(files) != len(manifest.SourceFiles)+1 {
		return execution.Result{}, ErrRuntimeArtifact
	}
	var invocation plugin.WorkerRequest
	bundled := r.capability.Kind == plugin.ScheduledTypeScriptKind
	if bundled {
		invocation, err = plugin.DecodeScheduledRequest(request, manifest)
		if err != nil {
			return execution.Result{}, err
		}
	} else {
		if manifest.Declarative || len(manifest.Routes) != 1 || len(manifest.Dependencies)+len(manifest.OptionalDependencies)+len(manifest.PluginImports)+len(manifest.Imports)+len(manifest.Observers)+len(manifest.Events)+len(manifest.Models) != 0 || manifest.Admin != nil || len(manifest.ConfigSchema) != 0 {
			return execution.Result{}, ErrRuntimeArtifact
		}
		route := manifest.Routes[0]
		if route.Operation != request.Artifact.Entrypoint || route.Capability != request.Auth.Capability || route.Helper != "" || route.ResourceField != "" || route.Function == "" {
			return execution.Result{}, ErrRuntimeArtifact
		}
		host := &plugin.InvocationContext{
			InvocationID: request.ID, ProjectVerified: true, Config: map[string]any{}, Available: []string{},
			Auth: plugin.InvocationAuth{Actor: safeActor(request.Auth.OriginalCaller), Caller: safeActor(request.Auth.Caller), Permissions: []plugin.InvocationPermission{{Capability: request.Auth.Capability, Operation: route.Operation, Resource: request.Auth.Resource}}},
		}
		invocation = plugin.WorkerRequest{Protocol: 1, ProjectID: request.ProjectID, Project: &plugin.ProjectSnapshot{ID: request.ProjectID, Active: true}, PluginID: manifest.ID, Version: manifest.Version, Operation: route.Operation, Payload: request.Input, HostContext: host}
	}
	policy := plugin.IsolationPolicy{FilesystemReadOnly: true, CPUMillis: int(TypeScriptCPU.Milliseconds()), MemoryBytes: TypeScriptMemory, ProcessLimit: 512, WallTime: 2 * time.Second}
	launcher, script := filepath.Join(r.root, "scripts", "run-sandboxed-typescript-worker.sh"), filepath.Join(r.root, "scripts", "typescript-worker.mjs")
	var worker *plugin.TypeScriptWorker
	if r.capability.Profile == MacOSProfile {
		launcher = filepath.Join(r.root, "scripts", "run-macos-typescript-worker.sh")
		worker, err = plugin.NewManifestMacOSTypeScriptWorker(launcher, script, filepath.Join(dir, "execution.json"), root, request.ProjectID, manifest, policy)
	} else {
		worker, err = plugin.NewManifestBubblewrapTypeScriptWorker(launcher, script, filepath.Join(dir, "execution.json"), root, request.ProjectID, manifest, policy)
	}
	if err != nil {
		return execution.Result{}, err
	}
	if bundled {
		if err = worker.BindBundledRuntime(root); err != nil {
			return execution.Result{}, err
		}
	}
	if err = worker.StartContext(ctx); err != nil {
		return execution.Result{}, err
	}
	defer func() {
		_ = worker.Stop()
		if runErr == nil {
			out.Usage = worker.ProcessUsage()
		}
	}()
	result := worker.HandleContext(ctx, invocation)
	if err = ctx.Err(); err != nil {
		return execution.Result{}, err
	}
	if bundled {
		return plugin.EncodeScheduledResult(invocation, result)
	}
	if !result.Success || result.ErrorCode != "" || result.Protocol != 1 || len(result.Observations) != 0 || !json.Valid(result.Payload) {
		return execution.Result{Status: execution.Failed, Error: &execution.Failure{Code: "runtime_error"}}, nil
	}
	return execution.Result{Status: execution.Succeeded, Output: result.Payload}, nil
}

func safeActor(principal string) plugin.InvocationActor {
	kind, id, ok := strings.Cut(principal, ":")
	if ok && (kind == "user" || kind == "operator") && id != "" {
		return plugin.InvocationActor{Kind: kind, ID: id, Authenticated: true, Level: "authenticated"}
	}
	if principal == "host" || (ok && kind == "plugin" && id != "") {
		if principal == "host" {
			kind = "host"
		}
		return plugin.InvocationActor{Kind: kind, ID: id, Level: "system"}
	}
	return plugin.InvocationActor{Kind: "public", Level: "anonymous"}
}

// Probe runs host-authored code inside the same sandbox before advertising a
// runtime. Missing sandbox prerequisites therefore never create fake capacity.
func (r *TypeScriptRuntime) Probe(ctx context.Context) error {
	if r.capability.Kind == plugin.ScheduledTypeScriptKind {
		return r.probePlugin(ctx)
	}
	files := []execution.BundleFile{
		{Path: "plugin.yaml", Data: []byte("id: neuron-probe\nversion: 1.0.0\nhost_api: 1\nentrypoint: index.mts\nsource_files:\n  - index.mts\ncapabilities:\n  - health\nroutes:\n  - operation: health\n    function: health\n    capability: health\n")},
		{Path: "index.mts", Data: []byte("export function health(): { accepted: boolean } { const memory = new WebAssembly.Memory({ initial: 1, maximum: 2 }); return { accepted: memory.buffer.byteLength === 65536 }; }\n")},
	}
	bundle, err := execution.BuildBundle(files, 4096)
	if err != nil {
		return err
	}
	request := execution.Request{ID: "probe", TraceID: "probe", ProjectID: "neuron-probe", Artifact: execution.Artifact{ID: "neuron-probe", Version: "1.0.0", Digest: execution.ArtifactDigest(bundle), Entrypoint: "health"}, Runtime: r.capability, Input: json.RawMessage("null"), Deadline: time.Now().Add(15 * time.Second), Limits: execution.Limits{CPUTime: TypeScriptCPU, MemoryBytes: TypeScriptMemory, InputBytes: 4096, OutputBytes: 4096}, Auth: execution.Authorization{ProjectID: "neuron-probe", OriginalCaller: "host", Caller: "host", Capability: "health"}}
	result, err := r.Run(ctx, request, files)
	if err != nil {
		return err
	}
	if result.Status != execution.Succeeded || string(result.Output) != "{\"accepted\":true}" {
		return errors.New("isolated Neuron startup probe failed")
	}
	return nil
}
