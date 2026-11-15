package app

import (
	"encoding/json"
	"net/http"

	"brain/internal/admin"
	"brain/internal/operator"
)

type adminCommand struct {
	Operation  string         `json:"operation"`
	ProjectID  string         `json:"project_id"`
	PluginID   string         `json:"plugin_id"`
	Version    string         `json:"version"`
	Capability string         `json:"capability"`
	Resource   string         `json:"resource"`
	Values     map[string]any `json:"values"`
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
		var err error
		switch command.Operation {
		case "plugin.enable":
			err = service.EnablePlugin(actor, command.ProjectID, command.PluginID, command.Version)
		case "plugin.remove":
			err = service.RemovePlugin(actor, command.ProjectID, command.PluginID, command.Version)
		case "plugin.grant":
			err = service.GrantCapability(actor, command.ProjectID, command.PluginID, command.Version, command.Capability)
		case "plugin.grant-resource":
			err = service.GrantResource(actor, command.ProjectID, command.PluginID, command.Version, command.Capability, command.Resource)
		case "plugin.configure":
			err = service.ConfigurePlugin(actor, command.ProjectID, command.PluginID, command.Version, command.Values)
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
