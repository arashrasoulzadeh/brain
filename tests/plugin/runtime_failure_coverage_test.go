package plugin_test

import (
	"database/sql"
	"encoding/json"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"brain/internal/plugin"
)

func TestModelCRUDValidationAndCorruptStorage(t *testing.T) {
	path := filepath.Join(t.TempDir(), "models.sqlite")
	storage, err := plugin.OpenSQLiteStorage(path)
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()
	registry := plugin.NewRegistry(1)
	if err := registry.BindSQLiteStorage(storage); err != nil {
		t.Fatal(err)
	}
	m := plugin.Manifest{ID: "record-test", Version: "1", HostAPI: 1, Capabilities: []string{"records"}, Models: map[string]plugin.ModelSpec{"item": {Version: 1, Fields: map[string]string{"name": "string", "count": "integer", "active": "boolean"}, Initial: "draft", Transitions: map[string]plugin.ModelTransition{"publish": {From: "draft", To: "published"}}}}}
	for _, action := range []string{"create", "read", "list", "update", "delete", "transition"} {
		route := &plugin.ModelRoute{Name: "item", Action: action}
		if action == "transition" {
			route.Transition = "publish"
		}
		m.Routes = append(m.Routes, plugin.Route{Operation: action, Function: action, Capability: "records", Helper: "host.model", Model: route})
	}
	for _, err := range []error{registry.Register(m), registry.SetReadiness(m.ID, m.Version, true), registry.Install("alpha", m.ID, m.Version), registry.Grant("alpha", m.ID, m.Version, "records"), registry.Enable("alpha", m.ID, m.Version)} {
		if err != nil {
			t.Fatal(err)
		}
	}
	if err := registry.BindWorker("alpha", m.ID, m.Version, &contractWorker{ready: true}); err != nil {
		t.Fatal(err)
	}
	invoke := func(action, payload string) (plugin.WorkerResponse, error) {
		return registry.Invoke(plugin.WorkerRequest{Protocol: 1, ProjectID: "alpha", PluginID: m.ID, Version: m.Version, Operation: action, Payload: []byte(payload)})
	}
	const fields = `{"name":"first","count":1,"active":true}`
	const create = `{"id":"one","fields":` + fields + `}`
	for _, payload := range []string{"{", `{"id":"bad/id"}`, `{"id":"one","fields":{}}`, `{"id":"one","fields":{"name":true,"count":1,"active":true}}`, `{"id":"one","fields":{"name":"x","count":1.5,"active":true}}`, `{"id":"one","fields":{"name":"x","count":null,"active":true}}`, `{"id":"one","fields":{"name":"x","count":1,"active":"true"}}`, `{"id":"one","fields":{"name":"` + strings.Repeat("x", 1025) + `","count":1,"active":true}}`} {
		if _, err := invoke("create", payload); !errors.Is(err, plugin.ErrInvalidWorkerRequest) {
			t.Fatalf("invalid model create: %v", err)
		}
	}
	created, err := invoke("create", create)
	if err != nil || !created.Success {
		t.Fatalf("create: %v", err)
	}
	if _, err := invoke("create", create); !errors.Is(err, plugin.ErrModelConflict) {
		t.Fatalf("duplicate create: %v", err)
	}
	var record struct{ Revision string }
	if err := json.Unmarshal(created.Payload, &record); err != nil || record.Revision == "" {
		t.Fatalf("revision: %+v %v", record, err)
	}
	if _, err := invoke("list", `{"after":"bad/id"}`); !errors.Is(err, plugin.ErrInvalidWorkerRequest) {
		t.Fatalf("invalid cursor: %v", err)
	}
	listed, err := invoke("list", `{}`)
	if err != nil || !strings.Contains(string(listed.Payload), `"id":"one"`) {
		t.Fatalf("list: %s %v", listed.Payload, err)
	}
	if _, err := invoke("update", `{"id":"missing"}`); !errors.Is(err, plugin.ErrModelConflict) {
		t.Fatalf("missing update: %v", err)
	}
	if _, err := invoke("update", `{"id":"one","revision":"`+record.Revision+`","fields":{}}`); !errors.Is(err, plugin.ErrInvalidWorkerRequest) {
		t.Fatalf("invalid update: %v", err)
	}
	updated, err := invoke("update", `{"id":"one","revision":"`+record.Revision+`","fields":{"name":"second","count":2,"active":false}}`)
	if err != nil || !strings.Contains(string(updated.Payload), "second") {
		t.Fatalf("update: %s %v", updated.Payload, err)
	}
	if err := json.Unmarshal(updated.Payload, &record); err != nil {
		t.Fatal(err)
	}
	published, err := invoke("transition", `{"id":"one","revision":"`+record.Revision+`"}`)
	if err != nil || !strings.Contains(string(published.Payload), "published") {
		t.Fatalf("transition: %s %v", published.Payload, err)
	}
	if err := json.Unmarshal(published.Payload, &record); err != nil {
		t.Fatal(err)
	}
	if _, err := invoke("transition", `{"id":"one","revision":"`+record.Revision+`"}`); !errors.Is(err, plugin.ErrModelConflict) {
		t.Fatalf("invalid transition: %v", err)
	}
	deleted, err := invoke("delete", `{"id":"one","revision":"`+record.Revision+`"}`)
	if err != nil || !strings.Contains(string(deleted.Payload), `"deleted":true`) {
		t.Fatalf("delete: %s %v", deleted.Payload, err)
	}
	if _, err := invoke("read", `{"id":"one"}`); !errors.Is(err, plugin.ErrModelConflict) {
		t.Fatalf("deleted record read: %v", err)
	}
	if _, err := invoke("create", create); err != nil {
		t.Fatal(err)
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	for _, statement := range []string{`UPDATE plugin_model_records SET schema_version=2`, `UPDATE plugin_model_records SET schema_version=1, fields_json='not json'`, `UPDATE plugin_model_records SET fields_json='{}'`} {
		if _, err := db.Exec(statement); err != nil {
			t.Fatal(err)
		}
		for _, operation := range []string{"read", "list"} {
			if _, err := invoke(operation, `{"id":"one"}`); err == nil {
				t.Fatalf("corrupt model accepted: %s", operation)
			}
		}
	}
}

type coverageWorker struct {
	ready       bool
	start, stop error
}

func (w *coverageWorker) Start() error { w.ready = w.start == nil; return w.start }
func (w *coverageWorker) Stop() error {
	if w.stop == nil {
		w.ready = false
	}
	return w.stop
}
func (w *coverageWorker) Ready() bool { return w.ready }

func TestSupervisorFailedRecoveryAndPartialShutdown(t *testing.T) {
	s := plugin.NewSupervisor()
	failure := errors.New("worker unavailable")
	policy := plugin.RecoveryPolicy{MaxRestarts: 1}
	if !errors.Is(s.Start("broken", &coverageWorker{start: failure}), failure) {
		t.Fatal("start failure ignored")
	}
	if !errors.Is(s.StartRecovering("broken", nil, policy), plugin.ErrWorkerRecoveryUnavailable) || !errors.Is(s.StartRecovering("broken", func() plugin.Worker { return nil }, policy), plugin.ErrWorkerRecoveryUnavailable) {
		t.Fatal("invalid factory accepted")
	}
	if !errors.Is(s.StartRecovering("broken", func() plugin.Worker { return &coverageWorker{start: failure} }, policy), failure) {
		t.Fatal("failed initial start ignored")
	}
	if !errors.Is(s.Recover("missing"), plugin.ErrWorkerNotRunning) {
		t.Fatal("missing worker recovered")
	}
	first := &coverageWorker{stop: failure}
	if err := s.Start("alpha:rooms@1", first); err != nil {
		t.Fatal(err)
	}
	if !errors.Is(s.Recover("alpha:rooms@1"), plugin.ErrWorkerRecoveryUnavailable) || !errors.Is(s.StartRecovering("alpha:rooms@1", func() plugin.Worker { return &coverageWorker{} }, policy), plugin.ErrWorkerAlreadyRunning) {
		t.Fatal("invalid recovery accepted")
	}
	second := &coverageWorker{}
	if err := s.Start("beta:rooms@1", second); err != nil {
		t.Fatal(err)
	}
	if !errors.Is(s.StopPluginVersion("", "1"), plugin.ErrWorkerNotRunning) || !errors.Is(s.StopPluginVersion("rooms", "1"), failure) || !s.Ready("alpha:rooms@1") || s.Ready("beta:rooms@1") {
		t.Fatal("partial shutdown did not preserve failed worker")
	}
	first.stop = nil
	if err := s.Stop("alpha:rooms@1"); err != nil {
		t.Fatal(err)
	}
	crashed := &coverageWorker{}
	calls := 0
	factory := func() plugin.Worker {
		calls++
		if calls == 1 {
			return crashed
		}
		if calls == 2 {
			return nil
		}
		return &coverageWorker{start: failure}
	}
	if err := s.StartRecovering("recover", factory, policy); err != nil {
		t.Fatal(err)
	}
	crashed.ready = false
	if !errors.Is(s.Recover("recover"), plugin.ErrWorkerRecoveryUnavailable) || !errors.Is(s.Recover("recover"), failure) || !errors.Is(s.Recover("recover"), plugin.ErrWorkerRestartLimit) {
		t.Fatal("failed recovery did not enforce restart limit")
	}
}
