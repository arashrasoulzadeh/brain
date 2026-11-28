# Function detail

Route: `/{project}/ui/functions/{name}`

Edits source, permissions, publication, deployments, analytics, and logs. Requires Functions to be ready and enabled. Uses the shared admin sidebar and the source config is the only method authority.

Use the editor for source/version updates and source validation; publish returns a route and method. Permissions and visibility are stored settings, while actual caller authorization is host-enforced. Logs show execution status, request ID, durations and step details. Deployment/analytics panels reflect the current draft/log adapter; do not assume a distributed deployment or telemetry store.

See the [admin route index](README.md) and [HTTP API](../http-api.md).

Publication and permission/visibility changes require explicit confirmation. Editing settings preserves unsaved source; failed host validation is not labeled as successful. The logs view refreshes without overlapping requests and clears stale results if access is lost.

The editor displays Monaco diagnostics separately from authoritative host validation. Unsaved source is marked and may be restored with confirmation; browser unload warns while local changes are unsaved. Source-save success is distinguished from publication or refresh failure. Execution logs can be searched and filtered by outcome, with a manual refresh, timestamps and p95 latency calculated from the currently loaded logs only.

The Test tab sends current (including unsaved) source and JSON input through the existing operator-authorized functions.run endpoint. It does not publish or save changes, and it may create execution logs. Results are marked stale when input or source changes. The editor can update the positive source version and add explicit capability names; host authorization remains authoritative. Publication displays the active immutable version and its method independently of the saved draft. Changed source using an occupied version is automatically published at the next unused version; the editor displays the actual returned version. Log methods are shown only when recorded; inaccessible logs are not treated as an empty successful dataset. Publishing enables the public visibility flag under the existing host contract; access policy still controls calls.

## Source-first editor workspace

The Editor view opens with the source at full available width, a compact action bar (Validate, Save, Publish), and saved/unsaved status. Settings opens a collapsible inspector for version, method, visibility and configured permissions. Focus mode expands the editor vertically and temporarily hides the inspector; the inspector stacks below source at narrower widths. The test console opens beneath the editor, shares the Test view's input/result state and runs only when explicitly requested. Switching views does not discard editor or test state. The toolbar supports find, JavaScript formatting, word wrap, dark/light editor theme, line/column and diagnostic navigation, and Ctrl/Cmd+S when the editor has focus and saving is allowed. Monaco diagnostics remain advisory; host preflight and execution are distinct. These controls do not add a plugin bridge or change host authority. Source version and permissions remain mutable configuration, not a deployment snapshot.
