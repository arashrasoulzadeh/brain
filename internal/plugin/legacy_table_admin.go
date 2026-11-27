package plugin

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
)

var ErrTableRevision = errors.New("table changed; read a fresh page before editing")

// These operations use durable snapshots, not the worker's local fixture store.
func needsTablePage(request WorkerRequest) bool {
	if request.PluginID != "tables" || request.Operation != "tables.read" {
		return false
	}
	var input map[string]json.RawMessage
	if json.Unmarshal(request.Payload, &input) != nil {
		return false
	}
	_, offset := input["offset"]
	_, limit := input["limit"]
	return offset || limit
}

func requiresDurableAdminSnapshot(request WorkerRequest) bool {
	return request.PluginID == "arena" && request.Operation == "arena.admin.interrupt" ||
		request.PluginID == "rooms" && (request.Operation == "rooms.admin.close" || request.Operation == "rooms.admin.remove-member") ||
		request.PluginID == "analytics" && (request.Operation == "analytics.history" || request.Operation == "analytics.summary") ||
		request.PluginID == "tables" && (request.Operation == "tables.update" || request.Operation == "tables.delete" || tableSchemaOperation(request.Operation) || needsTablePage(request))
}

// Keep the exact stored JSON for the compare-and-swap and avoid introducing
// number conversions into unchanged rows.
func tableSnapshot(storage *SQLiteStorage, projectID, table string) (string, []json.RawMessage, error) {
	var encoded string
	err := storage.db.QueryRow("SELECT rows_json FROM plugin_table_rows WHERE project_id=? AND table_name=?", projectID, table).Scan(&encoded)
	if errors.Is(err, sql.ErrNoRows) {
		return "", []json.RawMessage{}, nil
	}
	if err != nil {
		return "", nil, err
	}
	var rows []json.RawMessage
	if err := json.Unmarshal([]byte(encoded), &rows); err != nil {
		return "", nil, err
	}
	return encoded, rows, nil
}

func tableRevision(projectID, table, encoded string) string {
	sum := sha256.Sum256([]byte(projectID + "\x00" + table + "\x00" + encoded))
	return hex.EncodeToString(sum[:])
}

func prepareTableAdmin(storage *SQLiteStorage, request WorkerRequest) (WorkerRequest, func(sqliteExecutor) error, error) {
	if storage == nil {
		return request, nil, ErrStoredResourceMissing
	}
	var input map[string]json.RawMessage
	if json.Unmarshal(request.Payload, &input) != nil || input == nil {
		return request, nil, ErrInvalidWorkerRequest
	}
	var table string
	if json.Unmarshal(input["table"], &table) != nil || table == "" || strings.TrimSpace(table) != table || strings.ContainsRune(table, 0) || len(table) > 1024 {
		return request, nil, ErrInvalidWorkerRequest
	}
	var options map[string]any
	if json.Unmarshal(request.Payload, &options) != nil {
		return request, nil, ErrInvalidWorkerRequest
	}
	encoded, rows, err := tableSnapshot(storage, request.ProjectID, table)
	if err != nil {
		return request, nil, err
	}
	revision := tableRevision(request.ProjectID, table, encoded)
	if request.Operation == "tables.read" {
		offset, err := boundedStorageInt(options, "offset", 0, 0, 1000000000)
		if err != nil {
			return request, nil, err
		}
		limit, err := boundedStorageInt(options, "limit", 25, 1, 100)
		if err != nil {
			return request, nil, err
		}
		type indexedRow struct {
			Index int             `json:"row_index"`
			Row   json.RawMessage `json:"row"`
		}
		page := make([]indexedRow, 0, limit)
		size := 0
		for i := offset; i < len(rows) && len(page) < limit; i++ {
			// Leave room for the request envelope, JSON escaping and metadata.
			if size+len(rows[i]) > 200*1024 {
				if len(page) == 0 {
					return request, nil, ErrInvalidWorkerRequest
				}
				break
			}
			size += len(rows[i])
			page = append(page, indexedRow{Index: i, Row: rows[i]})
		}
		metadata := map[string]any{"revision": revision, "total_rows": len(rows), "offset": offset}
		if offset+len(page) < len(rows) {
			metadata["next_offset"] = offset + len(page)
		}
		data, err := json.Marshal(map[string]any{"table": table, "rows": page, "page": metadata})
		request.Payload = data
		return request, nil, err
	}
	var expected string
	if json.Unmarshal(input["revision"], &expected) != nil || expected != revision || encoded == "" {
		return request, nil, ErrTableRevision
	}
	index, err := boundedStorageInt(options, "row_index", -1, 0, 1000000000)
	if err != nil || index < 0 || index >= len(rows) {
		return request, nil, ErrInvalidWorkerRequest
	}
	switch request.Operation {
	case "tables.update":
		var fields map[string]json.RawMessage
		if json.Unmarshal(input["row"], &fields) != nil || fields == nil {
			return request, nil, ErrInvalidWorkerRequest
		}
		rows[index] = input["row"]
	case "tables.delete":
		rows = append(rows[:index], rows[index+1:]...)
	default:
		return request, nil, ErrInvalidWorkerRequest
	}
	next, err := json.Marshal(rows)
	if err != nil {
		return request, nil, err
	}
	result := map[string]any{"table": table, "row_index": index, "revision": tableRevision(request.ProjectID, table, string(next)), "total_rows": len(rows), "updated": request.Operation == "tables.update", "deleted": request.Operation == "tables.delete"}
	payload, err := json.Marshal(map[string]any{"projectId": request.ProjectID, "table": table, "result": result})
	if err != nil {
		return request, nil, err
	}
	request.Payload = payload
	return request, func(db sqliteExecutor) error {
		outcome, err := db.Exec("UPDATE plugin_table_rows SET rows_json=? WHERE project_id=? AND table_name=? AND rows_json=?", string(next), request.ProjectID, table, encoded)
		if err != nil {
			return err
		}
		count, err := outcome.RowsAffected()
		if err != nil {
			return err
		}
		if count != 1 {
			return ErrTableRevision
		}
		return nil
	}, nil
}
