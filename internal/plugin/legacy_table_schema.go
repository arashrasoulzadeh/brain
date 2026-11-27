package plugin

import (
	"encoding/json"
	"sort"
	"strings"
)

func tableSchemaOperation(op string) bool {
	switch op {
	case "tables.schema.profile", "tables.admin.create", "tables.admin.rename-column", "tables.admin.drop-column", "tables.admin.add-column":
		return true
	default:
		return false
	}
}

type tableColumnProfile struct {
	Name    string   `json:"name"`
	Types   []string `json:"types"`
	Present int      `json:"present"`
	Missing int      `json:"missing"`
}

func storedJSONType(raw json.RawMessage) string {
	switch raw[0] {
	case '"':
		return "string"
	case '{':
		return "object"
	case '[':
		return "array"
	case 't', 'f':
		return "boolean"
	case 'n':
		return "null"
	default:
		return "number"
	}
}

// Schema operations apply to the existing JSON snapshot store. No SQL
// identifiers or statements come from the caller, and no relational schema
// is implied by this observed profile.
func prepareTableSchema(storage *SQLiteStorage, request WorkerRequest) (WorkerRequest, func(sqliteExecutor) error, error) {
	if storage == nil {
		return request, nil, ErrStoredResourceMissing
	}
	var input map[string]json.RawMessage
	if json.Unmarshal(request.Payload, &input) != nil || input == nil {
		return request, nil, ErrInvalidWorkerRequest
	}
	readName := func(field string) string {
		var value string
		if json.Unmarshal(input[field], &value) != nil || value == "" || strings.TrimSpace(value) != value || strings.ContainsRune(value, 0) || len(value) > 256 {
			return ""
		}
		return value
	}
	table := readName("table")
	if table == "" {
		return request, nil, ErrInvalidWorkerRequest
	}
	encoded, rows, err := tableSnapshot(storage, request.ProjectID, table)
	if err != nil {
		return request, nil, err
	}
	revision := tableRevision(request.ProjectID, table, encoded)
	result := map[string]any{"table": table, "revision": revision, "exists": encoded != "", "total_rows": len(rows)}
	var effect func(sqliteExecutor) error
	switch request.Operation {
	case "tables.schema.profile":
		var options map[string]any
		if json.Unmarshal(request.Payload, &options) != nil {
			return request, nil, ErrInvalidWorkerRequest
		}
		offset, err := boundedStorageInt(options, "offset", 0, 0, 1000000000)
		if err != nil {
			return request, nil, err
		}
		limit, err := boundedStorageInt(options, "limit", 50, 1, 100)
		if err != nil {
			return request, nil, err
		}
		profiles := map[string]*tableColumnProfile{}
		for _, raw := range rows {
			var fields map[string]json.RawMessage
			if json.Unmarshal(raw, &fields) != nil || fields == nil {
				return request, nil, ErrModelSchema
			}
			for name, value := range fields {
				column := profiles[name]
				if column == nil {
					column = &tableColumnProfile{Name: name}
					profiles[name] = column
				}
				column.Present++
				kind := storedJSONType(value)
				found := false
				for _, existing := range column.Types {
					found = found || existing == kind
				}
				if !found {
					column.Types = append(column.Types, kind)
				}
			}
		}
		names := make([]string, 0, len(profiles))
		for name := range profiles {
			names = append(names, name)
		}
		sort.Strings(names)
		page := make([]tableColumnProfile, 0, limit)
		size := 0
		for i := offset; i < len(names) && len(page) < limit; i++ {
			column := profiles[names[i]]
			column.Missing = len(rows) - column.Present
			sort.Strings(column.Types)
			value, err := json.Marshal(column)
			if err != nil {
				return request, nil, err
			}
			if size+len(value) > 200*1024 {
				if len(page) == 0 {
					return request, nil, ErrInvalidWorkerRequest
				}
				break
			}
			size += len(value)
			page = append(page, *column)
		}
		result["columns"], result["total_columns"], result["offset"] = page, len(names), offset
		if offset+len(page) < len(names) {
			result["next_offset"] = offset + len(page)
		}
	case "tables.admin.create":
		if encoded != "" {
			return request, nil, ErrModelConflict
		}
		result["exists"], result["created"], result["revision"] = true, true, tableRevision(request.ProjectID, table, "[]")
		effect = func(db sqliteExecutor) error {
			outcome, err := db.Exec("INSERT INTO plugin_table_rows(project_id,table_name,rows_json) VALUES(?,?,?) ON CONFLICT(project_id,table_name) DO NOTHING", request.ProjectID, table, "[]")
			if err != nil {
				return err
			}
			count, err := outcome.RowsAffected()
			if err != nil {
				return err
			}
			if count != 1 {
				return ErrModelConflict
			}
			return nil
		}
	case "tables.admin.rename-column", "tables.admin.drop-column", "tables.admin.add-column":
		expected := readName("revision")
		if encoded == "" || expected != revision {
			return request, nil, ErrTableRevision
		}
		// Bulk edits are bounded and atomic. Larger datasets need a migration
		// mechanism, not an unbounded synchronous administration request.
		if len(encoded) > 4*1024*1024 || len(rows) > 10000 {
			return request, nil, ErrInvalidWorkerRequest
		}
		column := readName("column")
		target := readName("new_column")
		if column == "" || (request.Operation == "tables.admin.rename-column" && (target == "" || target == column)) {
			return request, nil, ErrInvalidWorkerRequest
		}
		affected := 0
		for i, raw := range rows {
			var fields map[string]json.RawMessage
			if json.Unmarshal(raw, &fields) != nil || fields == nil {
				return request, nil, ErrModelSchema
			}
			value, present := fields[column]
			switch request.Operation {
			case "tables.admin.add-column":
				if present {
					return request, nil, ErrModelConflict
				}
				value, present = input["default"]
				if !present {
					value = json.RawMessage("null")
				}
				fields[column] = value
			case "tables.admin.rename-column":
				if _, collision := fields[target]; collision {
					return request, nil, ErrModelConflict
				}
				if !present {
					continue
				}
				delete(fields, column)
				fields[target] = value
			case "tables.admin.drop-column":
				if !present {
					continue
				}
				delete(fields, column)
			}
			next, err := json.Marshal(fields)
			if err != nil {
				return request, nil, err
			}
			rows[i] = next
			affected++
		}
		if affected == 0 {
			return request, nil, ErrStoredResourceMissing
		}
		next, err := json.Marshal(rows)
		if err != nil {
			return request, nil, err
		}
		if len(next) > 4*1024*1024 {
			return request, nil, ErrInvalidWorkerRequest
		}
		result["affected_rows"], result["revision"] = affected, tableRevision(request.ProjectID, table, string(next))
		effect = func(db sqliteExecutor) error {
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
		}
	default:
		return request, nil, ErrInvalidWorkerRequest
	}
	payload, err := json.Marshal(map[string]any{"projectId": request.ProjectID, "table": table, "result": result})
	request.Payload = payload
	return request, effect, err
}
