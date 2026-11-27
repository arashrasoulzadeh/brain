# Login and control plane

Route: `/_system/ui`

Start the Go host and Next.js admin with [local development commands](../runtime-operations.md). Enter operator credentials; the admin stores its short-lived bearer in browser session storage, not as a product-user credential. Use only development credentials in a local environment. A restored session is checked against the server and expired sessions return to login.

Select an active project. Project selection is stored per browser session for global navigation; a project-specific URL overrides it and never falls back to another project. project-scoped requests must not overwrite the current project's views with stale results. The shared sidebar shows platform routes and authorized plugin contributions. The login language switch supports English and Persian with direction changes.

Sign out to revoke the server session when reachable, and clear local session/project state. If the host is unavailable, clear the browser session and follow operator recovery procedures. Do not expose this admin origin or operator credentials to public product clients.

See [admin routes](README.md) and [HTTP authentication](../http-api.md).
