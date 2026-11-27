package plugin

import (
	"crypto/rand"
	"errors"
	"fmt"
	"testing"
	"time"
)

func TestCoverageHostObservationRejections(t *testing.T) {
	r := NewRegistry(1)
	m := Manifest{ID: "publisher", Version: "1", Events: []EventSpec{{Name: "event", Source: "host", Fields: map[string]string{"value": "string"}}}}
	r.registered["publisher@1"] = Registered{Manifest: m}
	r.BindProjectResolver(func(string) (ProjectSnapshot, error) { return ProjectSnapshot{}, ErrProjectInvalid })
	r.PublishHostObservation("coverage", "publisher", "event", map[string]any{"value": "test"})
	r.BindProjectResolver(nil)
	r.PublishHostObservation("coverage", "publisher", "event", map[string]any{"value": make(chan int)})
	m.Version = "2"
	r.registered["publisher@2"] = Registered{Manifest: m}
	r.PublishHostObservation("coverage", "publisher", "event", map[string]any{"value": "test"})
	if revision := r.AdminObservationRevision("coverage", []string{"publisher.event"}); revision != 0 {
		t.Fatalf("rejected publication advanced revision: %d", revision)
	}
}

func TestCoverageObservationTargetLimits(t *testing.T) {
	r := NewRegistry(1)
	m := Manifest{ID: "publisher", Version: "1", Events: []EventSpec{{Name: "event", Fields: map[string]string{}}}}
	pub := WorkerRequest{Protocol: 1, ProjectID: "coverage", PluginID: "publisher", Version: "1", Operation: "publish"}
	intent := ObservationIntent{Event: "event", Payload: []byte("{}")}
	// Install more authorized observers than can fit in one delivery plan.
	for i := 0; i < maxObservationDeliveries+1; i++ {
		id := fmt.Sprintf("observer-%02d", i)
		key := id + "@1"
		manifest := Manifest{ID: id, Version: "1", Capabilities: []string{"observe.publisher.event"}, Observers: []ObserverSpec{{Plugin: "publisher", Event: "event", Function: "observe", Capability: "observe.publisher.event"}}}
		r.registered[key] = Registered{Manifest: manifest}
		r.readiness[key] = true
		r.installations["coverage:"+key] = Installation{ProjectID: "coverage", PluginID: id, Version: "1", Enabled: true, Grants: map[string]bool{"observe.publisher.event": true}}
	}
	r.installations["coverage:missing@1"] = Installation{ProjectID: "coverage", PluginID: "missing", Version: "1", Enabled: true}
	r.active["coverage:observer-00"] = "2"
	if got := len(r.observationTargets(pub, "event")); got != maxObservationDeliveries {
		t.Fatalf("target count: %d", got)
	}
	delete(r.active, "coverage:observer-00")
	if got := len(r.planObservations(pub, m, []ObservationIntent{intent}, "")); got != maxObservationDeliveries {
		t.Fatalf("delivery count: %d", got)
	}
	original := rand.Reader
	rand.Reader = coverageFailEntropy{}
	got := r.planObservations(pub, m, []ObservationIntent{intent}, "")
	rand.Reader = original
	if len(got) != 0 {
		t.Fatal("event planned without random identity")
	}
	r.recordCommittedAdminObservations(pub, []ObservationIntent{intent}, "")
	if got := r.AdminObservationRevision("coverage", []string{"publisher.event"}); got != 0 {
		t.Fatal("unregistered publisher advanced revision")
	}
	if got := r.planWorkerObservations(pub, []ObservationIntent{intent}); len(got) != 0 {
		t.Fatal("unregistered publisher admitted")
	}
}

func TestCoverageObservationQueuePermanentFailureRollback(t *testing.T) {
	s := coverageStorage(t)
	r := NewRegistry(1)
	r.bindObservationStorage(s)
	now := time.Now()
	delivery := observationDelivery{ID: "bad", Target: WorkerRequest{ProjectID: "coverage", PluginID: "observer", Version: "1", Payload: []byte("{")}}
	if err := s.enqueueObservation(delivery, now); err != nil {
		t.Fatal(err)
	}
	if _, err := s.db.Exec("CREATE TRIGGER reject_delivery_delete BEFORE DELETE ON plugin_observation_queue BEGIN SELECT RAISE(ABORT,'test failure'); END"); err != nil {
		t.Fatal(err)
	}
	if count, err := r.DrainObservations(now, 1); count != 0 || err == nil {
		t.Fatalf("failed deletion reported as completed: %d %v", count, err)
	}
	if _, err := s.db.Exec("DROP TRIGGER reject_delivery_delete"); err != nil {
		t.Fatal(err)
	}
	if count, err := r.DrainObservations(now, 1); count != 1 || err != nil {
		t.Fatalf("permanent failure was not removed: %d %v", count, err)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	if err := s.enqueueObservation(delivery, now); err == nil {
		t.Fatal("closed store accepted delivery")
	}
	r.executeObservationPlan([]observationDelivery{delivery})
	if events := r.Audit(); len(events) == 0 || events[len(events)-1].Action != "observation_enqueue_failed" {
		t.Fatal("enqueue failure not audited")
	}
	if _, _, err := r.validateQueuedObservation(delivery); !errors.Is(err, ErrInvalidWorkerRequest) {
		t.Fatal(err)
	}
}
