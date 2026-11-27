# HTTP API reference

This is the HTTP surface mounted by the current application in [app.go](../internal/app/app.go), not a speculative client gateway. Plugin operations are declared separately in their manifests. The global Next.js admin UI lives under `/_system/ui`; project pages live under `/{project}/ui/...` and their API calls use `/{project}/admin/...`. No legacy root aliases exist.

## Endpoints

| Route | Authentication | Input and successful result |
| --- | --- | --- |
| `GET /_system/livez` | None | Liveness; body `ok` followed by a newline. |
| `GET /_system/readyz` | None | Dependency, worker, and storage readiness. |
| `POST /_system/admin/session` | Operator credentials | JSON `username`, `secret`; returns `session`. |
| `DELETE /_system/admin/session` | Operator bearer | Revoke the operator session. |
| `GET /_system/admin/projects` | Operator bearer | List project IDs, names, slugs, active flags, settings. |
| `GET /_system/admin/plugins` | Operator bearer | Global plugin catalog and installation inventory. |
| `GET /_system/admin/audit` | Operator bearer | Server audit events. |
| `GET /_system/admin/system` | Operator bearer | Global capabilities and system manifests. |
| `POST /_system/admin/actions` | Operator bearer | Server-wide lifecycle and project-creation commands. |
| `GET /{project}/admin/plugins` | Operator bearer | Installed plugin catalog within resolved project. |
| `GET /{project}/admin/functions` | Operator bearer | Function drafts within resolved project. |
| `POST /{project}/admin/functions` | Operator bearer | Function draft and publication commands. |
| `GET /{project}/admin/function-logs` | Operator bearer | Function execution logs (optional `name`). |
| `GET /{project}/admin/routes` | Operator bearer | Live manifest-declared route discovery; each returned `path` includes the canonical project slug (or ID). |
| `POST /{project}/admin/files/blob` | Operator bearer | Confirmed raw file upload; project limit up to 1 GB. |
| `GET /{project}/admin/files/blob` | Operator bearer | Download an authorized file as a binary attachment. |
| `POST /{project}/admin/actions` | Operator bearer | Project plugin, grant, invocation and data commands. |
| `ANY /{project}/functions/{function_name}` | Anonymous or authorized bearer | Published functions; only their declared GET, POST, PUT or PATCH method is accepted. |
| `POST /{project}/ws` | Project user bearer | Request a one-time ticket for a declared socket route. |
| `GET /{project}/ws` | Project connection ticket | Upgrade to a single project-scoped worker route. |
| `ANY /{project}/{resource...}` | Manifest-declared | Dynamic plugin route (the HTTP method, path and authentication must be declared in `plugin.yaml`). |

The first segment is **mandatory**. It is a project ID or its unique slug; `_system` is reserved and cannot be created as a project. Unknown projects and missing installations return 404; inactive projects return 403. The host resolves the immutable project ID before invoking workers and rejects contradictory project IDs from query arguments, headers, paths and command bodies. Public plugin invocations cannot inherit operator privileges. Display names are not URL identifiers. Discovery and published function URLs use the slug when configured, otherwise the immutable ID. Both identifiers resolve to the same project; clients must not prepend the namespace again to a discovered `path`.

Operator sessions expire eight hours after issuance; use `Authorization: Bearer <session>`. They are distinct from project-user tokens. Never ship operator credentials or sessions in a product client.

The shared HTTP envelope retains nonempty successful JSON responses. An empty successful admin/public response becomes HTTP 200 with `{"status":"success","message":"Request completed without a response body."}` when the underlying handler returned 204. Errors preserve HTTP status and return `{"status":"error","code":"Forbidden","message":"..."}` (code is HTTP status text, not a worker error code). Liveness/readiness use the actual health body/status; errors still pass through the shared envelope.

## Admin actions

