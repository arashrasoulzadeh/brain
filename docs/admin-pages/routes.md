# Routes

Route: `/{project}/ui/routes`

Provides the route playground and discovered plugin routes. Requires at least one active plugin. Uses the shared admin sidebar.

The route inventory is resolved from `GET /{project}/admin/routes`, showing **only active project installations** with their HTTP method, complete project-prefixed path (slug or ID), capability, function and authentication mode. Select a declared plugin operation, validate JSON and confirm the exact discovered version in the invocation dialog. It displays the worker result separately from the input; a worker failure is not reported as transport success. The request uses plugin.invoke, not a generic public endpoint. Decode the worker payload and check worker success independently of HTTP status. Never paste real credentials into a playground example.

See the [admin route index](README.md) and [HTTP API](../http-api.md).
