# Plugins

Route: `/{project}/ui/plugins`

Shows registered and installed plugin versions. This is the lifecycle entry point for installing, marking ready, granting capabilities, and enabling plugins. Uses the shared admin sidebar.

Select the intended active project before installing. The Install dialog lists every declared capability and required dependency for the exact version. Confirm Install and enable to grant the listed capabilities, start the worker and persist an enabled installation in one operation. Cancellation does not mutate state. A stale capability list, missing enabled dependency or worker failure prevents installation. Resource access rules remain an additional check; installing Access Management can require explicit policy rules. The lower-level staged install, readiness, grant and enable API remains available for controlled upgrades. The sidebar refreshes after installation and lifecycle changes. Removing a grant or disabling a version can stop its worker. Review confirmation and audit; disabled detail pages remain reachable so the installation can be repaired.

See the [admin route index](README.md) and [HTTP API](../http-api.md).

The Enable action confirms the exact required dependency versions it will enable, including transitive dependencies. Missing or ungranted dependencies must be staged, made ready and granted explicitly first; optional dependencies are never auto-enabled. Fresh install and enable retains its existing requirement that dependencies already be enabled.
