package plugin_test

import (
	"errors"
	"path/filepath"
	"testing"

	"brain/internal/plugin"
)

func TestRejectedRegistryInvocationsDoNotWriteStorage(t *testing.T) {
	const projectID = "project"
	const version = "1.0.0"
	cases := []struct {
		name      string
		operation string
		protocol  int
		installed bool
		enabled   bool
		ready     bool
		granted   bool
		worker    bool
		want      error
	}{
		{name: "not installed", operation: "tables.write", protocol: 1, want: plugin.ErrNotInstalled},
		{name: "disabled", operation: "tables.write", protocol: 1, installed: true, want: plugin.ErrDisabled},
		{name: "not ready", operation: "tables.write", protocol: 1, installed: true, enabled: true, granted: true, worker: true, want: plugin.ErrNotReady},
		{name: "revoked grant", operation: "tables.write", protocol: 1, installed: true, enabled: true, ready: true, worker: true, want: plugin.ErrGrantDenied},
		{name: "wrong protocol", operation: "tables.write", protocol: 2, installed: true, enabled: true, ready: true, granted: true, worker: true, want: plugin.ErrInvalidWorkerRequest},
		{name: "undeclared route", operation: "tables.write", protocol: 1, installed: true, enabled: true, ready: true, granted: true, worker: true, want: plugin.ErrGrantDenied},
		{name: "no worker", operation: "tables.write", protocol: 1, installed: true, enabled: true, ready: true, granted: true, want: plugin.ErrWorkerNotReady},
		{name: "accepted", operation: "tables.write", protocol: 1, installed: true, enabled: true, ready: true, granted: true, worker: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			storage, err := plugin.OpenSQLiteStorage(filepath.Join(t.TempDir(), "storage.sqlite"))
			if err != nil {
				t.Fatal(err)
			}
			defer storage.Close()
			registry := plugin.NewRegistry(1)
			manifest := plugin.Manifest{
				ID: "tables", Version: version, HostAPI: 1, Capabilities: []string{"tables.write"},
				Routes: []plugin.Route{{Operation: "tables.write", Function: "write", Capability: "tables.write"}},
			}
			if tc.name == "undeclared route" {
				manifest.Routes = nil
			}
			if err := registry.Register(manifest); err != nil {
				t.Fatal(err)
			}
			registry.BindSQLiteStorage(storage)
			if tc.installed {
				if err := registry.Install(projectID, manifest.ID, version); err != nil {
					t.Fatal(err)
				}
			}
			if tc.enabled {
				if err := registry.SetReadiness(manifest.ID, version, true); err != nil {
					t.Fatal(err)
				}
				if err := registry.Grant(projectID, manifest.ID, version, "tables.write"); err != nil {
					t.Fatal(err)
				}
				if err := registry.Enable(projectID, manifest.ID, version); err != nil {
					t.Fatal(err)
				}
				if !tc.granted {
					if err := registry.Revoke(projectID, manifest.ID, version, "tables.write"); err != nil {
						t.Fatal(err)
					}
				}
				if !tc.ready {
					if err := registry.SetReadiness(manifest.ID, version, false); err != nil {
						t.Fatal(err)
					}
				}
			}
			worker := &countingWorker{}
			if tc.worker {
				if err := registry.BindWorker(projectID, manifest.ID, version, worker); err != nil {
					t.Fatal(err)
				}
			}
			response, err := registry.Invoke(plugin.WorkerRequest{
				Protocol: tc.protocol, ProjectID: projectID, PluginID: manifest.ID, Version: version,
				Operation: tc.operation, Payload: []byte(`{"table":"messages","row":{"id":"example"}}`),
			})
			if !errors.Is(err, tc.want) {
				t.Fatalf("invoke error = %v, want %v", err, tc.want)
			}
			rows, err := storage.Rows(projectID, "messages")
			if err != nil {
				t.Fatal(err)
			}
			wantRows := 0
			if tc.want == nil {
				wantRows = 1
				if !response.Success {
					t.Fatal("authorized invocation did not succeed")
				}
			}
			if len(rows) != wantRows || worker.calls != wantRows {
				t.Fatalf("stored rows = %d, worker calls = %d, want %d", len(rows), worker.calls, wantRows)
			}
		})
	}
}
