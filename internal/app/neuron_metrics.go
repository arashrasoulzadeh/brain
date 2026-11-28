package app

import (
	"crypto/sha256"
	"crypto/subtle"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"brain/internal/cortex"
)

// SetMetricsToken configures a dedicated read-only scrape credential before
// serving HTTP. An empty token disables the endpoint.
func (s *Server) SetMetricsToken(token string) error {
	if token != "" && (len(token) < 32 || len(token) > 4096 || strings.ContainsAny(token, " \t\r\n")) {
		return errors.New("metrics token must contain 32–4096 non-whitespace bytes")
	}
	s.metricsTokenHash = sha256.Sum256([]byte(token))
	s.metricsEnabled = token != ""
	return nil
}

func (s *Server) neuronMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	if !s.metricsEnabled {
		http.NotFound(w, r)
		return
	}
	token, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
	sum := sha256.Sum256([]byte(token))
	if !ok || subtle.ConstantTimeCompare(sum[:], s.metricsTokenHash[:]) != 1 {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	if s.neuronScheduler == nil {
		http.Error(w, "Neuron scheduler unavailable", http.StatusServiceUnavailable)
		return
	}
	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	_, _ = w.Write([]byte(formatNeuronMetrics(s.neuronScheduler.Snapshot())))
}

func formatNeuronMetrics(snapshot cortex.Snapshot) string {
	var out strings.Builder
	metric := func(name, kind, help string) string {
		name = "brain_cortex_" + name
		fmt.Fprintf(&out, "# HELP %s %s\n# TYPE %s %s\n", name, help, name, kind)
		return name
	}
	sample := func(name, labels string, value any) {
		fmt.Fprintf(&out, "%s%s %v\n", name, labels, value)
	}
	label := func(id string) string {
		return "{neuron=\"" + strings.NewReplacer("\\", "\\\\", "\n", "\\n", "\"", "\\\"").Replace(id) + "\"}"
	}
	sample(metric("connected_neurons", "gauge", "Authenticated connected Neurons."), "", len(snapshot.Neurons))
	sample(metric("running_executions", "gauge", "Active Cortex assignments."), "", snapshot.Running)
	sample(metric("queued_executions", "gauge", "Pending Cortex assignments."), "", snapshot.Queued)
	sample(metric("assigned_total", "counter", "Assignments during this Brain process lifetime."), "", snapshot.Executions.Assigned)
	name := metric("execution_outcomes_total", "counter", "Fenced terminal assignment outcomes during this Brain process lifetime.")
	sample(name, "{outcome=\"succeeded\"}", snapshot.Executions.Succeeded)
	sample(name, "{outcome=\"failed\"}", snapshot.Executions.Failed)
	sample(name, "{outcome=\"cancelled\"}", snapshot.Executions.Cancelled)
	sample(name, "{outcome=\"lost\"}", snapshot.Executions.Lost)
	for _, field := range []struct {
		name, help string
		value      func(cortex.NeuronSnapshot) int
	}{
		{"neuron_slots", "Advertised execution slots.", func(n cortex.NeuronSnapshot) int { return n.Slots }},
		{"neuron_running", "Occupied execution slots.", func(n cortex.NeuronSnapshot) int { return n.Running }},
		{"neuron_available_slots", "Available non-draining slots.", func(n cortex.NeuronSnapshot) int {
			if n.Draining {
				return 0
			}
			return max(0, n.Slots-n.Running)
		}},
	} {
		name := metric(field.name, "gauge", field.help)
		for _, n := range snapshot.Neurons {
			sample(name, label(n.ID), field.value(n))
		}
	}
	// Report gauges describe the last 128 cluster-wide outcomes, not counters.
	report := snapshot.PerformanceReport()
	for _, field := range []struct {
		name, help string
		value      func(cortex.NeuronPerformance) (any, bool)
	}{
		{"neuron_recent_completed", "Completed assignments in retained recent history.", func(p cortex.NeuronPerformance) (any, bool) { return p.Completed, true }},
		{"neuron_recent_cpu_samples", "Executions reporting positive CPU time in retained history.", func(p cortex.NeuronPerformance) (any, bool) { return p.CPUSamples, true }},
		{"neuron_recent_memory_samples", "Executions reporting positive peak RSS in retained history.", func(p cortex.NeuronPerformance) (any, bool) { return p.MemorySamples, true }},
		{"neuron_recent_cpu_seconds", "Sum of measured execution CPU seconds in retained history.", func(p cortex.NeuronPerformance) (any, bool) { return p.TotalCPUSeconds, p.CPUSamples > 0 }},
		{"neuron_recent_average_cpu_seconds", "Mean measured execution CPU seconds in retained history.", func(p cortex.NeuronPerformance) (any, bool) { return p.AverageCPUSeconds, p.CPUSamples > 0 }},
		{"neuron_recent_average_peak_memory_bytes", "Mean measured execution peak RSS in retained history.", func(p cortex.NeuronPerformance) (any, bool) { return p.AveragePeakMemoryBytes, p.MemorySamples > 0 }},
		{"neuron_recent_max_peak_memory_bytes", "Maximum measured execution peak RSS in retained history.", func(p cortex.NeuronPerformance) (any, bool) { return p.MaxPeakMemoryBytes, p.MemorySamples > 0 }},
		{"neuron_recent_average_wait_seconds", "Mean queue wait in retained history.", func(p cortex.NeuronPerformance) (any, bool) { return p.AverageWaitMillis / 1000, p.Completed > 0 }},
		{"neuron_recent_average_duration_seconds", "Mean assignment duration in retained history.", func(p cortex.NeuronPerformance) (any, bool) { return p.AverageDurationMillis / 1000, p.Completed > 0 }},
		{"neuron_recent_p95_duration_seconds", "Nearest-rank p95 assignment duration in retained history.", func(p cortex.NeuronPerformance) (any, bool) { return p.P95DurationMillis / 1000, p.Completed > 0 }},
	} {
		name := metric(field.name, "gauge", field.help)
		for _, p := range report {
			if value, ok := field.value(p); ok {
				sample(name, label(p.ID), value)
			}
		}
	}
	name = metric("neuron_recent_outcomes", "gauge", "Terminal outcomes by Neuron in retained history.")
	for _, p := range report {
		for _, pair := range []struct {
			outcome string
			count   int
		}{{"succeeded", p.Succeeded}, {"failed", p.Failed}, {"cancelled", p.Cancelled}, {"lost", p.Lost}} {
			sample(name, strings.TrimSuffix(label(p.ID), "}") + ",outcome=\"" + pair.outcome + "\"}", pair.count)
		}
	}
	return out.String()
}
