# Admin page contract

Every authenticated admin page uses the shared persistent shell in `admin-nextjs/app/components/admin-shell.tsx`:

- `AdminPageFrame` registers its heading and checks required-plugin access; the persistent shell owns the layout.
- The [shared theme](theme.md) defines compact navigation, tables, forms and focus states.
- `AdminSidebar` is the only sidebar implementation.
- `AdminHeader` and `AdminFooter` are shared as well.
- Plugin-backed pages are visible and accessible only when their plugin is ready and enabled.
- The Plugins catalog and plugin detail pages remain accessible so an operator can install and enable a plugin.

The page notes below are the route-by-route contract used for manual verification.

## Route inventory

These are actual Next.js page patterns. The URL's first segment is authoritative for project pages: either the immutable ID or unique URL slug. A missing, unknown or inactive project never falls back to another saved project. Global login, project management and audit remain in the reserved system namespace.

| Route | Guide | Behavior |
| --- | --- | --- |
| `/_system/ui` | [Login and control plane](login.md) | Sign in and select a project. |
| `/_system/ui/projects` | [Projects](projects.md) | Create, select, activate, deactivate and uninstall projects. |
| `/_system/ui/audit` | [Audit](audit.md) | Global lifecycle and delivery audit history. |
| `/_system/ui/workers` | [Neurons](workers.md) | Global execution capacity and Neuron status. |
| `/[project]/ui` | [Plugins](plugins.md) | Project landing page. |
| `/[project]/ui/dashboard` | [Dashboard](dashboard.md) | Project installation and worker overview. |
| `/[project]/ui/plugins` | [Plugins](plugins.md) | Installed and registered version catalog. |
| `/[project]/ui/plugins/[plugin]` | [Plugin detail](plugin-detail.md) | Exact plugin@version settings and grants. |
| `/[project]/ui/system` | [System](system.md) | Capability and manifest metadata. |
| `/[project]/ui/routes` | [Routes](routes.md) | Discovered operations and playground. |
| `/[project]/ui/monitoring` | [Monitoring](monitoring.md) | Readiness, worker and lifecycle state. |
| `/[project]/ui/workers` | [Neurons](workers.md) | Compatibility link and project worker recovery. |
| `/[project]/ui/functions` | [Functions](functions.md) | Function list for active project. |
| `/[project]/ui/functions/new` | [Create](function-new.md) | Draft editor. |
| `/[project]/ui/functions/[name]` | [Function detail](function-detail.md) | Editor, publication, settings, logs. |
| `/[project]/ui/[legacy]` | [Manifest aliases](extensions.md) | Dynamic alias for contributed pages (including [database](database.md), [files](files.md), [live sessions](live-sessions.md) and [workflows](workflows.md)). |
| `/[project]/ui/extensions/[plugin]/[version]/[page]` | [Extensions](extensions.md) | Exact authorized contribution, including generated models. |

All project-backed actions still need a valid active project, exact installation, declared action and grant. The root redirects to the global login; unprefixed project UI pages are not mounted. All routes share the [theme](theme.md) and the [host HTTP contract](../http-api.md). The route inventory is checked against the actual page files during the documentation build.


## Bundled plugin workspaces

The sidebar is populated from the selected project's eligible manifests. These destinations render the shared admin shell, inventory, action forms, typed results and model editors, with trusted specialized controls where necessary. The URL prefix below is `/{project}/ui/`.

| Plugin | Destination | Operator workflow |
| --- | --- | --- |
| Files | `files` | Inventory, inspect, choose a local file, upload, download, publish. |
| Tables | `database` | Table inventory and search, inferred schema, bounded row pages, confirmed append, revision-checked replacement and deletion. |
| Auth | `auth` | Bounded account metadata and session revocation; no password/token inspection. |
| Access management | `access-management` | Read, create and revoke resource rules. |
| Rooms | `rooms` | Room inventory, details, membership and result operations. |
| Matchmaking | `matchmaking` | Ticket inventory, status, queue, assign, cancel. |
| Arena | `arena` | Inspect, start, record operator input and finish matches. |
| Invites | `invitations` | Create invitations, redacted invitation list and cancellation. |
| Subscriptions | `subscriptions` | Channel metadata and ticket revocation. |
| Analytics | `analytics` | Validated capture, exact-name event history and project event-count aggregates. |
| Games | `games` | Read-only supported game modes. |
| PvP | `pvp` | Read-only runtime capability discovery. |
| Functions | `functions`, `workflows` | Function editor, publication, logs and declared runtime operations. |
| Dashboard | `dashboard` | [Project overview](dashboard.md), installation status, Manage links and trusted lifecycle screens. |
| Example template | Generated model contribution | Declared item records, state transitions and name normalization. |

The old `live-sessions` URL remains a compatibility alias, not a separate sidebar entry. Individual plugin pages remain independently accessible even when they also appear in extension tabs. New bundled plugins must explicitly declare an operator workspace or be covered by an existing trusted/generated UI; the manifest inventory test detects omissions.

Every enabled plugin has a project-scoped sidebar destination: declared workspaces, the trusted Dashboard or Functions workspace, or the plugin's own Manage page if it has no declared action UI. Manage always exposes installation and capability settings, plus declared configuration fields. No arbitrary settings are invented for plugins without a configuration schema.

The TypeScript coverage workflow also executes the real Chromium and packaged-runtime integration scripts under V8 coverage and merges those profiles with the unit and DOM profiles. The per-file 100% lines/functions/branches/statements gate remains unchanged; executing an integration script does not waive uncovered failure branches.
