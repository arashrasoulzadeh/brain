# Manifest-driven admin pages

The admin host accepts `admin.version: 1` and declarative `admin.pages` in plugin.yaml. Pages define actions referencing declared routes, typed fields, optional result components, optional `legacy_route` aliases, and optional cross-plugin `extends` targets. No plugin-specific React imports are needed for these pages.

- Action fields support string, number, boolean and JSON. Host validation rejects undeclared fields and missing required values.
- Result components support table (filtering and client-side pagination), detail, metric, bar chart, JSON, code, logs, grid and tabs. Paths are safe own-property lookups; no plugin-provided script or HTML is executed.
- Cross-plugin extensions link to independently authorized pages. A plugin cannot grant itself access to another plugin.
- Read-only queries use `plugin.admin.query`; confirmed writes use `plugin.admin.invoke`. Every invocation rechecks the operator, active project, exact enabled/ready installation, grants and declared operation.
- Pages can subscribe to their declared events with `refresh_on`. The browser polls a bounded, project-scoped revision feed every five seconds after a successful read, and only replays the last read-only query on changes. Writes are never replayed. Events are best-effort, in-memory invalidations and no payload is exposed.
- The database, files, workflows and live-session consoles now use manifest aliases. A new plugin can contribute a custom alias without modifying Next.js. The specialized function editor and trusted platform lifecycle screens remain separate.
- Plugin settings generate configuration forms from the existing config schema; only changed fields are submitted and existing values are not exposed.

Arena, invites and subscriptions declare routes without matching exported runtime handlers; they do not get speculative action pages. Auth user/session operations are not automatically exposed as admin controls. Dedicated resource list APIs, richer embedded extension slots, durable event replay and full end-to-end deployment verification are follow-up work.
