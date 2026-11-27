package plugin_test

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"brain/internal/plugin"
)

func TestObservationContractAndFanoutBounds(t *testing.T) {
	publisher := plugin.Manifest{ID: "publisher", Version: "1", HostAPI: 1, Capabilities: []string{"write"}, Routes: []plugin.Route{{Operation: "write", Function: "write", Capability: "write", ResourceField: "key"}}, Events: []plugin.EventSpec{{Name: "changed", Fields: map[string]string{"name": "string"}}}}
	subscriber := plugin.Manifest{ID: "subscriber", Version: "1", HostAPI: 1, Capabilities: []string{"observe.publisher.changed"}, Observers: []plugin.ObserverSpec{{Plugin: "publisher", Event: "changed", Function: "receive", Capability: "observe.publisher.changed"}}}
	for _, event := range []plugin.EventSpec{{Name: "invalid/name"}, {Name: "changed", Source: "foreign"}, {Name: "changed", Fields: map[string]string{"bad-key": "unknown"}}} {
		m := publisher
		m.Events = []plugin.EventSpec{event}
		if err := plugin.ValidateManifestContract(m); !errors.Is(err, plugin.ErrInvalidManifest) {
			t.Fatalf("invalid event: %v", err)
		}
	}
	bad := subscriber
	bad.Observers = append(append([]plugin.ObserverSpec{}, subscriber.Observers...), subscriber.Observers...)
	if err := plugin.ValidateManifestContract(bad); !errors.Is(err, plugin.ErrInvalidManifest) {
		t.Fatalf("duplicate observer: %v", err)
	}
	r := plugin.NewRegistry(1)
	if err := r.Register(publisher); err != nil {
		t.Fatal(err)
	}
	var intents []plugin.ObservationIntent
	installAccessTestPlugin(t, r, "alpha", publisher, &storageCommitWorker{handle: func(req plugin.WorkerRequest) plugin.WorkerResponse {
		return plugin.WorkerResponse{Protocol: 1, Success: true, Observations: intents}
	}})
	delivered := 0
	for i := 0; i < 33; i++ {
		m := subscriber
		m.ID = fmt.Sprintf("subscriber%d", i)
		if err := r.Register(m); err != nil {
			t.Fatal(err)
		}
		installAccessTestPlugin(t, r, "alpha", m, &storageCommitWorker{handle: func(req plugin.WorkerRequest) plugin.WorkerResponse {
			delivered++
			return plugin.WorkerResponse{Protocol: 1, Success: true}
		}})
	}
	req := plugin.WorkerRequest{Protocol: 1, ProjectID: "alpha", PluginID: publisher.ID, Version: "1", Operation: "write", Payload: []byte(`{"key":"one"}`)}
	invoke := func() {
		t.Helper()
		if response, err := r.Invoke(req); err != nil || !response.Success || len(response.Observations) != 0 {
			t.Fatalf("publisher: %+v %v", response, err)
		}
	}
	for _, payload := range []string{"", "{", "null", `{"name":42}`, `{"name":true}`, `{"name":"` + strings.Repeat("x", 1025) + `"}`, `{"name":"` + strings.Repeat("x", 17000) + `"}`} {
		intents = []plugin.ObservationIntent{{Event: "changed", Payload: json.RawMessage(payload)}}
		invoke()
	}
	if delivered != 0 {
		t.Fatal("invalid event delivered")
	}
	valid := plugin.ObservationIntent{Event: "changed", Payload: json.RawMessage(`{"name":"example"}`)}
	intents = make([]plugin.ObservationIntent, 17)
	for i := range intents {
		intents[i] = valid
	}
	invoke()
	if delivered != 0 {
		t.Fatal("oversized event batch delivered")
	}
	intents = []plugin.ObservationIntent{valid}
	invoke()
	if delivered != 32 {
		t.Fatalf("fanout=%d, want 32", delivered)
	}
}

