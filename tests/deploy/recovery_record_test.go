package deploy_test

import (
	"errors"
	"testing"
	"time"

	"brain/internal/deploy"
)

func TestRecoveryRecordStoresDurationAndRetention(t *testing.T) {
	path := t.TempDir() + "/recovery.json"
	started := time.Unix(100, 0).UTC()
	completed := started.Add(5 * time.Minute)
	record := deploy.RecoveryRecord{
		ReleaseSHA:     "0123456789abcdef0123456789abcdef01234567",
		StartedAt:      started,
		CompletedAt:    completed,
		RetentionHours: 168,
	}
	if err := deploy.RecordRecovery(path, record); err != nil {
		t.Fatal(err)
	}
	read, err := deploy.ReadRecovery(path)
	if err != nil || read.Duration() != 5*time.Minute {
		t.Fatalf("record = %#v, err = %v", read, err)
	}
	if err := deploy.RecordRecovery(path, deploy.RecoveryRecord{ReleaseSHA: record.ReleaseSHA, StartedAt: completed, CompletedAt: started, RetentionHours: 1}); !errors.Is(err, deploy.ErrInvalidRecoveryRecord) {
		t.Fatal("invalid recovery timestamps were accepted")
	}
}
