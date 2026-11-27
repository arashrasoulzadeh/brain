package plugin_test

import (
	"errors"
	"path/filepath"
	"testing"

	"brain/internal/plugin"
)

type storageCommitWorker struct {
	handle func(plugin.WorkerRequest) plugin.WorkerResponse
}

func (w *storageCommitWorker) Start() error { return nil }
func (w *storageCommitWorker) Stop() error  { return nil }
func (w *storageCommitWorker) Ready() bool  { return true }
func (w *storageCommitWorker) Handle(request plugin.WorkerRequest) plugin.WorkerResponse {
	return w.handle(request)
}

func TestRegistryCommitsStorageOnlyAfterSuccessfulAuthorizedExecution(t *testing.T) {
	for _, name := range []string{"worker failure", "revoked during invocation", "success"} {
		t.Run(name, func(t *testing.T) {
			storage, err := plugin.OpenSQLiteStorage(filepath.Join(t.TempDir(), "storage.sqlite"))
			if err != nil {
				t.Fatal(err)
			}
			defer storage.Close()
			registry := plugin.NewRegistry(1)
			manifest := plugin.Manifest{
				ID: "tables", Version: "1.0.0", HostAPI: 1, Capabilities: []string{"tables.write"},
				Routes: []plugin.Route{{Operation: "tables.write", Function: "write", Capability: "tables.write"}},
			}
			if err := registry.Register(manifest); err != nil {
				t.Fatal(err)
			}
			registry.BindSQLiteStorage(storage)
			if err := registry.Install("project", manifest.ID, manifest.Version); err != nil {
				t.Fatal(err)
			}
			if err := registry.SetReadiness(manifest.ID, manifest.Version, true); err != nil {
				t.Fatal(err)
			}
			if err := registry.Grant("project", manifest.ID, manifest.Version, "tables.write"); err != nil {
				t.Fatal(err)
			}
			if err := registry.Enable("project", manifest.ID, manifest.Version); err != nil {
				t.Fatal(err)
			}
			worker := &storageCommitWorker{handle: func(request plugin.WorkerRequest) plugin.WorkerResponse {
				rows, err := storage.Rows("project", "messages")
				if err != nil || len(rows) != 0 {
					t.Fatalf("storage changed before execution: %v, %v", rows, err)
				}
				if name == "worker failure" {
					return plugin.WorkerResponse{Protocol: 1, ErrorCode: "rejected"}
				}
				if name == "revoked during invocation" {
					if err := registry.Revoke("project", manifest.ID, manifest.Version, "tables.write"); err != nil {
						t.Fatal(err)
					}
				}
				return plugin.WorkerResponse{Protocol: 1, Success: true}
			}}
			if err := registry.BindWorker("project", manifest.ID, manifest.Version, worker); err != nil {
				t.Fatal(err)
			}
			response, err := registry.Invoke(plugin.WorkerRequest{
				Protocol: 1, ProjectID: "project", PluginID: manifest.ID, Version: manifest.Version,
				Operation: "tables.write", Payload: []byte(`{"table":"messages","row":{"id":"example"}}`),
			})
			if name == "revoked during invocation" {
				if !errors.Is(err, plugin.ErrGrantDenied) {
					t.Fatalf("revoked invocation error = %v", err)
				}
			} else if err != nil {
				t.Fatal(err)
			}
			if name == "worker failure" && (response.Success || response.ErrorCode != "rejected") {
				t.Fatalf("worker failure was not preserved: %#v", response)
			}
			rows, err := storage.Rows("project", "messages")
			want := 0
			if name == "success" {
				want = 1
			}
			if err != nil || len(rows) != want {
				t.Fatalf("rows = %v, err = %v, want %d rows", rows, err, want)
			}
		})
	}
}
