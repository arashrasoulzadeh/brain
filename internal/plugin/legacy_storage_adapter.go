package plugin

import (
	"encoding/json"
	"strings"
	"time"
)

type legacyStoragePlanner struct {
	storage *SQLiteStorage
}

func (p legacyStoragePlanner) Supports(pluginID string) bool {
	switch pluginID {
	case "files", "matchmaking", "rooms", "analytics", "tables":
		return true
	default:
		return false
	}
}

func (p legacyStoragePlanner) Prepare(request WorkerRequest) (WorkerRequest, func() error, error) {
	prepared, effect, err := prepareLegacyStorageRequest(p.storage, request)
	if err != nil || effect == nil {
		return prepared, nil, err
	}
	return prepared, func() error { return effect(p.storage.db) }, nil
}

// prepareWithObservations binds the existing storage effect and outbox inserts
// to one transaction. Neither is visible if the other fails.
func (p legacyStoragePlanner) prepareWithObservations(request WorkerRequest) (WorkerRequest, func([]observationDelivery) error, error) {
	prepared, effect, err := prepareLegacyStorageRequest(p.storage, request)
	if err != nil {
		return prepared, nil, err
	}
	return prepared, func(deliveries []observationDelivery) error {
		if effect == nil && len(deliveries) == 0 {
			return nil
		}
		tx, err := p.storage.db.Begin()
		if err != nil {
			return err
		}
		defer tx.Rollback()
		for _, delivery := range deliveries {
			if err := enqueueSQLiteObservation(tx, delivery, time.Now()); err != nil {
				return err
			}
		}
		if effect != nil {
			if err := effect(tx); err != nil {
				return err
			}
		}
		return tx.Commit()
	}, nil
}

