# Functions

Route: `/{project}/ui/functions`

The project-scoped function catalog lists authorized functions, with summary counts, name search, publication and visibility filters, 15-row pagination, method, version and creation/publication timestamps. Refresh reloads the authorized project data; no cross-project cache is used. Select a function to open its code and test workspace, or choose Create function.

Requires Functions to be ready and enabled. Access Management filters the list under functions.inspect. Configured public visibility does not guarantee caller access. The current backend is not an immutable version or deployment store.

See [create function](function-new.md), [function detail](function-detail.md) and [function runtime](../functions-runtime.md).
