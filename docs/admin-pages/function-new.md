# Create function

Route: `/{project}/ui/functions/new`

Creates a function draft. Requires Functions to be ready and enabled. The HTTP method is declared in source with `export const config = { method: "POST" }`. Uses the shared admin sidebar.

The form takes name, source, positive version, permissions and public-route setting. Author JavaScript-compatible module source with the method config and default handler. Saving creates a draft and navigates to its detail page; validate and test separately. Keep public disabled until ready to expose the route.

See the [admin route index](README.md) and [HTTP API](../http-api.md).
