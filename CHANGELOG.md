# Changelog

Automatically generated from commits on main.

## Unreleased
- `557658d8` — 2026-11-27 21:15:03 +0330 — Merge pull request #46 from arashrasoulzadeh/feat/confirmed-plugin-install
  Files: 
- `03ac9885` — 2026-11-27 21:13:19 +0330 — Document confirmed install command and Dashboard route in inventory
  Files: docs/admin-pages/README.md
- `73531656` — 2026-11-27 21:13:16 +0330 — Document confirmed install command and Dashboard route in inventory
  Files: docs/http-api.md
- `ad138e98` — 2026-11-27 21:07:34 +0330 — Test actual sidebar links and safeguard active version and project scope
  Files: admin-nextjs/app/components/plugin-install-dialog.tsx,admin-nextjs/app/dashboard/view.tsx admin-nextjs/app/view.tsx,internal/plugin/registry_install_activate.go scripts/test-admin-browser.mjs,tests/admin/admin-menu.test.ts
- `10226493` — 2026-11-27 21:07:15 +0330 — Confirm plugin capabilities and install active workers with live sidebar refresh
  Files: admin-nextjs/app/[project]/ui/dashboard/page.tsx,admin-nextjs/app/components/admin-chrome.tsx admin-nextjs/app/components/admin-config-form.tsx,admin-nextjs/app/components/admin-shell.tsx admin-nextjs/app/components/legacy-admin-page.tsx,admin-nextjs/app/components/plugin-catalog-panel.tsx admin-nextjs/app/components/plugin-install-dialog.tsx,admin-nextjs/app/dashboard/view.tsx admin-nextjs/app/lib/admin-api.ts,admin-nextjs/app/lib/admin-menu.ts admin-nextjs/app/lib/admin-navigation.ts,admin-nextjs/app/lib/admin-registry.ts admin-nextjs/app/lib/admin-types.ts,admin-nextjs/app/lib/plugin-install.ts admin-nextjs/app/plugins/[plugin]/view.tsx,admin-nextjs/app/view.tsx admin-nextjs/app/workers/view.tsx,docs/admin-pages/README.md docs/admin-pages/dashboard.md,docs/admin-pages/plugins.md internal/admin/lifecycle.go,internal/app/admin.go internal/plugin/registry_install_activate.go,internal/plugin/registry_install_activate_test.go scripts/test-admin-browser.mjs,scripts/test-admin-live.mjs
- `c26b3f3d` — 2026-11-27 17:21:01 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `048b932a` — 2026-11-27 20:50:06 +0330 — Merge pull request #44 from arashrasoulzadeh/feat/plugin-action-dialogs
  Files: 
- `35136672` — 2026-11-27 20:49:32 +0330 — Merge main and preserve session-expiry and plugin dialog regression tests; align plugin admin documentation
  Files: docs/admin-pages/README.md,docs/admin-pages/files.md docs/admin-pages/plugin-detail.md,scripts/test-admin-browser.mjs
- `385259dd` — 2026-11-27 17:19:00 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `1544e17b` — 2026-11-27 20:48:49 +0330 — Merge pull request #43 from arashrasoulzadeh/fix/admin-expired-session-layout
  Files: 
- `42ff350a` — 2026-11-27 20:26:07 +0330 — Persist file metadata and complete per-project plugin management audit
  Files: admin-nextjs/app/components/admin-config-form.tsx,admin-nextjs/app/components/admin-file-controls.tsx admin-nextjs/app/components/admin-inventory.tsx,admin-nextjs/app/components/admin-renderer.tsx admin-nextjs/app/lib/admin-types.ts,admin-nextjs/app/plugins/[plugin]/view.tsx internal/plugin/admin_inventory.go,internal/plugin/file_metadata.go internal/plugin/file_metadata_test.go,internal/plugin/legacy_admin_inventory.go internal/plugin/legacy_storage_adapter.go,internal/plugin/manifest.go internal/plugin/plugin.go,internal/plugin/registry_catalog.go internal/plugin/sqlite_storage.go,plugins/access-management/plugin.yaml plugins/analytics/plugin.yaml,plugins/arena/plugin.yaml plugins/auth/plugin.yaml,plugins/dashboard/plugin.yaml plugins/files/plugin.yaml,plugins/files/routes.ts plugins/functions/plugin.yaml,plugins/invites/plugin.yaml plugins/matchmaking/plugin.yaml,plugins/rooms/plugin.yaml plugins/subscriptions/plugin.yaml,plugins/tables/plugin.yaml plugins/template/plugin.yaml,scripts/test-admin-live.mjs
