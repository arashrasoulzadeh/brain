# Audit log

Route: `/_system/ui/audit`

Shows operator and plugin lifecycle audit records. Uses the shared admin sidebar.

Use the audit filter to inspect project/plugin events, lifecycle failures and observation delivery. Audit entries record actions and metadata, not arbitrary observation payloads. For failed delivery, distinguish queued/retry/failed/enqueue_failed; check subscriber grant, active project and storage health. Audit history is not a substitute for a durable external compliance archive.

See the [admin route index](README.md) and [HTTP API](../http-api.md).
