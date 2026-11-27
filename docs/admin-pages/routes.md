# Routes

Route: `/routes`

Provides the route playground and discovered plugin routes. Requires at least one active plugin. Uses the shared admin sidebar.

The route inventory is resolved from `GET /{project}/admin/routes`, showing **only active project installations** with their HTTP method, project URL, capability, function and authentication mode. Select a declared plugin operation, validate JSON, choose an active project and invoke as the operator. The request uses plugin.invoke, not a generic public endpoint. Decode the worker payload and check worker success independently of HTTP status. Never paste real credentials into a playground example.

See the [admin route index](README.md) and [HTTP API](../http-api.md).
