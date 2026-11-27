# Workflows

Route: `/{project}/ui/workflows`

This route is a manifest-driven alias. The current registered manifests do not declare a workflows page, so the console shows Page unavailable until an installed, authorized plugin contributes that alias. Enabling Functions alone does not implement a workflow scheduler.

A future plugin can declare a `legacy_route: workflows` page with its own operations, capabilities, fields and result renderers; the shared host validates those declarations. The existing project-bound workflow and schedule helpers are local primitives, not a deployed scheduler or queue service.

See [manifest-driven extensions](extensions.md) and [declarative runtime](../declarative-runtime.md).
