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

The current draft and execution-log adapter writes `.local/function-drafts.json` and `.local/function-logs.json` relative to the host working directory. Protect both files and include them in backup/restore alongside SQLite. The adapter is not yet a distributed version store. A published function now selects an immutable source, method, version and SHA-256 digest from its retained publication history (up to 128 distinct versions per function). Editing a draft does not change the live handler or HTTP method; publish a new version to change either. Publishing changed source with a version already pinned automatically allocates the next version above the highest retained publication. Identical source at the same version reuses its pin. The returned version is authoritative; existing pins are never overwritten. The 128-version history limit still applies. Public visibility can be revoked independently. Existing legacy public records are adopted from their first known source because historical bytes were not retained. Public execution fails closed if the durable publication is missing or corrupt. The runtime itself is still the host VM until distributed Functions migration. Logs record request ID, duration, status and workflow steps. Do not log or return secrets.

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

## Cortex published execution

In managed Cortex development or explicitly configured production Cortex, published public/user functions run on an isolated Neuron with no local fallback. Publication source and HTTP method are pinned independently of drafts. The user-function runtime exposes pure helpers and trusted identity, but not operator-only plugin reads. Operator draft tests retain their existing bridge pending migration. Publication history is stored in the existing Brain-local function metadata file; this is not multi-host storage. Full regression verification is pending.

### Troubleshooting published calls

The development admin origin on port 3000 proxies the same public function path to Brain (default port 18082). Use the published method and JSON input where required. Do not send an operator session as a public user bearer. If Access Management is installed, an anonymous call needs an allow rule for principal `public`, the installed Functions plugin version, operation and capability `functions.call`, and the function name as resource. Explicit denies still win. Missing/private functions and denied callers return the same 404 to conceal protected routes. After successful admission, unavailable execution capacity returns 503, an execution deadline returns 504, and execution/result failures return 500 with a safe message. Inspect the authenticated function Logs and Neurons views for further diagnostics; no raw worker error is exposed to public callers.

## Scheduled operator draft bridge

In Cortex mode, operator draft tests also execute on an isolated Neuron; `await context.plugins.call(operation, payload)` and `await context.execute(operation, payload)` use the same host-authorized read-only bridge. Only saved permissions and manifest-opted-in read operations are eligible. The parent stays in its slot; the child needs another compatible free slot. With a single slot or saturated capacity, the call fails promptly rather than queueing indefinitely. Calls are bounded to 16 per draft execution and 256 KiB per direction, with one active call at a time and the parent deadline. Public functions cannot invoke this bridge. Legacy explicit runtime mode retains its old adapter. General function-to-function execution and deeper continuation trees are not implemented.
