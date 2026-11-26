package plugin

import "errors"

var ErrAdminInventoryUnavailable = errors.New("admin inventory unavailable")

// AdminInventorySpec declares a bounded read-only resource list. The trusted
// storage provider, not the plugin manifest, owns namespace and SQL selection.
type AdminInventorySpec struct {
	Kind        string `yaml:"kind" json:"kind"`
	Capability  string `yaml:"capability" json:"capability"`
	Title       string `yaml:"title" json:"title"`
	SelectField string `yaml:"select_field,omitempty" json:"select_field,omitempty"`
}

func (s AdminInventorySpec) valid(m Manifest, page AdminPage) bool {
	if (s.Kind != "state_keys" && s.Kind != "file_keys" && s.Kind != "table_names") || s.Title == "" || len(s.Title) > 120 {
		return false
	}
	declared := false
	for _, capability := range m.Capabilities {
		if capability == s.Capability {
			declared = true
		}
	}
	if !declared {
		return false
	}
	if s.SelectField == "" {
		return true
	}
	for _, action := range page.Actions {
		for _, field := range action.Fields {
			if field.Name == s.SelectField && field.Type == "string" {
				return true
			}
		}
	}
	return false
}

type AdminInventoryRow struct {
	Key       string `json:"key"`
	SizeBytes *int64 `json:"size_bytes,omitempty"`
	Published *bool  `json:"published,omitempty"`
}

type AdminInventoryPage struct {
	Rows []AdminInventoryRow `json:"rows"`
	Next string              `json:"next,omitempty"`
}

// AdminInventoryProvider is implemented by trusted host storage adapters.
// It is not part of the untrusted worker protocol.
type AdminInventoryProvider interface {
	AdminInventory(projectID, pluginID, kind, capability, after string, limit int) (AdminInventoryPage, error)
}

func (r *Registry) AdminInventory(projectID, pluginID, version, pageID, after string, limit int) (AdminInventoryPage, error) {
	if !isValidProjectID(projectID) || len(after) > 65536 || limit < 1 || limit > 100 {
		return AdminInventoryPage{}, ErrInvalidWorkerRequest
	}
	r.storagePlannerMu.RLock()
	provider, ok := r.storagePlanner.(AdminInventoryProvider)
	r.storagePlannerMu.RUnlock()
	if !ok {
		return AdminInventoryPage{}, ErrAdminInventoryUnavailable
	}
	r.readinessMu.RLock()
	defer r.readinessMu.RUnlock()
	r.installationMu.RLock()
	defer r.installationMu.RUnlock()
	if err := r.LifecycleStorageHealth(); err != nil {
		return AdminInventoryPage{}, err
	}
	key := pluginID + "@" + version
	registered, exists := r.registered[key]
	if !exists {
		return AdminInventoryPage{}, ErrNotRegistered
	}
	installation, exists := r.installations[projectID+":"+key]
	if !exists || !installation.Enabled || installation.Draining || !r.readiness[key] {
		return AdminInventoryPage{}, ErrDisabled
	}
	page, exists := registered.Manifest.AdminPageSpec(pageID)
	if !exists || page.Inventory == nil {
		return AdminInventoryPage{}, ErrGrantDenied
	}
	for _, capability := range registered.Manifest.Capabilities {
		if !installation.Grants[capability] {
			return AdminInventoryPage{}, ErrGrantDenied
		}
	}
	// A broad enumeration is not authorized by a resource-specific grant.
	if !installation.Grants[page.Inventory.Capability] || len(installation.Resources[page.Inventory.Capability]) > 0 {
		return AdminInventoryPage{}, ErrGrantDenied
	}
	return provider.AdminInventory(projectID, pluginID, page.Inventory.Kind, page.Inventory.Capability, after, limit)
}
