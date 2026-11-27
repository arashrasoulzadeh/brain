# Projects

Route: `/projects`

Shows project lifecycle state and project actions. Uses the shared admin sidebar. This page is available after operator login and does not require a product plugin.

Changing projects immediately clears the previous project's installation and function-draft views. Project-scoped requests are generation-checked, so a delayed response cannot overwrite a newer selection or a newer refresh. Creating the first active project selects it automatically; removing a different project preserves the current active selection.
