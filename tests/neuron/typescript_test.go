package neuron_test

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"brain/internal/execution"
	"brain/internal/neuron"
	"brain/internal/plugin"
)

const standaloneManifest = "id: sample\nversion: 1.0.0\nhost_api: 1\nentrypoint: src/index.mts\nsource_files:\n  - src/index.mts\ncapabilities:\n  - call\nroutes:\n  - operation: call\n    function: call\n    capability: call\n"

func standaloneRequest(t *testing.T, runner *neuron.TypeScriptRuntime, files []execution.BundleFile) execution.Request {
	t.Helper()
	bundle, err := execution.BuildBundle(files, execution.MaxBundleBytes)
	if err != nil {
		t.Fatal(err)
	}
	return execution.Request{ID: "test", TraceID: "trace", ProjectID: "alpha", Artifact: execution.Artifact{ID: "sample", Version: "1.0.0", Digest: execution.ArtifactDigest(bundle), Entrypoint: "call"}, Runtime: runner.Capability(), Input: json.RawMessage("null"), Deadline: time.Now().Add(5 * time.Second), Limits: execution.Limits{CPUTime: neuron.TypeScriptCPU, MemoryBytes: neuron.TypeScriptMemory, InputBytes: 4096, OutputBytes: 4096}, Auth: execution.Authorization{ProjectID: "alpha", OriginalCaller: "user:42", Caller: "plugin:sample", Capability: "call"}}
}

func standaloneFiles(manifest, source string) []execution.BundleFile {
	return []execution.BundleFile{{Path: "plugin.yaml", Data: []byte(manifest)}, {Path: "src/index.mts", Data: []byte(source)}}
}

func TestStandaloneTypeScriptRejectsUntrustedArtifactsBeforeLaunch(t *testing.T) {
	root := t.TempDir()
	if _, err := neuron.NewTypeScriptRuntime(filepath.Join(root, "missing"), true); err == nil {
		t.Fatal("missing trusted runtime directory accepted")
	}
	runner, err := neuron.NewTypeScriptRuntime(root, true)
	if err != nil {
		t.Fatal(err)
	}
	files := standaloneFiles(standaloneManifest, "export function call() { return 42; }")
	valid := standaloneRequest(t, runner, files)
	for _, tc := range []struct {
		name   string
		mutate func(*execution.Request)
		want   error
	}{
		{"wrong project", func(r *execution.Request) { r.Auth.ProjectID = "beta" }, execution.ErrInvalidRequest},
		{"wrong runtime", func(r *execution.Request) { r.Runtime.Profile = "unisolated" }, plugin.ErrInvalidIsolationPolicy},
		{"wrong CPU limit", func(r *execution.Request) { r.Limits.CPUTime /= 2 }, plugin.ErrInvalidIsolationPolicy},
		{"wrong memory limit", func(r *execution.Request) { r.Limits.MemoryBytes /= 2 }, plugin.ErrInvalidIsolationPolicy},
		{"too much input", func(r *execution.Request) { r.Limits.InputBytes = 2 << 20 }, plugin.ErrInvalidIsolationPolicy},
		{"too much output", func(r *execution.Request) { r.Limits.OutputBytes = 2 << 20 }, plugin.ErrInvalidIsolationPolicy},
		{"digest substitution", func(r *execution.Request) { r.Artifact.Digest = execution.ArtifactDigest([]byte("other")) }, execution.ErrArtifactMismatch},
		{"wrong plugin", func(r *execution.Request) { r.Artifact.ID = "other" }, neuron.ErrRuntimeArtifact},
		{"wrong version", func(r *execution.Request) { r.Artifact.Version = "2.0.0" }, neuron.ErrRuntimeArtifact},
		{"wrong entrypoint", func(r *execution.Request) { r.Artifact.Entrypoint = "other" }, neuron.ErrRuntimeArtifact},
		{"wrong capability", func(r *execution.Request) { r.Auth.Capability = "other" }, neuron.ErrRuntimeArtifact},
		{"expired", func(r *execution.Request) { r.Deadline = time.Now().Add(-time.Second) }, context.DeadlineExceeded},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := valid
			tc.mutate(&r)
			if _, err := runner.Run(context.Background(), r, files); !errors.Is(err, tc.want) {
				t.Fatal(err)
			}
		})
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := runner.Run(ctx, valid, files); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if _, err := runner.Run(context.Background(), valid, []execution.BundleFile{{Path: "../escape"}}); !errors.Is(err, execution.ErrInvalidBundle) {
		t.Fatal(err)
	}
	for _, manifest := range []string{
		"not a manifest",
		strings.Replace(standaloneManifest, "host_api: 1", "host_api: 2", 1),
		standaloneManifest + "dependencies:\n  - other@1.0.0\n",
		standaloneManifest + "imports:\n  - node:fs\n",
		strings.Replace(standaloneManifest, "    capability: call", "    capability: call\n    helper: sdk.dispatch", 1),
	} {
		bad := standaloneFiles(manifest, "export function call() { return 42; }")
		if _, err := runner.Run(context.Background(), standaloneRequest(t, runner, bad), bad); !errors.Is(err, neuron.ErrRuntimeArtifact) {
			t.Fatal(manifest, err)
		}
	}
	undeclared := standaloneFiles(standaloneManifest, "import fs from 'node:fs';\nexport function call() { return fs; }")
	if _, err := runner.Run(context.Background(), standaloneRequest(t, runner, undeclared), undeclared); !errors.Is(err, plugin.ErrInvalidManifest) {
		t.Fatal(err)
	}
}

