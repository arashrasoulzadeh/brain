package plugin

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestCoverageInviteCommitExpiry(t *testing.T) {
	s := coverageStorage(t)
	for _, operation := range []string{"invites.accept", "invites.cancel"} {
		t.Run(operation, func(t *testing.T) {
			code := "INVITE-coverage-code"
			id := "invite-" + operation
			expires := time.Now().Add(500 * time.Millisecond)
			coverageState(t, s, inviteNamespace, id, storedInvite{ID: id, ProjectID: "coverage", RoomID: "room", TargetPlayerID: "player", CodeDigest: authTokenDigest(code), ExpiresAt: expires.UnixMilli()})
			req := coverageRequest(operation, `{"roomId":"room","inviteId":"`+id+`","playerId":"player","code":"`+code+`"}`)
			_, commit, err := prepareInviteStorageEffect(s, req)
			if err != nil || commit == nil {
				t.Fatalf("invite admission: %v", err)
			}
			time.Sleep(time.Until(expires) + 10*time.Millisecond)
			if err := commit(s.db); !errors.Is(err, errInviteUnavailable) {
				t.Fatalf("expired invitation committed: %v", err)
			}
			if value, err := s.State("coverage", inviteNamespace, id); err != nil || value == nil || value["used"] != false {
				t.Fatalf("expired invitation mutated: %v %v", value, err)
			}
		})
	}
}

func TestCoverageSubscriptionCommitExpiry(t *testing.T) {
	s := coverageStorage(t)
	token := strings.Repeat("a", 32)
	expires := time.Now().Add(500 * time.Millisecond)
	coverageState(t, s, subscriptionTickets, authTokenDigest(token), storedSubscriptionTicket{ProjectID: "coverage", Channel: subscriptionChannel, Operation: "subscriptions.open", ExpiresAt: expires.UnixMilli()})
	_, commit, err := prepareSubscriptionStorageEffect(s, coverageRequest("subscriptions.close", `{"channel":"rooms.connected","ticket":"`+token+`"}`))
	if err != nil || commit == nil {
		t.Fatalf("ticket admission: %v", err)
	}
	time.Sleep(time.Until(expires) + 10*time.Millisecond)
	if err := commit(s.db); !errors.Is(err, errSubscriptionTicket) {
		t.Fatalf("expired ticket committed: %v", err)
	}
}

func TestCoverageStorageReadFailure(t *testing.T) {
	s := coverageStorage(t)
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	if _, _, err := prepareSubscriptionStorageEffect(s, coverageRequest("subscriptions.open", `{"channel":"rooms.connected"}`)); err == nil {
		t.Fatal("closed stream store accepted")
	}
	if _, err := s.dueObservations(time.Now(), 1); err == nil {
		t.Fatal("closed observation store accepted")
	}
	if _, err := s.ObservationReceipts("coverage", "sample", 1); err == nil {
		t.Fatal("closed receipt store accepted")
	}
	r := NewRegistry(1)
	if err := r.BindSQLiteStorage(s); err == nil {
		t.Fatal("closed registry store accepted")
	}
}
