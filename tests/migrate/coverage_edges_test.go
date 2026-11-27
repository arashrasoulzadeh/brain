package migrate_test

import (
	"context"
	"errors"
	"testing"

	"brain/internal/migrate"
)

func TestMigrationRejectsMissingApplyAndRetriesFailedMark(t *testing.T) {
	ctx := context.Background()
	if err := migrate.NewRunner([]migrate.Migration{{Version: 1}}, nil).Run(ctx); err == nil {
		t.Fatal("missing apply function accepted")
	}
	failure := errors.New("mark unavailable")
	applications, marks := 0, 0
	runner := migrate.NewRunner([]migrate.Migration{{
		Version: 1,
		Apply: func(context.Context) error {
			applications++
			return nil
		},
	}}, func(context.Context, int) error {
		marks++
		if marks == 1 {
			return failure
		}
		return nil
	})
	if err := runner.Run(ctx); !errors.Is(err, failure) {
		t.Fatalf("mark failure: %v", err)
	}
	if err := runner.Run(ctx); err != nil {
		t.Fatal(err)
	}
	if applications != 2 || marks != 2 {
		t.Fatalf("apply and mark counts: %d, %d", applications, marks)
	}
	if err := migrate.NewRunner(migrate.PlatformMigrations, nil).Run(ctx); err != nil {
		t.Fatal(err)
	}
}
