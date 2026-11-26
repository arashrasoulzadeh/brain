# Changelog

Automatically generated from commits on main.

## Unreleased
- `7224a1a` — 2026-11-27 02:01:39 +0330 — fix(ci): align sandbox fixture with worker route contract
  Files: scripts/test-sandboxed-typescript-worker.sh
- `9036e93` — 2026-11-16 12:40:59 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `24bf1f1` — 2026-11-16 15:39:15 +0330 — silence error toasts from background function-draft polling and add CLAUDE.md
  Files: CLAUDE.md,admin-nextjs/app/components/toast-provider.tsx admin-nextjs/app/page.tsx
- `222dd25` — 2026-11-16 04:00:43 +0330 — show raw function execution records
  Files: admin-nextjs/app/functions/[name]/page.tsx,admin-nextjs/app/styles.css
- `bf4bdc9` — 2026-11-16 04:00:02 +0330 — isolate function editor in Monaco container
  Files: admin-nextjs/app/functions/function-editor.tsx,admin-nextjs/app/styles.css
- `64e5032` — 2026-11-16 03:59:01 +0330 — make Monaco text layer visible
  Files: admin-nextjs/app/functions/function-editor.tsx,admin-nextjs/app/styles.css
- `023651a` — 2026-11-16 03:57:49 +0330 — fit Monaco and identify function executions
  Files: admin-nextjs/app/styles.css,internal/admin/functions.go
- `25064bf` — 2026-11-16 03:54:16 +0330 — fit Monaco editor to function workspace
  Files: admin-nextjs/app/functions/function-editor.tsx,admin-nextjs/app/styles.css
- `90c0225` — 2026-11-16 03:53:26 +0330 — capture real sandbox latency breakdown
  Files: admin-nextjs/app/functions/function-editor.tsx,internal/admin/functions.go
- `af4f73e` — 2026-11-16 03:51:45 +0330 — use Monaco for function code editing
  Files: admin-nextjs/app/functions/function-editor.tsx,admin-nextjs/app/styles.css admin-nextjs/package-lock.json,admin-nextjs/package.json
- `887da3e` — 2026-11-16 03:49:48 +0330 — poll logs without overwriting editor drafts
  Files: admin-nextjs/app/functions/[name]/page.tsx
- `f07da06` — 2026-11-16 03:48:19 +0330 — show compatibility latency for legacy logs
  Files: admin-nextjs/app/functions/[name]/page.tsx
- `fe7902c` — 2026-11-16 03:47:00 +0330 — simplify function editor input
  Files: admin-nextjs/app/functions/[name]/page.tsx,admin-nextjs/app/functions/function-editor.tsx admin-nextjs/app/styles.css
- `cc83f16` — 2026-11-16 03:45:32 +0330 — remove function editor autocomplete
  Files: admin-nextjs/app/functions/function-editor.tsx
- `a6d1f58` — 2026-11-16 03:44:11 +0330 — trace function calls from edge to response
  Files: admin-nextjs/app/functions/function-editor.tsx,internal/admin/functions.go
- `8e3f0f5` — 2026-11-16 03:42:24 +0330 — restore reliable function editor typing
  Files: admin-nextjs/app/functions/function-editor.tsx,admin-nextjs/app/styles.css
- `8c6f950` — 2026-11-16 03:41:28 +0330 — show rich function log details in modal
  Files: admin-nextjs/app/functions/[name]/page.tsx,admin-nextjs/app/styles.css
- `2f0c0c7` — 2026-11-16 03:39:31 +0330 — separate compact log list from details
  Files: admin-nextjs/app/functions/[name]/page.tsx,admin-nextjs/app/styles.css
- `b39facf` — 2026-11-16 03:36:21 +0330 — refresh and paginate function logs
  Files: admin-nextjs/app/functions/[name]/page.tsx,admin-nextjs/app/styles.css
- `811e023` — 2026-11-16 03:33:42 +0330 — record function workflow latency details
  Files: admin-nextjs/app/functions/[name]/page.tsx,admin-nextjs/app/styles.css internal/admin/functions.go
- `c899520` — 2026-11-16 03:23:15 +0330 — fix editor completions and function law checks
  Files: admin-nextjs/app/functions/function-editor.tsx,admin-nextjs/app/styles.css
- `d6ae06f` — 2026-11-16 03:22:27 +0330 — validate function source before publish
  Files: admin-nextjs/app/functions/[name]/page.tsx,admin-nextjs/app/functions/function-editor.tsx admin-nextjs/app/styles.css
- `e0ae963` — 2026-11-16 03:21:20 +0330 — expand function execution logs
  Files: admin-nextjs/app/functions/[name]/page.tsx,admin-nextjs/app/styles.css
- `07beafc` — 2026-11-16 03:19:29 +0330 — add rich function editor experience
  Files: admin-nextjs/app/functions/function-editor.tsx,admin-nextjs/app/styles.css
- `158b524` — 2026-11-16 03:16:13 +0330 — keep login outside persistent admin shell
  Files: admin-nextjs/app/components/admin-shell.tsx,admin-nextjs/app/page.tsx tests/admin/admin-shell.test.ts
- `34415f7` — 2026-11-15 16:19:58 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
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
