package app

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"brain/internal/admin"
	"brain/internal/operator"
	"brain/internal/plugin"
)

type operatorLogin struct {
	Username string `json:"username"`
	Secret   string `json:"secret"`
}

func OperatorSessionHandler(sessions *operator.Store, digest func(string) string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if sessions == nil || digest == nil || r.Method != http.MethodPost {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		var login operatorLogin
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 16<<10)).Decode(&login); err != nil {
			http.Error(w, "invalid operator login", http.StatusBadRequest)
			return
		}
		if err := sessions.Authenticate(login.Username, login.Secret, digest); err != nil {
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}
		tokenBytes := make([]byte, 32)
		if _, err := rand.Read(tokenBytes); err != nil {
			http.Error(w, "operator session unavailable", http.StatusInternalServerError)
			return
		}
		token, err := sessions.IssueSession(func() string { return hex.EncodeToString(tokenBytes) })
		if err != nil {
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"session": token})
	})
}

func AuthenticatedAdminHandler(service *admin.LifecycleService, sessions *operator.Store) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if sessions == nil {
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}
		const prefix = "Bearer "
		header := r.Header.Get("Authorization")
		if !strings.HasPrefix(header, prefix) {
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}
		actor, ok := sessions.ValidateSession(strings.TrimPrefix(header, prefix))
		if !ok {
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}
		AdminHandler(service).ServeHTTP(w, r.WithContext(operator.WithActor(r.Context(), actor)))
	})
}

func OperatorLogoutHandler(sessions *operator.Store) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if sessions == nil || r.Method != http.MethodDelete {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		const prefix = "Bearer "
		header := r.Header.Get("Authorization")
		if !strings.HasPrefix(header, prefix) || sessions.RevokeSession(strings.TrimPrefix(header, prefix)) != nil {
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
}

func AuthenticatedProjectsHandler(service *admin.LifecycleService, sessions *operator.Store) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || service == nil || sessions == nil {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		const prefix = "Bearer "
		header := r.Header.Get("Authorization")
		if !strings.HasPrefix(header, prefix) {
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}
		actor, ok := sessions.ValidateSession(strings.TrimPrefix(header, prefix))
		if !ok {
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}
		projects, err := service.ListProjects(actor)
		if err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(projects)
	})
}

func AuthenticatedPluginsHandler(service *admin.LifecycleService, sessions *operator.Store) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || service == nil || sessions == nil {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		const prefix = "Bearer "
		header := r.Header.Get("Authorization")
		if !strings.HasPrefix(header, prefix) {
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}
		actor, ok := sessions.ValidateSession(strings.TrimPrefix(header, prefix))
		if !ok {
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}
		manifests, installations, err := service.ListPlugins(actor, r.URL.Query().Get("project_id"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"manifests": manifests, "installations": installations})
	})
}

func AuthenticatedAuditHandler(service *admin.LifecycleService, sessions *operator.Store) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || service == nil || sessions == nil {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		const prefix = "Bearer "
		header := r.Header.Get("Authorization")
		if !strings.HasPrefix(header, prefix) {
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}
		actor, ok := sessions.ValidateSession(strings.TrimPrefix(header, prefix))
		if !ok {
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}
		records, err := service.AuditRecords(actor)
		if err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(records)
	})
}

func AuthenticatedSystemHandler(service *admin.LifecycleService, sessions *operator.Store) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || service == nil || sessions == nil {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		const prefix = "Bearer "
		header := r.Header.Get("Authorization")
		if !strings.HasPrefix(header, prefix) {
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}
		actor, ok := sessions.ValidateSession(strings.TrimPrefix(header, prefix))
		if !ok {
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}
		manifests, err := service.SystemCapabilities(actor)
		if err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(manifests)
	})
}

type adminCommand struct {
	Operation       string          `json:"operation"`
	ProjectID       string          `json:"project_id"`
	PluginID        string          `json:"plugin_id"`
	Version         string          `json:"version"`
	Ready           bool            `json:"ready"`
	Capability      string          `json:"capability"`
	Resource        string          `json:"resource"`
	Key             string          `json:"key"`
	Values          map[string]any  `json:"values"`
	Confirmed       bool            `json:"confirmed"`
	Revision        uint64          `json:"revision"`
	Dependents      []string        `json:"dependents"`
	Name            string          `json:"name"`
	Source          string          `json:"source"`
	Input           json.RawMessage `json:"input"`
	Payload         json.RawMessage `json:"payload"`
	TargetOperation string          `json:"target_operation"`
	FunctionVersion int             `json:"function_version"`
	Public          bool            `json:"public"`
	Permissions     []string        `json:"permissions"`
	Method          string          `json:"method"`
}

func publicFunctionURL(r *http.Request, projectID, functionName string) string {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	return scheme + "://" + r.Host + "/public/" + url.PathEscape(projectID) + "/" + url.PathEscape(functionName)
}

