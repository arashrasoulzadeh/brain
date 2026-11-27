package deploy_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"brain/internal/deploy"
)

func TestDeploymentRecordsRejectInvalidAndUnavailableFiles(t *testing.T) {
	sha := strings.Repeat("a", 40)
	image := "brain@sha256:" + strings.Repeat("b", 64)
	manifest := deploy.ReleaseManifest{
		CommitSHA: sha, Backend: image, Compiler: image, Runner: image,
		PackagedIsolationVerified: true, PackagedRevocationVerified: true,
	}
	object := deploy.ObjectVersion{ProjectID: "alpha", ObjectKey: "file", VersionID: "one"}
	objects := deploy.ObjectVersionManifest{Objects: []deploy.ObjectVersion{object}}
	now := time.Now().UTC()
	recovery := deploy.RecoveryRecord{ReleaseSHA: sha, StartedAt: now, CompletedAt: now, RetentionHours: 1}
	for _, tc := range []struct {
		name       string
		write      func(string) error
		read       func(string) error
		invalid    string
		invalidErr error
	}{
		{"healthy", func(p string) error { return deploy.RecordHealthyManifest(p, manifest) }, func(p string) error { _, err := deploy.ReadHealthyManifest(p); return err }, "{}", deploy.ErrInvalidReleaseManifest},
		{"objects", func(p string) error { return deploy.RecordObjectVersionManifest(p, objects) }, func(p string) error { _, err := deploy.ReadObjectVersionManifest(p); return err }, `{"objects":[{"project_id":"","object_key":"file","version_id":"one"}]}`, deploy.ErrInvalidObjectVersionManifest},
		{"recovery", func(p string) error { return deploy.RecordRecovery(p, recovery) }, func(p string) error { _, err := deploy.ReadRecovery(p); return err }, "{}", deploy.ErrInvalidRecoveryRecord},
		{"version", func(p string) error { return deploy.RecordDeployedVersion(p, sha) }, func(p string) error { _, err := deploy.ReadDeployedVersion(p); return err }, "bad", deploy.ErrInvalidReleaseSHA},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			missing := filepath.Join(dir, "absent", "record")
			if err := tc.write(missing); err == nil {
				t.Fatal("record written to missing parent")
			}
			if err := tc.read(missing); !errors.Is(err, os.ErrNotExist) {
				t.Fatalf("missing record: %v", err)
			}
			destination := filepath.Join(dir, "directory")
			if err := os.Mkdir(destination, 0o700); err != nil {
				t.Fatal(err)
			}
			if err := tc.write(destination); err == nil {
				t.Fatal("record replaced a directory")
			}
			entries, err := os.ReadDir(dir)
			if err != nil || len(entries) != 1 {
				t.Fatalf("temporary record leaked: %v, %v", entries, err)
			}
			path := filepath.Join(dir, "record")
			for _, invalid := range []string{"not json", tc.invalid} {
				if err := os.WriteFile(path, []byte(invalid), 0o600); err != nil {
					t.Fatal(err)
				}
				if err := tc.read(path); !errors.Is(err, tc.invalidErr) {
					t.Fatalf("invalid record: %v", err)
				}
			}
			if err := tc.write(path); err != nil {
				t.Fatal(err)
			}
			if err := tc.read(path); err != nil {
				t.Fatal(err)
			}
		})
	}
	if err := deploy.ValidateObjectVersionManifest(deploy.ObjectVersionManifest{Objects: []deploy.ObjectVersion{{}}}); !errors.Is(err, deploy.ErrInvalidObjectVersionManifest) {
		t.Fatalf("empty object identity: %v", err)
	}
	outsideJSONRange := time.Date(10000, 1, 1, 0, 0, 0, 0, time.UTC)
	recovery.StartedAt, recovery.CompletedAt = outsideJSONRange, outsideJSONRange
	if err := deploy.RecordRecovery(filepath.Join(t.TempDir(), "record"), recovery); err == nil {
		t.Fatal("invalid JSON timestamp accepted")
	}
}

func TestDeploymentValidationAndLockFailurePaths(t *testing.T) {
	dir := t.TempDir()
	lock := deploy.NewLock(filepath.Join(dir, "lock"))
	if err := lock.Release(); err != nil {
		t.Fatal(err)
	}
	if err := lock.Acquire(); err != nil {
		t.Fatal(err)
	}
	if err := lock.Acquire(); !errors.Is(err, deploy.ErrAlreadyLocked) {
		t.Fatalf("repeated acquisition: %v", err)
	}
	if err := lock.Release(); err != nil {
		t.Fatal(err)
	}
	if err := deploy.NewLock(filepath.Join(dir, "missing", "lock")).Acquire(); err == nil {
		t.Fatal("lock acquired without parent")
	}
	if err := deploy.ValidateComposeConfig([]byte("services: [")); !errors.Is(err, deploy.ErrInvalidComposeConfig) {
		t.Fatalf("invalid compose: %v", err)
	}
	if _, err := deploy.SelectRollbackManifest(deploy.ReleaseManifest{}, true); !errors.Is(err, deploy.ErrInvalidReleaseManifest) {
		t.Fatalf("invalid rollback manifest: %v", err)
	}
	if err := deploy.PauseAssignmentsForDeployment(nil, "alpha"); !errors.Is(err, deploy.ErrDeploymentOrder) {
		t.Fatalf("missing pauser: %v", err)
	}
	for _, steps := range [][]deploy.DeploymentStep{
		{deploy.StepAcquireLock},
		{deploy.StepPullImages, deploy.StepPauseAssignments, deploy.StepDrainRooms},
	} {
		if err := deploy.ValidateDeploymentSequence(steps); !errors.Is(err, deploy.ErrDeploymentOrder) {
			t.Fatalf("invalid deployment steps: %v", err)
		}
	}
	now := time.Now()
	if _, err := deploy.EvaluateManualRoomDrain(deploy.ManualRoomDrainRequest{}, 0, now); !errors.Is(err, deploy.ErrInvalidRoomDrainRequest) {
		t.Fatalf("empty drain: %v", err)
	}
	req := deploy.ManualRoomDrainRequest{DeploymentID: "one", RequestedAt: now, Deadline: now.Add(time.Minute), RoomCount: 2}
	for _, count := range []int{-1, 3} {
		if _, err := deploy.EvaluateManualRoomDrain(req, count, now); !errors.Is(err, deploy.ErrInvalidRoomDrainRequest) {
			t.Fatalf("invalid drain count: %v", err)
		}
	}
}
