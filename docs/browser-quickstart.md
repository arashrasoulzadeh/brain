# Browser and JavaScript client

The [JavaScript client](../clients/javascript/src/client.ts) takes an application-supplied, generic `ClientTransport`. It is a transport contract, not a built-in browser gateway. The production Go HTTP server does not expose generic user RPC URLs; see [HTTP API](http-api.md).

```ts
import { JavaScriptClient, type ClientTransport } from "./clients/javascript/src/index.js";

// Supply a real, trusted project-aware transport from your application.
// Do not implement it by embedding an operator token.
function createClient(transport: ClientTransport) {
  return new JavaScriptClient("dev", transport);
}
```

The client validates project IDs and exposes `signIn(username, password)` (returns token and username) and `session(token)` (returns username). The durable auth adapter requires a username in the host identifier grammar, not an arbitrary email address. The host checks the project and user independently. Treat returned tokens as credentials; secure cookie storage and CSRF protection must be integrated at the actual application gateway.

For a published Function, send the method and URL returned by publication with the appropriate JSON input; optionally send a verified project-user bearer. See [function runtime](functions-runtime.md).

The separate [admin client](../clients/javascript/src/admin.ts) provides `adminLogin(apiOrigin, username, secret)` and `adminAction(apiOrigin, session, action)`. It is for trusted operator tooling; the latter resolves to void, not the returned plugin payload. The full HTTP API describes result handling. Never bundle operator credentials into a user-facing browser.

Run the JavaScript and plugin tests from the root with `npm ci`, `npm run typecheck` and `npm run test:ts`.
