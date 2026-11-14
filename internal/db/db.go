package db

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"brain/internal/project"
)

var ErrNotReady = errors.New("database not ready")
var ErrDDLForbidden = errors.New("runtime role cannot execute DDL")
var ErrProjectContextRequired = errors.New("project context required for runtime SQL")
var ErrProjectContextMismatch = errors.New("runtime SQL project context mismatch")

type Role string

const (
	RuntimeRole   Role = "runtime"
	MigrationRole Role = "migration"
)

func AllowDDL(role Role) error {
	if role != MigrationRole {
		return ErrDDLForbidden
	}
	return nil
}

type Opener func(driver, dsn string) (*sql.DB, error)

type Owner struct {
	db *sql.DB
}

func Open(driver, dsn string, opener Opener) (*Owner, error) {
	db, err := opener(driver, dsn)
	if err != nil {
		return nil, err
	}
	return &Owner{db: db}, nil
}

func (o *Owner) Ready(ctx context.Context, timeout time.Duration) error {
	if o == nil || o.db == nil {
		return ErrNotReady
	}
	checkCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	if err := o.db.PingContext(checkCtx); err != nil {
		return ErrNotReady
	}
	return nil
}

// WaitReady retries a readiness check until it succeeds or the context ends.
func WaitReady(ctx context.Context, check func(context.Context) error, interval time.Duration) error {
	if err := check(ctx); err == nil {
		return nil
	}
	timer := time.NewTimer(interval)
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
			if err := check(ctx); err == nil {
				return nil
			}
			timer.Reset(interval)
		}
	}
}

func (o *Owner) Close() error {
	if o == nil || o.db == nil {
		return nil
	}
	return o.db.Close()
}

func (o *Owner) ExecProjectContext(ctx context.Context, projectID, query string, args ...any) (sql.Result, error) {
	if err := validateProjectContext(ctx, projectID); err != nil {
		return nil, err
	}
	if o == nil || o.db == nil {
		return nil, ErrNotReady
	}
	return o.db.ExecContext(ctx, query, args...)
}

func (o *Owner) QueryProjectContext(ctx context.Context, projectID, query string, args ...any) (*sql.Rows, error) {
	if err := validateProjectContext(ctx, projectID); err != nil {
		return nil, err
	}
	if o == nil || o.db == nil {
		return nil, ErrNotReady
	}
	return o.db.QueryContext(ctx, query, args...)
}

func validateProjectContext(ctx context.Context, projectID string) error {
	if ctx == nil {
		return ErrProjectContextRequired
	}
	caller, ok := project.FromContext(ctx)
	if !ok || caller.ID == "" {
		return ErrProjectContextRequired
	}
	if projectID == "" || caller.ID != projectID {
		return ErrProjectContextMismatch
	}
	return nil
}
