package operator_test

import (
	"errors"
	"testing"
	"time"

	"brain/internal/operator"
)

func TestBootstrapCreatesSeparateOperatorAccount(t *testing.T) {
	store := &operator.Store{}
	account, err := store.Bootstrap("admin", "secret", func(value string) string { return "digest:" + value })
	if err != nil {
		t.Fatal(err)
	}
	if account.Username != "admin" || account.SecretDigest != "digest:secret" {
		t.Fatalf("account = %#v", account)
	}
	if account.SecretDigest == "secret" {
		t.Fatal("raw secret was stored")
	}
}

func TestMFAAcceptsCurrentAndAdjacentTOTPWindow(t *testing.T) {
	store := &operator.Store{}
	_, _ = store.Bootstrap("admin", "secret", func(string) string { return "digest" })
	if err := store.EnableMFA("JBSWY3DPEHPK3PXP"); err != nil {
		t.Fatal(err)
	}
	now := time.Unix(1700000000, 0)
	if !store.VerifyMFA("324550", now) {
		t.Fatal("known TOTP code was rejected")
	}
	if store.VerifyMFA("000000", now) {
		t.Fatal("invalid TOTP code was accepted")
	}
}

func TestBootstrapRunsOnlyOnce(t *testing.T) {
	store := &operator.Store{}
	_, _ = store.Bootstrap("admin", "secret", func(string) string { return "digest" })
	if _, err := store.Bootstrap("other", "secret", func(string) string { return "digest" }); !errors.Is(err, operator.ErrAlreadyBootstrapped) {
		t.Fatalf("Bootstrap() = %v", err)
	}
}

func TestAuthenticateAcceptsOnlyMatchingOperatorCredentials(t *testing.T) {
	store := &operator.Store{}
	digest := func(value string) string { return "digest:" + value }
	_, _ = store.Bootstrap("admin", "secret", digest)
	if err := store.Authenticate("admin", "secret", digest); err != nil {
		t.Fatal(err)
	}
	if err := store.Authenticate("admin", "wrong", digest); !errors.Is(err, operator.ErrInvalidCredentials) {
		t.Fatalf("wrong secret = %v", err)
	}
	if err := store.Authenticate("other", "secret", digest); !errors.Is(err, operator.ErrInvalidCredentials) {
		t.Fatalf("wrong user = %v", err)
	}
}

func TestAuthenticateRejectsMissingAccount(t *testing.T) {
	if err := (&operator.Store{}).Authenticate("admin", "secret", func(string) string { return "digest" }); !errors.Is(err, operator.ErrInvalidCredentials) {
		t.Fatalf("missing account = %v", err)
	}
}
