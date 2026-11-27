# Changelog

Automatically generated from commits on main.

## Unreleased
- `c70d1bda` — 2026-11-28 00:36:14 +0330 — Merge pull request #50 from arashrasoulzadeh/admin/functions-catalog-rework
  Files: 
- `a81a7997` — 2026-11-28 00:20:16 +0330 — Rebuild project Functions catalog and preserve access boundaries
  Files: admin-nextjs/app/components/admin-chrome.tsx,admin-nextjs/app/functions/view.tsx admin-nextjs/app/lib/admin-registry.ts,admin-nextjs/app/lib/function-inspection.ts admin-nextjs/app/theme.css,docs/admin-pages/functions.md tests/admin-ui/pages.test.mjs,tests/admin/function-inspection.test.ts
- `2d02d97e` — 2026-11-27 20:36:02 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `43e57815` — 2026-11-28 00:05:48 +0330 — Merge PR #48: Expand admin operations
  Files: 
- `bda81846` — 2026-11-27 22:55:27 +0330 — Test admin storage page ceilings, malformed snapshots and actor boundaries
  Files: internal/plugin/admin_storage_boundary_test.go
- `2e27284a` — 2026-11-27 22:53:41 +0330 — Correct boundary test formatting and inactive-project fixture
  Files: internal/app/project_http_boundary_test.go
- `8866c01d` — 2026-11-27 22:52:15 +0330 — test: cover project HTTP identity binding, malformed requests and worker response boundaries
  Files: internal/app/project_http.go,internal/app/project_http_boundary_test.go
- `1d38948b` — 2026-11-27 22:50:08 +0330 — test: cover account transaction rollback, stale sign-in and managed inventory search boundaries
  Files: internal/plugin/auth_admin_failure_test.go,internal/plugin/confirmed_install_boundary_test.go tests/plugin/admin_inventory_search_policy_test.go
- `72fbe27e` — 2026-11-27 22:48:38 +0330 — fix: make room closure and membership mutations compare-and-swap safe
  Files: docs/plugin-guides/rooms.md,internal/plugin/legacy_room_admin_test.go internal/plugin/legacy_storage_adapter.go
- `aecf6446` — 2026-11-27 22:47:23 +0330 — test: cover confirmed install admission, dependency and persistence failure boundaries
  Files: internal/plugin/confirmed_install_boundary_test.go
- `86772c54` — 2026-11-27 22:45:40 +0330 — Exercise shared admin shell, dashboard and revoked extension access
  Files: admin-nextjs/app/components/admin-chrome.tsx,admin-nextjs/app/extensions/[project]/[plugin]/[version]/[page]/view.tsx tests/admin-ui/pages.test.mjs,tests/admin-ui/workspaces.test.mjs
- `667120e9` — 2026-11-27 22:44:08 +0330 — Cover schema migration errors and fix zero-time analytics bounds
  Files: internal/plugin/legacy_analytics_timeline.go,internal/plugin/legacy_analytics_timeline_test.go internal/plugin/legacy_table_schema_test.go
- `0bc3ec14` — 2026-11-27 22:42:53 +0330 — Test binary file streaming, rollback and authenticated proxy boundaries
  Files: admin-nextjs/app/lib/plugin-version-review.ts,internal/plugin/file_binary_storage_test.go scripts/test-runtime-api.mjs,tests/admin-ui/file-proxy.test.mjs tests/admin-ui/pages.test.mjs
- `73dbd881` — 2026-11-27 22:40:25 +0330 — ci: collect genuine V8 coverage from packaged runtime and browser integration tests
  Files: .github/workflows/admin-browser.yml,.github/workflows/coverage.yml docs/admin-pages/README.md,internal/plugin/registry_versions.go
- `eafc2b98` — 2026-11-27 22:39:12 +0330 — test: cover staged plugin upgrades and exact-schema and room authorization in admin runtime
  Files: docs/http-api.md,tests/admin-ui/pages.test.mjs tests/admin/admin-plugin-operations.test.ts
- `ef840672` — 2026-11-27 22:37:41 +0330 — feat: staged plugin version management, dependency review and catalog health filters
  Files: admin-nextjs/app/components/plugin-catalog-panel.tsx,admin-nextjs/app/components/plugin-install-dialog.tsx admin-nextjs/app/lib/plugin-version-review.ts,admin-nextjs/app/plugins/[plugin]/view.tsx docs/admin-pages/plugin-detail.md,tests/admin/plugin-version-review.test.ts
