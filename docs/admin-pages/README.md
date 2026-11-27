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

These are actual Next.js page patterns, not a promise that each alias has a contributed product page. Every plugin-backed action still needs a valid active project, exact installation, declared action and grant.

| Route | Guide | Behavior |
| --- | --- | --- |
| `/` | [Login and control plane](login.md) | Sign in and select a project. |
| `/projects` | [Projects](projects.md) | Create, select, activate, deactivate and uninstall projects. |
| `/plugins` | [Plugins](plugins.md) | Installed and registered version catalog. |
| `/plugins/[plugin]` | [Plugin detail](plugin-detail.md) | Exact plugin@version settings and grants. |
| `/audit` | [Audit](audit.md) | Lifecycle and delivery audit history. |
| `/system` | [System](system.md) | Capability and manifest metadata. |
| `/routes` | [Routes](routes.md) | Discovered operations and playground. |
| `/monitoring` | [Monitoring](monitoring.md) | Readiness, worker and lifecycle state. |
| `/workers` | [Workers](workers.md) | Worker inspection and startup controls. |
| `/functions` | [Functions](functions.md) | Function list for active project. |
| `/functions/new` | [Create](function-new.md) | Draft editor. |
| `/functions/[name]` | [Function detail](function-detail.md) | Editor, publication, settings, logs. |
| `/database` | [Database](database.md) | Manifest alias for table operations. |
| `/files` | [Files](files.md) | Manifest alias for file operations. |
| `/live-sessions` | [Live sessions](live-sessions.md) | Room, matchmaking and extension contributions. |
| `/workflows` | [Workflows](workflows.md) | Reserved manifest alias, unavailable without a contribution. |
| `/[legacy]` | [Manifest aliases](extensions.md) | Dynamic alias for contributed pages. |
| `/extensions/[project]/[plugin]/[version]/[page]` | [Extensions](extensions.md) | Exact authorized contribution, including generated models. |

All routes share the [theme](theme.md) and the [host HTTP contract](../http-api.md). The route inventory is checked against the actual page files during the documentation build.
