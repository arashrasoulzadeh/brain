package operator_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"brain/internal/operator"
)

func TestOperatorAccountRecoveryAndSessionFailurePaths(t *testing.T) {
	now := time.Unix(1000, 0)
	store := operator.NewStoreWithClock(func() time.Time { return now })
	digest := func(value string) string { return "digest:" + value }
	if _, ok := store.Account(); ok {
		t.Fatal("empty store has an account")
	}
	if _, ok := store.ValidateSession(""); ok {
		t.Fatal("empty session accepted")
	}
	if _, err := store.IssueSession(func() string { return "token" }); !errors.Is(err, operator.ErrInvalidCredentials) {
		t.Fatalf("unauthenticated issuance: %v", err)
	}
	for _, name := range []string{"", " "} {
		if _, err := store.Bootstrap(name, "secret", digest); !errors.Is(err, operator.ErrInvalidBootstrap) {
			t.Fatalf("invalid bootstrap: %v", err)
		}
	}
	if _, err := store.Bootstrap("admin", "secret", digest); err != nil {
		t.Fatal(err)
	}
	account, ok := store.Account()
	if !ok || account.Username != "admin" {
		t.Fatal("account missing")
	}
	account.Username = "changed"
	if original, _ := store.Account(); original.Username != "admin" {
		t.Fatal("account snapshot modified store")
	}
	if err := store.Authenticate("admin", "secret", digest); err != nil {
		t.Fatal(err)
	}
	if _, err := store.CreateRecoveryToken(func() string { return "" }); !errors.Is(err, operator.ErrInvalidCredentials) {
		t.Fatalf("empty recovery token: %v", err)
	}
	if err := store.Recover("missing", "new", digest); !errors.Is(err, operator.ErrInvalidCredentials) {
		t.Fatalf("missing recovery credential: %v", err)
	}
	token, err := store.CreateRecoveryToken(func() string { return "recovery" })
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Recover("wrong", "new", digest); !errors.Is(err, operator.ErrInvalidCredentials) {
		t.Fatalf("incorrect recovery credential: %v", err)
	}
	if err := store.Recover(token, "new", digest); err != nil {
		t.Fatal(err)
	}
	if err := store.Recover(token, "again", digest); !errors.Is(err, operator.ErrInvalidCredentials) {
		t.Fatalf("replayed recovery credential: %v", err)
	}
	if err := store.Authenticate("admin", "new", digest); err != nil {
		t.Fatal(err)
	}
	if _, err := store.IssueSession(nil); !errors.Is(err, operator.ErrInvalidCredentials) {
		t.Fatalf("missing token source: %v", err)
	}
	if _, err := store.IssueSession(func() string { return "" }); !errors.Is(err, operator.ErrInvalidSession) {
		t.Fatalf("empty token: %v", err)
	}
	first, err := store.IssueSession(func() string { return "first" })
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.IssueSession(func() string { return first }); !errors.Is(err, operator.ErrInvalidSession) {
		t.Fatalf("duplicate token: %v", err)
	}
	now = now.Add(operator.DefaultSessionLifetime)
	if _, err := store.IssueSession(func() string { return first }); err != nil {
		t.Fatalf("expired session was not pruned: %v", err)
	}
	if err := store.RevokeSession(""); !errors.Is(err, operator.ErrInvalidSession) {
		t.Fatalf("empty revocation: %v", err)
	}
	if _, ok := operator.ActorFromContext(nil); ok {
		t.Fatal("nil context contained an actor")
	}
	ctx := operator.WithActor(context.Background(), operator.Actor{ID: "admin", Kind: operator.ActorOperator})
	if actor, ok := operator.ActorFromContext(ctx); !ok || actor.ID != "admin" {
		t.Fatalf("actor context: %#v %t", actor, ok)
	}
}

func TestMFARejectsMalformedEnrollmentAndEmptyCode(t *testing.T) {
	store := &operator.Store{}
	digest := func(value string) string { return "digest:" + value }
	if _, err := store.Bootstrap("admin", "secret", digest); err != nil {
		t.Fatal(err)
	}
	if err := store.Authenticate("admin", "secret", digest); err != nil {
		t.Fatal(err)
	}
	for _, secret := range []string{"", " ", "invalid!", "A"} {
		if err := store.EnableMFA(secret); !errors.Is(err, operator.ErrInvalidCredentials) {
			t.Fatalf("malformed MFA secret was accepted: %v", err)
		}
		if store.VerifyMFA("", time.Unix(1700000000, 0)) {
			t.Fatal("empty MFA code was accepted")
		}
	}
	if err := store.EnableMFA("JBSWY3DPEHPK3PXP"); err != nil {
		t.Fatal(err)
	}
	if store.VerifyMFA("", time.Unix(1700000000, 0)) {
		t.Fatal("empty MFA code was accepted for valid enrollment")
	}
	if !store.VerifyMFA("324550", time.Unix(1700000000, 0)) {
		t.Fatal("valid MFA code was rejected")
	}
}
