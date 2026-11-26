package plugin

import (
	"encoding/json"
	"time"
)

func (s *SQLiteStorage) initializeObservationStorage() error {
	_, err := s.db.Exec(`
CREATE TABLE IF NOT EXISTS plugin_observation_queue (
  project_id TEXT NOT NULL, target_plugin TEXT NOT NULL, target_version TEXT NOT NULL,
  observation_id TEXT NOT NULL, delivery_json TEXT NOT NULL, attempts INTEGER NOT NULL,
  due_at INTEGER NOT NULL, PRIMARY KEY(project_id, target_plugin, target_version, observation_id)
);
CREATE INDEX IF NOT EXISTS plugin_observation_due ON plugin_observation_queue(due_at);
CREATE TABLE IF NOT EXISTS plugin_observation_receipts (
  project_id TEXT NOT NULL, plugin_id TEXT NOT NULL, observation_id TEXT NOT NULL,
  delivered_at INTEGER NOT NULL, payload_json TEXT NOT NULL,
  PRIMARY KEY(project_id, plugin_id, observation_id)
);
CREATE INDEX IF NOT EXISTS plugin_observation_receipts_time ON plugin_observation_receipts(project_id, plugin_id, delivered_at);
`)
	return err
}

func (s *SQLiteStorage) enqueueObservation(delivery observationDelivery, now time.Time) error {
	return enqueueSQLiteObservation(s.db, delivery, now)
}

func enqueueSQLiteObservation(db sqliteExecutor, delivery observationDelivery, now time.Time) error {
	encoded, err := json.Marshal(delivery)
	if err != nil {
		return err
	}
	// The guarded INSERT is one statement: concurrent publishers cannot race
	// separate capacity checks and overfill the durable queue.
	result, err := db.Exec(`INSERT INTO plugin_observation_queue
(project_id,target_plugin,target_version,observation_id,delivery_json,attempts,due_at)
SELECT ?,?,?,?,?,0,? WHERE
(SELECT count(*) FROM plugin_observation_queue) < ? AND
(SELECT count(*) FROM plugin_observation_queue WHERE project_id=?) < ?`,
		delivery.Target.ProjectID, delivery.Target.PluginID, delivery.Target.Version, delivery.ID,
		string(encoded), now.UnixMilli(), maxPendingObservations, delivery.Target.ProjectID, maxPendingProjectObservations)
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count != 1 {
		return ErrObservationCapacity
	}
	return nil
}

func (s *SQLiteStorage) dueObservations(now time.Time, limit int) ([]observationDelivery, error) {
	rows, err := s.db.Query(`SELECT delivery_json,attempts FROM plugin_observation_queue WHERE due_at<=? ORDER BY due_at, observation_id LIMIT ?`, now.UnixMilli(), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	deliveries := make([]observationDelivery, 0, limit)
	for rows.Next() {
		var encoded string
		var attempts int
		if err := rows.Scan(&encoded, &attempts); err != nil {
			return nil, err
		}
		var delivery observationDelivery
		if err := json.Unmarshal([]byte(encoded), &delivery); err != nil {
			return nil, err
		}
		delivery.Attempts = attempts
		deliveries = append(deliveries, delivery)
	}
	return deliveries, rows.Err()
}

func (s *SQLiteStorage) retryObservation(delivery observationDelivery, now time.Time) error {
	target := delivery.Target
	delay := time.Second * time.Duration(1<<uint(delivery.Attempts))
	_, err := s.db.Exec(`UPDATE plugin_observation_queue SET attempts=?,due_at=? WHERE project_id=? AND target_plugin=? AND target_version=? AND observation_id=?`,
		delivery.Attempts+1, now.Add(delay).UnixMilli(), target.ProjectID, target.PluginID, target.Version, delivery.ID)
	return err
}

func (s *SQLiteStorage) finishObservation(delivery observationDelivery, persist bool, now time.Time) error {
	target := delivery.Target
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if persist {
		if _, err := tx.Exec(`INSERT OR IGNORE INTO plugin_observation_receipts(project_id,plugin_id,observation_id,delivered_at,payload_json) VALUES(?,?,?,?,?)`,
			target.ProjectID, target.PluginID, delivery.ID, now.UnixMilli(), string(target.Payload)); err != nil {
			return err
		}
	}
	if _, err := tx.Exec(`DELETE FROM plugin_observation_queue WHERE project_id=? AND target_plugin=? AND target_version=? AND observation_id=?`,
		target.ProjectID, target.PluginID, target.Version, delivery.ID); err != nil {
		return err
	}
	return tx.Commit()
}

// ObservationReceipts returns only the named project's retained, explicitly
// persisted observer envelopes. Callers must authorize access before exposing
// this host storage primitive outside the registry.
func (s *SQLiteStorage) ObservationReceipts(projectID, pluginID string, limit int) ([]Observation, error) {
	if !isValidProjectID(projectID) || pluginID == "" || limit < 1 || limit > 1000 {
		return nil, ErrInvalidWorkerRequest
	}
	rows, err := s.db.Query(`SELECT payload_json FROM plugin_observation_receipts WHERE project_id=? AND plugin_id=? ORDER BY delivered_at DESC, observation_id DESC LIMIT ?`, projectID, pluginID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	events := make([]Observation, 0)
	for rows.Next() {
		var encoded string
		if err := rows.Scan(&encoded); err != nil {
			return nil, err
		}
		var event Observation
		if err := json.Unmarshal([]byte(encoded), &event); err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	return events, rows.Err()
}
