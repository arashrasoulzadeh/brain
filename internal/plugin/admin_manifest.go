package plugin

import (
	"encoding/json"
	"regexp"
	"strings"
)

// AdminSpec is declarative UI metadata, not an authorization grant.
type AdminSpec struct {
	Version int         `yaml:"version" json:"version"`
	Pages   []AdminPage `yaml:"pages" json:"pages"`
}

type AdminPage struct {
	ID          string        `yaml:"id" json:"id"`
	Title       string        `yaml:"title" json:"title"`
	Actions     []AdminAction `yaml:"actions" json:"actions"`
	Extends     string        `yaml:"extends,omitempty" json:"extends,omitempty"`
	RefreshOn   []string      `yaml:"refresh_on,omitempty" json:"refresh_on,omitempty"`
	LegacyRoute string        `yaml:"legacy_route,omitempty" json:"legacy_route,omitempty"`
}

type AdminAction struct {
	Title     string       `yaml:"title" json:"title"`
	Operation string       `yaml:"operation" json:"operation"`
	Fields    []AdminField `yaml:"fields,omitempty" json:"fields,omitempty"`
	ReadOnly  bool         `yaml:"read_only,omitempty" json:"read_only,omitempty"`
}

type AdminField struct {
	Name     string `yaml:"name" json:"name"`
	Type     string `yaml:"type" json:"type"`
	Required bool   `yaml:"required,omitempty" json:"required,omitempty"`
}

var adminTarget = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_.-]{0,63}/[a-z][a-z0-9-]{0,63}$`)

var adminPageID = regexp.MustCompile(`^[a-z][a-z0-9-]{0,63}$`)

func validateAdminSpec(m Manifest) error {
	if m.Admin == nil {
		return nil
	}
	if m.Admin.Version != 1 || len(m.Admin.Pages) == 0 || len(m.Admin.Pages) > 64 {
		return ErrInvalidManifest
	}
	pages := map[string]bool{}
	for _, page := range m.Admin.Pages {
		if !adminPageID.MatchString(page.ID) || strings.TrimSpace(page.Title) == "" || pages[page.ID] || len(page.Actions) == 0 || len(page.Actions) > 64 {
			return ErrInvalidManifest
		}
		if page.Extends != "" && (!adminTarget.MatchString(page.Extends) || strings.HasPrefix(page.Extends, m.ID+"/")) {
			return ErrInvalidManifest
		}
		if page.LegacyRoute != "" && (!adminPageID.MatchString(page.LegacyRoute) || reservedAdminAlias(page.LegacyRoute)) {
			return ErrInvalidManifest
		}
		if len(page.RefreshOn) > 16 {
			return ErrInvalidManifest
		}
		events := map[string]bool{}
		for _, name := range page.RefreshOn {
			if events[name] || !m.adminObservationAllowed(name) {
				return ErrInvalidManifest
			}
			events[name] = true
		}
		pages[page.ID] = true
		actions := map[string]bool{}
		for _, action := range page.Actions {
			if _, ok := m.Route(action.Operation); !ok || strings.TrimSpace(action.Title) == "" || actions[action.Operation] {
				return ErrInvalidManifest
			}
			if len(action.Fields) > 32 {
				return ErrInvalidManifest
			}
			fields := map[string]bool{}
			for _, field := range action.Fields {
				if !validResponseKey(field.Name) || field.Name == "projectId" || field.Name == "project_id" || fields[field.Name] || (field.Type != "string" && field.Type != "number" && field.Type != "boolean" && field.Type != "json") {
					return ErrInvalidManifest
				}
				fields[field.Name] = true
			}
			actions[action.Operation] = true
		}
	}
	return nil
}

func (m Manifest) AdminOperation(pageID, operation string) bool {
	if m.Admin == nil || m.Admin.Version != 1 {
		return false
	}
	for _, page := range m.Admin.Pages {
		if page.ID != pageID {
			continue
		}
		for _, action := range page.Actions {
			if action.Operation == operation {
				return true
			}
		}
	}
	return false
}

// ValidateAdminPayload enforces declared form fields at the host boundary.
func (m Manifest) ValidateAdminPayload(pageID, operation string, payload json.RawMessage) error {
	if !m.AdminOperation(pageID, operation) {
		return ErrGrantDenied
	}
	for _, page := range m.Admin.Pages {
		if page.ID != pageID {
			continue
		}
		for _, action := range page.Actions {
			if action.Operation != operation {
				continue
			}
			if len(action.Fields) == 0 && !action.ReadOnly {
				if !json.Valid(payload) {
					return ErrInvalidWorkerRequest
				}
				return nil
			}
			var values map[string]any
			if json.Unmarshal(payload, &values) != nil || values == nil {
				return ErrInvalidWorkerRequest
			}
			for _, field := range action.Fields {
				value, exists := values[field.Name]
				if !exists {
					if field.Required {
						return ErrInvalidWorkerRequest
					}
					continue
				}
				valid := false
				switch field.Type {
				case "string":
					text, ok := value.(string)
					valid = ok && (!field.Required || strings.TrimSpace(text) != "")
				case "number":
					_, valid = value.(float64)
				case "boolean":
					_, valid = value.(bool)
				case "json":
					valid = !field.Required || value != nil
				}
				if !valid {
					return ErrInvalidWorkerRequest
				}
				delete(values, field.Name)
			}
			if len(values) != 0 {
				return ErrInvalidWorkerRequest
			}
			return nil
		}
	}
	return ErrGrantDenied
}

func (m Manifest) adminObservationAllowed(name string) bool {
	for _, event := range m.Events {
		if m.ID+"."+event.Name == name {
			return true
		}
	}
	for _, observer := range m.Observers {
		if observer.Plugin+"."+observer.Event == name {
			return true
		}
	}
	return false
}

func (m Manifest) AdminPageSpec(pageID string) (AdminPage, bool) {
	if m.Admin != nil && m.Admin.Version == 1 {
		for _, page := range m.Admin.Pages {
			if page.ID == pageID {
				return page, true
			}
		}
	}
	return AdminPage{}, false
}

func (m Manifest) AdminReadOnly(pageID, operation string) bool {
	page, ok := m.AdminPageSpec(pageID)
	if ok {
		for _, action := range page.Actions {
			if action.Operation == operation {
				return action.ReadOnly
			}
		}
	}
	return false
}

func reservedAdminAlias(route string) bool {
	switch route {
	case "projects", "plugins", "system", "audit", "routes", "monitoring", "workers", "functions", "extensions":
		return true
	}
	return false
}
