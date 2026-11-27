package app

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"brain/internal/admin"
	"brain/internal/operator"
	"brain/internal/plugin"
	"brain/internal/project"
)

func projectIdentity(r *http.Request) (project.Project, bool) {
	p, ok := project.FromContext(r.Context())
	return p, ok && p.ID != "" && p.Active
}

// parseRequestJSONObject rejects duplicate fields and trailing JSON documents.
// RawMessage also preserves exact JSON numeric values when binding the
// authoritative project ID into an operator command or plugin invocation.
func parseRequestJSONObject(data []byte) (map[string]json.RawMessage, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	start, err := decoder.Token()
	if err != nil || start != json.Delim('{') {
		return nil, errors.New("JSON object required")
	}
	fields := make(map[string]json.RawMessage)
	for decoder.More() {
		token, err := decoder.Token()
		if err != nil {
			return nil, err
		}
		key, ok := token.(string)
		if !ok {
			return nil, errors.New("invalid object property")
		}
		if _, repeated := fields[key]; repeated {
			return nil, errors.New("duplicate object property")
		}
		var raw json.RawMessage
		if err := decoder.Decode(&raw); err != nil {
			return nil, err
		}
		fields[key] = raw
	}
	end, err := decoder.Token()
	if err != nil || end != json.Delim('}') {
		return nil, errors.New("malformed JSON object")
	}
	var extra json.RawMessage
	if decoder.Decode(&extra) != io.EOF {
		return nil, errors.New("trailing JSON")
	}
	return fields, nil
}

func checkedRawIdentity(p project.Project, fields map[string]json.RawMessage) bool {
	for _, key := range []string{"project_id", "projectId", "project"} {
		raw, exists := fields[key]
		if !exists {
			continue
		}
		var value string
		if json.Unmarshal(raw, &value) != nil || value != p.ID && value != p.Slug {
			return false
		}
	}
	return true
}

// AuthenticatedProjectCommandHandler authenticates before parsing input, so
// unauthenticated callers never inspect command validation or project data.
func AuthenticatedProjectCommandHandler(next http.Handler, sessions *operator.Store) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, authorized := operatorActorFromRequest(sessions, r); !authorized {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		ProjectCommandHandler(next).ServeHTTP(w, r)
	})
}

// ProjectCommandHandler binds the command's project ID to host-resolved context
// before an existing operator or functions handler sees request JSON.
func ProjectCommandHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p, ok := projectIdentity(r)
		if !ok {
			http.NotFound(w, r)
			return
		}
		data, err := io.ReadAll(io.LimitReader(r.Body, 1<<20+1))
		if err != nil || len(data) > 1<<20 {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		command, err := parseRequestJSONObject(data)
		if err != nil || !checkedRawIdentity(p, command) {
			http.Error(w, "invalid project command", http.StatusBadRequest)
			return
		}
		boundID, _ := json.Marshal(p.ID)
		command["project_id"] = boundID
		bound, err := json.Marshal(command)
		if err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}
		r.Body = io.NopCloser(bytes.NewReader(bound))
		r.ContentLength = int64(len(bound))
		next.ServeHTTP(w, r)
	})
}

// ProjectQueryHandler binds the project ID for existing internal service
// adapters without trusting any caller-provided project ID.
func ProjectQueryHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p, ok := projectIdentity(r)
		if !ok {
			http.NotFound(w, r)
			return
		}
		clone := r.Clone(r.Context())
		query := clone.URL.Query()
		query.Set("project_id", p.ID)
		clone.URL.RawQuery = query.Encode()
		next.ServeHTTP(w, clone)
	})
}

func AuthenticatedProjectRoutesHandler(service *admin.LifecycleService, sessions *operator.Store) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p, ok := projectIdentity(r)
		if !ok {
			http.NotFound(w, r)
			return
		}
		if _, ok := operatorActorFromRequest(sessions, r); !ok {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		routes := service.ProjectHTTPRoutes(p.ID)
		namespace := p.Slug
		if namespace == "" {
			namespace = p.ID
		}
		for i := range routes {
			routes[i].Path = "/" + namespace + routes[i].Path
		}
		_ = json.NewEncoder(w).Encode(routes)
	})
}

