# Functions

Route: `/{project}/ui/functions`

The function catalog lists only drafts authorized for inspection in the active project. The Functions installation must be enabled, ready, not draining, and have its required grants. Counts summarize the loaded authorized catalog, not the whole server.

- Search names; filter independently by publication, visibility and declared HTTP method.
- Sort by name, creation time, last publication or source version. Results are paginated in groups of 20.
- Inspect source versions, permission counts, creation and publication dates. The host does not expose an update timestamp or immutable deployment history in this response.
- Select a function name to open its editor, settings, publication, analytics and logs; choose Create function to start a new draft.
- Refresh to reload from the server, or retry a failed request. The installation section links to plugin management; plugin and function versions are different.

Access Management filters drafts by the exact operator and function name under functions.inspect. A listed or published draft is not proof that its public endpoint is usable. Check configured visibility, declared method and functions.call policy separately. Publication refers to the current host record, not a deployment history.

See [create function](function-new.md), [function detail](function-detail.md) and [function runtime](../functions-runtime.md).