- `10ff7ca7` — 2026-11-27 20:16:54 +0330 — Use dialogs for management lifecycle actions and finish compatibility fixes
  Files: admin-nextjs/app/components/admin-file-controls.tsx,admin-nextjs/app/plugins/[plugin]/view.tsx internal/plugin/registry_catalog.go
- `654fb5f0` — 2026-11-27 20:16:20 +0330 — Use shared plugin action dialogs and project-scoped management settings
  Files: admin-nextjs/app/components/admin-config-form.tsx,admin-nextjs/app/components/admin-dialog.tsx admin-nextjs/app/components/admin-file-controls.tsx,admin-nextjs/app/components/admin-renderer.tsx admin-nextjs/app/components/model-panel.tsx,admin-nextjs/app/components/plugin-manage-link.tsx admin-nextjs/app/functions/view.tsx,admin-nextjs/app/lib/admin-registry.ts admin-nextjs/app/lib/admin-types.ts,admin-nextjs/app/plugins/[plugin]/view.tsx admin-nextjs/app/theme.css,docs/admin-pages/extensions.md docs/admin-pages/files.md,docs/admin-pages/plugin-detail.md internal/plugin/file_upload_limit.go,internal/plugin/file_upload_limit_test.go internal/plugin/plugin.go,internal/plugin/registry_catalog.go internal/plugin/registry_permissions.go,internal/plugin/registry_workers.go plugins/files/plugin.yaml,scripts/test-admin-browser.mjs scripts/test-admin-live.mjs
- `d1de78bb` — 2026-11-27 20:11:09 +0330 — Invalidate expired admin sessions centrally and show standalone login without protected shell
  Files: admin-nextjs/app/components/admin-shell.tsx,admin-nextjs/app/components/toast-provider.tsx admin-nextjs/app/lib/admin-api.ts,admin-nextjs/app/lib/admin-session-response.ts admin-nextjs/app/plugins/[plugin]/view.tsx,admin-nextjs/app/view.tsx admin-nextjs/app/workers/view.tsx,docs/admin-pages/login.md scripts/test-admin-browser.mjs,tests/admin/admin-session-response.test.ts
- `887ace77` — 2026-11-27 16:37:47 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `eee64e74` — 2026-11-27 20:07:36 +0330 — Merge pull request #42 from arashrasoulzadeh/feat/plugin-admin-workspaces
  Files: 
- `bf909810` — 2026-11-27 19:57:25 +0330 — Type the shared file action state transition as a string record
  Files: admin-nextjs/app/components/admin-renderer.tsx
- `056d1fe7` — 2026-11-27 19:56:24 +0330 — Clear prior object key when starting an upload and identify stored representation size
  Files: admin-nextjs/app/components/admin-inventory.tsx,admin-nextjs/app/components/admin-renderer.tsx
- `ccda373c` — 2026-11-27 19:55:54 +0330 — Add individual plugin workspaces and trusted file upload/download controls
  Files: admin-nextjs/app/components/admin-file-controls.tsx,admin-nextjs/app/components/admin-renderer.tsx admin-nextjs/app/components/legacy-admin-page.tsx,admin-nextjs/app/lib/admin-file-data.ts admin-nextjs/app/lib/admin-menu.ts,admin-nextjs/app/lib/admin-registry.ts admin-nextjs/app/lib/admin-sdk.ts,admin-nextjs/app/theme.css docs/admin-pages/README.md,docs/admin-pages/extensions.md docs/admin-pages/files.md,docs/admin-pages/live-sessions.md docs/plugin-guides/files.md,plugins/access-management/plugin.yaml plugins/analytics/plugin.yaml,plugins/arena/plugin.yaml plugins/auth/plugin.yaml,plugins/games/plugin.yaml plugins/invites/plugin.yaml,plugins/matchmaking/plugin.yaml plugins/pvp/plugin.yaml,plugins/rooms/plugin.yaml plugins/subscriptions/plugin.yaml,scripts/test-admin-live.mjs tests/admin/admin-file-data.test.ts,tests/admin/admin-menu.test.ts tests/admin/admin-registry.test.ts,tests/plugin/all_manifests_test.go
