package app_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