- `4d15b090` — 2026-11-27 22:36:39 +0330 — fix: persist exact promoted plugin versions and explicit upgrade capability approvals
  Files: admin-nextjs/app/lib/admin-types.ts,docs/admin-pages/plugin-detail.md internal/admin/lifecycle.go,internal/app/admin.go internal/plugin/plugin.go,internal/plugin/registry_catalog.go internal/plugin/registry_permissions.go,internal/plugin/registry_storage.go internal/plugin/registry_versions.go,tests/plugin/version_persistence_test.go
- `ad7402d9` — 2026-11-27 22:34:38 +0330 — feat: time-filtered analytics history and bounded UTC event timelines
  Files: docs/plugin-guides/analytics.md,internal/plugin/legacy_analytics_read.go internal/plugin/legacy_analytics_timeline.go,internal/plugin/legacy_analytics_timeline_test.go internal/plugin/legacy_storage_adapter.go,internal/plugin/legacy_table_admin.go plugins/analytics/plugin.yaml,plugins/sdk-typescript/src/manifest-operations.generated.d.ts scripts/test-runtime-api.mjs
- `48371786` — 2026-11-27 22:33:54 +0330 — feat: inspect observed database schemas and apply revision-checked bulk field migrations
  Files: docs/admin-pages/database.md,docs/plugin-guides/tables.md internal/plugin/legacy_storage_adapter.go,internal/plugin/legacy_table_admin.go internal/plugin/legacy_table_schema.go,internal/plugin/legacy_table_schema_test.go plugins/sdk-typescript/src/manifest-operations.generated.d.ts,plugins/tables/index.ts plugins/tables/plugin.yaml,plugins/tables/routes.ts scripts/test-runtime-api.mjs
- `ff7734dc` — 2026-11-27 22:31:16 +0330 — test: exercise arena interruption authorization and game rule metadata
  Files: docs/admin-pages/live-sessions.md,tests/arena/arena-routes.test.ts tests/games/games.test.ts
- `c1833538` — 2026-11-27 22:30:41 +0330 — feat: arena interruption, match status inventory and game rule inspection
  Files: docs/plugin-guides/arena.md,docs/plugin-guides/games.md internal/plugin/legacy_admin_inventory.go,internal/plugin/legacy_arena_admin_test.go internal/plugin/legacy_arena_storage.go,internal/plugin/legacy_table_admin.go plugins/arena/index.ts,plugins/arena/plugin.yaml plugins/arena/state-machine.ts,plugins/games/index.ts plugins/games/plugin.yaml,plugins/sdk-typescript/src/manifest-operations.generated.d.ts scripts/test-runtime-api.mjs
- `d5951f5d` — 2026-11-27 22:27:50 +0330 — test: exercise account suspension and room closure through the packaged admin runtime
  Files: docs/admin-pages/live-sessions.md,internal/plugin/legacy_auth_storage.go internal/plugin/legacy_storage_adapter.go,scripts/test-runtime-api.mjs tests/auth/durable-dispatch.test.ts
- `3df8c96c` — 2026-11-27 22:27:14 +0330 — feat: durable room administration and authorized live-session inventory metadata
  Files: admin-nextjs/app/components/admin-inventory.tsx,admin-nextjs/app/lib/admin-types.ts docs/plugin-guides/matchmaking.md,docs/plugin-guides/rooms.md internal/plugin/admin_inventory.go,internal/plugin/legacy_admin_inventory.go internal/plugin/legacy_room_admin_test.go,internal/plugin/legacy_storage_adapter.go internal/plugin/legacy_table_admin.go,plugins/rooms/index.ts plugins/rooms/plugin.yaml,plugins/rooms/routes.ts plugins/sdk-typescript/src/manifest-operations.generated.d.ts
- `bab0f0a9` — 2026-11-27 22:26:20 +0330 — feat: account search, redacted session inspection and atomic account suspension
  Files: docs/plugin-guides/auth.md,internal/plugin/legacy_auth_admin.go internal/plugin/legacy_auth_admin_test.go,internal/plugin/legacy_auth_storage.go plugins/auth/plugin.yaml,plugins/auth/service.ts plugins/sdk-typescript/src/manifest-operations.generated.d.ts
- `548138ce` — 2026-11-27 22:23:31 +0330 — test: cover route filtering and invalid payloads in browser and function search in DOM
  Files: admin-nextjs/app/functions/function-editor.tsx,admin-nextjs/app/view.tsx scripts/test-admin-browser.mjs,tests/admin-ui/pages.test.mjs
