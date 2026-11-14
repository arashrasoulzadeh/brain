package project_test

import (
	"context"
	"sync"
	"testing"

	"brain/internal/project"
)

func TestCallerContextIsBoundToTransactionContext(t *testing.T) {
	tx := project.WithCaller(context.Background(), "dev", "user-1")
	caller, ok := project.CallerFromContext(tx)
	if !ok || caller.ProjectID != "dev" || caller.ActorID != "user-1" {
		t.Fatalf("caller = %#v, %v", caller, ok)
	}
	if _, ok := project.CallerFromContext(context.Background()); ok {
		t.Fatal("caller leaked outside transaction context")
	}
}

func TestCallerFromNilContextFailsClosed(t *testing.T) {
	if caller, ok := project.CallerFromContext(nil); ok || caller != (project.Caller{}) {
		t.Fatalf("nil context = %#v, %v", caller, ok)
	}
}

func TestCallerContextRejectsMalformedProjectIdentity(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("malformed caller project was accepted")
		}
	}()
	_ = project.WithCaller(context.Background(), "dev:other", "actor")
}

func TestCallerLeaseRejectsMalformedProjectIdentity(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("malformed caller lease project was accepted")
		}
	}()
	_ = project.NewCallerLease("dev:other", "actor")
}

func TestCallerLeaseCanBeConsumedOnlyOnceConcurrently(t *testing.T) {
	lease := project.NewCallerLease("dev", "actor")
	var group sync.WaitGroup
	results := make(chan bool, 32)
	for index := 0; index < 32; index++ {
		group.Add(1)
		go func() {
			defer group.Done()
			_, ok := lease.Consume()
			results <- ok
		}()
	}
	group.Wait()
	close(results)
	consumed := 0
	for ok := range results {
		if ok {
			consumed++
		}
	}
	if consumed != 1 {
		t.Fatalf("lease consumed %d times, want once", consumed)
	}
}

func TestCallerLeaseCannotBeReused(t *testing.T) {
	lease := project.NewCallerLease("dev", "user-1")
	if _, ok := lease.Consume(); !ok {
		t.Fatal("first caller lease rejected")
	}
	if _, ok := lease.Consume(); ok {
		t.Fatal("pooled caller lease reused")
	}
}
