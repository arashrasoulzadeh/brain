package plugin

import (
	"database/sql"
	"encoding/json"
	"errors"
	"slices"
	"time"
)

const subscriptionChannel = "rooms.connected"
const subscriptionTickets = "subscriptions.tickets"
const subscriptionStreams = "subscriptions.streams"
const subscriptionRetention = 256

var errSubscriptionTicket = errors.New("invalid subscription ticket")

type storedSubscriptionTicket struct {
	ProjectID string `json:"projectId"`
	Channel   string `json:"channel"`
	Operation string `json:"operation"`
	ExpiresAt int64  `json:"expiresAt"`
}

type subscriptionConnection struct {
	RoomID      string `json:"roomId"`
	Reconnected bool   `json:"reconnected"`
}

type storedSubscriptionEvent struct {
	Sequence int64                  `json:"sequence"`
	Value    subscriptionConnection `json:"value"`
}

type storedSubscriptionStream struct {
	Sequence int64                     `json:"sequence"`
	Events   []storedSubscriptionEvent `json:"events"`
	Seen     []string                  `json:"seen"`
}

func subscriptionStream(storage *SQLiteStorage, project string) (storedSubscriptionStream, string, bool, error) {
	state := storedSubscriptionStream{Events: []storedSubscriptionEvent{}, Seen: []string{}}
	var previous string
	err := storage.db.QueryRow(`SELECT value_json FROM plugin_state WHERE project_id=? AND namespace=? AND state_key=?`, project, subscriptionStreams, subscriptionChannel).Scan(&previous)
	if errors.Is(err, sql.ErrNoRows) {
		return state, "", false, nil
	}
	if err != nil {
		return state, "", false, err
	}
	if json.Unmarshal([]byte(previous), &state) != nil || state.Sequence < 0 || state.Sequence > 9007199254740991 ||
		state.Events == nil || state.Seen == nil || len(state.Events) > subscriptionRetention || len(state.Seen) > subscriptionRetention*2 {
		return state, "", false, ErrModelSchema
	}
	for i, event := range state.Events {
		if event.Sequence != state.Sequence-int64(len(state.Events))+int64(i)+1 || event.Value.RoomID == "" {
			return state, "", false, ErrModelSchema
		}
	}
	return state, previous, true, nil
}

