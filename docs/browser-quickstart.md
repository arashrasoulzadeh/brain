# Browser quickstart

Use the typed client with a project-scoped transport:

```ts
import { JavaScriptClient } from "../clients/javascript/src/index.js";

const client = new JavaScriptClient("dev", async (operation, projectId, payload) => {
  const response = await fetch("/api/" + operation, {
    method: "POST",
    credentials: "include",
    headers: { "content-type": "application/json" },
    body: JSON.stringify({ projectId, payload }),
  });
  if (!response.ok) throw new Error("request failed");
  return response.json();
});

const session = await client.signIn("user@example.test", "password");
await client.session(session.token);
```

The dashboard quickstart sequence is `create-session`, `load-project`, then
`open-dashboard`. Use first-party secure cookies for browser sessions and keep
operator credentials in the trusted dashboard host.
