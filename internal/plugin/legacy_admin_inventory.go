package plugin

import "encoding/json"

// Keep legacy storage knowledge in the compatibility adapter. The admin
// registry and frontend never branch on product plugin names.
func (p legacyStoragePlanner) AdminInventory(projectID, pluginID, kind, capability, after string, limit int) (AdminInventoryPage, error) {
	return p.adminInventory(projectID, pluginID, kind, capability, after, limit, nil)
}

func (p legacyStoragePlanner) AdminInventoryScoped(projectID, pluginID, kind, capability, after string, limit int, filter AdminInventoryFilter) (AdminInventoryPage, error) {
	return p.adminInventory(projectID, pluginID, kind, capability, after, limit, &filter)
}

func (p legacyStoragePlanner) adminInventory(projectID, pluginID, kind, capability, after string, limit int, filter *AdminInventoryFilter) (AdminInventoryPage, error) {
	if p.storage == nil {
		return AdminInventoryPage{}, ErrAdminInventoryUnavailable
	}
	var query, key string
	var args []any
	switch {
	case pluginID == "rooms" && kind == "state_keys" && capability == "rooms.result",
		pluginID == "matchmaking" && kind == "state_keys" && capability == "matchmaking.status":
		query, key = `SELECT state_key FROM plugin_state WHERE project_id=? AND namespace=? AND state_key>?`, "state_key"
		args = []any{projectID, pluginID, after}
	case pluginID == "files" && kind == "file_keys" && capability == "files.metadata":
		query, key = `SELECT file_key,length(CAST(data AS BLOB)),published FROM plugin_file_rows WHERE project_id=? AND file_key>?`, "file_key"
		args = []any{projectID, after}
	case pluginID == "tables" && kind == "table_names" && capability == "tables.read":
		query, key = `SELECT table_name FROM plugin_table_rows WHERE project_id=? AND table_name>?`, "table_name"
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
			err = rows.Scan(&row.Key, &size, &published)
			row.SizeBytes, row.Published = &size, &published
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