// prepareLegacyStorageRequest is the compatibility adapter for existing product
// storage routes. Its writes are deferred to the registry commit boundary.
func prepareLegacyStorageRequest(storage *SQLiteStorage, request WorkerRequest) (WorkerRequest, func(sqliteExecutor) error, error) {
	var commit func(sqliteExecutor) error
	if storage == nil {
		return request, commit, nil
	}
	var payload map[string]any
	if json.Unmarshal(request.Payload, &payload) != nil || payload == nil {
		return request, commit, nil
	}
	if request.PluginID == "files" {
		key, _ := payload["key"].(string)
		key = strings.TrimSpace(key)
		if key == "" {
			return request, commit, nil
		}
		payload["key"] = key
		request.Payload, _ = json.Marshal(payload)
		if request.Operation == "files.upload" {
			if data, ok := payload["data"].(string); ok {
				commit = func(db sqliteExecutor) error {
					return putSQLiteFile(db, request.ProjectID, key, data, false)
				}
			}
		}
		if request.Operation == "files.publish" {
			data, _, err := storage.File(request.ProjectID, key)
			if err != nil {
				return request, nil, err
			}
			if data == "" {
				return request, nil, ErrInvalidWorkerRequest
			}
			if data != "" {
				commit = func(db sqliteExecutor) error {
					return putSQLiteFile(db, request.ProjectID, key, data, true)
				}
				payload["published"] = true
				request.Payload, _ = json.Marshal(payload)
			}
		}
		if request.Operation == "files.download" || request.Operation == "files.metadata" {
			data, published, err := storage.File(request.ProjectID, key)
			if err != nil {
				return request, nil, err
			}
			if data == "" {
				return request, nil, ErrInvalidWorkerRequest
			}
			if data != "" {
				payload["data"] = data
				payload["published"] = published
				request.Payload, _ = json.Marshal(payload)
			}
		}
		return request, commit, nil
	}
	if request.PluginID == "matchmaking" {
		id, _ := payload["id"].(string)
		id = strings.TrimSpace(id)
		if id == "" {
			return request, commit, nil
		}
		delete(payload, "ticket")
		payload["id"] = id
		state, err := storage.State(request.ProjectID, "matchmaking", id)
		if err != nil {
			return request, nil, err
		}
		if request.Operation == "matchmaking.queue" {
			state = map[string]any{"id": id, "status": "queued"}
			commit = func(db sqliteExecutor) error {
				return putSQLiteState(db, request.ProjectID, "matchmaking", id, state)
			}
		}
		if state == nil {
			return request, nil, ErrInvalidWorkerRequest
		}
		if request.Operation == "matchmaking.cancel" || request.Operation == "matchmaking.assign" {
			if state == nil {
				return request, commit, nil
			}
			if request.Operation == "matchmaking.cancel" {
				state["status"] = "cancelled"
			} else {
				state["status"] = "assigned"
			}
			commit = func(db sqliteExecutor) error {
				return putSQLiteState(db, request.ProjectID, "matchmaking", id, state)
			}
		}
		if state != nil {
			payload["ticket"] = state
			request.Payload, _ = json.Marshal(payload)
		}
		return request, commit, nil
	}
	if request.PluginID == "rooms" {
		id, _ := payload["id"].(string)
		id = strings.TrimSpace(id)
		if id == "" {
			return request, commit, nil
		}
		delete(payload, "room")
		payload["id"] = id
		state, err := storage.State(request.ProjectID, "rooms", id)
		if err != nil {
			return request, nil, err
		}
		if request.Operation == "rooms.create" {
			state = map[string]any{"id": id, "members": []any{}}
			commit = func(db sqliteExecutor) error {
				return putSQLiteState(db, request.ProjectID, "rooms", id, state)
			}
		}
		if state == nil {
			return request, nil, ErrInvalidWorkerRequest
		}
		if request.Operation == "rooms.join" || request.Operation == "rooms.reconnect" {
			if member, ok := payload["member"].(string); ok {
				member = strings.TrimSpace(member)
				if member == "" {
					return request, nil, ErrInvalidWorkerRequest
				}
				payload["member"] = member
				members, _ := state["members"].([]any)
				found := false
				for _, existing := range members {
					if existing == member {
						found = true
					}
				}
				if !found {
					members = append(members, member)
				}
				state["members"] = members
				commit = func(db sqliteExecutor) error {
					return putSQLiteState(db, request.ProjectID, "rooms", id, state)
				}
			}
		}
		if request.Operation == "rooms.result" {
			if value, ok := payload["result"]; ok {
				state["result"] = value
				commit = func(db sqliteExecutor) error {
					return putSQLiteState(db, request.ProjectID, "rooms", id, state)
				}
			}
		}
		payload["room"] = state
		request.Payload, _ = json.Marshal(payload)
		return request, commit, nil
	}
	if request.PluginID == "analytics" && (request.Operation == "analytics.capture" || request.Operation == "@observe/rooms/connected") {
		var event map[string]any
		if json.Unmarshal(request.Payload, &event) == nil && event != nil {
			if request.Operation == "@observe/rooms/connected" {
				if event["projectId"] != request.ProjectID || event["pluginId"] != "rooms" || event["event"] != "connected" {
					return request, nil, ErrInvalidWorkerRequest
				}
				event = map[string]any{"projectId": request.ProjectID, "name": "rooms.connected", "fields": event["payload"], "observationId": event["id"]}
			} else {
				event["projectId"] = request.ProjectID
				request.Payload, _ = json.Marshal(event)
			}
			state, err := storage.State(request.ProjectID, "analytics", "events")
			if err != nil {
				return request, nil, err
			}
			if state == nil {
				state = map[string]any{"events": []any{}}
			}
			events, _ := state["events"].([]any)
			if observationID, ok := event["observationId"].(string); ok && observationID != "" {
				for _, existing := range events {
					if record, ok := existing.(map[string]any); ok && record["observationId"] == observationID {
						return request, nil, nil
					}
				}
			}
			events = append(events, event)
			state["events"] = events
			commit = func(db sqliteExecutor) error {
				return putSQLiteState(db, request.ProjectID, "analytics", "events", state)
			}
		}
	}
	if request.PluginID != "tables" {
		return request, commit, nil
	}
	table, _ := payload["table"].(string)
	table = strings.TrimSpace(table)
	if table == "" {
		return request, commit, nil
	}
	payload["table"] = table
	request.Payload, _ = json.Marshal(payload)
	if request.Operation == "tables.write" {
		if row, ok := payload["row"].(map[string]any); ok {
			commit = func(db sqliteExecutor) error {
				return appendSQLiteRow(db, request.ProjectID, table, row)
			}
		}
	}
	if request.Operation == "tables.read" {
		rows, err := storage.Rows(request.ProjectID, table)
		if err != nil {
			return request, nil, err
		}
		payload["rows"] = rows
		request.Payload, _ = json.Marshal(payload)
	}
	return request, commit, nil
}