func AuthenticatedFunctionsHandler(service *admin.LifecycleService, sessions *operator.Store) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if service == nil || sessions == nil || !strings.HasPrefix(r.Header.Get("Authorization"), "Bearer ") {
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}
		actor, ok := sessions.ValidateSession(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "))
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
		actor, ok := sessions.ValidateSession(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "))
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
		result, err := service.PublicInvokeFunction(r.PathValue("project_id"), r.PathValue("function_name"), r.Method, input)
		if err != nil {
			http.Error(w, "public function unavailable", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(result)
	})
}

func AdminHandler(service *admin.LifecycleService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if service == nil || r.Method != http.MethodPost {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		actor, ok := operator.ActorFromContext(r.Context())
		if !ok {
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}
		var command adminCommand
		decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 64<<10))
		if err := decoder.Decode(&command); err != nil {
			http.Error(w, "invalid admin command", http.StatusBadRequest)
			return
		}
		if requiresAdminConfirmation(command.Operation) && !command.Confirmed {
			http.Error(w, "explicit confirmation required", http.StatusBadRequest)
			return
		}
		var err error
		switch command.Operation {
		case "project.create":
			err = service.CreateProject(actor, command.ProjectID)
		case "project.activate":
			err = service.SetProjectActive(actor, command.ProjectID, true)
		case "project.deactivate":
			err = service.SetProjectActive(actor, command.ProjectID, false)
		case "plugin.install":
			err = service.InstallPlugin(actor, command.ProjectID, command.PluginID, command.Version)
		case "plugin.readiness":
			err = service.SetPluginReadiness(actor, command.PluginID, command.Version, command.Ready)
		case "plugin.enable":
			err = service.EnablePlugin(actor, command.ProjectID, command.PluginID, command.Version)
		case "plugin.worker.start":
			err = service.StartPluginWorker(actor, command.ProjectID, command.PluginID, command.Version)
		case "plugin.remove":
			err = service.RemovePlugin(actor, command.ProjectID, command.PluginID, command.Version)
		case "plugin.grant":
			err = service.GrantCapability(actor, command.ProjectID, command.PluginID, command.Version, command.Capability)
		case "plugin.grant-resource":
			err = service.GrantResource(actor, command.ProjectID, command.PluginID, command.Version, command.Capability, command.Resource)
		case "plugin.configure":
			err = service.ConfigurePlugin(actor, command.ProjectID, command.PluginID, command.Version, command.Values)
		case "plugin.invoke":
			response, invokeErr := service.InvokePlugin(actor, admin.PluginInvocation{ProjectID: command.ProjectID, PluginID: command.PluginID, Version: command.Version, Operation: command.TargetOperation, Payload: command.Payload})
			if invokeErr != nil {
				http.Error(w, invokeErr.Error(), http.StatusForbidden)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(response)
			return
		case "plugin.revoke":
			err = service.RevokeCapability(actor, command.ProjectID, command.PluginID, command.Version, command.Capability)
		case "plugin.promote":
			err = service.PromotePlugin(actor, command.ProjectID, command.PluginID, command.Version)
		case "plugin.drain":
			err = service.DrainPlugin(actor, command.ProjectID, command.PluginID, command.Version)
		case "plugin.stop-immediate":
			err = service.StopPluginImmediately(actor, command.ProjectID, command.PluginID, command.Version)
		case "plugin.stop":
			err = service.StopPlugin(actor, command.ProjectID, command.PluginID, command.Version)
		case "plugin.resume":
			err = service.ResumePlugin(actor, command.ProjectID, command.PluginID, command.Version)
		case "plugin.disable.preview":
			preview, previewErr := service.PreviewDisable(actor, command.ProjectID, command.PluginID, command.Version)
			if previewErr != nil {
				http.Error(w, previewErr.Error(), http.StatusForbidden)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(preview)
			return
		case "plugin.disable.apply":
			err = service.ApplyDisable(actor, plugin.DisablePreview{ProjectID: command.ProjectID, PluginID: command.PluginID, Version: command.Version, Revision: command.Revision, Dependents: command.Dependents})
		case "plugin.recover":
			err = service.RecoverPlugin(actor, command.ProjectID, command.PluginID, command.Version)
		case "plugin.rollback":
			err = service.RollbackPlugin(actor, command.ProjectID, command.PluginID, command.Version)
		case "project.uninstall":
			err = service.UninstallProject(actor, command.ProjectID)
		case "durable-data.delete":
			err = service.DeleteDurableData(actor, command.ProjectID, command.PluginID, command.Key)
		default:
			http.Error(w, "unsupported admin operation", http.StatusBadRequest)
			return
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
}

func requiresAdminConfirmation(operation string) bool {
	switch operation {
	case "plugin.remove", "plugin.revoke", "plugin.stop-immediate", "plugin.stop", "plugin.rollback", "plugin.disable.apply", "project.deactivate", "project.uninstall", "durable-data.delete":
		return true
	default:
		return false
	}
}
