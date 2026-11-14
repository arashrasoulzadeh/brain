package migrate_test

import (
	"context"
	"testing"

	"brain/internal/migrate"
)

func TestRunnerAppliesEachMigrationOnce(t *testing.T) {
	applications := 0
	marked := 0
	runner := migrate.NewRunner([]migrate.Migration{{Version: 1, Name: "platform", Apply: func(context.Context) error { applications++; return nil }}}, func(context.Context, int) error { marked++; return nil })
	if err := runner.Run(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := runner.Run(context.Background()); err != nil {
		t.Fatal(err)
	}
	if applications != 1 || marked != 1 {
		t.Fatalf("applications=%d marked=%d, want one each", applications, marked)
	}
}

func TestRunnerStopsBeforeMarkingFailedMigration(t *testing.T) {
	marked := 0
	runner := migrate.NewRunner([]migrate.Migration{{Version: 1, Name: "broken", Apply: func(context.Context) error { return context.Canceled }}}, func(context.Context, int) error { marked++; return nil })
	if err := runner.Run(context.Background()); err == nil {
		t.Fatal("Run() succeeded for failed migration")
	}
	if marked != 0 {
		t.Fatalf("marked=%d, want zero", marked)
	}
}

func TestRunOneShotRunsTheMigrationSetOnce(t *testing.T) {
	applications := 0
	err := migrate.RunOneShot(context.Background(), []migrate.Migration{{
		Version: 1,
		Name:    "platform",
		Apply: func(context.Context) error {
			applications++
			return nil
		},
	}}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if applications != 1 {
		t.Fatalf("applications = %d, want 1", applications)
	}
}
