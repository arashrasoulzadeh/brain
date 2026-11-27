package plugin

import (
	"errors"
	"sort"
	"strings"
)

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
	Key         string `json:"key"`
	SizeBytes   *int64 `json:"size_bytes,omitempty"`
	Published   *bool  `json:"published,omitempty"`
	Filename    string `json:"filename,omitempty"`
	Extension   string `json:"extension,omitempty"`
	ContentType string `json:"content_type,omitempty"`
	StoredBytes *int64 `json:"stored_bytes,omitempty"`
	UpdatedAt   string `json:"updated_at,omitempty"`
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

// AdminInventoryFilter is a trusted snapshot of the admitted resource scope.
// Providers must filter before returning rows or pagination cursors.
type AdminInventoryFilter struct {
	AllowAll   bool
	DenyAll    bool
	Allowed    []string
	Denied     []string
	Restricted bool
	GrantKeys  []string
}

type AdminScopedInventoryProvider interface {
	AdminInventoryScoped(projectID, pluginID, kind, capability, after string, limit int, filter AdminInventoryFilter) (AdminInventoryPage, error)
}

// AdminSearchInventoryProvider must apply search after authorization and before pagination.
type AdminSearchInventoryProvider interface {
	AdminInventorySearch(projectID, pluginID, kind, capability, after, search string, limit int, filter *AdminInventoryFilter) (AdminInventoryPage, error)
}

// AdminInventory is the compatibility entry point for unmanaged projects.
func (r *Registry) AdminInventory(projectID, pluginID, version, pageID, after string, limit int) (AdminInventoryPage, error) {
	return r.adminInventory("", projectID, pluginID, version, pageID, after, "", limit)
}

// AdminInventoryAsOperator receives the authenticated identity from the host.
func (r *Registry) AdminInventoryAsOperator(actorID, projectID, pluginID, version, pageID, after string, limit int) (AdminInventoryPage, error) {
	if !validAccessIdentity(actorID) {
		return AdminInventoryPage{}, ErrActorDenied
	}
	return r.adminInventory("operator:"+actorID, projectID, pluginID, version, pageID, after, "", limit)
}

// AdminInventorySearchAsOperator preserves the same policy and grant checks as enumeration.
func (r *Registry) AdminInventorySearchAsOperator(actorID, projectID, pluginID, version, pageID, after, search string, limit int) (AdminInventoryPage, error) {
	if !validAccessIdentity(actorID) {
		return AdminInventoryPage{}, ErrActorDenied
	}
	return r.adminInventory("operator:"+actorID, projectID, pluginID, version, pageID, after, search, limit)
}

func (r *Registry) adminInventory(principal, projectID, pluginID, version, pageID, after, search string, limit int) (AdminInventoryPage, error) {
	search = strings.TrimSpace(search)
	if !isValidProjectID(projectID) || len(after) > 65536 || len(search) > 256 || strings.ContainsRune(search, 0) || limit < 1 || limit > 100 {
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
	if r.projectResolver != nil {
		model, err := r.projectResolver(projectID)
		if err != nil || model.ID != projectID || !model.Active {
			return AdminInventoryPage{}, ErrProjectInvalid
		}
	}
	managed, available := r.accessProviderLocked(projectID)
	if managed && (!available || principal == "") {
		return AdminInventoryPage{}, ErrActorDenied
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
	if !installation.Grants[page.Inventory.Capability] {
		return AdminInventoryPage{}, ErrGrantDenied
	}
	resources := installation.Resources[page.Inventory.Capability]
	if !managed {
		// Preserve the existing fail-closed behavior for unscoped providers.
		if len(resources) > 0 {
			return AdminInventoryPage{}, ErrGrantDenied
		}
		if search != "" {
			searcher, ok := provider.(AdminSearchInventoryProvider)
			if !ok {
				return AdminInventoryPage{}, ErrAdminInventoryUnavailable
			}
			return searcher.AdminInventorySearch(projectID, pluginID, page.Inventory.Kind, page.Inventory.Capability, after, search, limit, nil)
		}
		return provider.AdminInventory(projectID, pluginID, page.Inventory.Kind, page.Inventory.Capability, after, limit)
	}
	scoped, ok := provider.(AdminScopedInventoryProvider)
	storage, stored := r.storage.(*SQLiteStorage)
	route, declared := inventoryReadRoute(registered.Manifest, page)
	if !ok || !stored || storage == nil || !declared || pluginID == accessPluginID {
		return AdminInventoryPage{}, ErrActorDenied
	}
	// The installation read lock also excludes access-rule mutations. Resolve
	// the same exact principal/version/operation/capability rules as invocation.
	filter, err := storage.accessInventoryFilter(projectID, principal, pluginID, version, route)
	if err != nil {
		return AdminInventoryPage{}, ErrActorDenied
	}
	filter.Restricted = len(resources) > 0
	for key, granted := range resources {
		if granted {
			filter.GrantKeys = append(filter.GrantKeys, key)
		}
	}
	sort.Strings(filter.GrantKeys)
	if search != "" {
		searcher, ok := provider.(AdminSearchInventoryProvider)
		if !ok {
			return AdminInventoryPage{}, ErrAdminInventoryUnavailable
		}
		return searcher.AdminInventorySearch(projectID, pluginID, page.Inventory.Kind, page.Inventory.Capability, after, search, limit, &filter)
	}
	return scoped.AdminInventoryScoped(projectID, pluginID, page.Inventory.Kind, page.Inventory.Capability, after, limit, filter)
}

// Inventory access follows exactly one declared, read-only resource action.
func inventoryReadRoute(m Manifest, page AdminPage) (Route, bool) {
	var selected Route
	if page.Inventory == nil || page.Inventory.SelectField == "" {
		return selected, false
	}
	for _, action := range page.Actions {
		route, ok := m.Route(action.Operation)
		if !ok || !action.ReadOnly || route.Capability != page.Inventory.Capability || route.ResourceField != page.Inventory.SelectField || route.Model != nil || route.Call != nil || route.Helper == "host.admin" {
			continue
		}
		if selected.Operation != "" {
			return Route{}, false
		}
		selected = route
	}
	return selected, selected.Operation != ""
}

// Access-rule changes are serialized by the caller's installation lock.
func (s *SQLiteStorage) accessInventoryFilter(projectID, principal, id, version string, route Route) (AdminInventoryFilter, error) {
	filter := AdminInventoryFilter{Allowed: []string{}, Denied: []string{}, GrantKeys: []string{}}
	rows, err := s.db.Query(accessRuleScopeQuery+" ORDER BY record_id LIMIT 4097", projectID, accessPluginID, accessRuleModel, principal, id, version, route.Operation, route.Capability)
	if err != nil {
		return filter, err
	}
	defer rows.Close()
	count := 0
	for rows.Next() {
		count++
		if count > 4096 {
			return AdminInventoryFilter{}, ErrActorDenied
		}
		record, err := scanModel(rows, accessRuleSpec)
		if err != nil {
			return AdminInventoryFilter{}, err
		}
		resource := record.Fields["resource"].(string)
		if !validRuleResource(route, resource, true) {
			return AdminInventoryFilter{}, ErrActorDenied
		}
		allowed := record.Fields["allowed"].(bool)
		if resource == "*" {
			if allowed {
				filter.AllowAll = true
			} else {
				filter.DenyAll = true
			}
		} else if allowed {
			filter.Allowed = append(filter.Allowed, resource)
		} else {
			filter.Denied = append(filter.Denied, resource)
		}
	}
	return filter, rows.Err()
}
