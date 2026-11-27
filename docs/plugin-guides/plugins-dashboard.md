# Dashboard

## Developer

Dashboard lifecycle routes use `host.admin`; they are not executable plugin-worker functions. Use the trusted operator API and the [admin page guide](../admin-pages/README.md). Use the host-issued observation `lifecycleChanged` to refresh admin views rather than exposing operator credentials to plugin code.

## Administrator

Authenticate an operator before creating projects, installing plugins or changing grants. Preview destructive disable actions and confirm the exact project, plugin and version before applying them. Review audit records, readiness and drain state before promotion, rollback or removal. Do not grant general users access to operator endpoints.