- `e69aa54f` — 2026-11-27 16:19:42 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `c415bea9` — 2026-11-27 19:49:31 +0330 — Merge pull request #41 from arashrasoulzadeh/fix/canonical-project-http-paths
  Files: 
- `ddc4fd1e` — 2026-11-27 19:48:07 +0330 — Synchronize live admin navigation assertions with Next.js client routing
  Files: scripts/test-admin-live.mjs
- `bd2b5a66` — 2026-11-27 19:43:34 +0330 — Check mobile project identity after opening the collapsed navigation
  Files: scripts/test-admin-browser.mjs
- `14991857` — 2026-11-27 19:41:14 +0330 — Resolve NodeNext TypeScript imports in the Next.js production build
  Files: admin-nextjs/next.config.ts
- `e97b772a` — 2026-11-27 19:39:55 +0330 — Reject ambiguous project UI mounts and verify direct URL project authority in Chromium
  Files: admin-nextjs/app/components/admin-shell.tsx,docs/admin-pages/login.md docs/admins.md,scripts/test-admin-browser.mjs scripts/test-admin-live.mjs,tests/plugin/http_manifest_test.go
- `b1079469` — 2026-11-27 19:39:42 +0330 — Add URL authority and project navigation regressions and update admin route documentation
  Files: admin-nextjs/app/[project]/ui/layout.tsx,admin-nextjs/app/[project]/ui/template.tsx admin-nextjs/app/components/admin-chrome.tsx,admin-nextjs/app/functions/view.tsx admin-nextjs/app/lib/admin-api.ts,admin-nextjs/app/lib/admin-menu.ts admin-nextjs/next.config.ts,docs/admin-pages/README.md docs/admin-pages/audit.md,docs/admin-pages/database.md docs/admin-pages/files.md,docs/admin-pages/function-detail.md docs/admin-pages/function-new.md,docs/admin-pages/functions.md docs/admin-pages/live-sessions.md,docs/admin-pages/login.md docs/admin-pages/monitoring.md,docs/admin-pages/plugin-detail.md docs/admin-pages/plugins.md,docs/admin-pages/projects.md docs/admin-pages/routes.md,docs/admin-pages/system.md docs/admin-pages/workers.md,docs/admin-pages/workflows.md docs/http-api.md,docs/plugin-manifest.md scripts/documentation_inventory.py,scripts/test-admin-browser.mjs tests/admin/admin-api.test.ts,tests/admin/admin-menu.test.ts tests/admin/admin-registry.test.ts,tests/admin/admin-shell.test.ts
