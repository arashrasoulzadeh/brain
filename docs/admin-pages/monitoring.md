# Monitoring

Route: `/{project}/ui/monitoring`

Shows plugin readiness, enablement, worker, and drain health. Requires at least one active plugin. Uses the shared admin sidebar.

Use readiness, enablement and worker indicators together; an enabled version is not necessarily ready. Check project scope, grants, storage and worker startup audit. A green liveness check alone does not mean the host can accept traffic.

See the [admin route index](README.md) and [HTTP API](../http-api.md).

Health requires an enabled, ready, non-draining installation with all declared capability grants and explicit worker readiness. Unknown worker readiness is not healthy.
