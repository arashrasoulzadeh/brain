# Create function

Route: `/{project}/ui/functions/new`

Creates a function draft in the selected authorized project. The Functions installation must be enabled and ready.

Configure a unique name (1–64 characters, starting with a letter), positive source version, permissions, and public visibility. The source config is authoritative for the HTTP method. Starter templates (echo JSON, health check, array transform) can generate source with GET, POST, PUT, or PATCH; applying one requires confirmation and replaces the editor contents.

Use the Monaco editor, host preflight, and operator-authorized sandbox runner before creating a draft. The runner executes the current source with JSON input and may write logs, but does not publish or save the draft. It is not a public caller test. Return an object, array, or null.

Creating saves a draft and navigates to its detail page. Publishing is a separate action. The existing host create operation is not an atomic create-only uniqueness guarantee; use a unique name.

See the [admin route index](README.md) and [HTTP API](../http-api.md).

The shared source editor also provides find, JavaScript formatting, word wrapping, theme switching and line/column status. Its JavaScript language mode reflects the current runtime's lack of TypeScript transpilation.
