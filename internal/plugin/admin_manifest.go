package plugin

import (
	"regexp"
	"strings"
)

// AdminSpec is declarative UI metadata, not an authorization grant.
type AdminSpec struct {
	Version int         `yaml:"version" json:"version"`
	Pages   []AdminPage `yaml:"pages" json:"pages"`
}

type AdminPage struct {
	ID      string        `yaml:"id" json:"id"`
	Title   string        `yaml:"title" json:"title"`
	Actions []AdminAction `yaml:"actions" json:"actions"`
}

type AdminAction struct {
	Title     string `yaml:"title" json:"title"`
	Operation string `yaml:"operation" json:"operation"`
}

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
		pages[page.ID] = true
		actions := map[string]bool{}
		for _, action := range page.Actions {
			if _, ok := m.Route(action.Operation); !ok || strings.TrimSpace(action.Title) == "" || actions[action.Operation] {
				return ErrInvalidManifest
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
