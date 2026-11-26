package operator_test

import (
	"testing"
	"time"

	"brain/internal/operator"
)

func TestOperatorSessionExpiryIsAbsoluteAndRecheckedForActions(t *testing.T) {
	now := time.Unix(1000, 0)
	store := operator.NewStoreWithClock(func() time.Time { return now })
	digest := func(value string) string { return "digest:" + value }
	if _, err := store.Bootstrap("admin", "secret", digest); err != nil {
		t.Fatal(err)
	}
	first, err := store.AuthenticateSession("admin", "secret", digest, func() string { return "first" })
	if err != nil {
		t.Fatal(err)
	}
	now = now.Add(operator.DefaultSessionLifetime - time.Second)
	actor, ok := store.ValidateSession(first)
	if !ok {
		t.Fatal("session expired too early")
	}
	second, err := store.AuthenticateSession("admin", "secret", digest, func() string { return "second" })
	if err != nil {
		t.Fatal(err)
	}
	now = now.Add(time.Second)
	if _, ok := store.ValidateSession(first); ok {
		t.Fatal("session remained valid at its expiry boundary")
	}
	if err := store.AuthorizeAction(actor, "plugin.enable"); err == nil {
		t.Fatal("previously validated actor remained authorized after session expiry")
	}
	if _, ok := store.ValidateSession(second); !ok {
		t.Fatal("independent newer session expired with the older session")
	}
}
