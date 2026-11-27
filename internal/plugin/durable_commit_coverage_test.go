package plugin

import (
	"errors"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestDurableCoverageCapacityLimits(t *testing.T) {
	s := coverageStorage(t)
	expiration := time.Now().Add(time.Hour).UnixMilli()
	for _, tc := range []struct {
		namespace   string
		op, payload string
	}{
		{inviteNamespace, "invites.create", `{"roomId":"room","targetPlayerId":"player"}`},
		{subscriptionTickets, "subscriptions.open", `{"channel":"rooms.connected"}`},
	} {
		// Fill the real storage namespace without relying on random ticket IDs.
		_, err := s.db.Exec("WITH RECURSIVE n(i) AS (SELECT 1 UNION ALL SELECT i+1 FROM n WHERE i<4096) INSERT INTO plugin_state(project_id,namespace,state_key,value_json) SELECT ?,?,CAST(i AS TEXT),json_object('expiresAt',?,'channel',?) FROM n", "coverage", tc.namespace, expiration, subscriptionChannel)
		if err != nil {
			t.Fatal(err)
		}
		var effect func(sqliteExecutor) error
		if tc.namespace == inviteNamespace {
			_, effect, err = prepareInviteStorageEffect(s, coverageRequest(tc.op, tc.payload))
		} else {
			_, effect, err = prepareSubscriptionStorageEffect(s, coverageRequest(tc.op, tc.payload))
		}
		if err != nil || effect == nil {
			t.Fatalf("%s admission: %v", tc.op, err)
		}
		if err := effect(s.db); !errors.Is(err, ErrModelConflict) {
			t.Fatalf("%s capacity: %v", tc.op, err)
		}
		var count int
		if err := s.db.QueryRow("SELECT COUNT(*) FROM plugin_state WHERE project_id=? AND namespace=?", "coverage", tc.namespace).Scan(&count); err != nil || count != 4096 {
			t.Fatalf("%s changed full namespace: %d %v", tc.op, count, err)
		}
	}
}

func TestDurableCoverageStaleSubscriptionRevocation(t *testing.T) {
	s := coverageStorage(t)
	token := strings.Repeat("a", 32)
	ticket := storedSubscriptionTicket{ProjectID: "coverage", Channel: subscriptionChannel, Operation: "subscriptions.open", ExpiresAt: time.Now().Add(time.Hour).UnixMilli()}
	coverageState(t, s, subscriptionTickets, authTokenDigest(token), ticket)
	_, revoke, err := prepareSubscriptionStorageEffect(s, coverageRequest("subscriptions.revoke", `{"channel":"rooms.connected"}`))
	if err != nil || revoke == nil {
		t.Fatalf("revoke admission: %v", err)
	}
	failure := errors.New("injected SQL failure")
	for _, executor := range []coverageSQLExecutor{{execErr: failure}, {resultErr: failure}} {
		if err := revoke(executor); !errors.Is(err, failure) {
			t.Fatal(err)
		}
	}
	coverageState(t, s, subscriptionTickets, authTokenDigest(strings.Repeat("b", 32)), ticket)
	tx, err := s.db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	if err := revoke(tx); !errors.Is(err, ErrModelConflict) {
		t.Fatalf("stale revoke: %v", err)
	}
	if err := tx.Rollback(); err != nil {
		t.Fatal(err)
	}
	count, err := subscriptionTicketCount(s.db, "coverage", time.Now().UnixMilli(), true)
	if err != nil || count != 2 {
		t.Fatalf("stale transaction removed tickets: %d %v", count, err)
	}
	ticket.ProjectID = "other"
	coverageState(t, s, subscriptionTickets, authTokenDigest(token), ticket)
	payload := `{"channel":"rooms.connected","ticket":"` + token + `"}`
	if _, _, err := prepareSubscriptionStorageEffect(s, coverageRequest("subscriptions.receive", payload)); !errors.Is(err, errSubscriptionTicket) {
		t.Fatalf("cross-project ticket accepted: %v", err)
	}
}

func TestDurableCoverageObservationStorageFailures(t *testing.T) {
	s := coverageStorage(t)
	now := time.Now()
	delivery := observationDelivery{ID: "event", Target: WorkerRequest{ProjectID: "coverage", PluginID: "observer", Version: "1.0.0"}}
	failure := errors.New("injected SQL failure")
	for _, executor := range []coverageSQLExecutor{{execErr: failure}, {resultErr: failure}} {
		if err := enqueueSQLiteObservation(executor, delivery, now); !errors.Is(err, failure) {
			t.Fatal(err)
		}
	}
	if _, err := s.db.Exec("INSERT INTO plugin_observation_queue(project_id,target_plugin,target_version,observation_id,delivery_json,attempts,due_at) VALUES(?,?,?,?,?,?,?)", "coverage", "observer", "1.0.0", "bad", "{", 0, now.UnixMilli()); err != nil {
		t.Fatal(err)
	}
	if _, err := s.dueObservations(now, 10); err == nil {
		t.Fatal("corrupt queued observation accepted")
	}
	if _, err := s.db.Exec("INSERT INTO plugin_observation_receipts(project_id,plugin_id,observation_id,delivered_at,payload_json) VALUES(?,?,?,?,?)", "coverage", "observer", "bad", now.UnixMilli(), "{"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.ObservationReceipts("coverage", "observer", 10); err == nil {
		t.Fatal("corrupt receipt accepted")
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := s.dueObservations(now, 10); err == nil {
		t.Fatal("closed queue accepted")
	}
	if err := s.finishObservation(delivery, true, now); err == nil {
		t.Fatal("closed queue allowed commit")
	}
	if _, err := s.ObservationReceipts("coverage", "observer", 10); err == nil {
		t.Fatal("closed receipt store accepted")
	}
}

func TestDurableCoverageAuthMutationFailure(t *testing.T) {
	s := coverageStorage(t)
	p := legacyStoragePlanner{storage: s}
	_, commit, err := p.prepareAuthStorage(coverageRequest("auth.guest", "{}"))
	if err != nil || commit == nil {
		t.Fatalf("guest admission: %v", err)
	}
	// Abort only the mutation, after successful admission and transaction start.
	if _, err := s.db.Exec("CREATE TRIGGER reject_guest BEFORE INSERT ON plugin_state BEGIN SELECT RAISE(ABORT,'test failure'); END"); err != nil {
		t.Fatal(err)
	}
	token := strings.Repeat("a", 64)
	payload := []byte(`{"result":{"token":"` + token + `","username":"guest-test","userId":"coverage:guest-test","guest":true},"expiresAt":` + strconv.FormatInt(time.Now().Add(time.Hour).UnixMilli(), 10) + `}`)
	if err := commit(&WorkerResponse{Payload: payload}, nil); err == nil {
		t.Fatal("failed guest write accepted")
	}
	if got, err := s.State("coverage", authSessionsNamespace, authTokenDigest(token)); err != nil || got != nil {
		t.Fatalf("failed guest write persisted: %v %v", got, err)
	}
}
