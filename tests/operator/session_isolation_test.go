package operator_test

import (
	"fmt"
	"sync"
	"testing"

	"brain/internal/operator"
)

func TestOperatorSessionsAreIndependent(t *testing.T) {
	store := &operator.Store{}
	digest := func(value string) string { return "digest:" + value }
	if _, err := store.Bootstrap("admin", "secret", digest); err != nil {
		t.Fatal(err)
	}
	first, err := store.AuthenticateSession("admin", "secret", digest, func() string { return "first" })
	if err != nil {
		t.Fatal(err)
	}
	second, err := store.AuthenticateSession("admin", "secret", digest, func() string { return "second" })
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.AuthenticateSession("admin", "incorrect", digest, func() string { return "invalid" }); err == nil {
		t.Fatal("incorrect credentials issued a session")
	}
	actor, ok := store.ValidateSession(first)
	if !ok {
		t.Fatal("failed sign-in invalidated an existing session")
	}
	if err := store.AuthorizeAction(actor, "plugin.enable"); err != nil {
		t.Fatal(err)
	}
	if err := store.AuthorizeAction(operator.Actor{ID: "admin", Kind: operator.ActorOperator}, "plugin.enable"); err == nil {
		t.Fatal("HTTP login authorized an unbound actor")
	}
	if err := store.RevokeSession(first); err != nil {
		t.Fatal(err)
	}
	if err := store.AuthorizeAction(actor, "plugin.enable"); err == nil {
		t.Fatal("revoked session still authorized an action")
	}
	if _, ok := store.ValidateSession(second); !ok {
		t.Fatal("revoking one session invalidated another")
	}
}

func TestOperatorSessionsConcurrentAccess(t *testing.T) {
	store := &operator.Store{}
	digest := func(value string) string { return "digest:" + value }
	if _, err := store.Bootstrap("admin", "secret", digest); err != nil {
		t.Fatal(err)
	}
	var group sync.WaitGroup
	for i := 0; i < 32; i++ {
		group.Add(1)
		go func(i int) {
			defer group.Done()
			token, err := store.AuthenticateSession("admin", "secret", digest, func() string {
				return fmt.Sprintf("session-%d", i)
			})
			if err != nil {
				t.Error(err)
				return
			}
			actor, ok := store.ValidateSession(token)
			if !ok {
				t.Error("session unexpectedly invalid")
				return
			}
			if err := store.ExecuteAction(actor, "test.action", func() error { return nil }); err != nil {
				t.Error(err)
			}
			_ = store.Audit()
			if err := store.RevokeSession(token); err != nil {
				t.Error(err)
			}
		}(i)
	}
	group.Wait()
	if got := len(store.Audit()); got != 64 {
		t.Errorf("audit event count = %d, want 64", got)
	}
}

func TestRecoveryInvalidatesIssuedSessions(t *testing.T) {
	store := &operator.Store{}
	digest := func(value string) string { return "digest:" + value }
	if _, err := store.Bootstrap("admin", "secret", digest); err != nil {
		t.Fatal(err)
	}
	session, err := store.AuthenticateSession("admin", "secret", digest, func() string { return "session" })
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Authenticate("admin", "secret", digest); err != nil {
		t.Fatal(err)
	}
	recovery, err := store.CreateRecoveryToken(func() string { return "recovery" })
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Recover(recovery, "new-secret", digest); err != nil {
		t.Fatal(err)
	}
	if _, ok := store.ValidateSession(session); ok {
		t.Fatal("recovery left an existing session valid")
	}
}
