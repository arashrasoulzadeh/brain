# Changelog

Automatically generated from commits on main.

## Unreleased
- `e30c0038` — 2026-11-28 02:02:57 +0330 — Merge PR #54: distributed execution foundations
  Files: 
- `3aef94ac` — 2026-11-28 01:27:32 +0330 — docs: record distributed execution integration and verification blockers
  Files: docs/unified-execution-architecture.md
- `41270427` — 2026-11-28 01:26:56 +0330 — test: exercise Cortex Synapse and Neuron pipeline together
  Files: tests/distributed/pipeline_test.go
- `10f5fee9` — 2026-11-28 01:25:27 +0330 — feat(neuron): add verified worker-local execution engine
  Files: internal/neuron/engine.go,tests/neuron/engine_test.go
- `31da1f08` — 2026-11-28 01:23:39 +0330 — feat(synapse): stream and verify lease-bound artifact chunks
  Files: internal/synapse/artifact.go,tests/synapse/artifact_test.go
- `4f89cc6e` — 2026-11-28 01:21:21 +0330 — feat(synapse): require mutually authenticated TLS and stabilize queue selection
  Files: internal/cortex/scheduler.go,internal/synapse/tls.go tests/synapse/tls_test.go
- `e0cf8894` — 2026-11-28 01:19:27 +0330 — feat(synapse): define v1 protocol and fence execution assignments
  Files: internal/synapse/session.go,proto/synapse/v1/synapse.proto tests/synapse/session_test.go
- `74a5d05a` — 2026-11-28 01:15:50 +0330 — feat(cortex): schedule compatible neuron sessions with bounded queues and quotas
  Files: internal/cortex/scheduler.go,tests/cortex/scheduler_test.go
- `91309f63` — 2026-11-28 01:11:32 +0330 — docs: record execution implementation evidence and baseline coverage blockers
  Files: docs/unified-execution-architecture.md
- `6d563193` — 2026-11-28 01:11:04 +0330 — feat(execution): validate pinned assignments and reject cancelled results
  Files: internal/execution/dispatch.go,tests/execution/dispatch_test.go
- `20143421` — 2026-11-28 01:08:46 +0330 — feat(execution): pin artifact versions in a bounded scoped catalog
  Files: internal/execution/catalog.go,internal/execution/contract.go tests/execution/catalog_test.go
- `87060dfe` — 2026-11-28 01:06:49 +0330 — feat(execution): add canonical bounded artifact bundles
  Files: internal/execution/artifact.go,internal/execution/contract.go tests/execution/artifact_test.go
- `d2494253` — 2026-11-28 01:04:38 +0330 — feat(execution): introduce validated invocation contract and artifact digest checks
  Files: internal/execution/contract.go,tests/execution/contract_test.go
- `6fef5845` — 2026-11-28 01:02:11 +0330 — docs: audit unified execution architecture and plan migration
  Files: docs/unified-execution-architecture.md
- `5015ce8a` — 2026-11-27 21:22:54 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `d781058c` — 2026-11-28 00:50:52 +0330 — Merge pull request #53 from arashrasoulzadeh/functions/host-plugin-read-bridge
  Files: 
- `c94cbfbb` — 2026-11-28 00:50:48 +0330 — Merge pull request #52 from arashrasoulzadeh/admin/confirmed-dependency-enable
  Files: 
- `2caae147` — 2026-11-28 00:50:34 +0330 — Merge main and preserve both function studio and dependency regression tests
  Files: tests/admin-ui/pages.test.mjs
- `62ba9575` — 2026-11-28 00:50:01 +0330 — Merge pull request #51 from arashrasoulzadeh/admin/function-development-studio
  Files: 
- `ad86bc6e` — 2026-11-28 00:48:58 +0330 — Apply gofmt to new regression tests
  Files: tests/plugin/function_read_bridge_test.go
- `7bb316e2` — 2026-11-28 00:48:52 +0330 — Apply gofmt to new regression tests
  Files: tests/plugin/registry_enable_dependencies_test.go
- `09e7885d` — 2026-11-28 00:48:04 +0330 — Test function plugin metadata delegation and denied authority boundaries
  Files: tests/plugin/function_read_bridge_test.go
- `4cabff79` — 2026-11-28 00:47:04 +0330 — Add host-mediated read-only plugin bridge to operator function tests
  Files: docs/functions-runtime.md,docs/plugin-guides/functions.md internal/admin/function_plugin_bridge.go,internal/admin/functions.go internal/plugin/function_read.go,internal/plugin/manifest.go internal/plugin/registry_workers.go,internal/plugin/worker.go plugins/files/plugin.yaml
- `308b8764` — 2026-11-28 00:45:41 +0330 — Test exact dependency consent, project scope and atomic rollback
  Files: tests/plugin/registry_enable_dependencies_test.go
- `918082f8` — 2026-11-28 00:45:04 +0330 — Confirm exact required dependency activation before enabling plugins
  Files: admin-nextjs/app/components/plugin-catalog-panel.tsx,admin-nextjs/app/components/plugin-enable-dependencies.tsx admin-nextjs/app/lib/plugin-version-review.ts,admin-nextjs/app/plugins/[plugin]/view.tsx admin-nextjs/app/view.tsx,docs/admin-pages/plugin-detail.md docs/admin-pages/plugins.md,internal/admin/lifecycle.go internal/app/admin.go,internal/plugin/registry_enable_dependencies.go tests/admin-ui/pages.test.mjs
- `cb9fe3a3` — 2026-11-28 00:40:34 +0330 — Redesign function editor as a source-first IDE workspace
  Files: admin-nextjs/app/functions/[name]/view.tsx,admin-nextjs/app/functions/function-editor.tsx admin-nextjs/app/functions/function-workspace.tsx,admin-nextjs/app/theme.css docs/admin-pages/function-detail.md,docs/admin-pages/function-new.md tests/admin-ui/pages.test.mjs
- `06304377` — 2026-11-28 00:38:02 +0330 — Guard studio saves and distinguish unavailable logs
  Files: admin-nextjs/app/functions/[name]/view.tsx,admin-nextjs/app/functions/function-runner.tsx admin-nextjs/app/functions/new/view.tsx,tests/admin-ui/pages.test.mjs
- `6f97511d` — 2026-11-28 00:37:42 +0330 — Build function development studio with scoped sandbox execution
  Files: admin-nextjs/app/functions/[name]/view.tsx,admin-nextjs/app/functions/function-runner.tsx admin-nextjs/app/functions/new/view.tsx,admin-nextjs/app/lib/function-studio.ts admin-nextjs/app/theme.css,docs/admin-pages/function-detail.md docs/admin-pages/function-new.md,tests/admin-ui/pages.test.mjs tests/admin/function-inspection.test.ts
- `bce1988e` — 2026-11-27 21:06:25 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
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