func subscriptionTicketCount(db sqliteExecutor, project string, now int64, activeOnly bool) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM plugin_state WHERE project_id=? AND namespace=? AND json_extract(value_json,'$.channel')=?`
	args := []any{project, subscriptionTickets, subscriptionChannel}
	if activeOnly {
		query += ` AND json_extract(value_json,'$.expiresAt')>?`
		args = append(args, now)
	}
	err := db.QueryRow(query, args...).Scan(&count)
	return count, err
}

// The worker validates the admitted operation using host-owned snapshots.
// Ticket digests and the retained stream stay in separate durable namespaces.
func prepareSubscriptionStorageEffect(storage *SQLiteStorage, request WorkerRequest) (WorkerRequest, func(sqliteExecutor) error, error) {
	var input map[string]any
	if json.Unmarshal(request.Payload, &input) != nil || input == nil {
		return request, nil, ErrInvalidWorkerRequest
	}
	project := request.ProjectID
	stream, previous, exists, err := subscriptionStream(storage, project)
	if err != nil {
		return request, nil, err
	}
	snapshot := map[string]any{"stream": json.RawMessage(mustSubscriptionJSON(stream))}
	payload := map[string]any{"channel": subscriptionChannel, "_brainSubscriptions": snapshot}
	now := time.Now().UnixMilli()
	var effect func(sqliteExecutor) error
	if request.Operation == "@observe/rooms/connected" {
		if input["projectId"] != project || input["pluginId"] != "rooms" || input["event"] != "connected" {
			return request, nil, ErrInvalidWorkerRequest
		}
		id, _ := input["id"].(string)
		value, ok := input["payload"].(map[string]any)
		room, roomOK := value["roomId"].(string)
		reconnected, reconnectOK := value["reconnected"].(bool)
		if !ok || !roomOK || room == "" || len(room) > 1024 || !reconnectOK || len(id) > 256 {
			return request, nil, ErrInvalidWorkerRequest
		}
		payload = map[string]any{"id": id, "projectId": project, "pluginId": "rooms", "version": input["version"], "event": "connected", "payload": value, "_brainSubscriptions": snapshot}
		if id == "" || !slices.Contains(stream.Seen, id) {
			if stream.Sequence == 9007199254740991 {
				return request, nil, ErrModelConflict
			}
			stream.Sequence++
			stream.Events = append(stream.Events, storedSubscriptionEvent{Sequence: stream.Sequence, Value: subscriptionConnection{RoomID: room, Reconnected: reconnected}})
			if len(stream.Events) > subscriptionRetention {
				stream.Events = stream.Events[len(stream.Events)-subscriptionRetention:]
			}
			if id != "" {
				stream.Seen = append(stream.Seen, id)
				if len(stream.Seen) > subscriptionRetention*2 {
					stream.Seen = stream.Seen[len(stream.Seen)-subscriptionRetention*2:]
				}
			}
			next, err := json.Marshal(stream)
			if err != nil {
				return request, nil, err
			}
			effect = func(db sqliteExecutor) error {
				return compareAndSwapPluginState(db, project, subscriptionStreams, subscriptionChannel, previous, exists, string(next))
			}
		}
	} else {
		if input["channel"] != subscriptionChannel {
			return request, nil, ErrInvalidWorkerRequest
		}
		switch request.Operation {
		case "subscriptions.open":
			token, err := newModelRevision()
			if err != nil {
				return request, nil, err
			}
			record := storedSubscriptionTicket{ProjectID: project, Channel: subscriptionChannel, Operation: "subscriptions.open", ExpiresAt: now + 30000}
			snapshot["allocation"] = map[string]any{"ticket": token, "projectId": project, "channel": subscriptionChannel, "operation": record.Operation, "expiresAt": record.ExpiresAt}
			next, err := json.Marshal(record)
			if err != nil {
				return request, nil, err
			}
			effect = func(db sqliteExecutor) error {
				if _, err := db.Exec(`DELETE FROM plugin_state WHERE project_id=? AND namespace=? AND json_extract(value_json,'$.expiresAt')<=?`, project, subscriptionTickets, now); err != nil {
					return err
				}
				count, err := subscriptionTicketCount(db, project, now, false)
				if err != nil {
					return err
				}
				if count >= 4096 {
					return ErrModelConflict
				}
				return compareAndSwapPluginState(db, project, subscriptionTickets, authTokenDigest(token), "", false, string(next))
			}
		case "subscriptions.receive", "subscriptions.resync", "subscriptions.close":
			token, _ := input["ticket"].(string)
			if len(token) != 32 {
				return request, nil, errSubscriptionTicket
			}
			digest := authTokenDigest(token)
			var encoded string
			err := storage.db.QueryRow(`SELECT value_json FROM plugin_state WHERE project_id=? AND namespace=? AND state_key=?`, project, subscriptionTickets, digest).Scan(&encoded)
			if errors.Is(err, sql.ErrNoRows) {
				return request, nil, errSubscriptionTicket
			}
			if err != nil {
				return request, nil, err
			}
			var ticket storedSubscriptionTicket
			if json.Unmarshal([]byte(encoded), &ticket) != nil || ticket.ProjectID != project || ticket.Channel != subscriptionChannel || ticket.Operation != "subscriptions.open" || ticket.ExpiresAt <= now {
				return request, nil, errSubscriptionTicket
			}
			payload["ticket"] = token
			snapshot["ticket"] = map[string]any{"ticket": token, "projectId": project, "channel": subscriptionChannel, "operation": ticket.Operation, "expiresAt": ticket.ExpiresAt}
			if request.Operation == "subscriptions.receive" {
				payload["cursor"] = input["cursor"]
			}
			// Even a successful read rechecks the credential at the final
			// storage boundary so concurrent revocation cannot expose events.
			effect = func(db sqliteExecutor) error {
				if ticket.ExpiresAt <= time.Now().UnixMilli() {
					return errSubscriptionTicket
				}
				if request.Operation != "subscriptions.close" {
					var current string
					if err := db.QueryRow(`SELECT value_json FROM plugin_state WHERE project_id=? AND namespace=? AND state_key=?`, project, subscriptionTickets, digest).Scan(&current); err != nil || current != encoded {
						return errSubscriptionTicket
					}
					return nil
				}
				result, err := db.Exec(`DELETE FROM plugin_state WHERE project_id=? AND namespace=? AND state_key=? AND value_json=?`, project, subscriptionTickets, digest, encoded)
				if err != nil {
					return err
				}
				n, err := result.RowsAffected()
				if err != nil {
					return err
				}
				if n != 1 {
					return errSubscriptionTicket
				}
				return nil
			}
		case "subscriptions.inspect":
			count, err := subscriptionTicketCount(storage.db, project, now, true)
			if err != nil {
				return request, nil, err
			}
			snapshot["count"] = count
		case "subscriptions.revoke":
			count, err := subscriptionTicketCount(storage.db, project, now, false)
			if err != nil {
				return request, nil, err
			}
			snapshot["count"] = count
			effect = func(db sqliteExecutor) error {
				result, err := db.Exec(`DELETE FROM plugin_state WHERE project_id=? AND namespace=? AND json_extract(value_json,'$.channel')=?`, project, subscriptionTickets, subscriptionChannel)
				if err != nil {
					return err
				}
				n, err := result.RowsAffected()
				if err != nil {
					return err
				}
				if n != count {
					return ErrModelConflict
				}
				return nil
			}
		default:
			return request, nil, ErrInvalidWorkerRequest
		}
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return request, nil, err
	}
	request.Payload = encoded
	return request, effect, nil
}

// The typed stream contains only bounded JSON primitives, so it is safe to
// encode independently before computing the proposed mutation.
func mustSubscriptionJSON(state storedSubscriptionStream) []byte {
	encoded, _ := json.Marshal(state)
	return encoded
}
