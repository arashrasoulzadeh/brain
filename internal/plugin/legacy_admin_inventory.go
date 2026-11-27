package plugin

import "encoding/json"

// Keep legacy storage knowledge in the compatibility adapter. The admin
// registry and frontend never branch on product plugin names.
func (p legacyStoragePlanner) AdminInventory(projectID, pluginID, kind, capability, after string, limit int) (AdminInventoryPage, error) {
	return p.adminInventory(projectID, pluginID, kind, capability, after, "", limit, nil)
}

func (p legacyStoragePlanner) AdminInventoryScoped(projectID, pluginID, kind, capability, after string, limit int, filter AdminInventoryFilter) (AdminInventoryPage, error) {
	return p.adminInventory(projectID, pluginID, kind, capability, after, "", limit, &filter)
}

func (p legacyStoragePlanner) AdminInventorySearch(projectID, pluginID, kind, capability, after, search string, limit int, filter *AdminInventoryFilter) (AdminInventoryPage, error) {
	return p.adminInventory(projectID, pluginID, kind, capability, after, search, limit, filter)
}

func (p legacyStoragePlanner) adminInventory(projectID, pluginID, kind, capability, after, search string, limit int, filter *AdminInventoryFilter) (AdminInventoryPage, error) {
	if p.storage == nil {
		return AdminInventoryPage{}, ErrAdminInventoryUnavailable
	}
	var query, key string
	var args []any
	switch {
	case pluginID == "rooms" && kind == "state_keys" && capability == "rooms.result":
		query, key = `SELECT state_key, CASE WHEN json_extract(value_json,'$.closed')=1 THEN 'closed' ELSE 'open' END, COALESCE(json_array_length(value_json,'$.members'),0) FROM plugin_state WHERE project_id=? AND namespace=? AND state_key>?`, "state_key"
		args = []any{projectID, pluginID, after}
	case pluginID == "matchmaking" && kind == "state_keys" && capability == "matchmaking.status":
		query, key = `SELECT state_key, COALESCE(json_extract(value_json,'$.status'),'unknown') FROM plugin_state WHERE project_id=? AND namespace=? AND state_key>?`, "state_key"
		args = []any{projectID, pluginID, after}
	case pluginID == "arena" && kind == "state_keys" && capability == "arena.snapshot":
		query, key = `SELECT state_key FROM plugin_state WHERE project_id=? AND namespace=? AND state_key>?`, "state_key"
		args = []any{projectID, pluginID, after}
	case pluginID == "files" && kind == "file_keys" && capability == "files.metadata":
		query, key = `SELECT file_key,length(CAST(data AS BLOB)),published,metadata_json FROM plugin_file_rows WHERE project_id=? AND file_key>?`, "file_key"
		args = []any{projectID, after}
	case pluginID == "tables" && kind == "table_names" && capability == "tables.read":
		query, key = `SELECT DISTINCT table_name FROM plugin_table_rows WHERE project_id=? AND table_name>?`, "table_name"
		args = []any{projectID, after}
	default:
		return AdminInventoryPage{}, ErrGrantDenied
	}
	result := AdminInventoryPage{Rows: make([]AdminInventoryRow, 0, limit)}
	if filter != nil {
		if filter.DenyAll || (!filter.AllowAll && len(filter.Allowed) == 0) || (filter.Restricted && len(filter.GrantKeys) == 0) {
			return result, nil
		}
		allowed, _ := json.Marshal(filter.Allowed)
		denied, _ := json.Marshal(filter.Denied)
		grants, _ := json.Marshal(filter.GrantKeys)
		// The SQL column is selected above, never from a manifest or caller.
		// Only visible keys are eligible for a pagination cursor.
		query += " AND (? OR " + key + " IN (SELECT value FROM json_each(?)))"
		query += " AND NOT EXISTS (SELECT 1 FROM json_each(?) WHERE value=" + key + ")"
		query += " AND (? OR " + key + " IN (SELECT value FROM json_each(?)))"
		args = append(args, filter.AllowAll, string(allowed), string(denied), !filter.Restricted, string(grants))
	}
	if search != "" {
		// Literal substring matching, not LIKE wildcard matching. The column
		// and JSON paths are fixed host-owned SQL, and search is a bound value.
		query += " AND (instr(lower(" + key + "),lower(?))>0"
		args = append(args, search)
		if kind == "file_keys" {
			for _, field := range []string{"filename", "extension", "content_type"} {
				query += " OR instr(lower(CASE WHEN json_valid(metadata_json) THEN coalesce(json_extract(metadata_json, ?), '') ELSE '' END), lower(?))>0"
				args = append(args, "$."+field, search)
			}
		}
		query += ")"
	}
	query += " ORDER BY " + key
	if filter == nil {
		query += " LIMIT ?"
		args = append(args, limit+1)
	}
	rows, err := p.storage.db.Query(query, args...)
	if err != nil {
		return AdminInventoryPage{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var row AdminInventoryRow
		if kind == "file_keys" {
			var size int64
			var published bool
			var encoded string
			err = rows.Scan(&row.Key, &size, &published, &encoded)
			row.SizeBytes, row.Published = &size, &published
			if err == nil && encoded != "" {
				var meta FileMetadata
				err = json.Unmarshal([]byte(encoded), &meta)
				row.Filename, row.Extension, row.ContentType, row.UpdatedAt = meta.Filename, meta.Extension, meta.ContentType, meta.UpdatedAt
				row.SizeBytes, row.StoredBytes = &meta.SizeBytes, &meta.StoredBytes
			}
		} else if pluginID == "rooms" && kind == "state_keys" {
			var count int
			err = rows.Scan(&row.Key, &row.Status, &count)
			row.MemberCount = &count
		} else if pluginID == "matchmaking" && kind == "state_keys" {
			err = rows.Scan(&row.Key, &row.Status)
		} else {
			err = rows.Scan(&row.Key)
		}
		if err != nil {
			return AdminInventoryPage{}, err
		}
		if filter != nil && (!validAccessIdentity(row.Key) || row.Key == "*") {
			continue
		}
		if len(result.Rows) == limit {
			result.Next = result.Rows[limit-1].Key
			break
		}
		result.Rows = append(result.Rows, row)
	}
	if err := rows.Err(); err != nil {
		return AdminInventoryPage{}, err
	}
	return result, nil
}
