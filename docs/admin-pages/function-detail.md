# Function detail

Route: `/{project}/ui/functions/{name}`

Edits source, permissions, publication, deployments, analytics, and logs. Requires Functions to be ready and enabled. Uses the shared admin sidebar and the source config is the only method authority.

Use the editor for source/version updates and source validation; publish returns a route and method. Permissions and visibility are stored settings, while actual caller authorization is host-enforced. Logs show execution status, request ID, durations and step details. Deployment/analytics panels reflect the current draft/log adapter; do not assume an immutable deployment history or distributed telemetry store.

See the [admin route index](README.md) and [HTTP API](../http-api.md).

Publication and permission/visibility changes require explicit confirmation. Editing settings preserves unsaved source; failed host validation is not labeled as successful. The logs view refreshes without overlapping requests and clears stale results if access is lost.

The editor displays Monaco diagnostics separately from authoritative host validation. Unsaved source is marked and may be restored with confirmation; browser unload warns while local changes are unsaved. Source-save success is distinguished from publication or refresh failure. Execution logs can be searched and filtered by outcome, with a manual refresh, timestamps and p95 latency calculated from the currently loaded logs only.

The Test tab sends current (including unsaved) source and JSON input through the existing operator-authorized functions.run endpoint. It does not publish or save changes, and it may create execution logs. Results are marked stale when input or source changes. The editor can update the positive source version and add explicit capability names; host authorization remains authoritative. Publication displays the current mutable record and timestamp, not a separate deployed version. Log methods are shown only when recorded; inaccessible logs are not treated as an empty successful dataset. Publishing enables the public visibility flag under the existing host contract; access policy still controls calls.
