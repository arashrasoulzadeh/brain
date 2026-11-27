package plugin

import (
	"crypto/rand"
	"errors"
	"io"
	"testing"
	"time"
)

func coverageDelegationRegistry() (*Registry, WorkerRequest, WorkerRequest) {
	r := NewRegistry(1)
	call := &PluginCall{Plugin: "target", Version: "1", Operation: "read"}
	source := Manifest{ID: "source", Version: "1", Declarative: true, Dependencies: []string{"target@1"}, Capabilities: []string{"call.target.read"}, Routes: []Route{{Operation: "delegate", Capability: "call.target.read", Helper: "host.call", Call: call}}}
	target := Manifest{ID: "target", Version: "1", Declarative: true, Capabilities: []string{"read"}, Routes: []Route{{Operation: "read", Capability: "read", Helper: "sdk.pipeline", Pipeline: &PipelineSpec{Steps: []PipelineStep{{Helper: "value.identity", Input: "$input", Save: "value"}}, Return: "$steps.value"}}}}
	for _, manifest := range []Manifest{source, target} {
		key := manifest.ID + "@" + manifest.Version
		r.registered[key] = Registered{Manifest: manifest}
		r.readiness[key] = true
		r.installations["coverage:"+key] = Installation{ProjectID: "coverage", PluginID: manifest.ID, Version: manifest.Version, Enabled: true, Grants: map[string]bool{manifest.Capabilities[0]: true}}
		r.workers["coverage:"+key] = coverageStopFailureWorker{}
	}
	request := WorkerRequest{Protocol: 1, ProjectID: "coverage", PluginID: "source", Version: "1", Operation: "delegate", Payload: []byte("{}")}
	destination := WorkerRequest{Protocol: 1, ProjectID: "coverage", PluginID: "target", Version: "1", Operation: "read", Payload: []byte("{}")}
	return r, request, destination
}

func TestCoverageDelegationRejectsImpureAndInactiveTargets(t *testing.T) {
	r, source, target := coverageDelegationRegistry()
	if err := r.checkPluginCall(source, target); err != nil {
		t.Fatal(err)
	}
	for _, call := range []*PluginCall{nil, {Plugin: "source", Version: "1", Operation: "read"}, {Plugin: "!", Version: "1", Operation: "read"}, {Plugin: "target", Version: "", Operation: "read"}} {
		if call.valid(r.registered["source@1"].Manifest, "call.target.read") {
			t.Fatal("invalid delegation accepted")
		}
	}
	install := r.installations["coverage:source@1"]
	install.Draining = true
	install.DrainUntil = time.Now().Add(time.Hour)
	r.installations["coverage:source@1"] = install
	if err := r.checkPluginCall(source, target); !errors.Is(err, ErrGrantDenied) {
		t.Fatal(err)
	}
	install.Draining = false
	r.installations["coverage:source@1"] = install
	r.active["coverage:target"] = "2"
	if err := r.checkPluginCall(source, target); !errors.Is(err, ErrGrantDenied) {
		t.Fatal(err)
	}
	delete(r.active, "coverage:target")
	manifest := r.registered["target@1"].Manifest
	manifest.Declarative = false
	r.registered["target@1"] = Registered{Manifest: manifest}
	if err := r.checkPluginCall(source, target); !errors.Is(err, ErrGrantDenied) {
		t.Fatal(err)
	}
	manifest.Declarative = true
	manifest.Routes[0].Pipeline.Steps[0].Helper = "observe.fire"
	r.registered["target@1"] = Registered{Manifest: manifest}
	if err := r.checkPluginCall(source, target); !errors.Is(err, ErrGrantDenied) {
		t.Fatal(err)
	}
}

type coverageTransientReadyWorker struct {
	coverageStopFailureWorker
	readyCalls int
}

func (w *coverageTransientReadyWorker) Ready() bool {
	w.readyCalls++
	return w.readyCalls == 1
}

func TestCoverageDelegationInvocationFailures(t *testing.T) {
	r, source, _ := coverageDelegationRegistry()
	original := rand.Reader
	rand.Reader = coverageFailEntropy{}
	_, handled, err := r.invokePluginCall(source)
	rand.Reader = original
	if !handled || !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("entropy failure: %v %v", handled, err)
	}
	r.workers["coverage:target@1"] = &coverageTransientReadyWorker{}
	if _, handled, err := r.invokePluginCall(source); !handled || !errors.Is(err, ErrWorkerNotReady) {
		t.Fatalf("worker readiness race: %v %v", handled, err)
	}
}
