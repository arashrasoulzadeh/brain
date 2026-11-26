package plugin

import "hash/fnv"

func (r *Registry) BindSQLiteStorage(storage *SQLiteStorage) {
	r.workerMu.Lock()
	defer r.workerMu.Unlock()
	r.storage = storage
	installations, readiness, err := storage.LoadRegistryState()
	if err != nil {
		return
	}
	r.installationMu.Lock()
	r.readinessMu.Lock()
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
	r.readinessMu.Unlock()
	r.installationMu.Unlock()
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
func (r *Registry) lockStorageInvocation(request WorkerRequest) func() {
	if r.storage == nil {
		return func() {}
	}
	switch request.PluginID {
	case "files", "matchmaking", "rooms", "analytics", "tables":
	default:
		return func() {}
	}
	hash := fnv.New32a()
	_, _ = hash.Write([]byte(request.ProjectID + ":" + request.PluginID))
	lock := &r.storageInvocations[hash.Sum32()%uint32(len(r.storageInvocations))]
	lock.Lock()
	return lock.Unlock
}
