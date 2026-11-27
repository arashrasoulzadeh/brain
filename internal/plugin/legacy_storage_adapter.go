package plugin

import (
	"encoding/json"
	"errors"
	"strings"
)

var ErrStoredResourceMissing = errors.New("stored resource not found")

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
	return prepareLegacyStorageRequest(p.storage, request)
}

// Prepare only host-owned storage snapshots. Never let a caller provide the
// stored state consumed by the worker, even when the record does not exist.
func prepareLegacyStorageRequest(storage *SQLiteStorage, request WorkerRequest) (WorkerRequest, func() error, error) {
	if storage == nil {
		return request, nil, nil
	}
	var keyField string
	switch request.PluginID + "/" + request.Operation {
	case "files/files.upload", "files/files.publish", "files/files.download", "files/files.metadata":
		keyField = "key"
	case "matchmaking/matchmaking.queue", "matchmaking/matchmaking.cancel", "matchmaking/matchmaking.assign", "matchmaking/matchmaking.status":
		keyField = "id"
	case "rooms/rooms.create", "rooms/rooms.join", "rooms/rooms.reconnect", "rooms/rooms.result":
		keyField = "id"
	case "tables/tables.write", "tables/tables.read", "tables/tables.schema":
		keyField = "table"
	case "analytics/analytics.capture", "analytics/@observe/rooms/connected":
	default:
		return request, nil, nil
	}

	var input map[string]any
	if json.Unmarshal(request.Payload, &input) != nil || input == nil {
		return request, nil, ErrInvalidWorkerRequest
	}
	key := ""
	payload := map[string]any{}
	if keyField != "" {
		value, _ := input[keyField].(string)
		key = strings.TrimSpace(value)
		if key == "" {
			return request, nil, ErrInvalidWorkerRequest
		}
		payload[keyField] = key
	}
	var commit func() error
	switch request.PluginID {
	case "files":
		if request.Operation == "files.upload" {
			data, ok := input["data"].(string)
			if !ok || data == "" {
				return request, nil, ErrInvalidWorkerRequest
			}
			payload["data"] = data
			commit = func() error { return storage.PutFile(request.ProjectID, key, data, false) }
		} else {
			data, published, err := storage.File(request.ProjectID, key)
			if err != nil {
				return request, nil, err
			}
			if data == "" {
				return request, nil, ErrStoredResourceMissing
			}
			if request.Operation == "files.publish" {
				published = true
				commit = func() error { return storage.PutFile(request.ProjectID, key, data, true) }
			}
			payload["data"], payload["published"] = data, published
		}
	case "matchmaking":
		state, err := storage.State(request.ProjectID, "matchmaking", key)
		if err != nil {
			return request, nil, err
		}
		if request.Operation == "matchmaking.queue" {
			state = map[string]any{"id": key, "status": "queued"}
		}
		if state == nil {
			return request, nil, ErrStoredResourceMissing
		}
		if request.Operation == "matchmaking.cancel" {
			state["status"] = "cancelled"
		} else if request.Operation == "matchmaking.assign" {
			state["status"] = "assigned"
		}
		if request.Operation != "matchmaking.status" {
			commit = func() error { return storage.PutState(request.ProjectID, "matchmaking", key, state) }
		}
		payload["ticket"] = state
	case "rooms":
		state, err := storage.State(request.ProjectID, "rooms", key)
		if err != nil {
			return request, nil, err
		}
		if request.Operation == "rooms.create" {
			state = map[string]any{"id": key, "members": []any{}}
			commit = func() error { return storage.PutState(request.ProjectID, "rooms", key, state) }
		}
		if state == nil {
			return request, nil, ErrStoredResourceMissing
		}
		if request.Operation == "rooms.join" || request.Operation == "rooms.reconnect" {
			member, _ := input["member"].(string)
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
			commit = func() error { return storage.PutState(request.ProjectID, "rooms", key, state) }
		}
		if request.Operation == "rooms.result" {
			if value, ok := input["result"]; ok {
				state["result"] = value
				payload["result"] = value
				commit = func() error { return storage.PutState(request.ProjectID, "rooms", key, state) }
			}
		}
		payload["room"] = state
	case "analytics":
		var event map[string]any
		if request.Operation == "@observe/rooms/connected" {
			if input["projectId"] != request.ProjectID || input["pluginId"] != "rooms" || input["event"] != "connected" {
				return request, nil, ErrInvalidWorkerRequest
			}
			event = map[string]any{"projectId": request.ProjectID, "name": "rooms.connected", "fields": input["payload"]}
			payload = input
		} else {
			if identity, supplied := input["projectId"]; supplied && identity != request.ProjectID {
				return request, nil, ErrInvalidWorkerRequest
			}
			event = map[string]any{"projectId": request.ProjectID, "name": input["name"], "fields": input["fields"]}
			payload = event
		}
		state, err := storage.State(request.ProjectID, "analytics", "events")
		if err != nil {
			return request, nil, err
		}
		if state == nil {
			state = map[string]any{"events": []any{}}
		}
		events, _ := state["events"].([]any)
		state["events"] = append(events, event)
		commit = func() error { return storage.PutState(request.ProjectID, "analytics", "events", state) }
	case "tables":
		if request.Operation == "tables.write" {
			row, ok := input["row"].(map[string]any)
			if !ok || row == nil {
				return request, nil, ErrInvalidWorkerRequest
			}
			payload["row"] = row
			commit = func() error { return storage.Append(request.ProjectID, key, row) }
		} else {
			rows, err := storage.Rows(request.ProjectID, key)
			if err != nil {
				return request, nil, err
			}
			payload["rows"] = rows
		}
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return request, nil, err
	}
	request.Payload = encoded
	return request, commit, nil
}
