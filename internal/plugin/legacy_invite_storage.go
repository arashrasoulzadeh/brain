package plugin

import (
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

const inviteNamespace = "invites"

var errInviteUnavailable = errors.New("invite unavailable")

type storedInvite struct {
	ID             string `json:"id"`
	CodeDigest     string `json:"codeDigest,omitempty"`
	ProjectID      string `json:"projectId"`
	RoomID         string `json:"roomId"`
	TargetPlayerID string `json:"targetPlayerId"`
	ExpiresAt      int64  `json:"expiresAt"`
	Used           bool   `json:"used"`
}

type inviteSummary struct {
	ID             string `json:"id"`
	ProjectID      string `json:"projectId"`
	RoomID         string `json:"roomId"`
	TargetPlayerID string `json:"targetPlayerId"`
	ExpiresAt      int64  `json:"expiresAt"`
	Used           bool   `json:"used"`
}

func inviteID(value any) string {
	id, _ := value.(string)
	return strings.TrimSpace(id)
}

func validInviteScope(value string) bool {
	return len(value) <= 128 && isValidProjectID(value)
}

func prepareInviteStorageEffect(storage *SQLiteStorage, request WorkerRequest) (WorkerRequest, func(sqliteExecutor) error, error) {
	var input map[string]any
	if json.Unmarshal(request.Payload, &input) != nil || input == nil {
		return request, nil, ErrInvalidWorkerRequest
	}
	project := request.ProjectID
	room := inviteID(input["roomId"])
	if !validInviteScope(room) {
		return request, nil, ErrInvalidWorkerRequest
	}
	payload := map[string]any{"roomId": room}
	snapshot := map[string]any{}
	payload["_brainInvites"] = snapshot
	now := time.Now().UnixMilli()
	var effect func(sqliteExecutor) error
	switch request.Operation {
	case "invites.create":
		target := input["targetPlayerId"]
		if target == nil {
			target = input["target"]
		}
		player := inviteID(target)
		ttl := float64(60)
		if input["ttlSeconds"] != nil {
			value, ok := input["ttlSeconds"].(float64)
			if !ok || value < 1 || value > 86400 || value != float64(int(value)) {
				return request, nil, ErrInvalidWorkerRequest
			}
			ttl = value
		}
		if !validInviteScope(player) {
			return request, nil, ErrInvalidWorkerRequest
		}
		idPart, err := newModelRevision()
		if err != nil {
			return request, nil, err
		}
		codePart, err := newModelRevision()
		if err != nil {
			return request, nil, err
		}
		record := storedInvite{ID: project + ":invite-" + idPart, CodeDigest: authTokenDigest("INVITE-" + project + "-" + codePart), ProjectID: project, RoomID: room, TargetPlayerID: player, ExpiresAt: now + int64(ttl)*1000}
		if len(record.ID) > 256 || len("INVITE-"+project+"-"+codePart) > 256 {
			return request, nil, ErrInvalidWorkerRequest
		}
		payload["targetPlayerId"], payload["ttlSeconds"] = player, ttl
		snapshot["idPart"], snapshot["codePart"], snapshot["now"] = idPart, codePart, now
		next, err := json.Marshal(record)
		if err != nil {
			return request, nil, err
		}
		effect = func(db sqliteExecutor) error {
			if _, err := db.Exec(`DELETE FROM plugin_state WHERE project_id=? AND namespace=? AND json_extract(value_json,'$.expiresAt')<=?`, project, inviteNamespace, now); err != nil {
				return err
			}
			var count int
			if err := db.QueryRow(`SELECT COUNT(*) FROM plugin_state WHERE project_id=? AND namespace=?`, project, inviteNamespace).Scan(&count); err != nil {
				return err
			}
			if count >= 4096 {
				return ErrModelConflict
			}
			return compareAndSwapPluginState(db, project, inviteNamespace, record.ID, "", false, string(next))
		}
	case "invites.accept", "invites.cancel":
		id, _ := input["inviteId"].(string)
		code, _ := input["code"].(string)
		if request.Operation == "invites.cancel" {
			id = strings.TrimSpace(id)
			if code != "" {
				return request, nil, errInviteUnavailable
			}
		}
		if (id == "") == (code == "") || len(id) > 256 || len(code) > 256 {
			return request, nil, errInviteUnavailable
		}
		var previous string
		var err error
		if id != "" {
			err = storage.db.QueryRow(`SELECT value_json FROM plugin_state WHERE project_id=? AND namespace=? AND state_key=? AND json_extract(value_json,'$.roomId')=?`, project, inviteNamespace, id, room).Scan(&previous)
		} else {
			err = storage.db.QueryRow(`SELECT value_json FROM plugin_state WHERE project_id=? AND namespace=? AND json_extract(value_json,'$.codeDigest')=? AND json_extract(value_json,'$.roomId')=?`, project, inviteNamespace, authTokenDigest(code), room).Scan(&previous)
		}
		if errors.Is(err, sql.ErrNoRows) {
			return request, nil, errInviteUnavailable
		}
		if err != nil {
			return request, nil, err
		}
		var record storedInvite
		if json.Unmarshal([]byte(previous), &record) != nil || record.ProjectID != project || record.RoomID != room || record.Used || record.ExpiresAt <= now || record.ID == "" || (id != "" && record.ID != id) {
			return request, nil, errInviteUnavailable
		}
		// Supply only the single admitted invitation, never its stored digest.
		snapshot["records"] = []map[string]any{{"id": record.ID, "code": code, "projectId": project, "roomId": room, "targetPlayerId": record.TargetPlayerID, "expiresAt": record.ExpiresAt, "used": record.Used}}
		if request.Operation == "invites.accept" {
			player := inviteID(input["playerId"])
			if player != record.TargetPlayerID {
				return request, nil, errInviteUnavailable
			}
			payload["playerId"], payload["inviteId"], payload["code"] = player, id, code
			record.Used = true
			next, err := json.Marshal(record)
			if err != nil {
				return request, nil, err
			}
			effect = func(db sqliteExecutor) error {
				if record.ExpiresAt <= time.Now().UnixMilli() {
					return errInviteUnavailable
				}
				return compareAndSwapPluginState(db, project, inviteNamespace, record.ID, previous, true, string(next))
			}
		} else {
			payload["inviteId"] = id
			effect = func(db sqliteExecutor) error {
				if record.ExpiresAt <= time.Now().UnixMilli() {
					return errInviteUnavailable
				}
				result, err := db.Exec(`DELETE FROM plugin_state WHERE project_id=? AND namespace=? AND state_key=? AND value_json=?`, project, inviteNamespace, record.ID, previous)
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
		}
	case "invites.inspect":
		after, ok := input["after"].(string)
		if (input["after"] != nil && !ok) || len(after) > 256 {
			return request, nil, ErrInvalidWorkerRequest
		}
		rows, err := storage.db.Query(`SELECT state_key,json_extract(value_json,'$.targetPlayerId'),json_extract(value_json,'$.expiresAt'),json_extract(value_json,'$.used') FROM plugin_state WHERE project_id=? AND namespace=? AND json_extract(value_json,'$.roomId')=? AND state_key>? ORDER BY state_key LIMIT 100`, project, inviteNamespace, room, after)
		if err != nil {
			return request, nil, err
		}
		defer rows.Close()
		records := []inviteSummary{}
		for rows.Next() {
			record := inviteSummary{ProjectID: project, RoomID: room}
			if err := rows.Scan(&record.ID, &record.TargetPlayerID, &record.ExpiresAt, &record.Used); err != nil {
				return request, nil, err
			}
			records = append(records, record)
		}
		if err := rows.Err(); err != nil {
			return request, nil, err
		}
		snapshot["records"] = records
	default:
		return request, nil, ErrInvalidWorkerRequest
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return request, nil, err
	}
	request.Payload = encoded
	return request, effect, nil
}
