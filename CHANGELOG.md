# Changelog

Automatically generated from commits on main.

## Unreleased
- `782abbb` — 2026-11-15 19:49:42 +0330 — show precise function publish validation errors
  Files: admin-nextjs/app/functions/[name]/page.tsx,internal/admin/functions.go
- `fd03fa3` — 2026-11-15 16:18:15 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `59c25e6` — 2026-11-15 19:47:58 +0330 — show function publish status beside actions
  Files: admin-nextjs/app/functions/[name]/page.tsx
- `0c5e222` — 2026-11-15 16:17:07 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `3ae1cd1` — 2026-11-15 19:46:50 +0330 — add explicit function draft save action
  Files: admin-nextjs/app/functions/[name]/page.tsx
- `b19933b` — 2026-11-15 16:14:52 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `fd5fc6b` — 2026-11-15 19:44:37 +0330 — persist admin shell across page navigation
  Files: admin-nextjs/app/components/admin-chrome.tsx,admin-nextjs/app/components/admin-shell.tsx admin-nextjs/app/functions/[name]/page.tsx,admin-nextjs/app/layout.tsx admin-nextjs/app/page.tsx,tests/admin/admin-shell.test.ts
- `7d37674` — 2026-11-15 16:10:52 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `fa169ca` — 2026-11-15 19:40:38 +0330 — keep audit navigation consistent across pages
  Files: admin-nextjs/app/components/admin-chrome.tsx,admin-nextjs/app/page.tsx tests/admin/admin-shell.test.ts
- `017fae2` — 2026-11-15 16:09:19 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `c46295c` — 2026-11-15 19:39:04 +0330 — test shared admin shell across pages
  Files: tests/admin/admin-shell.test.ts
- `9a06b67` — 2026-11-15 16:06:12 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `969581e` — 2026-11-15 19:35:56 +0330 — document admin pages and shared shell
  Files: docs/admin-pages/README.md,docs/admin-pages/audit.md docs/admin-pages/database.md,docs/admin-pages/files.md docs/admin-pages/function-detail.md,docs/admin-pages/function-new.md docs/admin-pages/functions.md,docs/admin-pages/live-sessions.md docs/admin-pages/monitoring.md,docs/admin-pages/plugin-detail.md docs/admin-pages/plugins.md,docs/admin-pages/projects.md docs/admin-pages/routes.md,docs/admin-pages/system.md docs/admin-pages/workers.md,docs/admin-pages/workflows.md
- `fe9c9f7` — 2026-11-15 16:04:47 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `e1e28f6` — 2026-11-15 19:34:27 +0330 — unify admin sidebar and persist plugin lifecycle
  Files: admin-nextjs/app/functions/[name]/page.tsx,internal/plugin/plugin.go internal/plugin/sqlite_storage.go,tests/plugin/lifecycle_persistence_test.go
- `e761fbc` — 2026-11-15 15:59:14 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `f91cdd3` — 2026-11-15 19:28:51 +0330 — explain public function method mismatch
  Files: internal/admin/functions.go,internal/app/admin.go tests/admin/functions_test.go
- `8a9aafb` — 2026-11-15 15:57:06 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `76c9c49` — 2026-11-15 19:26:48 +0330 — remove editable method control from function detail
  Files: admin-nextjs/app/functions/[name]/page.tsx
- `2d154c1` — 2026-11-15 15:56:02 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `1b9b96f` — 2026-11-15 19:25:41 +0330 — complete public function routes and source methods
  Files: admin-nextjs/app/functions/[name]/page.tsx,admin-nextjs/app/functions/new/page.tsx admin-nextjs/app/page.tsx,admin-nextjs/next.config.ts internal/admin/functions.go,internal/app/admin.go tests/admin/functions_test.go
- `c009559` — 2026-11-15 15:48:42 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `a03dc58` — 2026-11-15 19:18:20 +0330 — enforce plugin lifecycle on function APIs
  Files: internal/admin/lifecycle.go,internal/app/admin.go tests/admin/functions_test.go
- `f81fe4b` — 2026-11-15 15:46:54 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `b24f115` — 2026-11-15 19:16:22 +0330 — guard plugin pages by lifecycle state
  Files: admin-nextjs/app/components/admin-chrome.tsx,admin-nextjs/app/functions/[name]/page.tsx admin-nextjs/app/functions/new/page.tsx,admin-nextjs/app/functions/page.tsx admin-nextjs/app/page.tsx,admin-nextjs/app/workflows/page.tsx
