package synapse_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"brain/internal/execution"
	"brain/internal/synapse"
)

func TestHostCallLeaseAdmissionAndBudget(t *testing.T) {
	s := newSession(t, 1, 2, 0, 0)
	r := req("draft")
	r.Runtime.Kind = execution.UserFunctionRuntimeKind
	r.Artifact.Entrypoint = "functions.invoke"
	r.Auth.Capability, r.Auth.Resource = "functions.invoke", "sample"
	r.Artifact.ProjectID = r.ProjectID
	r.Auth.OriginalCaller, r.Auth.Caller = "operator:root", "operator:root"
	ctx := execution.WithHostHandler(context.Background(), func(_ context.Context, op string, input json.RawMessage) (json.RawMessage, error) {
		return input, nil
	})
	done := submit(s, ctx, r)
	a := next(t, s).Assignment
	input := json.RawMessage(`{"id":"one"}`)
	if _, _, err := s.BeginHostCall(a.Lease, 1, "files.inspect", input); !errors.Is(err, synapse.ErrProtocol) {
		t.Fatal("unaccepted lease admitted", err)
	}
	if err := s.Accept(a.Lease); err != nil {
		t.Fatal(err)
	}
	for i := uint64(1); i <= execution.MaxHostCalls; i++ {
		callCtx, handler, err := s.BeginHostCall(a.Lease, i, "files.inspect", input)
		if err != nil {
			t.Fatal(err)
		}
		if parent, ok := execution.ParentFromContext(callCtx); !ok || parent != execution.ScopeOf(r) {
			t.Fatal("wrong parent identity")
		}
		if output, err := handler(callCtx, "files.inspect", input); err != nil || string(output) != string(input) {
			t.Fatal("wrong response", err)
		}
		if _, _, err := s.BeginHostCall(a.Lease, i+1, "files.inspect", input); !errors.Is(err, synapse.ErrProtocol) {
			t.Fatal("concurrent call admitted", err)
		}
		if err := s.Complete(a.Lease, success(r.ID)); !errors.Is(err, synapse.ErrProtocol) {
			t.Fatal("parent completed while child active", err)
		}
		if !s.FinishHostCall(a.Lease, i) || s.FinishHostCall(a.Lease, i) {
			t.Fatal("call completion not fenced")
		}
	}
	if _, _, err := s.BeginHostCall(a.Lease, execution.MaxHostCalls+1, "files.inspect", input); !errors.Is(err, synapse.ErrProtocol) {
		t.Fatal("budget exceeded", err)
	}
	if err := s.Complete(a.Lease, success(r.ID)); err != nil {
		t.Fatal(err)
	}
	if got := receive(t, done); got.err != nil {
		t.Fatal(got.err)
	}
	if _, _, err := s.BeginHostCall(a.Lease, 1, "files.inspect", input); !errors.Is(err, synapse.ErrProtocol) {
		t.Fatal("stale lease admitted", err)
	}
}

func TestPublicLeaseCannotRequestHostCalls(t *testing.T) {
	s := newSession(t, 1, 2, 0, 0)
	r := req("public")
	r.Runtime.Kind = execution.UserFunctionRuntimeKind
	r.Artifact.Entrypoint, r.Auth.Capability = "functions.call", "functions.call"
	r.Auth.OriginalCaller, r.Auth.Caller = "public", "public"
	ctx := execution.WithHostHandler(context.Background(), func(context.Context, string, json.RawMessage) (json.RawMessage, error) {
		t.Error("public host handler invoked")
		return nil, nil
	})
	done := submit(s, ctx, r)
	a := next(t, s).Assignment
	if err := s.Accept(a.Lease); err != nil {
		t.Fatal(err)
	}
	if _, _, err := s.BeginHostCall(a.Lease, 1, "files.inspect", json.RawMessage("null")); !errors.Is(err, synapse.ErrProtocol) {
		t.Fatal(err)
	}
	if err := s.Complete(a.Lease, success(r.ID)); err != nil {
		t.Fatal(err)
	}
	if got := receive(t, done); got.err != nil {
		t.Fatal(got.err)
	}
}
