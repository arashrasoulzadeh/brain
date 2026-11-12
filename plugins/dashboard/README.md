# Dashboard plugin

The dashboard plugin owns trusted operator project and plugin administration. It is not part of the core host.

## Supported v0.1 operations

- `project.create`
- `plugin.install`
- `plugin.grant`
- `plugin.enable`

The plugin creates typed host requests. The Go host remains responsible for authentication, authorization, project isolation, persistence, and effect commits.

## Grants

- `project.manage`
- `plugin.install`
- `plugin.grant`
- `plugin.enable`

## Lifecycle

The host validates this manifest, configuration, readiness, project grants, and lifecycle state before exposing dashboard operations.
