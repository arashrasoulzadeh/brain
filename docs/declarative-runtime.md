# Declarative plugin runtime

This branch introduces a manifest-first runtime alongside the existing plugin implementations.

- `declarative: true` permits plugins without executable files. Only bounded responses, pipelines, model operations and fixed pure dependency calls are supported.
- `sdk.pipeline` resolves explicitly named input and previous-step values and invokes allowlisted helpers. Pure helper names are shared with the TypeScript SDK and user-function VM.
- `host.model` uses project/plugin/model-scoped SQLite records, strict fields, bounded list pages, explicit schema versions and compare-and-swap revisions. State changes use declared transitions.
- `host.call` delegates to a fixed source-free dependency response or pure pipeline. Both installations, grants and active versions are checked; scoped grants, draining, nested calls and observation effects are rejected.
- Workers receive host-built project, invocation, configuration and dependency context. Dependency availability is a snapshot, not an authorization grant.
- Observation delivery is bounded and best effort after successful commit. Room connection observations are persisted by analytics.
- The admin plugin settings page generates model forms and configuration inputs for the selected active project.
- Generate contracts with `brain plugin-contract examples/declarative-plugin/plugin.yaml markdown` or `brain plugin-contract examples/declarative-plugin/plugin.yaml typescript`.

## Example

See [the source-free catalog](../examples/declarative-plugin/plugin.yaml). It is an example, not automatically installed in existing deployments. To register it, place the directory under the configured plugin root and follow the normal project installation and grant process.

## Remaining migration boundaries

Existing product-specific storage adapters and complex business logic are retained for compatibility. General effectful dependency calls, durable observation delivery, all legacy route migrations, PostgreSQL model persistence, schema migration tooling, and complete dashboard navigation generation require further implementation. The user-function VM shares pure helpers only; this does not imply full host API access or OS-level isolation.
