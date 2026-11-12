package operator_test

import (
	"errors"
	"testing"

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

func TestBootstrapRunsOnlyOnce(t *testing.T) {
	store := &operator.Store{}
	_, _ = store.Bootstrap("admin", "secret", func(string) string { return "digest" })
	if _, err := store.Bootstrap("other", "secret", func(string) string { return "digest" }); !errors.Is(err, operator.ErrAlreadyBootstrapped) {
		t.Fatalf("Bootstrap() = %v", err)
	}
}
