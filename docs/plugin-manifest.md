# HTTP routes in plugin.yaml

The Go host owns one HTTP router. Every production plugin route is declared in the same manifest as the plugin's operations, function bindings, capability grants, dependency versions and worker entrypoint. A plugin never registers its own root routes or constructs an absolute project prefix.

```yaml
id: rooms
version: 0.1.0
host_api: 1
mount: /rooms
capabilities:
  - rooms.read
  - rooms.create
routes:
  - operation: rooms.list
    function: listRooms
    capability: rooms.read
    method: GET
    path: /
    auth: user
  - operation: rooms.create
    function: createRoom
    capability: rooms.create
    method: POST
    path: /
    auth: user
```

The example is illustrative; the actual bundled Rooms manifest and source files remain the authority for exported operations. The mount starts with one slash and contains one lowercase URL-safe resource name; reserved mounts include `/_system`, `/admin`, `/functions`, `/ws` and `/ui`. A relative route path begins with `/` and may contain literal lowercase path segments and named `{id}` parameters. Percent encoding, ambiguous template overlaps, duplicate parameter names, traversal, trailing repeated slashes, unsupported HTTP methods, undeclared capabilities and invalid function bindings are rejected when loading manifests.

Allowed methods: `GET`, `POST`, `PUT`, `PATCH`, `DELETE`. Allowed authentication: `operator` (authenticated operator service), `user` (the installed project Auth plugin resolves a scoped session), or `public` (anonymous, with the broker's public principal and policy requirements). Authentication cannot be omitted from a published HTTP route. Existing routes without `method`, `path`, and `auth` remain internal plugin operations and are **not** published.

The host mounts `GET /{project}/rooms` or `POST /{project}/rooms`; it resolves the project ID/slug before the dispatch. Request query, route and JSON fields are combined only after rejecting conflicts, and client-provided project identity is never forwarded as authority. The worker receives the immutable project ID and a host-issued invocation context; the registry rechecks lifecycle and authorization before side effects commit.

The registry resolves only enabled, ready and non-draining installed versions. Promotion and rollback change the HTTP binding immediately, without an HTTP server restart. Multiple enabled versions without a promoted winner fail closed. The operator can discover active routes at `GET /{project}/admin/routes`, with each method, path, operation, function, capability, authentication and version.

Migration: Add explicit mount and HTTP route metadata to every published plugin, install and grant it per project, migrate callers to `/{project}/...`, and remove root endpoint assumptions. The former `/public/{project_id}/{name}` function URLs become `/{project}/functions/{name}`. `/_system` is reserved for global administration and liveness/readiness; `/{project}/ui/...` is reserved for the project admin UI.
