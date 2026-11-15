# Changelog

Automatically generated from commits on main.

## Unreleased
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
- `e6ecbb1` — 2026-11-15 14:58:06 +0330 — audit plugin metadata and route permissions
  Files: tests/plugin/all_manifests_test.go
- `394aa31` — 2026-11-15 11:27:50 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `3a986e9` — 2026-11-15 14:57:35 +0330 — preflight custom function source safely
  Files: internal/admin/functions.go,tests/admin/functions_test.go
- `6d35c59` — 2026-11-15 11:27:02 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `3c615ba` — 2026-11-15 14:56:39 +0330 — add project scoped custom function drafts
  Files: admin-nextjs/app/page.tsx,admin-nextjs/app/plugins/[plugin]/page.tsx admin-nextjs/app/styles.css,internal/admin/functions.go internal/app/admin.go,internal/app/app.go
- `a9bb7f4` — 2026-11-15 11:19:04 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `9a9d3e9` — 2026-11-15 14:47:45 +0330 — match plugin page shell and prevent route flash
  Files: admin-nextjs/app/plugins/[plugin]/page.tsx
- `55142a5` — 2026-11-15 11:15:57 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `96dfb07` — 2026-11-15 14:45:40 +0330 — fix plugin settings build
  Files: admin-nextjs/app/plugins/[plugin]/page.tsx
- `56c0713` — 2026-11-15 11:13:41 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `d2f6b43` — 2026-11-15 14:43:24 +0330 — keep plugin settings inside admin workspace
  Files: admin-nextjs/app/plugins/[plugin]/page.tsx
- `d8af0b4` — 2026-11-15 11:12:38 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `a3601a8` — 2026-11-15 14:42:23 +0330 — simplify installed plugins table
  Files: admin-nextjs/app/page.tsx,admin-nextjs/app/styles.css
- `eeb02a3` — 2026-11-15 11:10:41 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `6a00395` — 2026-11-15 14:40:26 +0330 — fix admin login proxy and centering
  Files: admin-nextjs/app/styles.css,admin-nextjs/next.config.ts
- `d942b3b` — 2026-11-15 11:06:23 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `a5937ac` — 2026-11-15 14:36:08 +0330 — redirect unauthenticated plugin pages to login
  Files: admin-nextjs/app/plugins/[plugin]/page.tsx
- `98ce14a` — 2026-11-15 14:34:48 +0330 — add project scoped plugin settings pages
  Files: admin-nextjs/app/page.tsx,admin-nextjs/app/plugins/[plugin]/page.tsx
- `6360d74` — 2026-11-15 11:03:28 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `1ab7ef7` — 2026-11-15 14:33:12 +0330 — show readiness controls for installed plugins
  Files: admin-nextjs/app/page.tsx
- `335b2d2` — 2026-11-15 11:01:29 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `ada9840` — 2026-11-15 14:31:13 +0330 — expose plugin readiness controls
  Files: admin-nextjs/app/page.tsx
- `b391400` — 2026-11-15 10:49:02 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `eb9bea9` — 2026-11-15 14:18:45 +0330 — fix route playground and monitoring states
  Files: admin-nextjs/app/page.tsx,admin-nextjs/app/styles.css
