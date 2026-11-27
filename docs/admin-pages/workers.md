# Workers

Route: `/workers`

Shows plugin worker state and allows worker startup for enabled plugins. Requires at least one active plugin. Uses the shared admin sidebar.

Start only the intended project/version worker after lifecycle prerequisites are met. Readiness and worker readiness are distinct. If startup fails, inspect sandbox prerequisites and manifest sources; do not disable isolation or grant broad permissions as a workaround.

See the [admin route index](README.md) and [HTTP API](../http-api.md).