Send project creation, configuration, activation, deactivation, uninstall and registered plugin readiness commands to `POST /_system/admin/actions`. Send all other actions to `POST /{project}/admin/actions`, where the host resolves the project identity. This allows an operator to reactivate an inactive project without enabling project workloads. Maximum request body: 64 KiB. In the table, **identity** means `project_id`, `plugin_id`, and `version`. The operator identity comes from the verified session, never from JSON. All destructive confirmations are explicit booleans. A confirmation is not authorization.

| Operation | Fields besides operation | Behavior |
| --- | --- | --- |
| `project.create` | project_id, optional name, slug, values | Create immutable project and URL namespace. |
| `project.configure` | project_id, optional name, slug, values | Update display metadata, unique slug or project settings; immutable ID never changes. |
| `project.activate` | project_id | Activate project. |
| `project.deactivate` | project_id, confirmed | Deactivate project and restrict work. |
| `project.uninstall` | project_id, confirmed | Uninstall project; review data retention separately. |
| `plugin.install` | identity | Stage a registered version without grants or activation. |
| `plugin.install.activate` | identity, capabilities, confirmed | Verify exact capability consent and atomically install, grant, and start an enabled worker. |
| `plugin.readiness` | plugin_id, version, ready | Set registered version readiness. |
| `plugin.enable` | identity | Enable after readiness, dependencies and grants. |
| `plugin.worker.start` | identity | Start bound worker. |
| `plugin.remove` | identity, confirmed | Remove installation. |
| `plugin.grant` | identity, capability | Grant declared capability. |
| `plugin.grant-resource` | identity, capability, resource | Grant exact resource scope. |
| `plugin.configure` | identity, values | Set schema-checked configuration; integers must be valid. |
| `plugin.admin.inventory` | identity, page_id, after, limit | Page of manifest-declared inventory; default limit 50. |
| `plugin.admin.revision` | identity, page_id, target_operation, payload | Read observation revision. |
| `plugin.admin.invoke` | identity, page_id, target_operation, payload, confirmed | Invoke declared admin action. |
| `plugin.admin.query` | identity, page_id, target_operation, payload | Read-only declared admin action. |
| `plugin.invoke` | identity, target_operation, payload | Invoke declared plugin operation as operator. |
| `plugin.revoke` | identity, capability, confirmed | Revoke capability. |
| `plugin.promote` | identity | Promote eligible version. |
| `plugin.drain` | identity | Drain version. |
| `plugin.stop-immediate` | identity, confirmed | Stop without waiting for drain. |
| `plugin.stop` | identity, confirmed | Stop version. |
| `plugin.resume` | identity | Resume eligible version. |
| `plugin.disable.preview` | identity | Return revision and dependent installations. |
| `plugin.disable.apply` | identity, revision, dependents, confirmed | Apply unchanged preview; refresh stale preview. |
| `plugin.recover` | identity | Recover eligible version. |
| `plugin.rollback` | identity, confirmed | Roll back eligible version. |
| `durable-data.delete` | project_id, plugin_id, key, confirmed | Explicit durable data deletion. |

Example (local operator only, use a disposable project):

```sh
BASE=http://127.0.0.1:18082
SESSION=$(curl -fsS "$BASE/_system/admin/session" \
  -H 'Content-Type: application/json' \
  --data '{"username":"admin","secret":"secret"}' | jq -r .session)
curl -fsS "$BASE/_system/admin/actions" \
  -H "Authorization: Bearer $SESSION" -H 'Content-Type: application/json' \
  --data '{"operation":"project.create","project_id":"docs-demo"}'
```

Install the desired registered plugin, mark its version ready, grant its manifest capabilities, then enable. If Access Management is enrolled, its rules must independently permit the operation.

For a read-only Games discovery call after those steps, send:

```json
{"operation":"plugin.invoke","project_id":"docs-demo","plugin_id":"games","version":"0.1.0","target_operation":"games.modes","payload":{}}
```

