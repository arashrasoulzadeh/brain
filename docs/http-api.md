# HTTP API reference

This is the HTTP surface mounted by the current application in [app.go](../internal/app/app.go), not a speculative client gateway. Plugin operations are declared separately in their manifests. The Next.js admin UI lives under `/_system/ui`; its project API calls use `/{project}/admin/...`. No legacy root aliases exist.

## Endpoints

| Route | Authentication | Input and successful result |
| --- | --- | --- |
| `GET /_system/livez` | None | Liveness; body `ok\\n`. |
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
| `GET /{project}/admin/routes` | Operator bearer | Live manifest-declared route discovery. |
| `POST /{project}/admin/actions` | Operator bearer | Plugin, grants and project lifecycle commands. |
| `ANY /{project}/functions/{function_name}` | Anonymous or authorized bearer | Published functions; only their declared GET, POST, PUT or PATCH method is accepted. |
| `ANY /{project}/{resource...}` | Manifest-declared | Dynamic plugin route (the HTTP method, path and authentication must be declared in `plugin.yaml`). |

The first segment is **mandatory**. It is a project ID or its unique slug; `_system` is reserved and cannot be created as a project. Unknown projects and missing installations return 404; inactive projects return 403. The host resolves the immutable project ID before invoking workers and rejects contradictory project IDs from query arguments, headers, paths and command bodies. Public plugin invocations cannot inherit operator privileges.

Operator sessions expire eight hours after issuance; use `Authorization: Bearer <session>`. They are distinct from project-user tokens. Never ship operator credentials or sessions in a product client.

The shared HTTP envelope retains nonempty successful JSON responses. An empty successful admin/public response becomes HTTP 200 with `{"status":"success","message":"Request completed without a response body."}` when the underlying handler returned 204. Errors preserve HTTP status and return `{"status":"error","code":"Forbidden","message":"..."}` (code is HTTP status text, not a worker error code). Liveness/readiness use the actual health body/status; errors still pass through the shared envelope.

## Admin actions

Send JSON to `POST /_system/admin/actions`. Maximum request body: 64 KiB. In the table, **identity** means `project_id`, `plugin_id`, and `version`. The operator identity comes from the verified session, never from JSON. All destructive confirmations are explicit booleans. A confirmation is not authorization.

| Operation | Fields besides operation | Behavior |
| --- | --- | --- |
| `project.create` | project_id, optional name, slug, values | Create immutable project and URL namespace. |
| `project.configure` | project_id, optional name, slug, values | Update display metadata, unique slug or project settings; immutable ID never changes. |
| `project.activate` | project_id | Activate project. |
| `project.deactivate` | project_id, confirmed | Deactivate project and restrict work. |
| `project.uninstall` | project_id, confirmed | Uninstall project; review data retention separately. |
| `plugin.install` | identity | Install registered version for project. |
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

## Breaking migration\n\n- Rename all health probes to `/_system/livez` or `/_system/readyz`.\n- Rewrite old control-plane clients to `/_system/admin/...` and project commands to `/{project}/admin/...`.\n- Rewrite published functions from `/public/{project_id}/{function_name}` to `/{project}/functions/{function_name}`; existing URLs return 404.\n- Install and enable each plugin for each project and grant the declared capability before its HTTP route becomes available. Disable and promotion are reflected without restarting the host.\n- Existing durable project IDs remain immutable; database startup adds name, slug and settings columns and backfills `name=slug=id`. Check for slug conflicts before migrating data.\n\n## Not mounted as public HTTP

The generic worker RPC and SDK transport interfaces do not expose a network endpoint; only explicitly declared HTTP routes are published. In particular `/api/auth.signin` and `/api/rooms.join` are **not** routes mounted by this host. Use a trusted, project-authenticating adapter when implementing a product gateway. Never substitute an operator bearer for end-user authorization.