func TestQueuedObservationRejectsStaleAndTamperedEvents(t *testing.T) {
	for _, fault := range []string{"envelope", "publisher revoked", "target promoted", "project inactive", "event schema", "target revoked"} {
		t.Run(fault, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "db")
			storage, err := plugin.OpenSQLiteStorage(path)
			if err != nil {
				t.Fatal(err)
			}
			defer storage.Close()
			r := plugin.NewRegistry(1)
			if err := r.BindSQLiteStorage(storage); err != nil {
				t.Fatal(err)
			}
			publisher := plugin.Manifest{ID: "publisher", Version: "1", HostAPI: 1, Capabilities: []string{"write"}, Routes: []plugin.Route{{Operation: "write", Function: "write", Capability: "write", ResourceField: "key"}}, Events: []plugin.EventSpec{{Name: "changed", Fields: map[string]string{"name": "string"}}}}
			subscriber := plugin.Manifest{ID: "subscriber", Version: "1", HostAPI: 1, Capabilities: []string{"observe.publisher.changed"}, Observers: []plugin.ObserverSpec{{Plugin: "publisher", Event: "changed", Function: "receive", Capability: "observe.publisher.changed"}}}
			for _, m := range []plugin.Manifest{publisher, subscriber} {
				if err := r.Register(m); err != nil {
					t.Fatal(err)
				}
				installAccessTestPlugin(t, r, "alpha", m, &contractWorker{ready: true})
			}
			if err := r.BindWorker("alpha", publisher.ID, "1", &storageCommitWorker{handle: func(req plugin.WorkerRequest) plugin.WorkerResponse {
				return plugin.WorkerResponse{Protocol: 1, Success: true, Observations: []plugin.ObservationIntent{{Event: "changed", Payload: json.RawMessage(`{"name":"example"}`)}}}
			}}); err != nil {
				t.Fatal(err)
			}
			if err := r.BindWorker("alpha", subscriber.ID, "1", &storageCommitWorker{handle: func(req plugin.WorkerRequest) plugin.WorkerResponse {
				t.Fatal("stale event reached observer")
				return plugin.WorkerResponse{}
			}}); err != nil {
				t.Fatal(err)
			}
			req := plugin.WorkerRequest{Protocol: 1, ProjectID: "alpha", PluginID: publisher.ID, Version: "1", Operation: "write", Payload: []byte(`{"key":"one"}`)}
			if _, err := r.Invoke(req); err != nil {
				t.Fatal(err)
			}
			db, err := sql.Open("sqlite", path)
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			switch fault {
			case "envelope", "event schema":
				var raw string
				if err := db.QueryRow(`SELECT delivery_json FROM plugin_observation_queue`).Scan(&raw); err != nil {
					t.Fatal(err)
				}
				var delivery map[string]any
				if err := json.Unmarshal([]byte(raw), &delivery); err != nil {
					t.Fatal(err)
				}
				target := delivery["Target"].(map[string]any)
				payload, err := base64.StdEncoding.DecodeString(target["payload"].(string))
				if err != nil {
					t.Fatal(err)
				}
				var event map[string]any
				if err := json.Unmarshal(payload, &event); err != nil {
					t.Fatal(err)
				}
				if fault == "envelope" {
					event["projectId"] = "beta"
				} else {
					event["payload"] = map[string]any{"name": false}
				}
				payload, err = json.Marshal(event)
				if err != nil {
					t.Fatal(err)
				}
				target["payload"] = base64.StdEncoding.EncodeToString(payload)
				encoded, err := json.Marshal(delivery)
				if err != nil {
					t.Fatal(err)
				}
				if _, err := db.Exec(`UPDATE plugin_observation_queue SET delivery_json=?`, string(encoded)); err != nil {
					t.Fatal(err)
				}
			case "publisher revoked":
				if err := r.Revoke("alpha", publisher.ID, "1", "write"); err != nil {
					t.Fatal(err)
				}
			case "target revoked":
				if err := r.Revoke("alpha", subscriber.ID, "1", subscriber.Capabilities[0]); err != nil {
					t.Fatal(err)
				}
			case "project inactive":
				r.BindProjectResolver(func(id string) (plugin.ProjectSnapshot, error) {
					return plugin.ProjectSnapshot{ID: id, Active: false}, nil
				})
			case "target promoted":
				next := subscriber
				next.Version = "2"
				if err := r.Register(next); err != nil {
					t.Fatal(err)
				}
				installAccessTestPlugin(t, r, "alpha", next, &contractWorker{ready: true})
				if err := r.Promote("alpha", next.ID, next.Version); err != nil {
					t.Fatal(err)
				}
			}
			if n, err := r.DrainObservations(time.Now().Add(time.Minute), 10); err != nil || n != 1 {
				t.Fatalf("stale delivery not consumed: %d %v", n, err)
			}
			if n, err := r.DrainObservations(time.Now().Add(time.Hour), 10); err != nil || n != 0 {
				t.Fatalf("permanent failure retried: %d %v", n, err)
			}
		})
	}
}
