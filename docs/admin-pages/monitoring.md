# Monitoring

Route: `/monitoring`

Shows plugin readiness, enablement, worker, and drain health. Requires at least one active plugin. Uses the shared admin sidebar.

Use readiness, enablement and worker indicators together; an enabled version is not necessarily ready. Check project scope, grants, storage and worker startup audit. A green liveness check alone does not mean the host can accept traffic.

See the [admin route index](README.md) and [HTTP API](../http-api.md).
