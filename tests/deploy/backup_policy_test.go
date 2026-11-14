package deploy_test

import (
	"errors"
	"testing"
	"time"

	"brain/internal/deploy"
)

func TestBackupPolicyRequiresTwelveHourCadenceAndEighteenHourAlert(t *testing.T) {
	policy := deploy.BackupPolicy{Interval: 12 * time.Hour, AlertAfter: 18 * time.Hour}
	if err := policy.Validate(); err != nil {
		t.Fatal(err)
	}
	if policy.NextDue(time.Unix(100, 0)).Unix() != (time.Unix(100, 0).Add(12 * time.Hour)).Unix() {
		t.Fatal("backup due time was calculated incorrectly")
	}
	if err := (deploy.BackupPolicy{Interval: 11 * time.Hour, AlertAfter: 18 * time.Hour}).Validate(); !errors.Is(err, deploy.ErrInvalidBackupPolicy) {
		t.Fatal("short backup interval was accepted")
	}
	if !policy.IsStale(time.Unix(0, 0), time.Unix(0, 0).Add(19*time.Hour)) {
		t.Fatal("old verified backup was not marked stale")
	}
	if policy.IsStale(time.Unix(0, 0), time.Unix(0, 0).Add(17*time.Hour)) {
		t.Fatal("fresh verified backup was marked stale")
	}
	if !policy.MeetsDataLossTarget(time.Unix(0, 0), time.Unix(0, 0).Add(24*time.Hour)) {
		t.Fatal("backup at the 24-hour target was rejected")
	}
	if policy.MeetsDataLossTarget(time.Unix(0, 0), time.Unix(0, 0).Add(25*time.Hour)) {
		t.Fatal("backup beyond the 24-hour target was accepted")
	}
}
