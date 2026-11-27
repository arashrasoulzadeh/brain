# Projects

Route: `/_system/ui/projects`

Shows project lifecycle state and project actions. Uses the shared admin sidebar. This page is available after operator login and does not require a product plugin.

Changing projects immediately clears the previous project's installation and function-draft views. Project-scoped requests are generation-checked, so a delayed response cannot overwrite a newer selection or a newer refresh. Creating the first active project selects it automatically; removing a different project preserves the current active selection.

Create projects with an immutable ID, display name and unique URL slug. **Configure** can update the name, slug and JSON settings without changing the immutable ID used for grants, storage or audit trails. The public API routes use `/{project}/...`; the operator UI itself is under `/_system/ui`.

Configuration and activate/deactivate/uninstall actions use a project-scoped confirmation dialog with pending and error states. Cancelling leaves the project unchanged.
