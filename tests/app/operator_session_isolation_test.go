package app_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"brain/internal/admin"
	"brain/internal/app"
	"brain/internal/operator"
	"brain/internal/plugin"
	"brain/internal/project"
)

func TestFailedOperatorLoginDoesNotInvalidateExistingSession(t *testing.T) {
	store := &operator.Store{}
	digest := func(value string) string { return "digest:" + value }
	if _, err := store.Bootstrap("admin", "secret", digest); err != nil {
		t.Fatal(err)
	}
	projects := project.NewStore()
	service := admin.NewLifecycleServiceWithProjects(store, plugin.NewRegistry(1), projects)
	handler := app.NewServerWithOperatorAdmin(service, store, digest).Handler()

	login := httptest.NewRecorder()
	handler.ServeHTTP(login, httptest.NewRequest(http.MethodPost, "/admin/session", strings.NewReader(`{"username":"admin","secret":"secret"}`)))
	if login.Code != http.StatusOK {
		t.Fatalf("login status = %d, body = %s", login.Code, login.Body.String())
	}
	var response struct {
		Session string `json:"session"`
	}
	if err := json.Unmarshal(login.Body.Bytes(), &response); err != nil || response.Session == "" {
		t.Fatalf("login response = %q, err = %v", login.Body.String(), err)
	}

	failed := httptest.NewRecorder()
	handler.ServeHTTP(failed, httptest.NewRequest(http.MethodPost, "/admin/session", strings.NewReader(`{"username":"admin","secret":"wrong"}`)))
	if failed.Code != http.StatusForbidden {
		t.Fatalf("failed login status = %d", failed.Code)
	}
	action := httptest.NewRequest(http.MethodPost, "/admin/actions", strings.NewReader(`{"operation":"project.create","project_id":"session-test"}`))
	action.Header.Set("Authorization", "Bearer "+response.Session)
	result := httptest.NewRecorder()
	handler.ServeHTTP(result, action)
	if result.Code < 200 || result.Code >= 300 {
		t.Fatalf("existing session action status = %d, body = %s", result.Code, result.Body.String())
	}
	if _, err := projects.Resolve("session-test"); err != nil {
		t.Fatal(err)
	}
}
