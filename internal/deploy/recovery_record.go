package deploy

import (
	"encoding/json"
	"errors"
	"os"
	"time"
)

var ErrInvalidRecoveryRecord = errors.New("invalid recovery record")

type RecoveryRecord struct {
	ReleaseSHA     string    `json:"release_sha"`
	StartedAt      time.Time `json:"started_at"`
	CompletedAt    time.Time `json:"completed_at"`
	RetentionHours int       `json:"retention_hours"`
}

func (r RecoveryRecord) Validate() error {
	if ValidateReleaseSHA(r.ReleaseSHA) != nil || r.StartedAt.IsZero() || r.CompletedAt.Before(r.StartedAt) || r.RetentionHours <= 0 {
		return ErrInvalidRecoveryRecord
	}
	return nil
}

func (r RecoveryRecord) Duration() time.Duration {
	return r.CompletedAt.Sub(r.StartedAt)
}

func RecordRecovery(path string, record RecoveryRecord) error {
	if err := record.Validate(); err != nil {
		return err
	}
	contents, err := json.Marshal(record)
	if err != nil {
		return err
	}
	temporaryPath := path + ".tmp"
	if err := os.WriteFile(temporaryPath, append(contents, '\n'), 0o600); err != nil {
		return err
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		_ = os.Remove(temporaryPath)
		return err
	}
	return nil
}

func ReadRecovery(path string) (RecoveryRecord, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return RecoveryRecord{}, err
	}
	var record RecoveryRecord
	if err := json.Unmarshal(contents, &record); err != nil {
		return RecoveryRecord{}, ErrInvalidRecoveryRecord
	}
	if err := record.Validate(); err != nil {
		return RecoveryRecord{}, err
	}
	return record, nil
}
