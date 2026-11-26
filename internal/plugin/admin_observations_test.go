package plugin

import "testing"

func TestAdminObservationRevisionIsProjectAndEventScoped(t *testing.T) {
	r := NewRegistry(1)
	r.recordAdminObservation("one", "rooms.connected")
	r.recordAdminObservation("two", "rooms.connected")
	r.recordAdminObservation("one", "rooms.closed")
	if got := r.AdminObservationRevision("one", []string{"rooms.connected"}); got != 1 {
		t.Fatalf("revision = %d, want 1", got)
	}
	if got := r.AdminObservationRevision("two", []string{"rooms.connected"}); got != 2 {
		t.Fatalf("revision = %d, want 2", got)
	}
	if got := r.AdminObservationRevision("three", []string{"rooms.connected"}); got != 0 {
		t.Fatalf("cross-project event exposed: %d", got)
	}
	for i := 0; i < len(r.adminObservations); i++ {
		r.recordAdminObservation("two", "rooms.closed")
	}
	if got := r.AdminObservationRevision("one", []string{"rooms.connected"}); got != 0 {
		t.Fatalf("bounded history was not evicted: %d", got)
	}
}
