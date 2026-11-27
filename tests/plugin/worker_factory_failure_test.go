package plugin_test

import (
	"errors"
	"testing"

	"brain/internal/plugin"
)

func TestWorkerFactoryFailureDoesNotEnableInstallation(t *testing.T) {
	failure := errors.New("worker unavailable")
	for _, fault := range []string{"factory error", "nil worker", "start error", "persistence", "unhealthy"} {
		t.Run(fault, func(t *testing.T) {
			r := plugin.NewRegistry(1)
			m := plugin.Manifest{ID: "sample", Version: "1", HostAPI: 1, Capabilities: []string{"read"}}
			store := &coverageStateStore{}
			for _, err := range []error{r.Register(m), r.BindRegistryStorage(store), r.Install("alpha", m.ID, m.Version), r.SetReadiness(m.ID, m.Version, true), r.Grant("alpha", m.ID, m.Version, "read")} {
				if err != nil {
					t.Fatal(err)
				}
			}
			worker := &lifecycleCoverageWorker{}
			calls := 0
			r.BindWorkerFactory(func(project, id, version string) (plugin.ContractWorker, error) {
				calls++
				switch fault {
				case "factory error":
					return nil, failure
				case "nil worker":
					return nil, nil
				case "start error":
					worker.start = failure
				case "persistence":
					store.fail = true
				}
				return worker, nil
			})
			want := failure
			switch fault {
			case "nil worker":
				want = plugin.ErrWorkerUnavailable
			case "persistence", "unhealthy":
				want = plugin.ErrLifecycleStorage
			}
			if fault == "unhealthy" {
				store.fail = true
				if err := r.FlushLifecycleState(); !errors.Is(err, plugin.ErrLifecycleStorage) {
					t.Fatal(err)
				}
			}
			if err := r.Enable("alpha", m.ID, m.Version); !errors.Is(err, want) {
				t.Fatalf("enable: %v, want %v", err, want)
			}
			status, err := r.Status("alpha", m.ID, m.Version)
			if err != nil || status.Enabled || status.WorkerReady || worker.ready {
				t.Fatalf("failed worker exposed: %+v %v", status, err)
			}
			if fault == "unhealthy" && calls != 0 {
				t.Fatal("unhealthy registry started worker")
			}
		})
	}
}

func TestWorkerRestartRejectsNilFactoryResult(t *testing.T) {
	r := plugin.NewRegistry(1)
	m := plugin.Manifest{ID: "sample", Version: "1", HostAPI: 1, Capabilities: []string{"read"}}
	if err := r.Register(m); err != nil {
		t.Fatal(err)
	}
	installAccessTestPlugin(t, r, "alpha", m, &contractWorker{ready: true})
	r.UnbindWorker("alpha", m.ID, m.Version)
	r.BindWorkerFactory(func(string, string, string) (plugin.ContractWorker, error) { return nil, nil })
	if err := r.StartWorker("alpha", m.ID, m.Version); !errors.Is(err, plugin.ErrWorkerUnavailable) {
		t.Fatalf("nil factory result: %v", err)
	}
	status, err := r.Status("alpha", m.ID, m.Version)
	if err != nil || status.WorkerReady {
		t.Fatalf("nil worker visible: %+v %v", status, err)
	}
}
