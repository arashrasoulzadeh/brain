package project

import (
	"context"
	"sync"
)

type Caller struct {
	ProjectID string
	ActorID   string
}
type callerContextKey struct{}

func WithCaller(ctx context.Context, projectID, actorID string) context.Context {
	if projectID == "" || actorID == "" {
		panic("caller context requires project and actor")
	}
	return context.WithValue(ctx, callerContextKey{}, Caller{ProjectID: projectID, ActorID: actorID})
}

func CallerFromContext(ctx context.Context) (Caller, bool) {
	if ctx == nil {
		return Caller{}, false
	}
	caller, ok := ctx.Value(callerContextKey{}).(Caller)
	return caller, ok
}

type CallerLease struct {
	mu     sync.Mutex
	caller Caller
	used   bool
}

func NewCallerLease(projectID, actorID string) *CallerLease {
	return &CallerLease{caller: Caller{ProjectID: projectID, ActorID: actorID}}
}

func (l *CallerLease) Consume() (Caller, bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.used || l.caller.ProjectID == "" || l.caller.ActorID == "" {
		return Caller{}, false
	}
	l.used = true
	return l.caller, true
}
