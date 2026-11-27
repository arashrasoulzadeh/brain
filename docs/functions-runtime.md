# Function development and publication

The current Functions authoring interface stores JavaScript-compatible module source and runs it in a restricted Node VM. It is not an arbitrary Node process, a TypeScript transpilation service, or the planned production Go/WASM compilation system.

## Write and validate a function

```js
export const config = { method: "POST" };
export default function handler(input, context) {
  return {
    project: context.project.id,
    name: context.helpers.call("string.trim", input.name),
    authenticated: context.auth.actor.authenticated
  };
}
```

Use a name matching `[A-Za-z][A-Za-z0-9_-]{0,63}`, source at most 256 KiB, and an exported config method of GET, POST, PUT or PATCH. The source preflight rejects host/network keywords and requires a default export. Actual module evaluation checks the default export is callable. Return a JSON object, array or null; scalar and malformed outputs are rejected. A validation pass does not mean execution will succeed.

In the admin, install/enable Functions and grant its declared capabilities, choose an active project, create a draft, validate and run with JSON input, then publish. Publication uses the method declared in source. Use the published route returned by the API; GET receives null input, other methods require JSON. The host derives project and authentication context. `context.helpers` exposes the allowlisted pure SDK helpers, not arbitrary imports or services.

## Permissions and safety

When Access Management is installed, authoring uses `functions.publish`, validation/run use `functions.invoke`, draft inspection uses `functions.inspect`, logs use `functions.logs`, and published calls use `functions.call`. The resource is the function name. Public callers require the public rule; user bearers require the actual verified user's rule. Invalid bearers never fall back to anonymous access.

The Node VM blocks imports and string/WASM code generation, imposes bounded evaluation and invocation deadlines and an outer subprocess deadline. VM isolation is not equivalent to an OS sandbox; do not advertise arbitrary untrusted code execution as production-safe. The Go/WASM runner primitives are separate and not wired as the authoring endpoint.

## Storage and operations

The current draft and execution-log adapter writes `.local/function-drafts.json` and `.local/function-logs.json` relative to the host working directory. Protect both files and include them in backup/restore alongside SQLite. The adapter is not a distributed version store; a publication is not an immutable deployment artifact. Logs record request ID, duration, status and workflow steps. Do not log or return secrets.

See the [HTTP command reference](http-api.md), [auth context](plugin-auth-context.md), and [function detail screen](admin-pages/function-detail.md).
\n## Project-first publication URLs\n\nPublished routes have the shape `/{project}/functions/{function_name}`, using the configured project slug when present. Publication and logs remain project-bound to the immutable ID; the previous `/public/{project_id}/{function_name}` URLs are deliberately removed. Deploy clients together with this breaking route change.\n
## Host-mediated plugin reads (operator sandbox tests)

`context.plugins.call(operation, payload)` returns a promise containing the authorized plugin's JSON response. Save the draft first with the explicit capability in its permissions. The host resolves the active exact plugin version in the same project and rechecks the operator's Functions invocation policy, plugin lifecycle, installation grants, resource policy and saved function permission. Only routes explicitly marked `function_read: true` in their manifest and backed by a strict read-only resource contract are admitted. The first supported operation is `files.metadata`; arbitrary operations, writes, nested delegation, caller-supplied project identity and public/user invocations are denied. No credentials or worker objects are exposed to source. The bridge is bounded to 16 calls and 256 KiB per frame within the existing execution deadlines. Node VM isolation is still not an OS security boundary.

Example: save a POST function with permissions `functions.invoke, files.metadata`, then test with `{"key":"images/avatar.png"}` using an operator allowed to read that exact key:

```js
export const config = { method: "POST" };
export default async function handler(input, context) {
  if (!input?.key) return { error: "File key is required" };
  return await context.plugins.call("files.metadata", { key: input.key });
}
```

This works in operator sandbox testing only; the public invocation endpoint deliberately does not inherit operator rights. Public/user delegation requires a separately designed caller policy and is not enabled by these grants.
