# Function detail

Route: `/{project}/ui/functions/{name}`

The function workspace separates Code & test, Publication, Settings, Logs and Analytics.

- Code & test: JavaScript Monaco editor with an editable plain-text fallback, editable version metadata, unsaved-source indicator, explicit restore and reload, source size limit, host preflight and real `functions.run` execution against JSON input (UI limit 64 KiB). Test output and elapsed request time appear beside the editor. Editing source or input invalidates the displayed test result. Tests run under the authenticated operator, do not publish, and are logged.
- Publication: current version, method, timestamp and project-slug endpoint. No deployment history or rollback is fabricated. Saving source or access settings of an already-published record can immediately affect the endpoint; the UI warns and requires confirmation for saved changes.
- Settings: edit permissions and configured visibility with confirmation. Actual caller authorization remains host-enforced.
- Logs and Analytics: project/function-filtered records, manual and optional five-second refresh, outcome/search filters, pagination, execution step details and raw record. Counts, success rate, average and p95 are calculated over loaded records, not an external telemetry system.

Host preflight is not execution. Source config is the method authority. Request failures do not imply that a mutation was rolled back; refresh before retrying uncertain writes. Browser unload and the workspace breadcrumb protect unsaved edits; unrelated sidebar navigation is not globally intercepted.

Requires Functions to be ready and enabled. See the [admin route index](README.md), [HTTP API](../http-api.md), and [function runtime](../functions-runtime.md).