func TestStandaloneTypeScriptLaunchFailureAndContext(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Linux launcher protocol fixture; real macOS sandbox is tested separately")
	}
	root := t.TempDir()
	scripts := filepath.Join(root, "scripts")
	if err := os.Mkdir(scripts, 0o700); err != nil {
		t.Fatal(err)
	}
	runner, err := neuron.NewTypeScriptRuntime(root, false)
	if err != nil {
		t.Fatal(err)
	}
	files := standaloneFiles(standaloneManifest, "export function call() { return 42; }")
	request := standaloneRequest(t, runner, files)
	if _, err := runner.Run(context.Background(), request, files); err == nil {
		t.Fatal("missing sandbox launcher accepted")
	}
	launcher := filepath.Join(scripts, "run-sandboxed-typescript-worker.sh")
	// Trusted fixture launcher only: test the Go JSON-RPC adapter and inspect the
	// context that it passes across the process boundary. Real OS isolation
	// and Node execution are independently exercised by packaged startup tests.
	logPath := filepath.Join(root, "request.json")
	script := "#!/bin/sh\nprintf '%s\\n' '{\"ready\":true}'\nIFS= read -r line || exit 1\nprintf '%s' \"$line\" > '" + logPath + "'\nprintf '%s\\n' '{\"jsonrpc\":\"2.0\",\"id\":\"call\",\"result\":{\"protocol\":1,\"success\":true,\"payload\":42}}'\nwhile IFS= read -r line; do :; done\n"
	if err := os.WriteFile(launcher, []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	for _, principals := range [][2]string{{"user:42", "plugin:sample"}, {"operator:root", "host"}, {"public", "unknown"}} {
		request.Auth.OriginalCaller, request.Auth.Caller = principals[0], principals[1]
		request.Deadline = time.Now().Add(5 * time.Second)
		result, err := runner.Run(context.Background(), request, files)
		if err != nil || result.Status != execution.Succeeded || string(result.Output) != "42" {
			t.Fatal(result, err)
		}
		data, err := os.ReadFile(logPath)
		if err != nil || !json.Valid(data) {
			t.Fatal(string(data), err)
		}
		var payload struct {
			Params struct {
				HostContext plugin.InvocationContext
			}
		}
		if err = json.Unmarshal(data, &payload); err != nil {
			t.Fatal(err)
		}
		auth := payload.Params.HostContext.Auth
		switch principals[0] {
		case "user:42":
			if auth.Actor.Kind != "user" || auth.Actor.ID != "42" || !auth.Actor.Authenticated || auth.Caller.Kind != "plugin" {
				t.Fatal(auth)
			}
		case "operator:root":
			if auth.Actor.Kind != "operator" || auth.Caller.Kind != "host" {
				t.Fatal(auth)
			}
		default:
			if auth.Actor.Kind != "public" || auth.Caller.Authenticated {
				t.Fatal(auth)
			}
		}
	}
	if err = os.WriteFile(launcher, []byte("#!/bin/sh\nprintf '%s\\n' '{\"ready\":true}'\nIFS= read -r line\nprintf '%s\\n' '{\"jsonrpc\":\"2.0\",\"id\":\"call\",\"result\":{\"protocol\":1,\"success\":false,\"errorCode\":\"private\"}}'\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	request.Deadline = time.Now().Add(5 * time.Second)
	result, err := runner.Run(context.Background(), request, files)
	if err != nil || result.Status != execution.Failed || result.Error == nil || result.Error.Code != "runtime_error" || result.Error.Message != "" {
		t.Fatal(result, err)
	}
	if err = os.WriteFile(launcher, []byte("#!/bin/sh\nwhile IFS= read -r line; do :; done\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	if _, err := runner.Run(ctx, request, files); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal(err)
	}
}

func TestStandaloneTypeScriptProbeFailsWithoutSandbox(t *testing.T) {
	runner, err := neuron.NewTypeScriptRuntime(t.TempDir(), true)
	if err != nil {
		t.Fatal(err)
	}
	if err := runner.Probe(context.Background()); err == nil {
		t.Fatal("missing runtime advertised as ready")
	}
}
