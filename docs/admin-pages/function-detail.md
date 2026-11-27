# Function detail

Route: `/{project}/ui/functions/{name}`

Edits source, permissions, publication, deployments, analytics, and logs. Requires Functions to be ready and enabled. Uses the shared admin sidebar and the source config is the only method authority.

Use the editor for source/version updates and source validation; publish returns a route and method. Permissions and visibility are stored settings, while actual caller authorization is host-enforced. Logs show execution status, request ID, durations and step details. Deployment/analytics panels reflect the current draft/log adapter; do not assume an immutable deployment history or distributed telemetry store.

See the [admin route index](README.md) and [HTTP API](../http-api.md).
