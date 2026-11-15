package plugin

import (
	"database/sql"
	"encoding/json"
	"errors"
	"strings"

	_ "modernc.org/sqlite"
)

var ErrStorageTableRequired = errors.New("storage table required")

type SQLiteStorage struct{ db *sql.DB }

func OpenSQLiteStorage(path string) (*SQLiteStorage, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	if _, err = db.Exec(`CREATE TABLE IF NOT EXISTS plugin_table_rows (project_id TEXT NOT NULL, table_name TEXT NOT NULL, rows_json TEXT NOT NULL, PRIMARY KEY(project_id, table_name))`); err != nil {
		_ = db.Close()
		return nil, err
	}
	return &SQLiteStorage{db: db}, nil
}

func (s *SQLiteStorage) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

func (s *SQLiteStorage) Rows(projectID, table string) ([]map[string]any, error) {
	if strings.TrimSpace(table) == "" {
		return nil, ErrStorageTableRequired
	}
	var encoded string
	err := s.db.QueryRow(`SELECT rows_json FROM plugin_table_rows WHERE project_id = ? AND table_name = ?`, projectID, table).Scan(&encoded)
	if errors.Is(err, sql.ErrNoRows) {
		return []map[string]any{}, nil
	}
	if err != nil {
		return nil, err
	}
	var rows []map[string]any
	return rows, json.Unmarshal([]byte(encoded), &rows)
}

func (s *SQLiteStorage) Append(projectID, table string, row map[string]any) error {
	rows, err := s.Rows(projectID, table)
	if err != nil {
		return err
	}
	rows = append(rows, row)
	encoded, err := json.Marshal(rows)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(`INSERT INTO plugin_table_rows(project_id, table_name, rows_json) VALUES (?, ?, ?) ON CONFLICT(project_id, table_name) DO UPDATE SET rows_json=excluded.rows_json`, projectID, table, string(encoded))
	return err
}
