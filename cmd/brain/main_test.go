package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestProductionServerComposesAuthenticatedAdminBoundary(t *testing.T) {
	t.Setenv("OPERATOR_USERNAME", "admin")
	t.Setenv("OPERATOR_SECRET", "secret")
	server := productionServer()
	login := httptest.NewRequest(http.MethodPost, "/admin/session", strings.NewReader(`{"username":"admin","secret":"secret"}`))
	loginResponse := httptest.NewRecorder()
	server.Handler().ServeHTTP(loginResponse, login)
	if loginResponse.Code != http.StatusOK {
		t.Fatalf("login status = %d, want %d", loginResponse.Code, http.StatusOK)
	}
	var payload struct {
		Session string `json:"session"`
	}
	if err := json.Unmarshal(loginResponse.Body.Bytes(), &payload); err != nil || payload.Session == "" {
		t.Fatalf("login payload = %q, err=%v", loginResponse.Body.String(), err)
	}
	action := httptest.NewRequest(http.MethodPost, "/admin/actions", strings.NewReader(`{"operation":"plugin.enable","project_id":"dev","plugin_id":"rooms","version":"1.0.0"}`))
	action.Header.Set("Authorization", "Bearer "+payload.Session)
	actionResponse := httptest.NewRecorder()
	server.Handler().ServeHTTP(actionResponse, action)
	if actionResponse.Code != http.StatusForbidden {
		t.Fatalf("admin action status = %d, want %d", actionResponse.Code, http.StatusForbidden)
	}
}
