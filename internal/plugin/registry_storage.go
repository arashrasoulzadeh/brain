package plugin

import "encoding/json"

func (r *Registry) BindSQLiteStorage(storage *SQLiteStorage) {
	r.workerMu.Lock()
	defer r.workerMu.Unlock()
	r.storage = storage
	installations, readiness, err := storage.LoadRegistryState()
	if err != nil {
		return
	}
	r.installationMu.Lock()
	r.readinessMu.Lock()
	for key, installation := range installations {
		if _, registered := r.registered[installation.PluginID+"@"+installation.Version]; registered {
			r.installations[key] = installation
		}
	}
	for key, ready := range readiness {
		if _, registered := r.registered[key]; registered {
			r.readiness[key] = ready
		}
	}
	r.readinessMu.Unlock()
	r.installationMu.Unlock()
}

func (r *Registry) persistLifecycleState() {
	if r.storage == nil {
		return
	}
	_ = r.storage.SaveRegistryState(r.installations, r.readiness)
}

func (r *Registry) prepareStorageRequest(request WorkerRequest) WorkerRequest {
	if r.storage == nil {
		return request
	}
	var payload map[string]any
	if json.Unmarshal(request.Payload, &payload) != nil {
		return request
	}
	if request.PluginID == "files" {
		key, _ := payload["key"].(string)
		if key == "" {
			return request
		}
		if request.Operation == "files.upload" {
			if data, ok := payload["data"].(string); ok {
				_ = r.storage.PutFile(request.ProjectID, key, data, false)
			}
		}
		if request.Operation == "files.publish" {
			if data, _, err := r.storage.File(request.ProjectID, key); err == nil && data != "" {
				_ = r.storage.PutFile(request.ProjectID, key, data, true)
				payload["published"] = true
				request.Payload, _ = json.Marshal(payload)
			}
		}
		if request.Operation == "files.download" || request.Operation == "files.metadata" {
			if data, published, err := r.storage.File(request.ProjectID, key); err == nil && data != "" {
				payload["data"] = data
				payload["published"] = published
				request.Payload, _ = json.Marshal(payload)
			}
		}
		return request
	}
	if request.PluginID == "matchmaking" {
		id, _ := payload["id"].(string)
		if id == "" {
			return request
		}
		state, _ := r.storage.State(request.ProjectID, "matchmaking", id)
		if request.Operation == "matchmaking.queue" {
			state = map[string]any{"id": id, "status": "queued"}
			_ = r.storage.PutState(request.ProjectID, "matchmaking", id, state)
		}
		if request.Operation == "matchmaking.cancel" || request.Operation == "matchmaking.assign" {
			if state == nil {
				return request
			}
			if request.Operation == "matchmaking.cancel" {
				state["status"] = "cancelled"
			} else {
				state["status"] = "assigned"
			}
			_ = r.storage.PutState(request.ProjectID, "matchmaking", id, state)
		}
		if state != nil {
			payload["ticket"] = state
			request.Payload, _ = json.Marshal(payload)
		}
		return request
	}
	if request.PluginID == "rooms" {
		id, _ := payload["id"].(string)
		if id == "" {
			return request
		}
		state, _ := r.storage.State(request.ProjectID, "rooms", id)
		if request.Operation == "rooms.create" {
			state = map[string]any{"id": id, "members": []any{}}
			_ = r.storage.PutState(request.ProjectID, "rooms", id, state)
		}
		if state == nil {
			return request
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
				_ = r.storage.PutState(request.ProjectID, "rooms", id, state)
			}
		}
		if request.Operation == "rooms.result" {
			if value, ok := payload["result"]; ok {
				state["result"] = value
				_ = r.storage.PutState(request.ProjectID, "rooms", id, state)
			}
		}
		payload["room"] = state
		request.Payload, _ = json.Marshal(payload)
		return request
	}
	if request.PluginID == "analytics" && request.Operation == "analytics.capture" {
		var event map[string]any
		if json.Unmarshal(request.Payload, &event) == nil {
			state, _ := r.storage.State(request.ProjectID, "analytics", "events")
			if state == nil {
				state = map[string]any{"events": []any{}}
			}
			events, _ := state["events"].([]any)
			events = append(events, event)
			state["events"] = events
			_ = r.storage.PutState(request.ProjectID, "analytics", "events", state)
		}
	}
	if request.PluginID != "tables" {
		return request
	}
	table, _ := payload["table"].(string)
	if table == "" {
		return request
	}
	if request.Operation == "tables.write" {
		if row, ok := payload["row"].(map[string]any); ok {
			_ = r.storage.Append(request.ProjectID, table, row)
		}
	}
	if request.Operation == "tables.read" {
		if rows, err := r.storage.Rows(request.ProjectID, table); err == nil {
			payload["rows"] = rows
			request.Payload, _ = json.Marshal(payload)
		}
	}
	return request
}

func (r *Registry) SaveDurableData(projectID, pluginID, key string, value []byte) error {
	if !isValidProjectID(projectID) || pluginID == "" || key == "" {
		return ErrProjectInvalid
	}
	r.durableMu.Lock()
	defer r.durableMu.Unlock()
	r.durableData[projectID+":"+pluginID+":"+key] = append([]byte(nil), value...)
	return nil
}

func (r *Registry) DurableData(projectID, pluginID, key string) ([]byte, bool) {
	r.durableMu.RLock()
	defer r.durableMu.RUnlock()
	value, ok := r.durableData[projectID+":"+pluginID+":"+key]
	return append([]byte(nil), value...), ok
}

func (r *Registry) DeleteDurableData(projectID, pluginID, key string) error {
	if !isValidProjectID(projectID) || pluginID == "" || key == "" {
		return ErrProjectInvalid
	}
	r.durableMu.Lock()
	defer r.durableMu.Unlock()
	delete(r.durableData, projectID+":"+pluginID+":"+key)
	r.appendAudit(AuditEvent{Action: "delete_durable_data", ProjectID: projectID, PluginID: pluginID, Detail: key})
	return nil
}