// PluginHTTPHandler resolves declared HTTP routes against the current
// installation snapshot, then delegates authorization and commit guards to
// the existing isolated worker invocation pipeline.
func PluginHTTPHandler(service *admin.LifecycleService, sessions *operator.Store) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p, ok := projectIdentity(r)
		if !ok {
			http.NotFound(w, r)
			return
		}
		_, relative, ok := strings.Cut(strings.TrimPrefix(r.URL.Path, "/"), "/")
		if !ok {
			http.NotFound(w, r)
			return
		}
		route, found := service.MatchHTTPRoute(p.ID, r.Method, "/"+relative)
		if !found {
			http.NotFound(w, r)
			return
		}
		credential := ""
		switch route.Auth {
		case "operator":
			actor, authenticated := operatorActorFromRequest(sessions, r)
			if !authenticated {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}
			credential = actor.ID
		case "user":
			token, valid := operatorBearerToken(r)
			if !valid {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}
			credential = token
		case "public":
		default:
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		fields := make(map[string]json.RawMessage)
		for key, values := range r.URL.Query() {
			if len(values) != 1 {
				http.Error(w, "invalid query", http.StatusBadRequest)
				return
			}
			encoded, _ := json.Marshal(values[0])
			fields[key] = encoded
		}
		if r.Method != http.MethodGet {
			input, err := io.ReadAll(io.LimitReader(r.Body, 1<<20+1))
			if err != nil || len(input) > 1<<20 {
				http.Error(w, "invalid request body", http.StatusBadRequest)
				return
			}
			if len(bytes.TrimSpace(input)) > 0 {
				body, err := parseRequestJSONObject(input)
				if err != nil {
					http.Error(w, "invalid JSON object", http.StatusBadRequest)
					return
				}
				for key, value := range body {
					if _, duplicate := fields[key]; duplicate {
						http.Error(w, "conflicting request fields", http.StatusBadRequest)
						return
					}
					fields[key] = value
				}
			}
		}
		for key, value := range route.Params {
			if supplied, exists := fields[key]; exists {
				var typed string
				if json.Unmarshal(supplied, &typed) != nil || typed != value {
					http.Error(w, "conflicting route parameter", http.StatusBadRequest)
					return
				}
			}
			encoded, _ := json.Marshal(value)
			fields[key] = encoded
		}
		if !checkedRawIdentity(p, fields) {
			http.Error(w, "conflicting project identity", http.StatusBadRequest)
			return
		}
		// Client-supplied identity is never forwarded to an isolated worker.
		delete(fields, "project_id")
		delete(fields, "projectId")
		delete(fields, "project")
		payload, err := json.Marshal(fields)
		if err != nil {
			http.Error(w, "invalid payload", http.StatusBadRequest)
			return
		}
		response, err := service.InvokeHTTPRoute(p.ID, route, payload, credential)
		if err != nil {
			status := http.StatusForbidden
			if errors.Is(err, plugin.ErrNotInstalled) || errors.Is(err, plugin.ErrDisabled) || errors.Is(err, plugin.ErrNotReady) || errors.Is(err, plugin.ErrNotRegistered) || errors.Is(err, plugin.ErrProjectInvalid) {
				status = http.StatusNotFound
			}
			http.Error(w, http.StatusText(status), status)
			return
		}
		if !response.Success || response.ErrorCode != "" {
			http.Error(w, "plugin invocation rejected", http.StatusBadRequest)
			return
		}
		if len(response.Payload) > 0 && !json.Valid(response.Payload) {
			http.Error(w, "invalid plugin response", http.StatusBadGateway)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if len(response.Payload) == 0 {
			_, _ = w.Write([]byte("{}"))
			return
		}
		_, _ = w.Write(response.Payload)
	})
}
