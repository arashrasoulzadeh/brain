package plugin_test

import (
	"path/filepath"
	"testing"

	"brain/internal/plugin"
)

func TestObservationPublishersRejectForgedHostState(t *testing.T) {
	storage, err := plugin.OpenSQLiteStorage(filepath.Join(t.TempDir(), "state.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()
	for _, tc := range []struct {
		id, operation, payload string
	}{
		{"rooms", "rooms.join", `{"id":"missing","member":"player","room":{"members":["player"]}}`},
		{"matchmaking", "matchmaking.assign", `{"id":"missing","ticket":{"status":"assigned"}}`},
		{"files", "files.publish", `{"key":"missing","published":true}`},
		{"tables", "tables.write", `{"project_id":"other","table":"messages","row":{"id":"one"}}`},
	} {
		t.Run(tc.operation, func(t *testing.T) {
			r := plugin.NewRegistry(1)
			m := plugin.Manifest{ID: tc.id, Version: "1", HostAPI: 1, Capabilities: []string{tc.operation}, Routes: []plugin.Route{{Operation: tc.operation, Function: "run", Capability: tc.operation}}}
			if err := r.Register(m); err != nil {
				t.Fatal(err)
			}
			if err := r.BindSQLiteStorage(storage); err != nil {
				t.Fatal(err)
			}
			if err := r.SetReadiness(m.ID, m.Version, true); err != nil {
				t.Fatal(err)
			}
			if err := r.Install("project", m.ID, m.Version); err != nil {
				t.Fatal(err)
			}
			if err := r.Grant("project", m.ID, m.Version, tc.operation); err != nil {
				t.Fatal(err)
			}
			if err := r.Enable("project", m.ID, m.Version); err != nil {
				t.Fatal(err)
			}
			if err := r.BindWorker("project", m.ID, m.Version, &storageCommitWorker{handle: func(plugin.WorkerRequest) plugin.WorkerResponse {
				t.Fatal("forged state reached worker")
				return plugin.WorkerResponse{}
			}}); err != nil {
				t.Fatal(err)
			}
			if _, err := r.Invoke(plugin.WorkerRequest{Protocol: 1, ProjectID: "project", PluginID: m.ID, Version: m.Version, Operation: tc.operation, Payload: []byte(tc.payload)}); err == nil {
				t.Fatal("forged state accepted")
			}
		})
	}
}
