# Plugins

Route: `/plugins`

Shows registered and installed plugin versions. This is the lifecycle entry point for installing, marking ready, granting capabilities, and enabling plugins. Uses the shared admin sidebar.

Select the intended active project before installing. Install the version, mark it ready, grant declared capabilities, then enable; access-management rules are an additional check. Removing a grant or disabling a version can stop its worker. Review confirmation and audit; disabled detail pages remain reachable so the installation can be repaired.

See the [admin route index](README.md) and [HTTP API](../http-api.md).