Worker responses use Go's exported field names: `Protocol`, `Success`, `Payload` (base64-encoded bytes containing JSON), `ErrorCode`, and optional `observations`. Check both HTTP status and `Success`; decode nonempty `Payload` before consuming it. Do not mistake a successful HTTP transport for a successful operation.

## Function commands

Send to `POST /{project}/admin/functions` (up to 1 MiB of JSON). All commands include `project_id` and `name`. The Functions installation must be enabled; applicable access rules use the actual function name.

| Operation | Additional fields | Result |
| --- | --- | --- |
| `functions.create` | source, function_version, public, permissions | Stored draft; version defaults to 1. |
| `functions.update` | source, function_version | Updated draft. |
| `functions.test` | source | Host preflight only, not proof of execution. |
| `functions.run` | source, input | Sandbox result or error and execution log. |
| `functions.permissions` | permissions, public | Save settings. |
| `functions.publish` | method (source config authoritative) | Publication version, method, timestamp and route. |

See [function development and publication](functions-runtime.md) for valid source, execution limits, authentication and examples. See [admin page contracts](admin-pages/README.md) for screens and [plugin communication](plugin-communication.md) for the internal RPC contract.

## Breaking migration

- Rename all health probes to `/_system/livez` or `/_system/readyz`.
- Rewrite old control-plane clients to `/_system/admin/...` and project commands to `/{project}/admin/...`.
- Rewrite published functions from `/public/{project_id}/{function_name}` to `/{project}/functions/{function_name}`; existing URLs return 404.
- Install and enable each plugin for each project and grant the declared capability before its HTTP route becomes available. Disable and promotion are reflected without restarting the host.
- Existing durable project IDs remain immutable; database startup adds name, slug and settings columns and backfills `name=slug=id`. Check for slug conflicts before migrating data.

## Not mounted as public HTTP

The generic worker RPC and SDK transport interfaces do not expose a network endpoint; only explicitly declared HTTP routes are published. In particular `/api/auth.signin` and `/api/rooms.join` are **not** routes mounted by this host. Use a trusted, project-authenticating adapter when implementing a product gateway. Never substitute an operator bearer for end-user authorization.

### WebSocket admission boundary

A manifest can opt a user-authenticated POST route into socket transport with `websocket_admission`, naming a separate user-authenticated worker operation using the same single path resource. Its worker must return `{"admitted":true}` only after membership validation. POST `/{project}/ws` with the project user bearer, same-host Origin and JSON `{"path":"/plugin/resource/command"}`. Use the returned ticket once within 60 seconds with WebSocket subprotocols `["brain.v1","ticket.<token>"]` on GET `/{project}/ws`. Do not put credentials in the URL. Send one JSON object per text message, omitting the resource and project identity. The host binds that resource, revalidates the session, exact installed route and worker admission for each invocation, and returns the worker response. Same-host HTTPS or loopback HTTP only; 64 KiB messages, 30-second idle timeout, 10-minute maximum connection lifetime and 256 concurrent connections. Tickets are process-local. Cross-origin, stale, unauthorized and replayed connections fail closed.


### Authenticated Files binary transport

`POST /{project}/admin/files/blob?version=0.1.0&key=object-key&filename=photo.jpg&contentType=image%2Fjpeg&confirmed=true` accepts raw bytes with an operator bearer token. The host checks the exact Files installation, operation, resource grant and access policy; it stages at most the configured size plus one byte, rechecks authorization and the size limit, and commits metadata and chunks together. The maximum setting is 1,000 MB (1 GB, decimal), default 100 MB. The stored `maxUploadBytes` remains a byte count. Excess size returns 413; concurrent staging is bounded. Existing 64 KiB admin JSON and 1 MiB worker limits are unchanged.

`GET /{project}/admin/files/blob?version=0.1.0&key=object-key` uses the existing Files download admission and returns an attachment as `application/octet-stream` with the original filename. Binary objects are represented by a `streamed: true` descriptor in the JSON download response, not an inline base64 body. Legacy objects remain readable. This is an admin-only trusted storage transport; it does not expose a public bucket URL.