- `2667b382` — 2026-11-27 22:23:05 +0330 — feat: improve function source state, catalog filters and execution log inspection
  Files: admin-nextjs/app/functions/[name]/view.tsx,admin-nextjs/app/functions/function-editor.tsx admin-nextjs/app/functions/view.tsx,admin-nextjs/app/lib/function-inspection.ts docs/admin-pages/function-detail.md,docs/admin-pages/functions.md tests/admin/function-inspection.test.ts
- `a1a0abe7` — 2026-11-27 22:21:53 +0330 — feat: searchable route discovery and immutable confirmed playground requests
  Files: admin-nextjs/app/components/routes-panel.tsx,admin-nextjs/app/lib/admin-playground.ts admin-nextjs/app/lib/project-requests.ts,admin-nextjs/app/view.tsx docs/admin-pages/routes.md,tests/admin/admin-playground.test.ts tests/admin/project-requests.test.ts
- `c6dd55cf` — 2026-11-27 22:05:16 +0330 — test: use namespaced CLI health endpoints
  Files: tests/cli/test_cli.py
- `8e80dda5` — 2026-11-27 22:04:44 +0330 — test: exercise database and analytics admin operations in packaged runtime and repair browser fixtures
  Files: docs/admin-pages/README.md,scripts/test-runtime-api.mjs tests/admin-ui/interactive.test.mjs,tests/browser/admin.test.mjs
- `abff53c3` — 2026-11-27 22:03:45 +0330 — feat: bounded database row pages and revision-checked row replacement and deletion
  Files: docs/admin-pages/database.md,docs/plugin-guides/tables.md internal/plugin/legacy_analytics_read.go,internal/plugin/legacy_storage_adapter.go internal/plugin/legacy_table_admin.go,internal/plugin/legacy_table_admin_test.go internal/plugin/registry_workers.go,plugins/sdk-typescript/src/manifest-operations.generated.d.ts plugins/tables/index.ts,plugins/tables/plugin.yaml plugins/tables/routes.ts,tests/admin/lifecycle_read_test.go tests/tables/admin-rows.test.ts
- `fc7d0298` — 2026-11-27 22:01:26 +0330 — feat: search authorized resource inventories across server-side pages
  Files: admin-nextjs/app/components/admin-inventory.tsx,admin-nextjs/app/lib/admin-sdk.ts docs/admin-pages/extensions.md,docs/http-api.md internal/admin/lifecycle.go,internal/app/admin.go internal/plugin/admin_inventory.go,internal/plugin/admin_inventory_test.go internal/plugin/legacy_admin_inventory.go,tests/admin-ui/api.test.mjs
- `55598dda` — 2026-11-27 22:00:44 +0330 — test: reconcile admin coverage fixtures with scoped routes and confirmation dialogs
  Files: tests/admin-ui/api.test.mjs,tests/admin-ui/components.test.mjs tests/admin-ui/dom-harness.mjs,tests/admin-ui/interactive.test.mjs tests/admin-ui/main-page.test.mjs,tests/admin-ui/pages.test.mjs tests/admin-ui/registry.test.mjs,tests/app/function_http_test.go tests/app/http_inventory_test.go
- `9d4e692b` — 2026-11-27 21:58:35 +0330 — Fix analytics fixture context and Go formatting
  Files: internal/plugin/legacy_analytics_read.go,internal/plugin/legacy_analytics_read_test.go internal/plugin/plugin.go,tests/analytics/coverage-edges.test.ts
- `9adbe19d` — 2026-11-27 21:56:31 +0330 — feat: expose project-scoped analytics history and event counts
  Files: docs/admin-pages/README.md,docs/plugin-guides/analytics.md internal/plugin/legacy_analytics_read.go,internal/plugin/legacy_analytics_read_test.go internal/plugin/legacy_storage_adapter.go,plugins/analytics/index.ts plugins/analytics/plugin.yaml,plugins/sdk-typescript/src/manifest-operations.generated.d.ts tests/analytics/coverage-edges.test.ts
- `461bd31a` — 2026-11-27 21:54:45 +0330 — feat: expose resource grant scopes and advanced admin lifecycle controls
  Files: admin-nextjs/app/dashboard/view.tsx,admin-nextjs/app/lib/admin-types.ts admin-nextjs/app/plugins/[plugin]/view.tsx,admin-nextjs/app/workers/view.tsx docs/admin-pages/dashboard.md,docs/admin-pages/plugin-detail.md internal/plugin/plugin.go,internal/plugin/registry_catalog.go tests/admin/lifecycle_read_test.go
