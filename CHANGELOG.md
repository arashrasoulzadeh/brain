# Changelog

Automatically generated from commits on main.

## Unreleased
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
- `0f9ea2c` — 2026-11-15 17:06:34 +0330 — implement tables admin routes
  Files: plugins/tables/index.ts,plugins/tables/plugin.yaml plugins/tables/tables-read.ts,plugins/tables/tables-schema.ts plugins/tables/tables-write.ts,tests/tables/tables.test.ts
- `e8b71c8` — 2026-11-15 13:35:51 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `cde7a13` — 2026-11-15 17:05:36 +0330 — add database operation console
  Files: admin-nextjs/app/page.tsx
- `3aa111b` — 2026-11-15 13:33:12 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `31ab0b1` — 2026-11-15 17:02:58 +0330 — report runtime worker health in admin
  Files: admin-nextjs/app/page.tsx,internal/plugin/plugin.go
- `f6f8c50` — 2026-11-15 13:32:03 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `3c63b7b` — 2026-11-15 17:01:43 +0330 — start plugin workers with dependency ordering
  Files: cmd/brain/main.go,internal/plugin/worker.go
- `c2508f8` — 2026-11-15 13:30:42 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `da82034` — 2026-11-15 17:00:29 +0330 — bind local workers when plugins enable
  Files: cmd/brain/main.go,internal/plugin/plugin.go internal/plugin/typescript_worker.go,scripts/typescript-worker.mjs
- `a94e65b` — 2026-11-15 13:28:48 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `dd18ac1` — 2026-11-15 16:58:27 +0330 — execute playground requests through host
  Files: admin-nextjs/app/page.tsx,internal/admin/lifecycle.go internal/app/admin.go,internal/plugin/plugin.go scripts/typescript-worker.mjs,tests/plugin/registry_invoke_test.go
- `65153c6` — 2026-11-15 13:25:05 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `fa151a6` — 2026-11-15 16:54:48 +0330 — route manifest functions through plugin workers
  Files: admin-nextjs/app/styles.css,internal/plugin/typescript_worker.go scripts/typescript-worker.mjs,tests/plugin/typescript-worker.test.ts
- `510f630` — 2026-11-15 11:37:14 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `4ea6c3e` — 2026-11-15 15:06:58 +0330 — show saved function drafts in admin
  Files: admin-nextjs/app/page.tsx
- `43c9061` — 2026-11-15 11:36:09 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `46bc82c` — 2026-11-15 15:05:55 +0330 — load project function drafts in admin
  Files: admin-nextjs/app/page.tsx
- `3192140` — 2026-11-15 11:35:22 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `f70dafa` — 2026-11-15 15:05:05 +0330 — persist custom function drafts locally
  Files: internal/admin/functions.go
- `d861dd5` — 2026-11-15 11:34:08 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `68003df` — 2026-11-15 15:03:50 +0330 — expose plugin capability grant state
  Files: admin-nextjs/app/plugins/[plugin]/page.tsx,internal/plugin/plugin.go
- `08659d6` — 2026-11-15 11:32:59 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `dddbb87` — 2026-11-15 15:02:43 +0330 — run custom functions in bounded vm
  Files: admin-nextjs/app/page.tsx,internal/admin/functions.go internal/app/admin.go
- `65a8d46` — 2026-11-15 11:29:35 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `a11cb82` — 2026-11-15 14:59:20 +0330 — limit custom function draft inputs
  Files: internal/admin/functions.go
- `ec22cd3` — 2026-11-15 11:29:09 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `3216e00` — 2026-11-15 14:58:55 +0330 — bind function drafts to active projects
  Files: internal/admin/functions.go
- `7e21adf` — 2026-11-15 11:28:21 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
