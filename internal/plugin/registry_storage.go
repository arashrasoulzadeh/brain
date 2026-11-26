package plugin

import "hash/fnv"

// RegistryStateStore is the persistence boundary for registry lifecycle metadata.
type RegistryStateStore interface {
	LoadRegistryState() (map[string]Installation, map[string]bool, error)
	SaveRegistryState(map[string]Installation, map[string]bool) error
}

// InvocationStoragePlanner prepares a scoped request and a deferred, atomic
// storage effect. Implementations must not write while preparing a request.
type InvocationStoragePlanner interface {
	Supports(pluginID string) bool
	Prepare(request WorkerRequest) (WorkerRequest, func() error, error)
}

// BindRegistryStorage restores persisted lifecycle state during host setup.
func (r *Registry) BindRegistryStorage(storage RegistryStateStore) error {
	r.readinessMu.Lock()
	defer r.readinessMu.Unlock()
	r.installationMu.Lock()
	defer r.installationMu.Unlock()
	if storage == nil {
		r.storage = nil
		return nil
	}
	installations, readiness, err := storage.LoadRegistryState()
	if err != nil {
		return err
	}
	r.storage = storage
	for key, installation := range installations {
		if _, registered := r.registered[installation.PluginID+"@"+installation.Version]; registered {
			r.installations[key] = installation
		}
	}
	for key, ready := range readiness {
		if _, registered := r.registered[key]; registered {
			r.readiness[key] = ready
		}
	}
	return nil
}

// BindInvocationStorage installs a trusted host-side storage adapter.
func (r *Registry) BindInvocationStorage(planner InvocationStoragePlanner) {
	r.storagePlannerMu.Lock()
	defer r.storagePlannerMu.Unlock()
	r.storagePlanner = planner
}

// BindSQLiteStorage preserves the existing setup API and compatibility routes.
func (r *Registry) BindSQLiteStorage(storage *SQLiteStorage) error {
	if storage == nil {
		if err := r.BindRegistryStorage(nil); err != nil {
			return err
		}
		r.BindInvocationStorage(nil)
		return nil
	}
	if err := r.BindRegistryStorage(storage); err != nil {
		return err
	}
	r.BindInvocationStorage(legacyStoragePlanner{storage: storage})
	return nil
}

func (r *Registry) persistLifecycleState() {
	if r.storage == nil {
		return
	}
	_ = r.storage.SaveRegistryState(r.installations, r.readiness)
}

func (r *Registry) SaveDurableData(projectID, pluginID, key string, value []byte) error {
	if !isValidProjectID(projectID) || pluginID == "" || key == "" {
		return ErrProjectInvalid
	}
	r.durableMu.Lock()
	defer r.durableMu.Unlock()
	r.durableData[projectID+":"+pluginID+":"+key] = append([]byte(nil), value...)
	return nil
}

func (r *Registry) DurableData(projectID, pluginID, key string) ([]byte, bool) {
	r.durableMu.RLock()
	defer r.durableMu.RUnlock()
	value, ok := r.durableData[projectID+":"+pluginID+":"+key]
	return append([]byte(nil), value...), ok
}

func (r *Registry) DeleteDurableData(projectID, pluginID, key string) error {
	if !isValidProjectID(projectID) || pluginID == "" || key == "" {
		return ErrProjectInvalid
	}
	r.durableMu.Lock()
	defer r.durableMu.Unlock()
	delete(r.durableData, projectID+":"+pluginID+":"+key)
	r.appendAudit(AuditEvent{Action: "delete_durable_data", ProjectID: projectID, PluginID: pluginID, Detail: key})
	return nil
}

// lockStorageInvocation bounds lock memory while serializing storage plans by
// project and plugin. Different versions share a stripe because they share data.
func (r *Registry) lockStorageInvocation(request WorkerRequest) (InvocationStoragePlanner, func()) {
	r.storagePlannerMu.RLock()
	planner := r.storagePlanner
	r.storagePlannerMu.RUnlock()
	if planner == nil || !planner.Supports(request.PluginID) {
		return nil, func() {}
	}
	hash := fnv.New32a()
	_, _ = hash.Write([]byte(request.ProjectID + ":" + request.PluginID))
	lock := &r.storageInvocations[hash.Sum32()%uint32(len(r.storageInvocations))]
	lock.Lock()
	return planner, lock.Unlock
}
