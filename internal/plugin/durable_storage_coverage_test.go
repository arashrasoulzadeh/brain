package plugin

import (
	"database/sql"
	"encoding/json"
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func coverageStorage(t *testing.T) *SQLiteStorage {
	t.Helper()
	s, err := OpenSQLiteStorage(filepath.Join(t.TempDir(), "state.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

func coverageRequest(op, payload string) WorkerRequest {
	return WorkerRequest{ProjectID: "coverage", Operation: op, Payload: []byte(payload)}
}

func coverageState(t *testing.T, s *SQLiteStorage, ns, key string, value any) {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.db.Exec("INSERT INTO plugin_state(project_id,namespace,state_key,value_json) VALUES(?,?,?,?) ON CONFLICT(project_id,namespace,state_key) DO UPDATE SET value_json=excluded.value_json", "coverage", ns, key, string(data)); err != nil {
		t.Fatal(err)
	}
}

func TestDurableCoverageArenaRejections(t *testing.T) {
	s := coverageStorage(t)
	for _, tc := range []struct {
		op, payload string
		want        error
	}{
		{"arena.start", "null", ErrInvalidWorkerRequest},
		{"arena.start", `{"roomId":" bad"}`, ErrInvalidWorkerRequest},
		{"arena.snapshot", `{"roomId":"missing"}`, ErrStoredResourceMissing},
	} {
		_, effect, err := prepareArenaStorageEffect(s, coverageRequest(tc.op, tc.payload))
		if !errors.Is(err, tc.want) || effect != nil {
			t.Fatalf("%s %s: %v", tc.op, tc.payload, err)
		}
	}
	coverageState(t, s, "arena", "bad", nil)
	_, _, err := prepareArenaStorageEffect(s, coverageRequest("arena.snapshot", `{"roomId":"bad"}`))
	if !errors.Is(err, ErrModelSchema) {
		t.Fatal(err)
	}
	coverageState(t, s, "arena", "room", map[string]any{"state": "active"})
	_, _, err = prepareArenaStorageEffect(s, coverageRequest("arena.input", `{"roomId":"room","playerId":"p"}`))
	if !errors.Is(err, ErrModelSchema) {
		t.Fatal(err)
	}
	_, _, err = prepareArenaStorageEffect(s, coverageRequest("arena.start", `{"roomId":"room"}`))
	if !errors.Is(err, ErrModelConflict) {
		t.Fatal(err)
	}
	req := coverageRequest("other", "null")
	if got, effect, err := prepareArenaStorageEffect(s, req); err != nil || effect != nil || string(got.Payload) != "null" {
		t.Fatalf("unexpected unsupported operation: %v", err)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	if _, _, err := prepareArenaStorageEffect(s, coverageRequest("arena.snapshot", `{"roomId":"room"}`)); err == nil {
		t.Fatal("closed database accepted")
	}
}

type coverageSQLResult struct{ err error }

func (r coverageSQLResult) LastInsertId() (int64, error) { return 0, r.err }
func (r coverageSQLResult) RowsAffected() (int64, error) { return 0, r.err }

type coverageSQLExecutor struct{ execErr, resultErr error }

func (e coverageSQLExecutor) Exec(string, ...any) (sql.Result, error) {
	return coverageSQLResult{e.resultErr}, e.execErr
}
func (e coverageSQLExecutor) QueryRow(string, ...any) *sql.Row { panic("unexpected query") }

func TestDurableCoverageCASFailures(t *testing.T) {
	failure := errors.New("injected SQL failure")
	for _, tc := range []struct {
		executor coverageSQLExecutor
		want     error
	}{
		{coverageSQLExecutor{execErr: failure}, failure},
		{coverageSQLExecutor{resultErr: failure}, failure},
		{coverageSQLExecutor{}, ErrModelConflict},
	} {
		for _, exists := range []bool{false, true} {
			if err := compareAndSwapPluginState(tc.executor, "coverage", "arena", "room", "old", exists, "new"); !errors.Is(err, tc.want) {
				t.Fatalf("CAS failure: got %v want %v", err, tc.want)
			}
		}
	}
}

func TestDurableCoverageInviteRejections(t *testing.T) {
	s := coverageStorage(t)
	for _, tc := range []struct {
		op, payload string
		want        error
	}{
		{"invites.create", "null", ErrInvalidWorkerRequest},
		{"invites.create", `{"roomId":"!"}`, ErrInvalidWorkerRequest},
		{"invites.create", `{"roomId":"room","target":"player","ttlSeconds":0}`, ErrInvalidWorkerRequest},
		{"invites.create", `{"roomId":"room","target":"!"}`, ErrInvalidWorkerRequest},
		{"invites.accept", `{"roomId":"room"}`, errInviteUnavailable},
		{"invites.cancel", `{"roomId":"room","code":"secret"}`, errInviteUnavailable},
		{"invites.accept", `{"roomId":"room","inviteId":"missing"}`, errInviteUnavailable},
		{"invites.accept", `{"roomId":"room","code":"missing"}`, errInviteUnavailable},
		{"invites.inspect", `{"roomId":"room","after":42}`, ErrInvalidWorkerRequest},
		{"other", `{"roomId":"room"}`, ErrInvalidWorkerRequest},
	} {
		_, effect, err := prepareInviteStorageEffect(s, coverageRequest(tc.op, tc.payload))
		if !errors.Is(err, tc.want) || effect != nil {
			t.Fatalf("%s %s: %v", tc.op, tc.payload, err)
		}
	}
	record := storedInvite{ID: "id", ProjectID: "coverage", RoomID: "room", TargetPlayerID: "player", ExpiresAt: time.Now().Add(time.Hour).UnixMilli()}
	coverageState(t, s, inviteNamespace, "id", record)
	if _, _, err := prepareInviteStorageEffect(s, coverageRequest("invites.accept", `{"roomId":"room","inviteId":"id","playerId":"wrong"}`)); !errors.Is(err, errInviteUnavailable) {
		t.Fatal(err)
	}
	_, cancel, err := prepareInviteStorageEffect(s, coverageRequest("invites.cancel", `{"roomId":"room","inviteId":"id"}`))
	if err != nil || cancel == nil {
		t.Fatalf("cancel admission: %v", err)
	}
	record.Used = true
	coverageState(t, s, inviteNamespace, "id", record)
	if err := cancel(s.db); !errors.Is(err, ErrModelConflict) {
		t.Fatalf("stale cancel: %v", err)
	}
	if _, _, err := prepareInviteStorageEffect(s, coverageRequest("invites.accept", `{"roomId":"room","inviteId":"id","playerId":"player"}`)); !errors.Is(err, errInviteUnavailable) {
		t.Fatal(err)
	}
	_, create, err := prepareInviteStorageEffect(s, coverageRequest("invites.create", `{"roomId":"room","target":"player"}`))
	if err != nil || create == nil {
		t.Fatalf("create admission: %v", err)
	}
	failure := errors.New("injected write failure")
	if err := create(coverageSQLExecutor{execErr: failure}); !errors.Is(err, failure) {
		t.Fatal(err)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	for _, op := range []string{"invites.accept", "invites.inspect"} {
		if _, _, err := prepareInviteStorageEffect(s, coverageRequest(op, `{"roomId":"room","inviteId":"id"}`)); err == nil {
			t.Fatalf("%s accepted closed database", op)
		}
	}
}

func TestDurableCoverageSubscriptionRejections(t *testing.T) {
	s := coverageStorage(t)
	for _, tc := range []struct {
		op, payload string
		want        error
	}{
		{"subscriptions.open", "null", ErrInvalidWorkerRequest},
		{"subscriptions.open", `{"channel":"wrong"}`, ErrInvalidWorkerRequest},
		{"subscriptions.receive", `{"channel":"rooms.connected","ticket":"short"}`, errSubscriptionTicket},
		{"subscriptions.receive", `{"channel":"rooms.connected","ticket":"01234567890123456789012345678901"}`, errSubscriptionTicket},
		{"other", `{"channel":"rooms.connected"}`, ErrInvalidWorkerRequest},
		{"@observe/rooms/connected", `{"projectId":"other"}`, ErrInvalidWorkerRequest},
		{"@observe/rooms/connected", `{"projectId":"coverage","pluginId":"rooms","event":"connected","payload":{"roomId":"room"}}`, ErrInvalidWorkerRequest},
	} {
		_, effect, err := prepareSubscriptionStorageEffect(s, coverageRequest(tc.op, tc.payload))
		if !errors.Is(err, tc.want) || effect != nil {
			t.Fatalf("%s %s: %v", tc.op, tc.payload, err)
		}
	}
	for _, state := range []storedSubscriptionStream{
		{Sequence: -1, Events: []storedSubscriptionEvent{}, Seen: []string{}},
		{Sequence: 1, Events: []storedSubscriptionEvent{{Sequence: 1}}, Seen: []string{}},
	} {
		coverageState(t, s, subscriptionStreams, subscriptionChannel, state)
		if _, _, _, err := subscriptionStream(s, "coverage"); !errors.Is(err, ErrModelSchema) {
			t.Fatal(err)
		}
	}
	stream := storedSubscriptionStream{Sequence: 9007199254740991, Events: []storedSubscriptionEvent{}, Seen: []string{}}
	coverageState(t, s, subscriptionStreams, subscriptionChannel, stream)
	observation := coverageRequest("@observe/rooms/connected", `{"projectId":"coverage","pluginId":"rooms","event":"connected","id":"new","payload":{"roomId":"room","reconnected":false}}`)
	if _, _, err := prepareSubscriptionStorageEffect(s, observation); !errors.Is(err, ErrModelConflict) {
		t.Fatal(err)
	}
	stream = storedSubscriptionStream{Sequence: subscriptionRetention, Events: []storedSubscriptionEvent{}, Seen: []string{}}
	for i := 1; i <= subscriptionRetention; i++ {
		stream.Events = append(stream.Events, storedSubscriptionEvent{Sequence: int64(i), Value: subscriptionConnection{RoomID: "room"}})
	}
	for i := 0; i < subscriptionRetention*2; i++ {
		stream.Seen = append(stream.Seen, "old")
	}
	coverageState(t, s, subscriptionStreams, subscriptionChannel, stream)
	_, commit, err := prepareSubscriptionStorageEffect(s, observation)
	if err != nil || commit == nil {
		t.Fatalf("observation admission: %v", err)
	}
	if err := commit(s.db); err != nil {
		t.Fatal(err)
	}
	got, _, _, err := subscriptionStream(s, "coverage")
	if err != nil || got.Sequence != int64(subscriptionRetention+1) || len(got.Events) != subscriptionRetention || len(got.Seen) != subscriptionRetention*2 || got.Seen[len(got.Seen)-1] != "new" {
		t.Fatalf("retention: %+v %v", got, err)
	}
	token := strings.Repeat("a", 32)
	ticket := storedSubscriptionTicket{ProjectID: "coverage", Channel: subscriptionChannel, Operation: "subscriptions.open", ExpiresAt: time.Now().Add(time.Hour).UnixMilli()}
	coverageState(t, s, subscriptionTickets, authTokenDigest(token), ticket)
	payload := `{"channel":"rooms.connected","ticket":"` + token + `"}`
	_, closeEffect, err := prepareSubscriptionStorageEffect(s, coverageRequest("subscriptions.close", payload))
	if err != nil || closeEffect == nil {
		t.Fatalf("close admission: %v", err)
	}
	failure := errors.New("injected failure")
	for _, executor := range []coverageSQLExecutor{{execErr: failure}, {resultErr: failure}} {
		if err := closeEffect(executor); !errors.Is(err, failure) {
			t.Fatal(err)
		}
	}
	if err := closeEffect(coverageSQLExecutor{}); !errors.Is(err, errSubscriptionTicket) {
		t.Fatal(err)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	if _, _, _, err := subscriptionStream(s, "coverage"); err == nil {
		t.Fatal("closed database accepted")
	}
}
