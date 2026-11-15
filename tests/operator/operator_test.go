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
	if err := store.Authenticate("admin", "secret", func(string) string { return "digest" }); err != nil {
		t.Fatal(err)
	}
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
	events := store.Audit()
	if len(events) != 3 || events[0].Action != "operator.signin" || events[0].Success != true || events[1].Success != false || events[2].Success != false {
		t.Fatalf("signin audit = %#v", events)
	}
}

func TestAuthenticateRejectsMissingAccount(t *testing.T) {
	if err := (&operator.Store{}).Authenticate("admin", "secret", func(string) string { return "digest" }); !errors.Is(err, operator.ErrInvalidCredentials) {
		t.Fatalf("missing account = %v", err)
	}
}

func TestStoreRejectsEndUserOperatorAction(t *testing.T) {
	store := &operator.Store{}
	_, _ = store.Bootstrap("admin", "secret", func(string) string { return "digest" })
	if err := store.AuthorizeAction(operator.Actor{ID: "user-1", Kind: operator.ActorUser}, "plugin.enable"); !errors.Is(err, operator.ErrOperatorRequired) {
		t.Fatalf("ordinary user authorization error = %v", err)
	}
}

func TestSensitiveOperatorSetupRequiresAuthentication(t *testing.T) {
	store := &operator.Store{}
	_, _ = store.Bootstrap("admin", "secret", func(string) string { return "digest" })
	if err := store.EnableMFA("JBSWY3DPEHPK3PXP"); !errors.Is(err, operator.ErrInvalidCredentials) {
		t.Fatalf("unauthenticated MFA enrollment = %v", err)
	}
	if _, err := store.CreateRecoveryToken(func() string { return "token" }); !errors.Is(err, operator.ErrInvalidCredentials) {
		t.Fatalf("unauthenticated recovery token = %v", err)
	}
}

func TestStoreRejectsForgedOperatorBeforeAuthentication(t *testing.T) {
	store := &operator.Store{}
	digest := func(string) string { return "digest" }
	_, _ = store.Bootstrap("admin", "secret", digest)
	if err := store.AuthorizeAction(operator.Actor{ID: "admin", Kind: operator.ActorOperator}, "plugin.enable"); !errors.Is(err, operator.ErrOperatorRequired) {
		t.Fatalf("unauthenticated operator authorization error = %v", err)
	}
	if err := store.Authenticate("admin", "secret", digest); err != nil {
		t.Fatal(err)
	}
	if err := store.AuthorizeAction(operator.Actor{ID: "admin", Kind: operator.ActorOperator}, "plugin.enable"); err != nil {
		t.Fatalf("authenticated operator authorization error = %v", err)
	}
}

func TestOperatorSessionIsAuthenticatedAndRevocable(t *testing.T) {
	store := &operator.Store{}
	digest := func(value string) string { return "digest:" + value }
	if _, err := store.Bootstrap("admin", "secret", digest); err != nil {
		t.Fatal(err)
	}
	if err := store.Authenticate("admin", "secret", digest); err != nil {
		t.Fatal(err)
	}
	token, err := store.IssueSession(func() string { return "operator-session" })
	if err != nil || token != "operator-session" {
		t.Fatalf("IssueSession() = %q, %v", token, err)
	}
	actor, ok := store.ValidateSession(token)
	if !ok || actor.ID != "admin" || actor.Kind != operator.ActorOperator {
		t.Fatalf("ValidateSession() = %#v, %v", actor, ok)
	}
	if err := store.RevokeSession(token); err != nil {
		t.Fatal(err)
	}
	if _, ok := store.ValidateSession(token); ok {
		t.Fatal("revoked operator session remained valid")
	}
}

func TestExecuteActionAuthorizesAndAuditsOperation(t *testing.T) {
	store := &operator.Store{}
	digest := func(string) string { return "digest" }
	_, _ = store.Bootstrap("admin", "secret", digest)
	if err := store.Authenticate("admin", "secret", digest); err != nil {
		t.Fatal(err)
	}
	actor := operator.Actor{ID: "admin", Kind: operator.ActorOperator}
	called := false
	if err := store.ExecuteAction(actor, "plugin.remove", func() error {
		called = true
		return nil
	}); err != nil || !called {
		t.Fatalf("authorized action = %v, called=%v", err, called)
	}
	if err := store.ExecuteAction(actor, "migration.apply", func() error { return errors.New("migration failed") }); err == nil {
		t.Fatal("failed operation was accepted")
	}
	events := store.Audit()
	if len(events) != 3 || events[1].Action != "plugin.remove" || !events[1].Success || events[2].Action != "migration.apply" || events[2].Success {
		t.Fatalf("action audit = %#v", events)
	}
}

func TestExecuteActionRejectsMissingOperation(t *testing.T) {
	store := &operator.Store{}
	if err := store.ExecuteAction(operator.Actor{}, "plugin.remove", nil); !errors.Is(err, operator.ErrInvalidOperatorAction) {
		t.Fatalf("missing operation = %v", err)
	}
}

func TestRecoveryRevokesAuthenticatedOperatorAccess(t *testing.T) {
	store := &operator.Store{}
	digest := func(value string) string { return "digest:" + value }
	_, _ = store.Bootstrap("admin", "old", digest)
	if err := store.Authenticate("admin", "old", digest); err != nil {
		t.Fatal(err)
	}
	token, err := store.CreateRecoveryToken(func() string { return "one-time-token" })
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Recover(token, "new", digest); err != nil {
		t.Fatal(err)
	}
	if err := store.AuthorizeAction(operator.Actor{ID: "admin", Kind: operator.ActorOperator}, "plugin.enable"); !errors.Is(err, operator.ErrOperatorRequired) {
		t.Fatalf("recovered operator authorization error = %v", err)
	}
}

func TestRecoveryTokenReplacesSecretAndIsSingleUse(t *testing.T) {
	store := &operator.Store{}
	digest := func(value string) string { return "digest:" + value }
	_, _ = store.Bootstrap("admin", "old", digest)
	if err := store.Authenticate("admin", "old", digest); err != nil {
		t.Fatal(err)
	}
	if err := store.EnableMFA("JBSWY3DPEHPK3PXP"); err != nil {
		t.Fatal(err)
	}
	token, err := store.CreateRecoveryToken(func() string { return "one-time-token" })
	if err != nil || token != "one-time-token" {
		t.Fatalf("CreateRecoveryToken() = %q, %v", token, err)
	}
	if err := store.Recover(token, "new", func(value string) string { return "digest:" + value }); err != nil {
		t.Fatal(err)
	}
	if err := store.Authenticate("admin", "new", digest); err != nil {
		t.Fatal("new secret rejected")
	}
	if store.VerifyMFA("000000", time.Unix(1700000000, 0)) {
		t.Fatal("MFA was not cleared")
	}
	events := store.Audit()
	if len(events) != 3 || events[0].Action != "operator.signin" || !events[0].Success || events[1].Action != "operator.recovery" || !events[1].Success || events[2].Action != "operator.signin" {
		t.Fatalf("recovery audit = %#v", events)
	}
	if events[1].Detail != "admin" {
		t.Fatal("recovery audit exposed sensitive data")
	}
}
