package app

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"

	"brain/internal/admin"
	"brain/internal/operator"
)

func publicFunctionURL(r *http.Request, projectID, functionName string) string {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	return scheme + "://" + r.Host + "/public/" + url.PathEscape(projectID) + "/" + url.PathEscape(functionName)
}

func AuthenticatedFunctionsHandler(service *admin.LifecycleService, sessions *operator.Store) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if service == nil || sessions == nil {
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}
		actor, ok := operatorActorFromRequest(sessions, r)
		if !ok {
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}
		projectID := r.URL.Query().Get("project_id")
		if r.Method == http.MethodGet {
			if err := service.RequireEnabledPlugin(actor, projectID, "functions"); err != nil {
				http.Error(w, err.Error(), http.StatusForbidden)
				return
			}
			drafts, err := service.ListFunctionDrafts(actor, projectID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(drafts)
			return
		}
		if r.Method != http.MethodPost {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		var command adminCommand
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&command); err != nil {
			http.Error(w, "invalid function request", http.StatusBadRequest)
			return
		}
		if err := service.RequireEnabledPlugin(actor, command.ProjectID, "functions"); err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		if command.Operation == "functions.test" {
			result, err := service.TestFunctionDraft(actor, command.ProjectID, command.Name, command.Source)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(result)
			return
		}
		if command.Operation == "functions.run" {
			result, err := service.RunFunctionDraft(actor, command.ProjectID, command.Name, command.Source, command.Input)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(result)
			return
		}
		if command.Operation == "functions.permissions" {
			if err := service.UpdateFunctionPermissions(actor, command.ProjectID, command.Name, command.Permissions, command.Public); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if command.Operation == "functions.update" {
			draft, err := service.UpdateFunctionDraft(actor, command.ProjectID, command.Name, command.Source, command.FunctionVersion)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(draft)
			return
		}
		if command.Operation == "functions.publish" {
			draft, err := service.PublishFunction(actor, command.ProjectID, command.Name, command.Method)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{"published": true, "project_id": command.ProjectID, "name": command.Name, "version": draft.Version, "method": draft.Method, "published_at": draft.PublishedAt, "route": publicFunctionURL(r, command.ProjectID, command.Name)})
			return
		}
		if command.Operation != "functions.create" {
			http.Error(w, "unsupported function operation", http.StatusBadRequest)
			return
		}
		draft, err := service.CreateFunctionDraftWithSettings(actor, command.ProjectID, command.Name, command.Source, command.FunctionVersion, command.Public, command.Permissions, command.Method)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(draft)
	})
}

func AuthenticatedFunctionLogsHandler(service *admin.LifecycleService, sessions *operator.Store) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		actor, ok := operatorActorFromRequest(sessions, r)
		if !ok {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		if err := service.RequireEnabledPlugin(actor, r.URL.Query().Get("project_id"), "functions"); err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		logs, err := service.ListFunctionLogs(actor, r.URL.Query().Get("project_id"), r.URL.Query().Get("name"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(logs)
	})
}

func PublicFunctionHandler(service *admin.LifecycleService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var input json.RawMessage
		if r.Method != http.MethodGet {
			if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&input); err != nil {
				http.Error(w, "invalid JSON input", http.StatusBadRequest)
				return
			}
		}
		var result map[string]any
		var err error
		if r.Header.Get("Authorization") != "" {
			token, ok := operatorBearerToken(r)
			if !ok {
				http.Error(w, "invalid authorization", http.StatusForbidden)
				return
			}
			result, err = service.PublicInvokeFunctionWithSession(r.PathValue("project_id"), r.PathValue("function_name"), r.Method, input, token)
		} else {
			result, err = service.PublicInvokeFunction(r.PathValue("project_id"), r.PathValue("function_name"), r.Method, input)
		}
		if err != nil {
			if errors.Is(err, admin.ErrFunctionMethodNotAllowed) {
				w.Header().Set("Allow", "GET, POST, PUT, PATCH")
				http.Error(w, err.Error(), http.StatusMethodNotAllowed)
				return
			}
			http.Error(w, "public function unavailable", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(result)
	})
}
