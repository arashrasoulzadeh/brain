package app_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"brain/internal/admin"
	"brain/internal/app"
	"brain/internal/cortex"
	"brain/internal/execution"
	"brain/internal/operator"
	"brain/internal/plugin"
	"brain/internal/project"
)

func TestNeuronStatusRequiresOperatorAndReportsRealScheduler(t *testing.T) {
	operators := &operator.Store{}
	digest := func(value string) string { return value }
	if _, err := operators.Bootstrap("admin", "secret", digest); err != nil {
		t.Fatal(err)
	}
	if err := operators.Authenticate("admin", "secret", digest); err != nil {
		t.Fatal(err)
	}
	token, err := operators.IssueSession(func() string { return "test-neuron-status-token" })
	if err != nil {
		t.Fatal(err)
	}
	service := admin.NewLifecycleServiceWithProjects(operators, plugin.NewRegistry(1), project.NewStore())
	server := app.NewServerWithOperatorAdmin(service, operators, digest)
	query := func(auth string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodGet, "/_system/admin/neurons", nil)
		req.Header.Set("Authorization", auth)
		w := httptest.NewRecorder()
		server.Handler().ServeHTTP(w, req)
		return w
	}
	for _, auth := range []string{"", "Bearer invalid", "Basic " + token} {
		if w := query(auth); w.Code != http.StatusForbidden {
			t.Fatal(w.Code, w.Body.String())
		}
	}
	type status struct {
		Configured bool
		Snapshot   cortex.Snapshot
	}
	read := func() status {
		t.Helper()
		w := query("Bearer " + token)
		var result status
		if w.Code != http.StatusOK || w.Header().Get("Cache-Control") != "no-store" {
			t.Fatal(w.Code, w.Body.String())
		}
		if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
			t.Fatal(err)
		}
		return result
	}
	if read().Configured {
		t.Fatal("unconfigured scheduler advertised as live")
	}
	catalog, err := execution.NewCatalog(1, 4096)
	if err != nil {
		t.Fatal(err)
	}
	scheduler, err := cortex.New(catalog, cortex.Config{MaxNeurons: 4, MaxProjectRunning: 4, HeartbeatTTL: time.Minute})
	if err != nil {
		t.Fatal(err)
	}
	server.SetNeuronScheduler(scheduler)
	if got := read(); !got.Configured || len(got.Snapshot.Neurons) != 0 || got.Snapshot.Running != 0 {
		t.Fatal(got)
	}
}

type controlTestExecutor struct{}

func (controlTestExecutor) Execute(context.Context, execution.Request) (execution.Result, error) {
	return execution.Result{}, nil
}

type controlTestManager struct {
	id, action string
}

func (m *controlTestManager) ManagedNeurons() []app.ManagedNeuron {
	return []app.ManagedNeuron{{ID: "managed", Running: false}}
}

func (m *controlTestManager) ControlNeuron(_ context.Context, id, action string) error {
	m.id, m.action = id, action
	return nil
}

func TestNeuronControlsRequireOperatorAndConfirmation(t *testing.T) {
	operators := &operator.Store{}
	digest := func(value string) string { return value }
	if _, err := operators.Bootstrap("admin", "secret", digest); err != nil {
		t.Fatal(err)
	}
	if err := operators.Authenticate("admin", "secret", digest); err != nil {
		t.Fatal(err)
	}
	token, err := operators.IssueSession(func() string { return "control-test-token" })
	if err != nil {
		t.Fatal(err)
	}
	service := admin.NewLifecycleServiceWithProjects(operators, plugin.NewRegistry(1), project.NewStore())
	server := app.NewServerWithOperatorAdmin(service, operators, digest)
	catalog, err := execution.NewCatalog(1, 4096)
	if err != nil {
		t.Fatal(err)
	}
	scheduler, err := cortex.New(catalog, cortex.Config{MaxNeurons: 4, MaxProjectRunning: 4, HeartbeatTTL: time.Minute})
	if err != nil {
		t.Fatal(err)
	}
	server.SetNeuronScheduler(scheduler)
	manager := &controlTestManager{}
	server.SetManagedNeuronController(manager)
	_, err = scheduler.Register(cortex.Registration{ID: "remote", Slots: 1, Runtimes: []execution.Runtime{{Kind: "typescript", Profile: "isolated"}}, Executor: controlTestExecutor{}})
	if err != nil {
		t.Fatal(err)
	}
	post := func(auth, body string, status int) {
		t.Helper()
		req := httptest.NewRequest(http.MethodPost, "/_system/admin/neurons", strings.NewReader(body))
		req.Header.Set("Authorization", auth)
		w := httptest.NewRecorder()
		server.Handler().ServeHTTP(w, req)
		if w.Code != status {
			t.Fatalf("status %d, want %d: %s", w.Code, status, w.Body.String())
		}
	}
	auth := "Bearer " + token
	post("", `{"id":"remote","action":"drain","confirmed":true}`, http.StatusForbidden)
	post(auth, `{"id":"remote","action":"drain"}`, http.StatusBadRequest)
	post(auth, `{"id":"remote","action":"drain","confirmed":true,"command":"bad"}`, http.StatusBadRequest)
	post(auth, `{"id":"missing","action":"drain","confirmed":true}`, http.StatusNotFound)
	post(auth, `{"id":"remote","action":"stop","confirmed":true}`, http.StatusConflict)
	post(auth, `{"id":"remote","action":"drain","confirmed":true}`, http.StatusOK)
	if n, ok := scheduler.Neuron("remote"); !ok || !n.OperatorDraining {
		t.Fatal(n)
	}
	post(auth, `{"id":"remote","action":"resume","confirmed":true}`, http.StatusOK)
	if n, ok := scheduler.Neuron("remote"); !ok || n.Draining {
		t.Fatal(n)
	}
	post(auth, `{"id":"managed","action":"start","confirmed":true}`, http.StatusOK)
	if manager.id != "managed" || manager.action != "start" {
		t.Fatal(manager)
	}
}
