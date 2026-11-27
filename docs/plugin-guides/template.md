# Example plugin

Worker, pipeline and model patterns. [Authoritative manifest](../../plugins/template/plugin.yaml).

## Developer

Use the template to learn sdk.context, typed observations, sdk.pipeline and host.model. Define schema and routes in YAML; never use a worker call to bypass model revision checks.

## Administrator

Install only in a disposable development project. Grant the example routes and observation capability if testing room events; avoid installing demonstration code in production.

## Example payloads

example.plugin.example: {}; example.normalize uses its manifest-declared input; example.items.create/read/list/update/delete/publish use the declared item model, ID and current revision for writes. See the manifest for exact fields.

Payload snippets are arguments to the declared operation, not standalone HTTP routes. Use the [HTTP invocation contract](../http-api.md) and exact manifest version/capability. Host-owned context, internal snapshots and grants must not be supplied by public callers.

## Limitations

Illustrative plugin, not a production business service.
