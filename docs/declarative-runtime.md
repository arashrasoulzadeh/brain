# Declarative plugin runtime

The host supports a manifest-first runtime alongside executable plugin implementations.

- `declarative: true` permits plugins without executable files. Only bounded responses, pipelines, model operations and fixed pure dependency calls are supported.
- `sdk.pipeline` resolves explicitly named input and previous-step values and invokes allowlisted helpers. Pure helper names are shared with the TypeScript SDK and user-function VM.
- `host.model` uses project/plugin/model-scoped SQLite records, strict fields, bounded list pages, explicit schema versions and compare-and-swap revisions. State changes use declared transitions.
- `host.call` delegates to a fixed source-free dependency response or pure pipeline. Both installations, grants and active versions are checked; scoped grants, draining, nested calls and observation effects are rejected.
- Workers receive host-built project, invocation, configuration and dependency context. Dependency availability is a snapshot, not an authorization grant.
- Storage-backed mutations and durable observation outbox writes are committed atomically where supported. Delivery is bounded, retried and reauthorized; other host and in-memory effects use best-effort post-commit enqueue. Room connection observations feed Analytics and Subscriptions.
- The admin plugin settings page generates model forms and configuration inputs for the selected active project.
- Generate contracts with `brain plugin-contract examples/declarative-plugin/plugin.yaml markdown` or `brain plugin-contract examples/declarative-plugin/plugin.yaml typescript`.

## Example

See [the source-free catalog](../examples/declarative-plugin/plugin.yaml). It is an example, not automatically installed in existing deployments. To register it, place the directory under the configured plugin root and follow the normal project installation and grant process.

## Resource-scoped legacy routes

Use `resource_field: id` (or the declared input key for the resource) on a route when its capability may have resource-scoped grants. The host checks the declared input before execution and again before commit. Model routes instead use the declared model name. Routes without a resource binding deny scoped grants rather than treating them as unrestricted. Dependency calls still require unrestricted grants.

The SQLite compatibility planner normalizes resource keys and supplies host-owned snapshots for file reads, room membership/results, matchmaking tickets and table rows/schema. Missing durable resources fail closed instead of accepting caller-provided state. Direct TypeScript exports also receive the host context as their second argument.

## Remaining migration boundaries

Existing product-specific storage adapters and complex business logic are retained for compatibility. General effectful dependency calls, distributed observation delivery, remaining legacy route migrations, PostgreSQL model persistence and schema migration tooling require further implementation. Manifest-driven admin pages, model forms, extension tabs, inventory and observation refresh are implemented; they do not expose arbitrary worker operations. The user-function VM shares pure helpers only; this does not imply full host API access or OS-level isolation.
