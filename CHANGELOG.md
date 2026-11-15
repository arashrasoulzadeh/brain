# Changelog

Automatically generated from commits on main.

## Unreleased
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
- `4f80ea3` — 2026-11-15 17:16:17 +0330 — add sqlite storage bridge for table routes
  Files: cmd/brain/main.go,internal/plugin/plugin.go internal/plugin/sqlite_storage.go,plugins/tables/tables-read.ts
- `95ffeb8` — 2026-11-15 13:44:55 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `84b45ea` — 2026-11-15 17:14:41 +0330 — make analytics capture host invokable
  Files: plugins/analytics/analytics-route-sink.ts,plugins/analytics/capture.ts plugins/analytics/index.ts,plugins/analytics/plugin.yaml
- `2b3c5a0` — 2026-11-15 13:44:09 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `cfdca95` — 2026-11-15 17:13:55 +0330 — implement function admin routes
  Files: plugins/functions/index.ts,plugins/functions/invoke-route.ts plugins/functions/plugin.yaml,plugins/functions/publish-route.ts
- `da5ea09` — 2026-11-15 13:43:38 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `e556070` — 2026-11-15 17:13:23 +0330 — implement room admin routes
  Files: plugins/rooms/index.ts,plugins/rooms/plugin.yaml plugins/rooms/room-route-store.ts,plugins/rooms/rooms-create-route.ts plugins/rooms/rooms-join-route.ts,plugins/rooms/rooms-reconnect-route.ts plugins/rooms/rooms-result-route.ts
- `210ad10` — 2026-11-15 13:42:59 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `c32d1c7` — 2026-11-15 17:12:45 +0330 — implement matchmaking admin routes
  Files: plugins/matchmaking/index.ts,plugins/matchmaking/matchmaking-assign-route.ts plugins/matchmaking/matchmaking-cancel-route.ts,plugins/matchmaking/matchmaking-queue-route.ts plugins/matchmaking/matchmaking-route-store.ts,plugins/matchmaking/matchmaking-status-route.ts plugins/matchmaking/plugin.yaml
- `7abd628` — 2026-11-15 13:41:57 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `70eb9f3` — 2026-11-15 17:11:36 +0330 — add workflow and live session consoles
  Files: admin-nextjs/app/page.tsx
- `42e9e40` — 2026-11-15 13:40:50 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `6c3dc6b` — 2026-11-15 17:10:36 +0330 — add files operation console
  Files: admin-nextjs/app/page.tsx
- `966ce9b` — 2026-11-15 13:40:01 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `094f8e4` — 2026-11-15 17:09:47 +0330 — implement file admin routes
  Files: plugins/files/file-route-store.ts,plugins/files/files-download.ts plugins/files/files-metadata.ts,plugins/files/files-publish.ts plugins/files/files-upload.ts,plugins/files/index.ts plugins/files/plugin.yaml,tests/files/files.test.ts
- `e581f88` — 2026-11-15 13:38:37 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `7f314b3` — 2026-11-15 17:07:48 +0330 — persist table operations within plugin worker
  Files: plugins/tables/index.ts,plugins/tables/plugin.yaml plugins/tables/table-store.ts,plugins/tables/tables-read.ts plugins/tables/tables-schema.ts,plugins/tables/tables-write.ts tests/tables/tables.test.ts
- `0d467e1` — 2026-11-15 13:36:48 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
