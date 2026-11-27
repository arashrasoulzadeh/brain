package app

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"brain/internal/admin"
	"brain/internal/operator"
	"brain/internal/plugin"
	"brain/internal/project"
)

func TestCoverageAdminCommandResponses(t *testing.T) {
	operators := &operator.Store{}
	digest := func(value string) string { return value }
	if _, err := operators.Bootstrap("admin", "secret", digest); err != nil {
		t.Fatal(err)
	}
	if err := operators.Authenticate("admin", "secret", digest); err != nil {
		t.Fatal(err)
	}
	actor := operator.Actor{Kind: operator.ActorOperator, ID: "admin"}
	registry := plugin.NewRegistry(1)
	manifest := plugin.Manifest{ID: "sample", Version: "1.0.0", HostAPI: 1, Declarative: true, Capabilities: []string{"read"}}
	for _, err := range []error{registry.Register(manifest), registry.Install("coverage", manifest.ID, manifest.Version), registry.Grant("coverage", manifest.ID, manifest.Version, "read"), registry.SetReadiness(manifest.ID, manifest.Version, true), registry.Enable("coverage", manifest.ID, manifest.Version)} {
		if err != nil {
			t.Fatal(err)
		}
	}
	service := admin.NewLifecycleServiceWithProjects(operators, registry, project.NewStore(project.Project{ID: "coverage", Active: true}))
	handler := AdminHandler(service)
	send := func(operation string, fields map[string]any) *httptest.ResponseRecorder {
		t.Helper()
		command := map[string]any{"operation": operation, "project_id": "coverage", "plugin_id": "sample", "version": "1.0.0", "confirmed": true}
		for key, value := range fields {
			command[key] = value
		}
		body, err := json.Marshal(command)
		if err != nil {
			t.Fatal(err)
		}
		req := httptest.NewRequest(http.MethodPost, "/admin/command", strings.NewReader(string(body)))
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, req.WithContext(operator.WithActor(req.Context(), actor)))
		return recorder
	}
	preview := send("plugin.disable.preview", nil)
	if preview.Code != http.StatusOK {
		t.Fatalf("disable preview: %d %s", preview.Code, preview.Body.String())
	}
	var decoded plugin.DisablePreview
	if err := json.Unmarshal(preview.Body.Bytes(), &decoded); err != nil || decoded.PluginID != manifest.ID {
		t.Fatalf("preview response: %+v %v", decoded, err)
	}
	if response := send("plugin.remove", nil); response.Code != http.StatusNoContent {
		t.Fatalf("remove: %d %s", response.Code, response.Body.String())
	}
}
