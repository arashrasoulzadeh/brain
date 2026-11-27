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
