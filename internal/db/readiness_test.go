package db_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"brain/internal/db"
)

func TestWaitReadyRecoversAfterTransientFailure(t *testing.T) {
	attempts := 0
	err := db.WaitReady(context.Background(), func(context.Context) error {
		attempts++
		if attempts < 3 {
			return errors.New("not ready")
		}
		return nil
	}, time.Millisecond)
	if err != nil || attempts != 3 {
		t.Fatalf("WaitReady() = %v after %d attempts", err, attempts)
	}
}

func TestWaitReadyStopsWhenContextExpires(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Millisecond)
	defer cancel()
	err := db.WaitReady(ctx, func(context.Context) error { return errors.New("offline") }, time.Millisecond)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("WaitReady() = %v, want deadline", err)
	}
}
