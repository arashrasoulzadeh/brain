# Host HTTP API

Use HTTPS in deployed environments. The routes below describe the Go host, not the Next.js proxy paths. Administrative routes are available only when the server is configured with an operator store and lifecycle service.

## Authentication and responses

POST a JSON object with `username` and `secret` to the operator session route. Use the returned `session` value as `Authorization: Bearer <session>` for all subsequent administrative requests. Never send or store credentials in logs, examples, or source control. Operator sessions and project user sessions are distinct.

Errors use HTTP status codes and a JSON object containing `status: "error"`, `code` and `message`. Data responses retain their documented JSON shape. Successful administrative mutations that would otherwise return an empty 204 response instead return HTTP 200 and a success message. Health checks are not authenticated; readiness may return 503 when required services or workers are unavailable.

## Endpoints

| Route | Authorization | Purpose |
| --- | --- | --- |
| `GET /livez` | None | Process liveness |
| `GET /readyz` | None | Configured readiness check with a two-second timeout |
| `POST /admin/session` | Operator credentials | Issue a session |
| `DELETE /admin/session` | Operator bearer token | Revoke that session |
| `GET /admin/projects` | Operator bearer token | List projects and active state |
| `GET /admin/plugins` | Operator bearer token | List manifests and installations; accepts `project_id` |
| `GET /admin/audit` | Operator bearer token | List audit records |
| `GET /admin/system` | Operator bearer token | List system manifests |
| `GET /admin/functions` | Operator bearer token | List drafts for `project_id`, subject to plugin access |
| `POST /admin/functions` | Operator bearer token | Create, edit, test, run or publish a draft, or update permissions |
| `GET /admin/function-logs` | Operator bearer token | List function logs for `project_id` and `name` |
| `POST /admin/actions` | Operator bearer token | Execute a lifecycle or plugin administration operation |
| `/public/{project_id}/{function_name}` | Publication and project policy | Invoke a published function using its configured method and, when required, project user bearer credentials |

## Administrative operations

Send the JSON body to `POST /admin/actions`. Specify the operation and the appropriate project, plugin and version fields. The host resolves these against the authorized installation, rather than trusting client-supplied grants. The request size limit is 64 KiB.

| Operation | Additional fields | Effect |
| --- | --- | --- |
| `project.create` | `project_id` | Create a project |
| `project.activate` | `project_id` | Activate a project |
| `project.deactivate` | `project_id`, `confirmed` | Deactivate a project |
| `project.uninstall` | `project_id`, `confirmed` | Uninstall project plugins |
| `plugin.install` | Project, plugin, version | Install declared version |
| `plugin.readiness` | Plugin, version, `ready` | Set version readiness |
| `plugin.enable` | Project, plugin, version | Enable authorized installation |
| `plugin.worker.start` | Project, plugin, version | Start its worker |
| `plugin.remove` | Project, plugin, version, `confirmed` | Remove installation |
| `plugin.grant` | Project, plugin, version, `capability` | Grant declared capability |
| `plugin.grant-resource` | Project, plugin, version, `capability`, `resource` | Grant scoped capability |
| `plugin.configure` | Project, plugin, version, `values` | Update validated configuration |
| `plugin.admin.inventory` | Project, plugin, version, `page_id`, `after`, `limit` | Read authorized paginated inventory; default limit 50 |
| `plugin.admin.revision` | Project, plugin, version, `page_id` | Read observation revision |
| `plugin.admin.invoke` | Project, plugin, version, `page_id`, `target_operation`, `payload`, `confirmed` | Execute declared admin action |
| `plugin.admin.query` | Project, plugin, version, `page_id`, `target_operation`, `payload` | Execute declared read-only admin action |
| `plugin.invoke` | Project, plugin, version, `target_operation`, `payload` | Invoke an authorized plugin operation |
| `plugin.revoke` | Project, plugin, version, `capability`, `confirmed` | Revoke capability |
| `plugin.promote` | Project, plugin, version | Promote version |
| `plugin.drain` | Project, plugin, version | Begin drain |
| `plugin.stop-immediate` | Project, plugin, version, `confirmed` | Stop immediately |
| `plugin.stop` | Project, plugin, version, `confirmed` | Stop installation |
| `plugin.resume` | Project, plugin, version | Recover installation |
| `plugin.disable.preview` | Project, plugin, version | Preview disable effects |
| `plugin.disable.apply` | Project, plugin, version, `revision`, `dependents`, `confirmed` | Apply matching preview |
| `plugin.recover` | Project, plugin, version | Recover installation |
| `plugin.rollback` | Project, plugin, version, `confirmed` | Roll back |
| `durable-data.delete` | Project, plugin, `key`, `confirmed` | Delete named durable data |

Project, plugin and version mean JSON fields `project_id`, `plugin_id`, and `version`. Confirmation fields must be true. Never automatically confirm destructive operations in response to an unrelated request.

Example read-only request:

```json
{
  "operation": "plugin.admin.query",
  "project_id": "demo",
  "plugin_id": "<installed-plugin>",
  "version": "<installed-version>",
  "page_id": "<declared-page>",
  "target_operation": "<declared-read-only-operation>",
  "payload": {}
}
```

Consult the [manifest reference](reference/plugins/index.md) for plugin-specific operations, permission requirements and payloads.

## Functions

Use `POST /admin/functions` with `project_id`, `name` and an operation below. The Functions plugin must be enabled and authorization is checked server-side. The body limit is 1 MiB.

| Operation | Additional fields |
| --- | --- |
| `functions.create` | `source`, `function_version`, `public`, `permissions`, `method` |
| `functions.update` | `source`, `function_version` |
| `functions.test` | `source` |
| `functions.run` | `source`, `input` |
| `functions.permissions` | `permissions`, `public` |
| `functions.publish` | `method` |

Published functions are reached through the public route using the configured HTTP method (GET, POST, PUT, or PATCH). Non-GET input must be JSON. If an Authorization header is supplied, the host validates it as a project user session, not an operator session. Publication, scope and function permissions are checked before invocation. Wrong methods return 405; unavailable functions return 404. See [function authoring](admin-pages/functions.md) for the UI workflow.
