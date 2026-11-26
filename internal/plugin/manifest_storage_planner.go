package plugin

import (
	"encoding/json"
	"strings"
)

// Manifest storage is bound to the registered version and its declared route.
// Only the registry may invoke it, after checking the route capability.
type manifestStoragePlanner struct {
	registry *Registry
	storage  *SQLiteStorage
}

func (p manifestStoragePlanner) Supports(id string) bool {
	p.registry.readinessMu.RLock()
	defer p.registry.readinessMu.RUnlock()
	for _, registered := range p.registry.registered {
		if registered.Manifest.ID != id {
			continue
		}
		for _, route := range registered.Manifest.Routes {
			if route.Storage != nil {
				return true
			}
		}
	}
	return false
}

func (p manifestStoragePlanner) Prepare(request WorkerRequest) (WorkerRequest, func() error, error) {
	p.registry.readinessMu.RLock()
	registered, ok := p.registry.registered[request.PluginID+"@"+request.Version]
	route, declared := registered.Manifest.Route(request.Operation)
	p.registry.readinessMu.RUnlock()
	if !ok || !declared || route.Storage == nil {
		return request, nil, nil
	}
	spec := route.Storage
	if !spec.valid() || !isValidProjectID(request.PluginID) {
		return request, nil, ErrInvalidManifest
	}
	var payload map[string]any
	if json.Unmarshal(request.Payload, &payload) != nil || payload == nil {
		return request, nil, nil
	}
	key := spec.FixedKey
	if spec.Key != "" {
		key, _ = payload[spec.Key].(string)
		key = strings.TrimSpace(key)
		payload[spec.Key] = key
	}
	if key == "" {
		return request, nil, nil
	}
	// Derived response fields are never accepted from the caller.
	if spec.Target != "" {
		delete(payload, spec.Target)
	}
	if spec.Action == "blob.read" {
		delete(payload, spec.Field)
	}
	project, namespace := request.ProjectID, request.PluginID
	var commit func() error
	switch spec.Action {
	case "blob.put":
		if data, ok := payload[spec.Input].(string); ok {
			commit = func() error { return p.storage.putScopedFile(project, namespace, key, data, false) }
		}
	case "blob.read", "blob.publish":
		data, published, err := p.storage.scopedFile(project, namespace, key)
		if err != nil {
			return request, nil, err
		}
		if data != "" {
			if spec.Action == "blob.publish" {
				commit = func() error { return p.storage.putScopedFile(project, namespace, key, data, true) }
				payload[spec.Target] = true
			} else {
				payload[spec.Target], payload[spec.Field] = data, published
			}
		}
	case "rows.append":
		if row, ok := payload[spec.Input].(map[string]any); ok {
			commit = func() error { return p.storage.appendScopedRow(project, namespace, key, row) }
		}
	case "rows.read":
		rows, err := p.storage.scopedRows(project, namespace, key)
		if err != nil {
			return request, nil, err
		}
		payload[spec.Target] = rows
	default:
		state, err := p.storage.State(project, namespace, key)
		if err != nil {
			return request, nil, err
		}
		state, changed := applyStateStorage(spec, state, payload)
		if changed {
			commit = func() error { return p.storage.PutState(project, namespace, key, state) }
		}
		if state != nil && spec.Target != "" {
			payload[spec.Target] = state
		}
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return request, nil, err
	}
	request.Payload = encoded
	return request, commit, nil
}

func applyStateStorage(spec *StorageSpec, state, payload map[string]any) (map[string]any, bool) {
	switch spec.Action {
	case "state.create":
		return storageValues(spec.Values, payload), true
	case "state.append":
		if state == nil {
			state = map[string]any{}
		}
		items, _ := state[spec.Field].([]any)
		state[spec.Field] = append(items, payload)
		return state, true
	}
	if state == nil {
		return nil, false
	}
	switch spec.Action {
	case "state.patch":
		if spec.Input != "" {
			if value, ok := payload[spec.Input]; ok {
				state[spec.Field] = value
				return state, true
			}
			return state, false
		}
		for field, value := range storageValues(spec.Values, payload) {
			state[field] = value
		}
		return state, true
	case "state.add_unique":
		if value, ok := payload[spec.Input].(string); ok {
			value = strings.TrimSpace(value)
			payload[spec.Input] = value
			if value == "" {
				return state, false
			}
			items, _ := state[spec.Field].([]any)
			for _, item := range items {
				if existing, ok := item.(string); ok && existing == value {
					return state, false
				}
			}
			state[spec.Field] = append(items, value)
			return state, true
		}
	}
	return state, false
}

func storageValues(values, payload map[string]any) map[string]any {
	state := make(map[string]any, len(values))
	for field, value := range values {
		if reference, ok := value.(string); ok && strings.HasPrefix(reference, "$input.") {
			state[field] = payload[strings.TrimPrefix(reference, "$input.")]
		} else if _, ok := value.([]any); ok {
			state[field] = []any{}
		} else {
			state[field] = value
		}
	}
	return state
}
