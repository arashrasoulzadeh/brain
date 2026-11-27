package plugin_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"brain/internal/plugin"
	"brain/internal/project"
)

type failingCoverageSocket struct{ closed bool }

func (s *failingCoverageSocket) Close() error {
	s.closed = true
	return errors.New("socket close failed")
}

func TestWebSocketAndTransportFailureBoundaries(t *testing.T) {
	for name, fn := range map[string]func(){
		"heartbeat": func() { plugin.NewHeartbeatTracker(0) },
		"queue":     func() { plugin.NewWebSocketSendQueue(0) },
	} {
		t.Run(name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatal("invalid transport limit accepted")
				}
			}()
			fn()
		})
	}
	if plugin.AllowedWebSocketOrigin("", []string{""}) {
		t.Fatal("empty origin accepted")
	}
	if (plugin.ProtocolError{Code: plugin.ErrorCode("unknown"), RequestID: "one", Message: "error"}).Valid() {
		t.Fatal("unknown protocol error code accepted")
	}
	if (plugin.RetryPolicy{MaxAttempts: 2, BaseDelay: time.Second}).Delay(0) != 0 {
		t.Fatal("zero retry attempt delayed")
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "socket")
	if err := os.Mkdir(path, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(path, "child"), []byte("fixture"), 0o600); err != nil {
		t.Fatal(err)
	}
	if listener, err := plugin.ListenUnix(path); err == nil {
		listener.Close()
		t.Fatal("nonempty directory replaced by Unix socket")
	}
	registry := plugin.NewWebSocketRegistry()
	first := &failingCoverageSocket{}
	registry.Register("first", first)
	registry.Unregister("first")
	if registry.Count() != 0 || first.closed {
		t.Fatal("unregister closed or retained a connection")
	}
	second := &failingCoverageSocket{}
	registry.Register("second", second)
	if err := registry.CloseAll(); err == nil || !second.closed || registry.Count() != 0 {
		t.Fatalf("close all failed to release sockets: %v", err)
	}
	handler := plugin.WebSocketEndpoint()
	for _, path := range []string{"/wrong", "/v1/projects/alpha/rooms/room/ws", "/alpha/ws"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		req.Header.Set("X-Project-ID", "alpha")
		req.Header.Set("Connection", "Upgrade")
		req.Header.Set("Upgrade", "websocket")
		req = req.WithContext(project.WithProject(req.Context(), project.Project{ID: "alpha", Active: true}))
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		want := http.StatusNotFound
		if path == "/alpha/ws" {
			want = http.StatusUnauthorized
		}
		if rec.Code != want {
			t.Fatalf("endpoint %s = %d, want %d", path, rec.Code, want)
		}
	}
	request := plugin.WorkerRequest{Protocol: 2, ProjectID: "alpha", PluginID: "plugin", Version: "1", Operation: "read"}
	if err := plugin.ValidateWorkerRequestForProject(request, 1, "alpha"); !errors.Is(err, plugin.ErrInvalidWorkerRequest) {
		t.Fatalf("invalid worker protocol: %v", err)
	}
}
