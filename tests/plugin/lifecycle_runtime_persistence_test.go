package plugin_test

import (
	"errors"
	"testing"

	"brain/internal/plugin"
)

type lifecycleRuntimeWorker struct {
	ready bool
}

func (w *lifecycleRuntimeWorker) Start() error { w.ready = true; return nil }
func (w *lifecycleRuntimeWorker) Stop() error  { w.ready = false; return nil }
func (w *lifecycleRuntimeWorker) Ready() bool  { return w.ready }
func (w *lifecycleRuntimeWorker) Handle(plugin.WorkerRequest) plugin.WorkerResponse {
	return plugin.WorkerResponse{Protocol: 1, Success: true}
}

func TestLifecycleRestrictionStopsBoundWorkerAndSurvivesRestart(t *testing.T) {
	for _, operation := range []string{"disable", "stop", "remove", "uninstall", "revoke"} {
		t.Run(operation, func(t *testing.T) {
			storage, err := plugin.OpenSQLiteStorage(t.TempDir() + "/lifecycle.sqlite")
			if err != nil {
				t.Fatal(err)
			}
			defer storage.Close()
			m := plugin.Manifest{ID: "sample", Version: "1.0.0", HostAPI: 1, Capabilities: []string{"read"}, Routes: []plugin.Route{{Operation: "read", Function: "read", Capability: "read"}}}
			r := plugin.NewRegistry(1)
			if err := r.Register(m); err != nil {
				t.Fatal(err)
			}
			if err := r.BindSQLiteStorage(storage); err != nil {
				t.Fatal(err)
			}
			if err := r.Install("project", m.ID, m.Version); err != nil {
				t.Fatal(err)
			}
			if err := r.SetReadiness(m.ID, m.Version, true); err != nil {
				t.Fatal(err)
			}
			if err := r.Grant("project", m.ID, m.Version, "read"); err != nil {
				t.Fatal(err)
			}
			worker := &lifecycleRuntimeWorker{}
			r.BindWorkerFactory(func(_, _, _ string) (plugin.ContractWorker, error) { return worker, nil })
			if err := r.Enable("project", m.ID, m.Version); err != nil {
				t.Fatal(err)
			}
			switch operation {
			case "disable":
				preview, err := r.PreviewDisable("project", m.ID, m.Version)
				if err != nil {
					t.Fatal(err)
				}
				if err := r.ApplyDisablePreview(preview); err != nil {
					t.Fatal(err)
				}
			case "stop":
				if err := r.StopImmediately("project", m.ID, m.Version); err != nil {
					t.Fatal(err)
				}
			case "remove":
				if err := r.Remove("project", m.ID, m.Version); err != nil {
					t.Fatal(err)
				}
			case "uninstall":
				if err := r.UninstallProject("project"); err != nil {
					t.Fatal(err)
				}
			case "revoke":
				if err := r.Revoke("project", m.ID, m.Version, "read"); err != nil {
					t.Fatal(err)
				}
			}
			if worker.Ready() {
				t.Fatal("restricted worker remains running")
			}
			restarted := plugin.NewRegistry(1)
			if err := restarted.Register(m); err != nil {
				t.Fatal(err)
			}
			if err := restarted.BindSQLiteStorage(storage); err != nil {
				t.Fatal(err)
			}
			want := plugin.ErrDisabled
			if operation == "remove" || operation == "uninstall" {
				want = plugin.ErrNotInstalled
			} else if operation == "revoke" {
				want = plugin.ErrGrantDenied
			}
			if err := restarted.Authorize("project", m.ID, m.Version, "read"); !errors.Is(err, want) {
				t.Fatalf("restriction lost after restart: %v, want %v", err, want)
			}
		})
	}
}
