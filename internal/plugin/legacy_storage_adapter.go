package plugin

import "encoding/json"

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

// prepareLegacyStorageRequest is the compatibility adapter for existing product
// storage routes. Its writes are deferred to the registry commit boundary.
func prepareLegacyStorageRequest(storage *SQLiteStorage, request WorkerRequest) (WorkerRequest, func() error, error) {
	var commit func() error
	if storage == nil {
		return request, commit, nil
	}
	var payload map[string]any
	if json.Unmarshal(request.Payload, &payload) != nil {
		return request, commit, nil
	}
	if request.PluginID == "files" {
		key, _ := payload["key"].(string)
		if key == "" {
			return request, commit, nil
		}
		if request.Operation == "files.upload" {
			if data, ok := payload["data"].(string); ok {
				commit = func() error { return storage.PutFile(request.ProjectID, key, data, false) }
			}
		}
		if request.Operation == "files.publish" {
			data, _, err := storage.File(request.ProjectID, key)
			if err != nil {
				return request, nil, err
			}
			if data != "" {
				commit = func() error { return storage.PutFile(request.ProjectID, key, data, true) }
				payload["published"] = true
				request.Payload, _ = json.Marshal(payload)
			}
		}
		if request.Operation == "files.download" || request.Operation == "files.metadata" {
			data, published, err := storage.File(request.ProjectID, key)
			if err != nil {
				return request, nil, err
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
		if id == "" {
			return request, commit, nil
		}
		state, err := storage.State(request.ProjectID, "matchmaking", id)
		if err != nil {
			return request, nil, err
		}
		if request.Operation == "matchmaking.queue" {
			state = map[string]any{"id": id, "status": "queued"}
			commit = func() error { return storage.PutState(request.ProjectID, "matchmaking", id, state) }
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
			commit = func() error { return storage.PutState(request.ProjectID, "matchmaking", id, state) }
		}
		if state != nil {
			payload["ticket"] = state
			request.Payload, _ = json.Marshal(payload)
		}
		return request, commit, nil
	}
	if request.PluginID == "rooms" {
		id, _ := payload["id"].(string)
		if id == "" {
			return request, commit, nil
		}
		state, err := storage.State(request.ProjectID, "rooms", id)
		if err != nil {
			return request, nil, err
		}
		if request.Operation == "rooms.create" {
			state = map[string]any{"id": id, "members": []any{}}
			commit = func() error { return storage.PutState(request.ProjectID, "rooms", id, state) }
		}
		if state == nil {
			return request, commit, nil
		}
		if request.Operation == "rooms.join" || request.Operation == "rooms.reconnect" {
			if member, ok := payload["member"].(string); ok {
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
				commit = func() error { return storage.PutState(request.ProjectID, "rooms", id, state) }
			}
		}
		if request.Operation == "rooms.result" {
			if value, ok := payload["result"]; ok {
				state["result"] = value
				commit = func() error { return storage.PutState(request.ProjectID, "rooms", id, state) }
			}
		}
		payload["room"] = state
		request.Payload, _ = json.Marshal(payload)
		return request, commit, nil
	}
	if request.PluginID == "analytics" && (request.Operation == "analytics.capture" || request.Operation == "@observe/rooms/connected") {
		var event map[string]any
		if json.Unmarshal(request.Payload, &event) == nil {
			if request.Operation == "@observe/rooms/connected" {
				if event["projectId"] != request.ProjectID || event["pluginId"] != "rooms" || event["event"] != "connected" {
					return request, nil, ErrInvalidWorkerRequest
				}
				event = map[string]any{"projectId": request.ProjectID, "name": "rooms.connected", "fields": event["payload"]}
			} else {
				event["projectId"] = request.ProjectID
			}
			state, err := storage.State(request.ProjectID, "analytics", "events")
			if err != nil {
				return request, nil, err
			}
			if state == nil {
				state = map[string]any{"events": []any{}}
			}
			events, _ := state["events"].([]any)
			events = append(events, event)
			state["events"] = events
			commit = func() error { return storage.PutState(request.ProjectID, "analytics", "events", state) }
		}
	}
	if request.PluginID != "tables" {
		return request, commit, nil
	}
	table, _ := payload["table"].(string)
	if table == "" {
		return request, commit, nil
	}
	if request.Operation == "tables.write" {
		if row, ok := payload["row"].(map[string]any); ok {
			commit = func() error { return storage.Append(request.ProjectID, table, row) }
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