- `9e476612` — 2026-11-27 19:37:03 +0330 — Move admin browser routes into explicit project and system UI namespaces
  Files: admin-nextjs/app/%5Fsystem/ui/audit/page.tsx,admin-nextjs/app/%5Fsystem/ui/page.tsx admin-nextjs/app/%5Fsystem/ui/projects/page.tsx,admin-nextjs/app/[legacy]/page.tsx admin-nextjs/app/[project]/ui/[legacy]/page.tsx,admin-nextjs/app/[project]/ui/extensions/[plugin]/[version]/[page]/page.tsx admin-nextjs/app/[project]/ui/functions/[name]/page.tsx,admin-nextjs/app/[project]/ui/functions/new/page.tsx admin-nextjs/app/[project]/ui/functions/page.tsx,admin-nextjs/app/[project]/ui/monitoring/page.tsx admin-nextjs/app/[project]/ui/page.tsx,admin-nextjs/app/[project]/ui/plugins/[plugin]/page.tsx admin-nextjs/app/[project]/ui/plugins/page.tsx,admin-nextjs/app/[project]/ui/routes/page.tsx admin-nextjs/app/[project]/ui/system/page.tsx,admin-nextjs/app/[project]/ui/workers/page.tsx admin-nextjs/app/audit/page.tsx,admin-nextjs/app/components/admin-chrome.tsx admin-nextjs/app/components/admin-shell.tsx,admin-nextjs/app/components/legacy-admin-page.tsx admin-nextjs/app/components/plugin-catalog-panel.tsx,admin-nextjs/app/database/page.tsx admin-nextjs/app/extensions/[project]/[plugin]/[version]/[page]/view.tsx,admin-nextjs/app/files/page.tsx admin-nextjs/app/functions/[name]/view.tsx,admin-nextjs/app/functions/new/view.tsx admin-nextjs/app/functions/page.tsx,admin-nextjs/app/functions/view.tsx admin-nextjs/app/lib/admin-api.ts,admin-nextjs/app/lib/admin-menu.ts admin-nextjs/app/lib/admin-registry.ts,admin-nextjs/app/live-sessions/page.tsx admin-nextjs/app/monitoring/page.tsx,admin-nextjs/app/plugins/[plugin]/view.tsx admin-nextjs/app/plugins/page.tsx,admin-nextjs/app/projects/page.tsx admin-nextjs/app/routes/page.tsx,admin-nextjs/app/system/page.tsx admin-nextjs/app/view.tsx,admin-nextjs/app/workers/view.tsx admin-nextjs/app/workflows/page.tsx,admin-nextjs/next.config.ts internal/plugin/manifest_http.go
- `ef4f2644` — 2026-11-27 19:32:37 +0330 — Verify project slug function URLs and document canonical route discovery
  Files: docs/admin-pages/routes.md,docs/http-api.md tests/admin/admin-api.test.ts,tests/app/function_http_coverage_test.go
- `3c68702e` — 2026-11-27 19:32:02 +0330 — Expose canonical project-scoped routes and migrate remaining admin HTTP callers
  Files: admin-nextjs/app/components/admin-config-form.tsx,admin-nextjs/app/components/routes-panel.tsx admin-nextjs/app/functions/[name]/page.tsx,admin-nextjs/app/workers/page.tsx internal/app/project_http.go,tests/app/project_plugin_http_test.go
- `ceef22f3` — 2026-11-27 15:57:58 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `05ba5ce0` — 2026-11-27 19:26:14 +0330 — Merge pull request #38 from arashrasoulzadeh/feat/project-scoped-http-and-plugin-routes
  Files: 
- `173da987` — 2026-11-27 19:25:20 +0330 — Exercise authenticated Arena admission and invalid durable snapshots
  Files: tests/arena/arena-routes.test.ts
- `71fd24bd` — 2026-11-27 19:24:23 +0330 — Format WebSocket gateway table test
  Files: tests/plugin/websocket_gateway_test.go
- `0046b6b9` — 2026-11-27 19:24:18 +0330 — Cover project namespace configuration and persistence on memory and SQLite stores
  Files: tests/project/url_namespace_test.go
- `97834de2` — 2026-11-27 19:23:55 +0330 — Cover WebSocket gateway boundary and rejected socket messages
  Files: tests/plugin/websocket_gateway_test.go
- `d249e585` — 2026-11-27 19:23:28 +0330 — Cover WebSocket upgrade tokens and conflicting project identity
  Files: tests/plugin/websocket_endpoint_test.go
- `ca5405fe` — 2026-11-27 19:23:21 +0330 — Require an exact Connection upgrade token for WebSocket admission
  Files: internal/plugin/websocket_endpoint.go
- `5995ee43` — 2026-11-27 19:21:47 +0330 — Test project URL validation and reject invalid admin namespaces before transport
  Files: tests/client/admin-client.test.ts
- `9dd994b1` — 2026-11-27 19:18:37 +0330 — Exercise actual WebSocket upgrades, membership, replay and durable actor binding
  Files: tests/plugin/websocket_gateway_test.go
