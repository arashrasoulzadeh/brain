# Plugin detail

Route: `/{project}/ui/plugins/{plugin}`

Shows one plugin version and lifecycle controls. It intentionally remains accessible while disabled so an operator can enable the plugin. Uses the shared admin sidebar.

The route parameter is the URL-encoded plugin ID and exact version separated by @. Mark ready, grant/revoke capabilities, enable and edit only changed config fields. Generated admin pages are listed only when eligible; follow links rather than bypassing authorization. Review dependent installations and active work before disabling or removing a version.

See the [admin route index](README.md) and [HTTP API](../http-api.md).

## Project settings and actions

Manage links from plugin workspaces target the exact `/{project}/ui/plugins/{plugin}@{version}` installation. Configuration patches are scoped by immutable project ID and version. The settings summary displays explicitly declared public_config values and which remaining fields are configured; hidden values can be replaced without revealing them. Files exposes its effective numeric upload limit. Declared settings can be persisted by project and version; runtime enforcement depends on each plugin's implementation. Manage also provides installation, readiness, lifecycle, capability and dependency-preview actions. Editing settings opens a dialog with explicit confirmation and a saving status. Registered plugins with no configuration schema have no editable settings.
