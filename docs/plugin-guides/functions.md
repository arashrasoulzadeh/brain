# Functions

Function authoring and invocation. [Authoritative manifest](../../plugins/functions/plugin.yaml).

## Developer

Manifest operations describe capabilities and route contracts; use the host authoring endpoint for actual restricted VM execution and publication. Function name is the access resource. The manifest's functions.call response is not itself the published-function HTTP executor.

## Administrator

Enable Functions, select a project, create and test a draft, set permissions and publish. Inspect redacted access and execution logs. Grant actual function-name scope when using Access Management.

## Example payloads

Use functions.create on /admin/functions with name, project_id and source such as export const config={method:"POST"}; export default input=>({input}); then test, run and publish. See [runtime](../functions-runtime.md).

Payload snippets are arguments to the declared operation, not standalone HTTP routes. Use the [HTTP invocation contract](../http-api.md) and exact manifest version/capability. Host-owned context, internal snapshots and grants must not be supplied by public callers.

## Limitations

Restricted JavaScript-compatible VM, not a full TypeScript compiler or OS sandbox. Local JSON drafts/logs are separate from SQLite; Go/WASM is not wired to the authoring endpoint.

Operator sandbox tests can call opted-in plugin reads through `context.plugins.call(operation, payload)`. Save the exact target capability on the function, and enable the target plugin with the matching resource policy. `files.metadata` is the first opted-in operation; public/user calls and writes are denied. See [runtime](../functions-runtime.md).
