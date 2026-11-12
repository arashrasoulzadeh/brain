# Dashboard plugin

The dashboard plugin owns trusted operator project and plugin administration. It is not part of the core host.

## Grants

- `project.manage`
- `plugin.install`
- `plugin.grant`
- `plugin.enable`

## Lifecycle

The host validates this manifest, configuration, readiness, project grants, and lifecycle state before exposing dashboard operations.
