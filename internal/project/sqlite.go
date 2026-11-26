package project

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

func NewSQLiteStore(path string) (*Store, error) {
	database, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	if _, err := database.Exec(`CREATE TABLE IF NOT EXISTS projects (id TEXT PRIMARY KEY, active INTEGER NOT NULL)`); err != nil {
		_ = database.Close()
		return nil, err
	}
	var count int
	if err := database.QueryRow(`SELECT COUNT(*) FROM projects`).Scan(&count); err == nil && count == 0 {
		legacyPath := filepath.Join(filepath.Dir(path), "projects.json")
		if data, readErr := os.ReadFile(legacyPath); readErr == nil {
			var projects []Project
			if json.Unmarshal(data, &projects) == nil {
				for _, item := range projects {
					if ValidateID(item.ID) == nil {
						_, _ = database.Exec(`INSERT OR IGNORE INTO projects (id, active) VALUES (?, ?)`, item.ID, item.Active)
					}
				}
			}
		}
	}
	return &Store{projects: make(map[string]Project), db: database}, nil
}

// CheckHealth verifies that durable project storage is configured and reachable.
func (s *Store) CheckHealth(ctx context.Context) error {
	if s == nil || s.db == nil {
		return errors.New("durable project storage unavailable")
	}
	return s.db.PingContext(ctx)
}

func (s *Store) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}
