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
