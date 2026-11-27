package plugin

import "database/sql"

type authSessionSummary struct {
	Username       string `json:"username"`
	ActiveSessions int    `json:"activeSessions"`
	EarliestExpiry int64  `json:"earliestExpiry,omitempty"`
	LatestExpiry   int64  `json:"latestExpiry,omitempty"`
}

func (s *SQLiteStorage) authSessionSummary(project, username string, now int64) (authSessionSummary, error) {
	result := authSessionSummary{Username: username}
	var first, last sql.NullInt64
	err := s.db.QueryRow(`SELECT COUNT(*), MIN(json_extract(value_json,'$.expiresAt')), MAX(json_extract(value_json,'$.expiresAt'))
		FROM plugin_state WHERE project_id=? AND namespace=? AND json_extract(value_json,'$.username')=? AND json_extract(value_json,'$.projectId')=? AND json_extract(value_json,'$.expiresAt')>?`, project, authSessionsNamespace, username, project, now).Scan(&result.ActiveSessions, &first, &last)
	result.EarliestExpiry, result.LatestExpiry = first.Int64, last.Int64
	return result, err
}
