package plugin

import (
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
)

// Arena state is keyed by the admitted project and room. The worker remains
// responsible for domain validation; only a successful invocation can commit
// its proposed state and observations. Never forward a caller-owned snapshot.
func prepareArenaStorageEffect(storage *SQLiteStorage, request WorkerRequest) (WorkerRequest, func(sqliteExecutor) error, error) {
	switch request.Operation {
	case "arena.start", "arena.input", "arena.snapshot", "arena.result", "arena.connect", "arena.player.input":
	default:
		return request, nil, nil
	}
	var input map[string]any
	if json.Unmarshal(request.Payload, &input) != nil || input == nil {
		return request, nil, ErrInvalidWorkerRequest
	}
	room, _ := input["roomId"].(string)
	if room != strings.TrimSpace(room) || !isValidProjectID(room) || len(room) > 128 {
		return request, nil, ErrInvalidWorkerRequest
	}
	var previous string
	err := storage.db.QueryRow(`SELECT value_json FROM plugin_state WHERE project_id=? AND namespace='arena' AND state_key=?`, request.ProjectID, room).Scan(&previous)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return request, nil, err
	}
	exists := err == nil
	var state map[string]any
	if exists {
		if json.Unmarshal([]byte(previous), &state) != nil || state == nil {
			return request, nil, ErrModelSchema
		}
	}
	if request.Operation == "arena.start" && exists {
		return request, nil, ErrModelConflict
	}
	if request.Operation != "arena.start" && !exists {
		return request, nil, ErrStoredResourceMissing
	}
	// Capture the original snapshot before constructing the proposed mutation.
	payload := map[string]any{"roomId": room, "_brainArena": state}
	if exists {
		payload["_brainArena"] = json.RawMessage(previous)
	}
	switch request.Operation {
	case "arena.start":
		payload["players"] = input["players"]
		state = map[string]any{"state": "active", "players": input["players"], "sequences": map[string]any{}}
	case "arena.input", "arena.player.input":
		player, _ := input["playerId"].(string)
		player = strings.TrimSpace(player)
		if request.Operation == "arena.player.input" {
			actor := invocationActor(request.accessPrincipal)
			if actor.Kind != "user" || !actor.Authenticated || !strings.HasPrefix(actor.ID, request.ProjectID+":") || (input["playerId"] != nil && input["playerId"] != actor.ID) {
				return request, nil, ErrActorDenied
			}
			player = actor.ID
		}
		payload["playerId"], payload["sequence"], payload["payload"] = player, input["sequence"], input["payload"]
		sequences, ok := state["sequences"].(map[string]any)
		if !ok {
			return request, nil, ErrModelSchema
		}
		sequences[player] = input["sequence"]
	case "arena.result":
		payload["winnerIds"], payload["completedAt"] = input["winnerIds"], input["completedAt"]
		completed, _ := input["completedAt"].(string)
		state["state"] = "finished"
		state["result"] = map[string]any{"winnerIds": input["winnerIds"], "completedAt": strings.TrimSpace(completed)}
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return request, nil, err
	}
	request.Payload = encoded
	if request.Operation == "arena.snapshot" || request.Operation == "arena.connect" {
		return request, nil, nil
	}
	// Freeze the proposed bytes. The CAS prevents stale writes across host
	// processes as well as duplicate starts. The outbox shares this transaction.
	next, err := json.Marshal(state)
	if err != nil {
		return request, nil, err
	}
	return request, func(db sqliteExecutor) error {
		return compareAndSwapPluginState(db, request.ProjectID, "arena", room, previous, exists, string(next))
	}, nil
}

// compareAndSwapPluginState is shared by durable snapshot adapters. The key
// and namespace come from the trusted planner, never from a worker response.
func compareAndSwapPluginState(db sqliteExecutor, projectID, namespace, key, previous string, exists bool, next string) error {
	var result sql.Result
	var err error
	if exists {
		result, err = db.Exec(`UPDATE plugin_state SET value_json=? WHERE project_id=? AND namespace=? AND state_key=? AND value_json=?`, next, projectID, namespace, key, previous)
	} else {
		result, err = db.Exec(`INSERT INTO plugin_state(project_id,namespace,state_key,value_json) VALUES(?,?,?,?) ON CONFLICT(project_id,namespace,state_key) DO NOTHING`, projectID, namespace, key, next)
	}
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n != 1 {
		return ErrModelConflict
	}
	return nil
}