- `58242df` — 2026-11-15 15:38:12 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `13d6415` — 2026-11-15 19:07:57 +0330 — fix local operator login setup
  Files: admin-nextjs/app/page.tsx,docs/admins.md scripts/start-local.sh
- `1ef4bac` — 2026-11-15 15:35:59 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `4ba054c` — 2026-11-15 19:04:00 +0330 — finish shared admin shell across all pages
  Files: admin-nextjs/app/components/admin-chrome.tsx,admin-nextjs/app/functions/new/page.tsx admin-nextjs/app/functions/page.tsx,admin-nextjs/app/plugins/[plugin]/page.tsx admin-nextjs/app/workers/page.tsx,admin-nextjs/app/workflows/page.tsx
- `9e392cc` — 2026-11-15 15:27:42 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `86f77a3` — 2026-11-15 18:57:17 +0330 — return json for every api response
  Files: internal/app/api_response.go
- `a512f26` — 2026-11-15 18:55:52 +0330 — make cross-cutting changes project-wide law
  Files: AGENTS.md,docs/developers.md
- `2f03e16` — 2026-11-15 15:24:48 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `31ab1e7` — 2026-11-15 18:54:33 +0330 — wire api envelope into server handler
  Files: internal/app/app.go
- `4a9a60b` — 2026-11-15 18:54:14 +0330 — standardize api messages and add admin toasts
  Files: admin-nextjs/app/components/toast-provider.tsx,admin-nextjs/app/layout.tsx internal/app/api_response.go,tests/app/admin_test.go
- `e038b0c` — 2026-11-15 15:22:27 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `f1cef48` — 2026-11-15 18:52:06 +0330 — apply tailwind component layer across admin ui
  Files: admin-nextjs/app/styles.css
- `539f29e` — 2026-11-15 15:19:14 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `c44ecc0` — 2026-11-15 18:48:53 +0330 — fix function publish flow and tailwind controls
  Files: admin-nextjs/app/functions/[name]/page.tsx,admin-nextjs/app/styles.css admin-nextjs/package-lock.json,admin-nextjs/package.json admin-nextjs/postcss.config.mjs
- `60556fc` — 2026-11-15 15:14:38 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `f48bead` — 2026-11-15 18:44:07 +0330 — fix function deployment method and source publishing
  Files: admin-nextjs/.gitignore,admin-nextjs/app/components/admin-chrome.tsx admin-nextjs/app/functions/[name]/page.tsx,admin-nextjs/app/functions/function-editor.tsx admin-nextjs/app/functions/new/page.tsx,admin-nextjs/app/functions/page.tsx admin-nextjs/app/page.tsx,admin-nextjs/app/plugins/[plugin]/page.tsx admin-nextjs/app/styles.css,admin-nextjs/app/workers/page.tsx admin-nextjs/next.config.ts,admin-nextjs/tsconfig.json internal/admin/functions.go,internal/admin/lifecycle.go internal/app/admin.go,internal/app/app.go internal/plugin/plugin.go,plugins/functions/call-route.ts plugins/functions/index.ts,plugins/functions/plugin.yaml tests/admin/functions_test.go,tests/functions/public-call.test.ts
- `7db93a4` — 2026-11-15 13:51:16 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `fa0fc43` — 2026-11-15 17:21:01 +0330 — persist analytics events in sqlite
  Files: internal/plugin/plugin.go
- `49424e3` — 2026-11-15 13:50:48 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `45b3a9a` — 2026-11-15 17:20:32 +0330 — persist room state in sqlite
  Files: internal/plugin/plugin.go,plugins/rooms/rooms-create-route.ts plugins/rooms/rooms-join-route.ts,plugins/rooms/rooms-reconnect-route.ts plugins/rooms/rooms-result-route.ts
- `5d01aa3` — 2026-11-15 13:49:58 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `7860a8d` — 2026-11-15 17:19:43 +0330 — persist matchmaking tickets in sqlite
  Files: internal/plugin/plugin.go,internal/plugin/sqlite_storage.go plugins/matchmaking/matchmaking-assign-route.ts,plugins/matchmaking/matchmaking-cancel-route.ts plugins/matchmaking/matchmaking-status-route.ts
- `3eaa5ae` — 2026-11-15 13:47:51 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `2d57b95` — 2026-11-15 17:17:36 +0330 — persist file routes in sqlite
  Files: internal/plugin/plugin.go,internal/plugin/sqlite_storage.go plugins/files/files-download.ts,plugins/files/files-metadata.ts plugins/files/files-publish.ts
- `02df869` — 2026-11-15 13:46:31 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
