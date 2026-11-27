package plugin_test

import (
	"errors"
	"reflect"
	"testing"

	"brain/internal/plugin"
)

type dependencyEnableWorker struct {
	ready bool
	fail error
}

func (w *dependencyEnableWorker) Start() error {
	if w.fail != nil {
		return w.fail
	}
	w.ready = true
	return nil
}
func (w *dependencyEnableWorker) Stop() error { w.ready = false; return nil }
func (w *dependencyEnableWorker) Ready() bool { return w.ready }
func (w *dependencyEnableWorker) Handle(req plugin.WorkerRequest) plugin.WorkerResponse {
	return plugin.WorkerResponse{Protocol: req.Protocol, Success: true}
}

func TestEnableExactDependenciesRequiresConsentAndRollsBack(t *testing.T) {
	r := plugin.NewRegistry(1)
	manifests := []plugin.Manifest{
		{ID: "leaf", Version: "3", HostAPI: 1, Capabilities: []string{"read"}},
		{ID: "child", Version: "2", HostAPI: 1, Capabilities: []string{"read"}, Dependencies: []string{"leaf@3"}},
		{ID: "parent", Version: "1", HostAPI: 1, Capabilities: []string{"read"}, Dependencies: []string{"child@2"}, OptionalDependencies: []string{"optional@1"}},
	}
	for _, m := range manifests {
		for _, err := range []error{r.Register(m), r.Install("one", m.ID, m.Version), r.SetReadiness(m.ID, m.Version, true), r.Grant("one", m.ID, m.Version, "read")} {
			if err != nil { t.Fatal(err) }
		}
	}
	var order []string
	var workers []*dependencyEnableWorker
	failure := errors.New("parent startup failed")
	fail := true
	r.BindWorkerFactory(func(project, id, version string) (plugin.ContractWorker, error) {
		if project != "one" { t.Fatalf("wrong project %s", project) }
		order = append(order, id+"@"+version)
		w := &dependencyEnableWorker{}
		if id == "parent" && fail { w.fail = failure }
		workers = append(workers, w)
		return w, nil
	})
	plan := []string{"leaf@3", "child@2"}
	for _, invalid := range [][]string{nil, {"child@2"}, {"child@2", "leaf@3"}, {"leaf@3", "child@2", "optional@1"}} {
		if err := r.EnableWithDependencies("one", "parent", "1", invalid); err == nil {
			t.Fatalf("invalid consent accepted: %v", invalid)
		}
	}
	if len(order) != 0 { t.Fatal("started workers without consent") }
	if err := r.EnableWithDependencies("two", "parent", "1", plan); err == nil {
		t.Fatal("cross-project activation accepted")
	}
	if err := r.EnableWithDependencies("one", "parent", "1", plan); !errors.Is(err, failure) {
		t.Fatalf("expected startup failure, got %v", err)
	}
	for _, m := range manifests {
		status, err := r.Status("one", m.ID, m.Version)
		if err != nil || status.Enabled { t.Fatalf("partial activation: %+v %v", status, err) }
	}
	for _, worker := range workers {
		if worker.Ready() { t.Fatal("partial worker left running") }
	}
	order = nil
	fail = false
	if err := r.EnableWithDependencies("one", "parent", "1", plan); err != nil { t.Fatal(err) }
	if !reflect.DeepEqual(order, []string{"leaf@3", "child@2", "parent@1"}) {
		t.Fatalf("wrong startup order %v", order)
	}
	for _, m := range manifests {
		status, err := r.Status("one", m.ID, m.Version)
		if err != nil || !status.Enabled || !status.WorkerReady { t.Fatalf("not enabled: %+v %v", status, err) }
	}
	if err := r.EnableWithDependencies("one", "parent", "1", plan); err == nil {
		t.Fatal("stale confirmation accepted")
	}
}
