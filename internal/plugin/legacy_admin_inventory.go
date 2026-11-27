package plugin

import "database/sql"

// Keep legacy storage knowledge in the compatibility adapter. The admin
// registry and frontend never branch on product plugin names.
func (p legacyStoragePlanner) AdminInventory(projectID, pluginID, kind, capability, after string, limit int) (AdminInventoryPage, error) {
	if p.storage == nil {
		return AdminInventoryPage{}, ErrAdminInventoryUnavailable
	}
	var rows *sql.Rows
	var err error
	switch {
	case pluginID == "rooms" && kind == "state_keys" && capability == "rooms.result",
		pluginID == "matchmaking" && kind == "state_keys" && capability == "matchmaking.status":
		rows, err = p.storage.db.Query(`SELECT state_key FROM plugin_state WHERE project_id=? AND namespace=? AND state_key>? ORDER BY state_key LIMIT ?`, projectID, pluginID, after, limit+1)
	case pluginID == "files" && kind == "file_keys" && capability == "files.metadata":
		rows, err = p.storage.db.Query(`SELECT file_key,length(CAST(data AS BLOB)),published FROM plugin_file_rows WHERE project_id=? AND file_key>? ORDER BY file_key LIMIT ?`, projectID, after, limit+1)
	case pluginID == "tables" && kind == "table_names" && capability == "tables.read":
		rows, err = p.storage.db.Query(`SELECT table_name FROM plugin_table_rows WHERE project_id=? AND table_name>? ORDER BY table_name LIMIT ?`, projectID, after, limit+1)
	default:
		return AdminInventoryPage{}, ErrGrantDenied
	}
	if err != nil {
		return AdminInventoryPage{}, err
	}
	defer rows.Close()
	result := AdminInventoryPage{Rows: make([]AdminInventoryRow, 0, limit)}
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
