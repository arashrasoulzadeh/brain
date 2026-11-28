package plugin

import (
	"encoding/json"

	"brain/internal/execution"
)

// BindArtifactPins initializes a host-only immutable artifact index in the
// same durable database as installation state, before registering artifacts.
func (s *SQLiteStorage) BindArtifactPins(catalog *execution.Catalog) error {
	if catalog == nil {
		return execution.ErrInvalidRequest
	}
	if _, err := s.db.Exec(`CREATE TABLE IF NOT EXISTS execution_artifact_pins (
		project_id TEXT NOT NULL, artifact_id TEXT NOT NULL, version TEXT NOT NULL,
		pin_json TEXT NOT NULL, PRIMARY KEY(project_id, artifact_id, version)
	)`); err != nil {
		return err
	}
	return catalog.BindPinStore(s)
}

// PinArtifact is called only by the trusted catalog after canonical bundle
// validation. Concurrent Brain processes cannot replace an existing pin.
func (s *SQLiteStorage) PinArtifact(key execution.ArtifactKey, pin execution.ArtifactPin) error {
	encoded, err := json.Marshal(pin)
	if err != nil {
		return err
	}
	if _, err := s.db.Exec(`INSERT INTO execution_artifact_pins(project_id, artifact_id, version, pin_json)
		VALUES (?, ?, ?, ?) ON CONFLICT(project_id, artifact_id, version) DO NOTHING`,
		key.ProjectID, key.ID, key.Version, string(encoded)); err != nil {
		return err
	}
	var existing string
	if err := s.db.QueryRow(`SELECT pin_json FROM execution_artifact_pins
		WHERE project_id=? AND artifact_id=? AND version=?`,
		key.ProjectID, key.ID, key.Version).Scan(&existing); err != nil {
		return err
	}
	if existing != string(encoded) {
		return execution.ErrArtifactConflict
	}
	return nil
}
