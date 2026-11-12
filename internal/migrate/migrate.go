package migrate

import (
	"context"
	"fmt"
)

type Migration struct {
	Version int
	Name    string
	Apply   func(context.Context) error
}

type Runner struct {
	migrations []Migration
	mark       func(context.Context, int) error
	applied    map[int]bool
}

func NewRunner(migrations []Migration, mark func(context.Context, int) error) *Runner {
	return &Runner{migrations: migrations, mark: mark, applied: map[int]bool{}}
}

func (r *Runner) Run(ctx context.Context) error {
	for _, migration := range r.migrations {
		if r.applied[migration.Version] {
			continue
		}
		if migration.Apply == nil {
			return fmt.Errorf("migration %d has no apply function", migration.Version)
		}
		if err := migration.Apply(ctx); err != nil {
			return fmt.Errorf("migration %d (%s): %w", migration.Version, migration.Name, err)
		}
		if r.mark != nil {
			if err := r.mark(ctx, migration.Version); err != nil {
				return fmt.Errorf("mark migration %d: %w", migration.Version, err)
			}
		}
		r.applied[migration.Version] = true
	}
	return nil
}
