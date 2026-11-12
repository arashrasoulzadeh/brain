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

func (o *Owner) Close() error {
	if o == nil || o.db == nil {
		return nil
	}
	return o.db.Close()
}
