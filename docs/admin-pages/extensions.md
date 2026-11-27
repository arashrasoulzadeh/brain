# Manifest-driven admin pages

The admin host accepts `admin.version: 1` and declarative `admin.pages` in plugin.yaml. Pages define actions referencing declared routes, typed fields, optional result components, optional `legacy_route` aliases, and optional cross-plugin `extends` targets. No plugin-specific React imports are needed for these pages.

- Action fields support string, number, boolean and JSON. Host validation rejects undeclared fields and missing required values.
- Result components support table (filtering and client-side pagination), detail, metric, bar chart, JSON, code, logs, grid and tabs. Paths are safe own-property lookups; no plugin-provided script or HTML is executed.
- Cross-plugin extensions link to independently authorized pages. A plugin cannot grant itself access to another plugin.
- Read-only queries use `plugin.admin.query`; confirmed writes use `plugin.admin.invoke`. Every invocation rechecks the operator, active project, exact enabled/ready installation, grants and declared operation.
- Pages can subscribe to their declared events with `refresh_on`. The browser checks the exact last read-only action and its resource permission every five seconds after a successful query. It only replays that query when a declared observation changes; denied checks clear the old result even without an observation. Writes are never replayed. Events are best-effort, in-memory invalidations and no payload is exposed.
- The database, files, workflows and live-session consoles now use manifest aliases. A new plugin can contribute a custom alias without modifying Next.js. The specialized function editor and trusted platform lifecycle screens remain separate.
- Plugin settings generate configuration forms from the existing config schema; only changed fields are submitted and existing values are not exposed.

Arena, invitations and subscriptions contribute operational pages. Authentication contributes operator-only account metadata and session revocation, never credential or token inspection. The host now supports durable observation delivery through its SQLite outbox; the browser invalidation feed remains bounded and in memory. Broader inventory providers and full end-to-end deployment verification remain follow-up work.

## Resource inventory and embedded tabs

Pages may declare an `inventory` with `kind`, `capability`, `title`, and an optional `select_field` referring to a string action field. The host resolves the declaration against the authenticated project, enabled version and grants on every page request. The legacy SQLite adapter provides room IDs, ticket IDs, table names and file metadata using cursor pagination (50 rows per request). It rejects cross-plugin namespaces and broad enumeration with a resource-specific grant. File bodies and state values are never returned by this endpoint. Custom storage adapters may implement the trusted inventory provider interface.

A target page can offer `extension_slots: [tab]`. Another plugin can request `extends_slot: tab` with its existing `extends` target. Embedded forms retain their own plugin identity, version and operation authorization. A target that does not offer a tab slot receives the existing link behavior instead. No plugin code is executed in the browser.

Inventory lists re-read their bounded, policy-filtered page every five seconds while visible, including when a policy change produces no observation. They refresh immediately after confirmed mutations, and provide manual refresh and recovery after transient failures.

## Generated model administration

Declared `host.model` routes automatically produce admin pages without additional YAML or React code. The host derives only declared actions, exposes their typed model schema, and distinguishes read/list queries from confirmed create/update/delete/transition operations. Model records load through the same authenticated admin SDK and retain optimistic revision checks. Plugin settings link to the generated pages instead of maintaining a duplicate form implementation.

With access-management installed, inventories require an authenticated operator and a unique declared read-only resource route. They filter rows by the exact operator/project/plugin/version/operation/capability rules, intersect with installation resource grants, honor explicit deny (including wildcard deny), and never return a denied key or cursor. Unavailable policy providers, ambiguous routes, unsupported storage providers and excessive policy sets fail closed. The legacy compatibility API remains unavailable in policy-managed projects.

## Additional plugin administration

Arena provides match inspection, start and finish. Invitations provide room-scoped inspection (redacted codes) and cancellation. Realtime subscriptions expose channel counts and cursor metadata, plus confirmed channel-wide ticket revocation without displaying any ticket. Authentication exposes bounded account metadata (registered users and active guests) and project/user-scoped session revocation. Administrative inspection routes check the effective operator and admitted capability in host-injected context; user and delegated plugin principals cannot use them. New capabilities require explicit installation grants and, for policy-managed projects, matching operator access rules.

These views reflect the existing worker-local runtime state of Auth, Invites, Arena and Subscriptions. They do not imply durable accounts, invitations or live channel state across worker restarts.
