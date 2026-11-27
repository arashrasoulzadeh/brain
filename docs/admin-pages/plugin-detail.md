# Plugin detail

Route: `/{project}/ui/plugins/{plugin}`

Shows one plugin version and lifecycle controls. It intentionally remains accessible while disabled so an operator can enable the plugin. Uses the shared admin sidebar.

The route parameter is the URL-encoded plugin ID and exact version separated by @. Mark ready, grant/revoke capabilities, enable and edit only changed config fields. Generated admin pages are listed only when eligible; follow links rather than bypassing authorization. Review dependent installations and active work before disabling or removing a version.

See the [admin route index](README.md) and [HTTP API](../http-api.md).