- `54f4e4ff` — 2026-11-27 21:53:32 +0330 — feat: improve admin audit, inventory, worker health and request isolation
  Files: admin-nextjs/app/components/admin-inventory.tsx,admin-nextjs/app/components/audit-panel.tsx admin-nextjs/app/components/monitoring-panel.tsx,admin-nextjs/app/lib/admin-observability.ts admin-nextjs/app/lib/admin-types.ts,admin-nextjs/app/view.tsx admin-nextjs/app/workers/view.tsx,docs/admin-pages/audit.md docs/admin-pages/extensions.md,docs/admin-pages/monitoring.md docs/admin-pages/workers.md,internal/admin/lifecycle.go internal/operator/operator.go,internal/plugin/registry_audit.go tests/admin/admin-observability.test.ts,tests/admin/lifecycle_read_test.go
- `dcec131f` — 2026-11-27 18:20:51 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `43d7023b` — 2026-11-27 21:48:36 +0330 — Merge pull request #14 from arashrasoulzadeh/quality/coverage-completion
  Files: 
- `e462a13e` — 2026-11-27 21:48:25 +0330 — Merge main into coverage completion and retain existing CI checks
  Files: .github/workflows/test.yml
- `efdc3992` — 2026-11-27 21:48:05 +0330 — Merge pull request #47 from arashrasoulzadeh/fix/admin-page-consistency
  Files: 
- `46fd8ff0` — 2026-11-27 21:47:53 +0330 — Merge main into admin page consistency and preserve file MB settings
  Files: admin-nextjs/app/components/admin-config-form.tsx,admin-nextjs/app/components/routes-panel.tsx
- `97aee320` — 2026-11-27 18:07:54 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `2aff993c` — 2026-11-27 21:37:44 +0330 — Merge pull request #45 from arashrasoulzadeh/feat/files-megabyte-limit
  Files: 
- `3865c2ba` — 2026-11-27 21:21:23 +0330 — Cover admin access, route failures and cancelled project edits in browser regression tests
  Files: admin-nextjs/app/view.tsx,docs/admin-pages/function-detail.md docs/admin-pages/projects.md,docs/admin-pages/routes.md scripts/test-admin-browser.mjs
- `ac357976` — 2026-11-27 21:20:25 +0330 — Isolate live admin server and allow bounded binary fixtures through CI relay
  Files: .github/workflows/admin-browser.yml,scripts/admin-ci-backend-relay.mjs scripts/test-admin-live.mjs
- `20cc1a03` — 2026-11-27 21:18:30 +0330 — Improve admin page access, lifecycle dialogs, route invocation and function editor consistency
  Files: admin-nextjs/app/components/admin-chrome.tsx,admin-nextjs/app/components/admin-config-form.tsx admin-nextjs/app/components/audit-panel.tsx,admin-nextjs/app/components/legacy-admin-page.tsx admin-nextjs/app/components/monitoring-panel.tsx,admin-nextjs/app/components/plugin-catalog-panel.tsx admin-nextjs/app/components/routes-panel.tsx,admin-nextjs/app/extensions/[project]/[plugin]/[version]/[page]/view.tsx admin-nextjs/app/functions/[name]/view.tsx,admin-nextjs/app/functions/function-editor.tsx admin-nextjs/app/functions/new/view.tsx,admin-nextjs/app/functions/view.tsx admin-nextjs/app/view.tsx,admin-nextjs/app/workers/view.tsx
- `699e98c7` — 2026-11-27 17:48:05 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `98e11603` — 2026-11-27 21:15:32 +0330 — Merge main and preserve confirmed installation, file transport and combined regression coverage
  Files: admin-nextjs/app/components/admin-config-form.tsx,docs/http-api.md scripts/test-admin-live.mjs
- `557658d8` — 2026-11-27 21:15:03 +0330 — Merge pull request #46 from arashrasoulzadeh/feat/confirmed-plugin-install
  Files: 
- `644bfc1e` — 2026-11-27 21:13:51 +0330 — Document binary routes and verify streaming through the admin host above proxy buffer limits
  Files: admin-nextjs/app/[project]/admin/files/blob/route.ts,admin-nextjs/app/components/admin-file-controls.tsx admin-nextjs/app/lib/admin-file-transfer.ts,docs/http-api.md internal/app/admin_files.go,scripts/test-admin-live.mjs
