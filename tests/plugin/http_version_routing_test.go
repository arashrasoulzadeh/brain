package plugin_test

import (
	"errors"
	"testing"

	"brain/internal/plugin"
)

func TestPromotedProjectHTTPVersionAndRollback(t *testing.T) {
	registry := plugin.NewRegistry(1)
	for _, version := range []string{"1.0.0", "2.0.0"} {
		m := plugin.Manifest{ID: "rooms", Version: version, HostAPI: 1, Mount: "/rooms", Capabilities: []string{"rooms.read"},
			Routes: []plugin.Route{{Method: "GET", Path: "/", Auth: "operator", Operation: "rooms.list", Function: "listRooms", Capability: "rooms.read"}}}
		if err := registry.Register(m); err != nil {
			t.Fatal(err)
		}
		if err := registry.Install("alpha", m.ID, version); err != nil {
			t.Fatal(err)
		}
		if err := registry.Grant("alpha", m.ID, version, "rooms.read"); err != nil {
			t.Fatal(err)
		}
		if err := registry.SetReadiness(m.ID, version, true); err != nil {
			t.Fatal(err)
		}
		if err := registry.Enable("alpha", m.ID, version); err != nil {
			t.Fatal(err)
		}
	}
	if _, ok := registry.MatchHTTPRoute("alpha", "GET", "/rooms"); ok {
		t.Fatal("unpromoted overlapping versions exposed")
	}
	if err := registry.Promote("alpha", "rooms", "1.0.0"); err != nil {
		t.Fatal(err)
	}
	if route, ok := registry.MatchHTTPRoute("alpha", "GET", "/rooms"); !ok || route.Version != "1.0.0" {
		t.Fatalf("promoted version 1: %#v, %v", route, ok)
	}
	if err := registry.Promote("alpha", "rooms", "2.0.0"); err != nil {
		t.Fatal(err)
	}
	if route, ok := registry.MatchHTTPRoute("alpha", "GET", "/rooms"); !ok || route.Version != "2.0.0" {
		t.Fatalf("promoted version 2: %#v, %v", route, ok)
	}
	if err := registry.Rollback("alpha", "rooms", "1.0.0"); err != nil {
		t.Fatal(err)
	}
	if route, ok := registry.MatchHTTPRoute("alpha", "GET", "/rooms"); !ok || route.Version != "1.0.0" {
		t.Fatalf("rollback version 1: %#v, %v", route, ok)
	}
	if _, ok := registry.MatchHTTPRoute("beta", "GET", "/rooms"); ok {
		t.Fatal("version leaked across projects")
	}
	selected, ok := registry.MatchHTTPRoute("alpha", "GET", "/rooms")
	if !ok {
		t.Fatal("rolled back version was not selected")
	}
	worker := &httpPromotionWorker{onHandle: func() {
		if err := registry.Promote("alpha", "rooms", "2.0.0"); err != nil {
			t.Error(err)
		}
	}}
	if err := registry.BindWorker("alpha", "rooms", "1.0.0", worker); err != nil {
		t.Fatal(err)
	}
	if _, err := registry.InvokeHTTPRoute("beta", selected, []byte(`{}`), "admin"); !errors.Is(err, plugin.ErrNotRegistered) {
		t.Fatalf("cross-project selection: %v", err)
	}
	if _, err := registry.InvokeHTTPRoute("alpha", selected, []byte(`{}`), "admin"); !errors.Is(err, plugin.ErrNotRegistered) {
		t.Fatalf("in-flight promotion did not reject stale response: %v", err)
	}
	if _, err := registry.InvokeHTTPRoute("alpha", selected, []byte(`{}`), "admin"); !errors.Is(err, plugin.ErrNotRegistered) || worker.calls != 1 {
		t.Fatalf("stale selection was admitted: %v; calls=%d", err, worker.calls)
	}
}

type httpPromotionWorker struct {
	onHandle func()
	calls    int
}

func (w *httpPromotionWorker) Start() error { return nil }
func (w *httpPromotionWorker) Stop() error  { return nil }
func (w *httpPromotionWorker) Ready() bool  { return true }
func (w *httpPromotionWorker) Handle(req plugin.WorkerRequest) plugin.WorkerResponse {
	w.calls++
	w.onHandle()
	return plugin.WorkerResponse{Protocol: req.Protocol, Success: true, Payload: []byte(`{}`)}
}

