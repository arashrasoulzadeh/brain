package plugin

import (
	"errors"
	"testing"
)

func TestCoverageModelAdmissionFailures(t *testing.T) {
	r := NewRegistry(1)
	spec := ModelSpec{Version: 1, Fields: map[string]string{"name": "string"}}
	route := Route{Operation: "create", Capability: "write", Helper: "host.model", Model: &ModelRoute{Name: "record", Action: "create"}}
	m := Manifest{ID: "sample", Version: "1", Models: map[string]ModelSpec{"record": spec}, Routes: []Route{route}, Capabilities: []string{"write"}}
	r.registered["sample@1"] = Registered{Manifest: m}
	r.readiness["sample@1"] = true
	req := WorkerRequest{Protocol: 1, ProjectID: "coverage", PluginID: "sample", Version: "1", Operation: "create", Payload: []byte(`{"id":"record","fields":{"name":"test"}}`)}
	if _, handled, err := r.invokeModel(req); !handled || !errors.Is(err, ErrNotInstalled) {
		t.Fatalf("uninstalled model: %v %v", handled, err)
	}
	i := Installation{ProjectID: "coverage", PluginID: "sample", Version: "1", Enabled: true, Grants: map[string]bool{"write": true}}
	r.installations["coverage:sample@1"] = i
	if _, handled, err := r.invokeModel(req); !handled || !errors.Is(err, ErrLifecycleStorage) {
		t.Fatalf("missing model storage: %v %v", handled, err)
	}
	i.Resources = map[string]map[string]bool{"write": {"other": true}}
	r.installations["coverage:sample@1"] = i
	if _, handled, err := r.invokeModel(req); !handled || !errors.Is(err, ErrGrantDenied) {
		t.Fatalf("out of scope model: %v %v", handled, err)
	}
}

func TestCoverageUnknownModelActionAndWriteFailure(t *testing.T) {
	s := coverageStorage(t)
	spec := ModelSpec{Version: 1, Fields: map[string]string{"name": "string"}}
	req := coverageRequest("create", `{"id":"record","fields":{"name":"test"}}`)
	req.PluginID = "sample"
	result, err := s.executeModel(req, spec, ModelRoute{Name: "record", Action: "create"})
	if err != nil {
		t.Fatal(err)
	}
	record := result.(modelRecord)
	req.Payload = []byte(`{"id":"record","revision":"` + record.Revision + `","fields":{"name":"updated"}}`)
	if _, err := s.executeModel(req, spec, ModelRoute{Name: "record", Action: "unknown"}); !errors.Is(err, ErrInvalidWorkerRequest) {
		t.Fatal(err)
	}
	if _, err := s.db.Exec("CREATE TRIGGER reject_model_update BEFORE UPDATE ON plugin_model_records BEGIN SELECT RAISE(ABORT,'test failure'); END"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.executeModel(req, spec, ModelRoute{Name: "record", Action: "update"}); err == nil {
		t.Fatal("failed update accepted")
	}
	got, err := s.executeModel(req, spec, ModelRoute{Name: "record", Action: "read"})
	if err != nil {
		t.Fatal(err)
	}
	if current := got.(modelRecord); current.Revision != record.Revision || current.Fields["name"] != "test" {
		t.Fatalf("failed write mutated model: %+v", current)
	}
}
