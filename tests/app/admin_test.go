package app_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"brain/internal/admin"
	"brain/internal/app"
	"brain/internal/operator"
	"brain/internal/plugin"
)

func TestAdminHandlerFailsClosedWithoutTrustedOperatorContext(t *testing.T) {
	service := admin.NewLifecycleService(&operator.Store{}, plugin.NewRegistry(1))
	req := httptest.NewRequest(http.MethodPost, "/admin/actions", strings.NewReader(`{"operation":"plugin.enable"}`))
	rec := httptest.NewRecorder()
	app.AdminHandler(service).ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}

func TestAdminHandlerDoesNotTrustActorFromBody(t *testing.T) {
	service := admin.NewLifecycleService(&operator.Store{}, plugin.NewRegistry(1))
	req := httptest.NewRequest(http.MethodPost, "/admin/actions", strings.NewReader(`{"operation":"plugin.enable","actor":{"id":"admin","kind":"operator"}}`))
	req = req.WithContext(context.Background())
	rec := httptest.NewRecorder()
	app.AdminHandler(service).ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}

func TestServerWiresAuthenticatedAdminCommand(t *testing.T) {
	operators := &operator.Store{}
	digest := func(string) string { return "digest" }
	_, _ = operators.Bootstrap("admin", "secret", digest)
	if err := operators.Authenticate("admin", "secret", digest); err != nil {
		t.Fatal(err)
	}
	service := admin.NewLifecycleService(operators, plugin.NewRegistry(1))
	actor := operator.Actor{ID: "admin", Kind: operator.ActorOperator}
	request := httptest.NewRequest(http.MethodPost, "/admin/actions", strings.NewReader(`{"operation":"plugin.enable","project_id":"dev","plugin_id":"rooms","version":"1.0.0"}`))
	request = request.WithContext(operator.WithActor(context.Background(), actor))
	recorder := httptest.NewRecorder()
	app.NewServerWithAdmin(service).Handler().ServeHTTP(recorder, request)
	if recorder.Code != http.StatusForbidden {
		t.Fatalf("admin command status = %d, want %d", recorder.Code, http.StatusForbidden)
	}
	if events := operators.Audit(); len(events) != 2 || events[1].Action != "plugin.enable" {
		t.Fatalf("admin command audit = %#v", events)
	}
}

func TestAdminHandlerRecognizesLifecycleOperations(t *testing.T) {
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
	for _, operation := range []string{
		"plugin.revoke", "plugin.promote", "plugin.drain", "plugin.stop-immediate",
		"plugin.recover", "plugin.rollback", "project.uninstall", "durable-data.delete",
	} {
		request := httptest.NewRequest(http.MethodPost, "/admin/actions", strings.NewReader(`{"operation":"`+operation+`","project_id":"dev","plugin_id":"rooms","version":"1.0.0","key":"state","confirmed":true}`))
		request = request.WithContext(operator.WithActor(context.Background(), actor))
		recorder := httptest.NewRecorder()
		app.AdminHandler(service).ServeHTTP(recorder, request)
		if recorder.Code == http.StatusBadRequest {
			t.Fatalf("operation %q was not recognized", operation)
		}
	}
}

func TestAdminHandlerRequiresConfirmationForDestructiveOperations(t *testing.T) {
	operators := &operator.Store{}
	digest := func(string) string { return "digest" }
	if _, err := operators.Bootstrap("admin", "secret", digest); err != nil {
		t.Fatal(err)
	}
	if err := operators.Authenticate("admin", "secret", digest); err != nil {
		t.Fatal(err)
	}
	service := admin.NewLifecycleService(operators, plugin.NewRegistry(1))
	request := httptest.NewRequest(http.MethodPost, "/admin/actions", strings.NewReader(`{"operation":"plugin.remove","project_id":"dev","plugin_id":"rooms","version":"1.0.0"}`))
	request = request.WithContext(operator.WithActor(context.Background(), operator.Actor{ID: "admin", Kind: operator.ActorOperator}))
	recorder := httptest.NewRecorder()
	app.AdminHandler(service).ServeHTTP(recorder, request)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("unconfirmed destructive operation status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
}
