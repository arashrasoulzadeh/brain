package plugin

import (
	"errors"
	"hash/fnv"
	"maps"
)

var ErrLifecycleStorage = errors.New("plugin lifecycle storage unavailable")

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
	r.lifecycleMu.Lock()
	defer r.lifecycleMu.Unlock()
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
	active := make(map[string]string)
	for _, installation := range installations {
		if !installation.Promoted {
			continue
		}
		key := installation.ProjectID + ":" + installation.PluginID
		if active[key] != "" {
			return ErrLifecycleStorage
		}
		if _, registered := r.registered[installation.PluginID+"@"+installation.Version]; !registered {
			return ErrLifecycleStorage
		}
		active[key] = installation.Version
	}
	r.active = active
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
		r.bindObservationStorage(nil)
		return nil
	}
	if err := r.BindRegistryStorage(storage); err != nil {
		return err
	}
	r.BindInvocationStorage(legacyStoragePlanner{storage: storage})
	r.bindObservationStorage(storage)
	return nil
}

// LifecycleStorageHealth reports a failed lifecycle write without exposing
// database errors to callers. Authorization remains closed until a full
// lifecycle snapshot is saved successfully.
func (r *Registry) LifecycleStorageHealth() error {
	if r.storageFailed.Load() {
		return ErrLifecycleStorage
	}
	return nil
}

// FlushLifecycleState retries a complete snapshot after storage is restored.
func (r *Registry) FlushLifecycleState() error {
	r.readinessMu.RLock()
	defer r.readinessMu.RUnlock()
	r.installationMu.Lock()
	defer r.installationMu.Unlock()
	return r.persistLifecycleState()
}

func (r *Registry) persistLifecycleState() error {
	if r.storage == nil {
		return r.LifecycleStorageHealth()
	}
	if err := r.storage.SaveRegistryState(r.installations, r.readiness); err != nil {
		r.storageFailed.Store(true)
		return ErrLifecycleStorage
	}
	r.storageFailed.Store(false)
	return nil
}

// cloneInstallation prevents tentative grant/configuration changes from
// mutating the last known installation through shared map references.
func cloneInstallation(i Installation) Installation {
	i.Grants = maps.Clone(i.Grants)
	i.Config = maps.Clone(i.Config)
	i.ApprovedCapabilities = maps.Clone(i.ApprovedCapabilities)
	i.Resources = maps.Clone(i.Resources)
	for capability, resources := range i.Resources {
		i.Resources[capability] = maps.Clone(resources)
	}
	return i
}

func (r *Registry) saveInstallationLocked(key string, candidate Installation) error {
	previous, exists := r.installations[key]
	r.installations[key] = candidate
	if err := r.persistLifecycleState(); err != nil {
		if exists {
			r.installations[key] = previous
		} else {
			delete(r.installations, key)
		}
		return err
	}
	return nil
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
