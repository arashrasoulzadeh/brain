package db

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

var ErrNotReady = errors.New("database not ready")

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
