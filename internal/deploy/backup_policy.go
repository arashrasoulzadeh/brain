package deploy

import (
	"errors"
	"time"
)

var ErrInvalidBackupPolicy = errors.New("invalid backup policy")

type BackupPolicy struct {
	Interval   time.Duration
	AlertAfter time.Duration
}

func (p BackupPolicy) Validate() error {
	if p.Interval < 12*time.Hour || p.AlertAfter < 18*time.Hour || p.AlertAfter < p.Interval {
		return ErrInvalidBackupPolicy
	}
	return nil
}

func (p BackupPolicy) NextDue(lastVerified time.Time) time.Time {
	return lastVerified.Add(p.Interval)
}
