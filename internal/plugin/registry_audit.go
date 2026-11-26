package plugin

type AuditEvent struct {
	Action, ProjectID, PluginID, Version, Detail string
}

func (r *Registry) appendAudit(event AuditEvent) {
	r.auditMu.Lock()
	defer r.auditMu.Unlock()
	r.audit = append(r.audit, event)
}

func (r *Registry) Audit() []AuditEvent {
	r.auditMu.RLock()
	defer r.auditMu.RUnlock()
	return append([]AuditEvent(nil), r.audit...)
}
