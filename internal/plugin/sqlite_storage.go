package plugin

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"

	"brain/internal/sqliteconfig"

	_ "modernc.org/sqlite"
)

var ErrStorageTableRequired = errors.New("storage table required")

type SQLiteStorage struct{ db *sql.DB }

type sqliteExecutor interface {
	Exec(query string, args ...any) (sql.Result, error)
	QueryRow(query string, args ...any) *sql.Row
}

type persistedRegistryState struct {
	Installations map[string]Installation `json:"installations"`
	Readiness     map[string]bool         `json:"readiness"`
}

func OpenSQLiteStorage(path string) (*SQLiteStorage, error) {
	db, err := sql.Open("sqlite", sqliteconfig.DSN(path))
	if err != nil {
		return nil, err
	}
	if _, err = db.Exec(`CREATE TABLE IF NOT EXISTS plugin_table_rows (project_id TEXT NOT NULL, table_name TEXT NOT NULL, rows_json TEXT NOT NULL, PRIMARY KEY(project_id, table_name)); CREATE TABLE IF NOT EXISTS plugin_file_rows (project_id TEXT NOT NULL, file_key TEXT NOT NULL, data TEXT NOT NULL, published INTEGER NOT NULL, PRIMARY KEY(project_id, file_key)); CREATE TABLE IF NOT EXISTS plugin_state (project_id TEXT NOT NULL, namespace TEXT NOT NULL, state_key TEXT NOT NULL, value_json TEXT NOT NULL, PRIMARY KEY(project_id, namespace, state_key)); CREATE TABLE IF NOT EXISTS plugin_model_records (project_id TEXT NOT NULL, plugin_id TEXT NOT NULL, model_name TEXT NOT NULL, record_id TEXT NOT NULL, schema_version INTEGER NOT NULL, revision TEXT NOT NULL, state TEXT NOT NULL, fields_json TEXT NOT NULL, PRIMARY KEY(project_id, plugin_id, model_name, record_id))`); err != nil {
		_ = db.Close()
		return nil, err
	}
	storage := &SQLiteStorage{db: db}
	if err := storage.initializeFileMetadata(); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := storage.initializeObservationStorage(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return storage, nil
}

func (s *SQLiteStorage) SaveRegistryState(installations map[string]Installation, readiness map[string]bool) error {
	return s.PutState("", "registry", "lifecycle", map[string]any{"installations": installations, "readiness": readiness})
}

func (s *SQLiteStorage) LoadRegistryState() (map[string]Installation, map[string]bool, error) {
	value, err := s.State("", "registry", "lifecycle")
	if err != nil || value == nil {
		return map[string]Installation{}, map[string]bool{}, err
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		return nil, nil, err
	}
	var state persistedRegistryState
	if err := json.Unmarshal(encoded, &state); err != nil {
		return nil, nil, err
	}
	return state.Installations, state.Readiness, nil
}

func (s *SQLiteStorage) PutState(projectID, namespace, key string, value map[string]any) error {
	return putSQLiteState(s.db, projectID, namespace, key, value)
}

func putSQLiteState(db sqliteExecutor, projectID, namespace, key string, value map[string]any) error {
	encoded, err := json.Marshal(value)
	if err != nil {
		return err
	}
	_, err = db.Exec(`INSERT INTO plugin_state(project_id,namespace,state_key,value_json) VALUES(?,?,?,?) ON CONFLICT(project_id,namespace,state_key) DO UPDATE SET value_json=excluded.value_json`, projectID, namespace, key, string(encoded))
	return err
}

func (s *SQLiteStorage) State(projectID, namespace, key string) (map[string]any, error) {
	var encoded string
	err := s.db.QueryRow(`SELECT value_json FROM plugin_state WHERE project_id=? AND namespace=? AND state_key=?`, projectID, namespace, key).Scan(&encoded)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var value map[string]any
	return value, json.Unmarshal([]byte(encoded), &value)
}

func (s *SQLiteStorage) PutFile(projectID, key, data string, published bool) error {
	return putSQLiteFile(s.db, projectID, key, data, published)
}

func putSQLiteFile(db sqliteExecutor, projectID, key, data string, published bool) error {
	_, err := db.Exec(`INSERT INTO plugin_file_rows(project_id,file_key,data,published) VALUES(?,?,?,?) ON CONFLICT(project_id,file_key) DO UPDATE SET data=excluded.data,published=excluded.published,metadata_json=''`, projectID, key, data, published)
	return err
}

func (s *SQLiteStorage) File(projectID, key string) (string, bool, error) {
	var data string
	var published bool
	err := s.db.QueryRow(`SELECT data,published FROM plugin_file_rows WHERE project_id=? AND file_key=?`, projectID, key).Scan(&data, &published)
	if errors.Is(err, sql.ErrNoRows) {
		return "", false, nil
	}
	return data, published, err
}

func (s *SQLiteStorage) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

func (s *SQLiteStorage) Rows(projectID, table string) ([]map[string]any, error) {
	return sqliteRows(s.db, projectID, table)
}

func sqliteRows(db sqliteExecutor, projectID, table string) ([]map[string]any, error) {
	if strings.TrimSpace(table) == "" {
		return nil, ErrStorageTableRequired
	}
	var encoded string
	err := db.QueryRow(`SELECT rows_json FROM plugin_table_rows WHERE project_id = ? AND table_name = ?`, projectID, table).Scan(&encoded)
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
	return appendSQLiteRow(s.db, projectID, table, row)
}

func appendSQLiteRow(db sqliteExecutor, projectID, table string, row map[string]any) error {
	rows, err := sqliteRows(db, projectID, table)
	if err != nil {
		return err
	}
	rows = append(rows, row)
	encoded, err := json.Marshal(rows)
	if err != nil {
		return err
	}
	_, err = db.Exec(`INSERT INTO plugin_table_rows(project_id, table_name, rows_json) VALUES (?, ?, ?) ON CONFLICT(project_id, table_name) DO UPDATE SET rows_json=excluded.rows_json`, projectID, table, string(encoded))
	return err
}

// CheckHealth verifies that the plugin database connection is reachable.
func (s *SQLiteStorage) CheckHealth(ctx context.Context) error {
	if s == nil || s.db == nil {
		return errors.New("plugin storage unavailable")
	}
	return s.db.PingContext(ctx)
}
