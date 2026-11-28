package app_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"brain/internal/app"
	"brain/internal/cortex"
	"brain/internal/execution"
)

func TestNeuronMetricsRequiresDedicatedCredential(t *testing.T) {
	server := app.NewServer()
	token := strings.Repeat("x", 32)
	scrape := func(auth string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodGet, "/_system/metrics", nil)
		req.Header.Set("Authorization", auth)
		w := httptest.NewRecorder()
		server.Handler().ServeHTTP(w, req)
		return w
	}
	if w := scrape(""); w.Code != http.StatusNotFound {
		t.Fatal(w.Code, w.Body.String())
	}
	for _, invalid := range []string{"short", strings.Repeat(" ", 32), strings.Repeat("x", 4097)} {
		if err := server.SetMetricsToken(invalid); err == nil {
			t.Fatal("invalid token accepted")
		}
	}
	if err := server.SetMetricsToken(token); err != nil {
		t.Fatal(err)
	}
	for _, auth := range []string{"", "Bearer invalid", "Basic " + token} {
		if w := scrape(auth); w.Code != http.StatusForbidden {
			t.Fatal(w.Code, w.Body.String())
		}
	}
	if w := scrape("Bearer " + token); w.Code != http.StatusServiceUnavailable {
		t.Fatal(w.Code, w.Body.String())
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
	id := "worker\"\\\nline"
	_, err = scheduler.Register(cortex.Registration{ID: id, Slots: 2, Runtimes: []execution.Runtime{{Kind: "typescript", Profile: "isolated"}}, Executor: controlTestExecutor{}})
	if err != nil {
		t.Fatal(err)
	}
	w := scrape("Bearer " + token)
	if w.Code != http.StatusOK || w.Header().Get("Content-Type") != "text/plain; version=0.0.4; charset=utf-8" || w.Header().Get("Cache-Control") != "no-store" {
		t.Fatal(w.Code, w.Header(), w.Body.String())
	}
	for _, want := range []string{
		"brain_cortex_connected_neurons 1\n",
		"brain_cortex_neuron_slots{neuron=\"worker\\\"\\\\\\nline\"} 2\n",
		"brain_cortex_neuron_recent_cpu_samples{",
		"# TYPE brain_cortex_assigned_total counter",
	} {
		if !strings.Contains(w.Body.String(), want) {
			t.Fatalf("missing %q in %s", want, w.Body.String())
		}
	}
	if strings.Contains(w.Body.String(), "brain_cortex_neuron_recent_cpu_seconds{") || strings.Contains(w.Body.String(), token) {
		t.Fatal("unreported usage or token exposed")
	}
	if err := scheduler.SetOperatorDrain(id, true); err != nil {
		t.Fatal(err)
	}
	if w := scrape("Bearer " + token); !strings.Contains(w.Body.String(), "brain_cortex_neuron_available_slots{neuron=\"worker\\\"\\\\\\nline\"} 0\n") {
		t.Fatal(w.Body.String())
	}
	if err := server.SetMetricsToken(""); err != nil {
		t.Fatal(err)
	}
	if w := scrape("Bearer " + token); w.Code != http.StatusNotFound {
		t.Fatal(w.Code, w.Body.String())
	}
}
