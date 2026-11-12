package plugin_test

import (
	"errors"
	"testing"

	"brain/internal/plugin"
)

type contractWorker struct{ ready bool }

func (w *contractWorker) Start() error { w.ready = true; return nil }
func (w *contractWorker) Stop() error  { w.ready = false; return nil }
func (w *contractWorker) Ready() bool  { return w.ready }
func (w *contractWorker) Handle(request plugin.WorkerRequest) plugin.WorkerResponse {
	return plugin.WorkerResponse{Protocol: request.Protocol, Success: true, Payload: request.Payload}
}

func TestBundledWorkerRunsSharedContractFixture(t *testing.T) {
	w := &contractWorker{ready: true}
	request := plugin.WorkerRequest{Protocol: 1, ProjectID: "dev", PluginID: "arena", Version: "1", Operation: "fixture", Payload: []byte(`{"ok":true}`)}
	response, err := plugin.Invoke(w, request, 1)
	if err != nil || !response.Success || string(response.Payload) != `{"ok":true}` {
		t.Fatalf("Invoke() = %#v, %v", response, err)
	}
	w.ready = false
	if !errors.Is(func() error { _, err := plugin.Invoke(w, request, 1); return err }(), plugin.ErrWorkerNotReady) {
		t.Fatal("unready worker accepted invocation")
	}
}
