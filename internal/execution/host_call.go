package execution

import (
	"context"
	"encoding/json"
	"errors"
)

const MaxHostCalls = 16
const MaxHostCallBytes = 256 << 10

var ErrHostCallDenied = errors.New("host call denied or unavailable")

// HostCallFunc is bound by trusted host or transport code, never serialized
// into an assignment or supplied by the user. It exposes no credentials.
type HostCallFunc func(context.Context, string, json.RawMessage) (json.RawMessage, error)

type hostHandlerKey struct{}
type hostCallerKey struct{}
type parentKey struct{}

func WithHostHandler(ctx context.Context, handler HostCallFunc) context.Context {
	return context.WithValue(ctx, hostHandlerKey{}, handler)
}

func HostHandler(ctx context.Context) HostCallFunc {
	handler, _ := ctx.Value(hostHandlerKey{}).(HostCallFunc)
	return handler
}

func WithHostCaller(ctx context.Context, caller HostCallFunc) context.Context {
	return context.WithValue(ctx, hostCallerKey{}, caller)
}

func CallHost(ctx context.Context, operation string, input json.RawMessage) (json.RawMessage, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	caller, _ := ctx.Value(hostCallerKey{}).(HostCallFunc)
	if caller == nil || !ValidHostCall(operation, input) {
		return nil, ErrHostCallDenied
	}
	return caller(ctx, operation, input)
}

func ValidHostCall(operation string, input json.RawMessage) bool {
	return present(operation) && len(operation) <= 128 && len(input) <= MaxHostCallBytes && json.Valid(input)
}

// ParentScope carries only host-selected ancestry. Child permissions remain
// independently checked by the target's normal registry authorization.
type ParentScope struct {
	ID, TraceID, ProjectID, OriginalCaller, Caller string
}

func ScopeOf(req Request) ParentScope {
	return ParentScope{ID: req.ID, TraceID: req.TraceID, ProjectID: req.ProjectID, OriginalCaller: req.Auth.OriginalCaller, Caller: req.Auth.Caller}
}

func WithParent(ctx context.Context, req Request) context.Context {
	return context.WithValue(ctx, parentKey{}, ScopeOf(req))
}

func ParentFromContext(ctx context.Context) (ParentScope, bool) {
	scope, ok := ctx.Value(parentKey{}).(ParentScope)
	return scope, ok
}