- `69a841a0` — 2026-11-27 19:18:04 +0330 — Vendor pinned coder WebSocket transport for offline reproducible builds
  Files: vendor/github.com/coder/websocket/LICENSE.txt,vendor/github.com/coder/websocket/README.md vendor/github.com/coder/websocket/accept.go,vendor/github.com/coder/websocket/close.go vendor/github.com/coder/websocket/compress.go,vendor/github.com/coder/websocket/conn.go vendor/github.com/coder/websocket/dial.go,vendor/github.com/coder/websocket/doc.go vendor/github.com/coder/websocket/errors.go,vendor/github.com/coder/websocket/frame.go vendor/github.com/coder/websocket/hijack.go,vendor/github.com/coder/websocket/internal/bpool/bpool.go vendor/github.com/coder/websocket/internal/errd/wrap.go,vendor/github.com/coder/websocket/internal/util/util.go vendor/github.com/coder/websocket/internal/wsjs/wsjs_js.go,vendor/github.com/coder/websocket/internal/xsync/go.go vendor/github.com/coder/websocket/mask.go,vendor/github.com/coder/websocket/mask_amd64.s vendor/github.com/coder/websocket/mask_arm64.s,vendor/github.com/coder/websocket/mask_asm.go vendor/github.com/coder/websocket/mask_go.go,vendor/github.com/coder/websocket/netconn.go vendor/github.com/coder/websocket/netconn_js.go,vendor/github.com/coder/websocket/netconn_notjs.go vendor/github.com/coder/websocket/read.go,vendor/github.com/coder/websocket/stringer.go vendor/github.com/coder/websocket/write.go,vendor/github.com/coder/websocket/ws_js.go vendor/modules.txt
- `1abba9c8` — 2026-11-27 19:16:11 +0330 — Implement authenticated arena membership admission and durable player inputs
  Files: docs/plugin-guides/arena.md,internal/plugin/legacy_arena_storage.go internal/plugin/websocket_gateway.go,plugins/arena/index.ts plugins/arena/plugin.yaml,plugins/arena/state-machine.ts plugins/sdk-typescript/src/manifest-operations.generated.d.ts,tests/arena/arena-routes.test.ts
- `558af223` — 2026-11-27 19:14:58 +0330 — Wire manifest-opted authenticated WebSocket gateway and bounded transport
  Files: docs/http-api.md,go.mod go.sum,internal/admin/lifecycle.go internal/app/api_response.go,internal/app/app.go internal/plugin/http_routes.go,internal/plugin/manifest.go internal/plugin/manifest_http.go,internal/plugin/websocket_gateway.go internal/plugin/websocket_ticket.go
- `8d82d580` — 2026-11-27 19:00:28 +0330 — Add bounded single-use WebSocket transport ticket binding
  Files: docs/http-api.md,internal/plugin/websocket_ticket.go tests/plugin/websocket_ticket_test.go
- `5efdf699` — 2026-11-27 18:55:17 +0330 — Update admin transport fixture and format routing regression test
  Files: tests/coverage/plugin-boundaries.test.ts,tests/plugin/http_version_routing_test.go
- `86e71f47` — 2026-11-27 18:54:29 +0330 — Revalidate manifest HTTP route selection at invocation and commit
  Files: CHANGELOG.md,internal/admin/lifecycle.go internal/plugin/http_routes.go,internal/plugin/model_runtime.go internal/plugin/registry_workers.go,internal/plugin/worker.go tests/plugin/http_version_routing_test.go
- `79a83e02` — 2026-11-27 18:52:06 +0330 — Enforce project command namespace and reserve global lifecycle operations
  Files: admin-nextjs/app/page.tsx,admin-nextjs/app/plugins/[plugin]/page.tsx clients/javascript/src/admin.ts,docs/http-api.md internal/app/admin.go,internal/app/app.go scripts/test-admin-browser.mjs,scripts/test-admin-live.mjs scripts/test-runtime-api.mjs,tests/app/admin_namespace_test.go tests/app/admin_test.go,tests/client/admin-client.test.ts
- `f15998af` — 2026-11-27 15:19:36 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `6e6115eb` — 2026-11-27 18:49:07 +0330 — Integrate merged manifest SDK improvements into project routing
  Files: docs/sdk-reference.md
- `c6f51829` — 2026-11-27 18:48:32 +0330 — Merge pull request #39 from arashrasoulzadeh/refactor/manifest-driven-runtime
  Files: 
- `07f5ee22` — 2026-11-27 18:44:19 +0330 — Use atomic CLI coverage mode consistently
  Files: cmd/manifest-sdk/main_test.go
