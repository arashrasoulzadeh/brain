package plugin_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"brain/internal/plugin"
)

func TestWorkerRejectsMalformedRequestsAndResponses(t *testing.T) {
	req := plugin.WorkerRequest{Protocol: 1, ProjectID: "alpha", PluginID: "sample", Version: "1", Operation: "read", Payload: []byte(`{}`)}
	stopped := plugin.NewTypeScriptWorker("true")
	if got := stopped.Handle(req); got.ErrorCode != plugin.ErrTypeScriptWorkerStopped.Error() {
		t.Fatalf("stopped worker: %+v", got)
	}
	missing := plugin.NewTypeScriptWorker(filepath.Join(t.TempDir(), "missing"))
	if err := missing.Start(); err == nil {
		t.Fatal("missing executable started")
	}
	for _, tc := range []struct {
		name, script, expected string
		timeout                time.Duration
	}{
		{"malformed", `IFS= read -r line; printf 'invalid\n'`, plugin.ErrInvalidRPCMessage.Error(), 0},
		{"rpc error", `IFS= read -r line; printf '%s\n' '{"errorCode":"denied"}'`, "denied", time.Second},
		{"protocol", `IFS= read -r line; printf '%s\n' '{"result":{"protocol":2,"success":true}}'`, plugin.ErrInvalidRPCMessage.Error(), time.Second},
		{"large", `IFS= read -r line; head -c 1048577 /dev/zero | tr '\000' x; printf '\n'`, plugin.ErrInvalidRPCMessage.Error(), time.Second},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := plugin.NewTypeScriptWorkerWithOptions(plugin.TypeScriptWorkerOptions{InvocationTimeout: tc.timeout}, "sh", "-c", tc.script)
			if err := w.Start(); err != nil {
				t.Fatal(err)
			}
			defer w.Stop()
			if err := w.Start(); !errors.Is(err, plugin.ErrTypeScriptWorkerStarted) {
				t.Fatalf("duplicate start: %v", err)
			}
			w.SetTrustedStartupDiagnostics(os.Stderr)
			bad := req
			bad.Payload = []byte("{")
			if w.Handle(bad).ErrorCode == "" {
				t.Fatal("malformed JSON sent")
			}
			bad.Payload = []byte(`"` + strings.Repeat("x", 1<<20) + `"`)
			if w.Handle(bad).ErrorCode != plugin.ErrInvalidRPCMessage.Error() {
				t.Fatal("oversized request sent")
			}
			if got := w.Handle(req); got.ErrorCode != tc.expected {
				t.Fatalf("response: %+v, want %s", got, tc.expected)
			}
		})
	}
	w := plugin.NewTypeScriptWorker("sh", "-c", "exit 0")
	if err := w.Start(); err != nil {
		t.Fatal(err)
	}
	if got := w.Handle(req); got.ErrorCode == "" || w.Ready() {
		t.Fatalf("dead subprocess: %+v", got)
	}
	if err := w.Start(); err == nil {
		t.Fatal("single-use subprocess restarted")
	}
}

func TestManifestWorkerStartupRejectionAndExit(t *testing.T) {
	dir := t.TempDir()
	m := coverageManifest()
	if err := os.WriteFile(filepath.Join(dir, "index.ts"), []byte("export function read() {}"), 0600); err != nil {
		t.Fatal(err)
	}
	descriptor := filepath.Join(dir, "descriptor.json")
	for _, tc := range []struct {
		project    string
		manifest   plugin.Manifest
		descriptor string
	}{
		{"", m, descriptor}, {"alpha", plugin.Manifest{}, descriptor}, {"alpha", m, dir},
	} {
		if _, err := plugin.NewManifestTypeScriptWorker("node", "worker", tc.descriptor, dir, tc.project, tc.manifest, plugin.IsolationPolicy{}); err == nil {
			t.Fatal("invalid manifest startup accepted")
		}
	}
	invalid := m
	invalid.SourceFiles = []string{"bad\nname.ts"}
	if _, err := plugin.NewManifestBubblewrapTypeScriptWorker("launcher", "worker", descriptor, dir, "alpha", invalid, plugin.IsolationPolicy{}); err == nil {
		t.Fatal("newline source accepted")
	}
	for _, tc := range []struct {
		script string
		ready  bool
	}{
		{"exit 1", false}, {"printf 'wrong\\n'", false}, {"printf '{\"ready\":true}\\n'", true},
	} {
		path := filepath.Join(t.TempDir(), "launcher")
		if err := os.WriteFile(path, []byte("#!/bin/sh\n"+tc.script+"\n"), 0700); err != nil {
			t.Fatal(err)
		}
		w, err := plugin.NewManifestBubblewrapTypeScriptWorker(path, "worker", descriptor, dir, "alpha", m, plugin.IsolationPolicy{})
		if err != nil {
			t.Fatal(err)
		}
		err = w.Start()
		if (err == nil) != tc.ready {
			t.Fatalf("handshake %q: %v", tc.script, err)
		}
		if tc.ready {
			deadline := time.Now().Add(time.Second)
			for w.Ready() && time.Now().Before(deadline) {
				time.Sleep(time.Millisecond)
			}
			if w.Ready() {
				t.Fatal("exited worker still ready")
			}
			if err := w.Stop(); err != nil {
				t.Fatal(err)
			}
		}
	}
}

func TestLowLevelProtocolAndHandshakeRejections(t *testing.T) {
	var runner plugin.WasmRunner
	if !errors.Is(runner.RunForProject(nil, "alpha", nil), plugin.ErrWasmProjectContextRequired) || !errors.Is(runner.RunForProject(context.Background(), "alpha", nil), plugin.ErrWasmProjectContextRequired) {
		t.Fatal("missing wasm context accepted")
	}
	if err := runner.Run(context.Background(), []byte("not wasm")); !errors.Is(err, plugin.ErrWasmModuleInvalid) {
		t.Fatalf("invalid wasm: %v", err)
	}
	if (plugin.ProtocolError{Code: plugin.ErrorUnauthorized, Message: "denied"}).Valid() {
		t.Fatal("missing request id accepted")
	}
	handler := plugin.HandshakeHandlerForProject(1, "sample", "alpha")
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusMethodNotAllowed {
		t.Fatalf("wrong method: %d", res.Code)
	}
	req = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"protocol":1,"serviceId":"sample","projectId":"alpha","pluginId":"sample","version":"1"} {}`))
	res = httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusUnauthorized {
		t.Fatalf("trailing JSON: %d", res.Code)
	}
}
