package admin_test

import (
	"errors"
	"testing"

	"brain/internal/admin"
	"brain/internal/operator"
	"brain/internal/plugin"
)

func TestLifecycleServiceRequiresAuthenticatedOperator(t *testing.T) {
	operators := &operator.Store{}
	digest := func(string) string { return "digest" }
	if _, err := operators.Bootstrap("admin", "secret", digest); err != nil {
		t.Fatal(err)
	}
	registry := plugin.NewRegistry(1)
	service := admin.NewLifecycleService(operators, registry)
	if err := service.EnablePlugin(operator.Actor{ID: "admin", Kind: operator.ActorOperator}, "dev", "rooms", "1.0.0"); !errors.Is(err, operator.ErrOperatorRequired) {
		t.Fatalf("unauthenticated lifecycle action = %v", err)
	}
	if events := operators.Audit(); len(events) != 1 || events[0].Action != "plugin.enable" || events[0].Success {
		t.Fatalf("authorization audit = %#v", events)
	}
}

func TestLifecycleServiceRoutesAuthorizedActionAndAuditsFailure(t *testing.T) {
	operators := &operator.Store{}
	digest := func(string) string { return "digest" }
	if _, err := operators.Bootstrap("admin", "secret", digest); err != nil {
		t.Fatal(err)
	}
	if err := operators.Authenticate("admin", "secret", digest); err != nil {
		t.Fatal(err)
	}
	service := admin.NewLifecycleService(operators, plugin.NewRegistry(1))
	actor := operator.Actor{ID: "admin", Kind: operator.ActorOperator}
	if err := service.RemovePlugin(actor, "dev", "rooms", "1.0.0"); !errors.Is(err, plugin.ErrNotInstalled) {
		t.Fatalf("lifecycle registry error = %v", err)
	}
	events := operators.Audit()
	if len(events) != 2 || events[1].Action != "plugin.remove" || events[1].Success {
		t.Fatalf("failure audit = %#v", events)
	}
}
