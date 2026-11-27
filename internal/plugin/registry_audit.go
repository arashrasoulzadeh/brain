package plugin

import "time"

type AuditEvent struct {
	Action, ProjectID, PluginID, Version, Detail string

	OccurredAt time.Time
}

func (r *Registry) appendAudit(event AuditEvent) {
	r.auditMu.Lock()
	defer r.auditMu.Unlock()
	event.OccurredAt = time.Now().UTC()
	r.audit = append(r.audit, event)
}

func (r *Registry) Audit() []AuditEvent {
	r.auditMu.RLock()
	defer r.auditMu.RUnlock()
	return append([]AuditEvent(nil), r.audit...)
}
