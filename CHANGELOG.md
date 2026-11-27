# Changelog

Automatically generated from commits on main.

## Unreleased
- `2432526a` — 2026-11-27 16:28:23 +0330 — Merge pull request #37 from arashrasoulzadeh/docs/current-complete-reference
  Files: 
- `e30d893f` — 2026-11-27 16:17:38 +0330 — Merge pull request #36 from arashrasoulzadeh/test/host-coverage-edge-cases
  Files: 
- `0c1edeb0` — 2026-11-27 16:16:55 +0330 — docs: recognize generic SDK types and direct exports in coverage audit
  Files: scripts/documentation_inventory.py
- `b3c62868` — 2026-11-27 16:16:22 +0330 — docs: expand every admin page guide and audit Next.js routes
  Files: docs/admin-pages/README.md,docs/admin-pages/audit.md docs/admin-pages/database.md,docs/admin-pages/files.md docs/admin-pages/function-detail.md,docs/admin-pages/function-new.md docs/admin-pages/functions.md,docs/admin-pages/live-sessions.md docs/admin-pages/login.md,docs/admin-pages/monitoring.md docs/admin-pages/plugin-detail.md,docs/admin-pages/plugins.md docs/admin-pages/routes.md,docs/admin-pages/system.md docs/admin-pages/workers.md,docs/admin-pages/workflows.md docs/plugin-guides/dashboard.md,docs/quality.md scripts/documentation_inventory.py,tests/documentation/test_inventory.py
- `3877b9fe` — 2026-11-27 16:14:41 +0330 — docs: add every plugin guide and enforce public contract documentation inventory
  Files: docs/plugin-guides/access-management.md,docs/plugin-guides/analytics.md docs/plugin-guides/arena.md,docs/plugin-guides/auth.md docs/plugin-guides/dashboard.md,docs/plugin-guides/declarative-plugin.md docs/plugin-guides/files.md,docs/plugin-guides/functions.md docs/plugin-guides/games.md,docs/plugin-guides/index.md docs/plugin-guides/invites.md,docs/plugin-guides/matchmaking.md docs/plugin-guides/pvp.md,docs/plugin-guides/rooms.md docs/plugin-guides/subscriptions.md,docs/plugin-guides/tables.md docs/plugin-guides/template.md,docs/quality.md scripts/build-docs.py,scripts/documentation_inventory.py tests/documentation/test_inventory.py
- `abe3a4d5` — 2026-11-27 12:44:19 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `cf684327` — 2026-11-27 16:13:33 +0330 — docs: refresh host, client, SDK, deployment and HTTP references against current main
  Files: clients/android/README.md,docs/admins.md docs/assets/brain.css,docs/browser-quickstart.md docs/declarative-runtime.md,docs/developers.md docs/flutter-quickstart.md,docs/functions-runtime.md docs/http-api.md,docs/index.md docs/restore-runbook.md,docs/runtime-operations.md docs/sdk-reference.md,mkdocs.yml plugins/README.md
- `743e3dce` — 2026-11-27 16:11:01 +0330 — test: verify close failures cannot be mistaken for lock release
  Files: internal/deploy/lock_failure_coverage_test.go
- `4abec123` — 2026-11-27 16:10:51 +0330 — test: exercise successful admin lifecycle and page response dispatch
  Files: internal/app/admin_edge_coverage_test.go
- `2f002ed0` — 2026-11-27 16:10:22 +0330 — test: use one invitation selector for expiry commit checks
  Files: internal/plugin/runtime_storage_edge_test.go
- `7f5f756b` — 2026-11-27 16:10:13 +0330 — test: cover late invitation expiry and storage read failures
  Files: internal/plugin/runtime_storage_edge_test.go
- `795f83eb` — 2026-11-27 16:10:04 +0330 — test: cover model admission and invalid action without durable mutation
  Files: internal/plugin/model_admission_edge_test.go
- `4a65e87b` — 2026-11-27 16:09:47 +0330 — test: cover manifest discovery, stale credentials, and registry admission failures
  Files: internal/plugin/host_edge_coverage_test.go
- `334bcded` — 2026-11-27 16:09:20 +0330 — Merge pull request #34 from arashrasoulzadeh/test/remaining-host-coverage
  Files: 
- `a0ad040c` — 2026-11-27 16:09:16 +0330 — Merge pull request #33 from arashrasoulzadeh/fix/sqlite-coverage-stability
  Files: 
- `e45fa348` — 2026-11-27 16:03:49 +0330 — test: cover draft persistence failures and request identity fallback
  Files: internal/admin/function_persistence_failure_test.go
- `988aedb5` — 2026-11-27 16:03:06 +0330 — test: cover empty API error and login entropy failure
  Files: internal/app/response_failure_coverage_test.go
- `a4c0cf12` — 2026-11-27 16:03:04 +0330 — refactor: remove redundant newline check after TrimSpace
  Files: internal/app/api_response.go
- `d555bcc9` — 2026-11-27 16:02:48 +0330 — test: cover storage planner identity rejection, duplicate events and atomic rollback
  Files: internal/plugin/legacy_adapter_failure_matrix_test.go
- `8f59f70d` — 2026-11-27 16:02:39 +0330 — test: cover ambiguous publishers, bounded deliveries and outbox rollback
  Files: internal/plugin/observation_failure_matrix_test.go
- `6c1c1690` — 2026-11-27 16:02:25 +0330 — test: cover delegated operation policy, worker readiness and entropy failures
  Files: internal/plugin/delegation_coverage_test.go
- `91887a1c` — 2026-11-27 16:02:05 +0330 — test: cover macOS opt-in and loopback restrictions across platforms
  Files: cmd/brain/sandbox_runtime_coverage_test.go
- `8a8f3c5c` — 2026-11-27 16:02:03 +0330 — refactor: isolate platform sandbox configuration for coverage without changing runtime selection
  Files: cmd/brain/main.go
- `7376c8d2` — 2026-11-27 16:01:51 +0330 — test: exercise macOS sandbox profile generation and fail-closed paths on all hosts
  Files: internal/plugin/macos_profile_coverage_test.go
- `80ce28bf` — 2026-11-27 16:01:40 +0330 — refactor: isolate macOS profile construction from runtime platform check
  Files: internal/plugin/macos_worker.go
- `e9cc3e94` — 2026-11-27 16:00:50 +0330 — test: verify lifecycle persistence survives temporary SQLite write contention
  Files: internal/plugin/sqlite_contention_test.go
- `3a97c86b` — 2026-11-27 16:00:48 +0330 — test: cover SQLite DSN preservation and per-connection timeout
  Files: internal/sqliteconfig/config_test.go
- `abf651e4` — 2026-11-27 16:00:38 +0330 — fix: apply per-connection busy timeout to shared SQLite database
  Files: internal/project/sqlite.go
- `4ff3ffee` — 2026-11-27 16:00:35 +0330 — fix: apply per-connection busy timeout to shared SQLite database
  Files: internal/plugin/sqlite_storage.go
- `ce35f373` — 2026-11-27 16:00:33 +0330 — fix: configure bounded SQLite busy handling on every connection
  Files: internal/sqliteconfig/config.go
- `6ffece22` — 2026-11-27 12:30:24 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `ae39ac29` — 2026-11-27 15:59:53 +0330 — Merge pull request #32 from arashrasoulzadeh/fix/native-macos-dev
  Files: .github/workflows/test.yml
- `a7746253` — 2026-11-27 15:59:51 +0330 — Merge pull request #30 from arashrasoulzadeh/test/durable-storage-coverage
  Files: 
- `5fe94dac` — 2026-11-27 15:13:27 +0330 — Use a valid temporary directory for Seatbelt initialization and add native launch probes
  Files: .github/workflows/test.yml,scripts/run-macos-typescript-worker.sh tests/plugin/macos_worker_test.go
- `d71d6616` — 2026-11-27 15:11:45 +0330 — Diagnose native worker startup and permit scoped executable operations
  Files: .github/workflows/test.yml,internal/plugin/macos_worker.go tests/plugin/macos_worker_test.go
- `fe04fd13` — 2026-11-27 15:11:35 +0330 — test: cover sandbox factory manifest and descriptor failure paths
  Files: cmd/brain/worker_factory_failure_test.go
- `24e7f2ff` — 2026-11-27 15:11:12 +0330 — test: exercise production startup failures with isolated process fixtures
  Files: cmd/brain/startup_runtime_failure_test.go
- `7092f906` — 2026-11-27 15:10:44 +0330 — test: cover ambiguous policy routes and corrupt access rule inventories
  Files: internal/plugin/access_inventory_failure_test.go
- `9ef231e4` — 2026-11-27 15:10:06 +0330 — Permit only required executable mappings and child cleanup in macOS sandbox
  Files: internal/plugin/macos_worker.go
- `ae95bfe5` — 2026-11-27 15:10:00 +0330 — test: verify lifecycle failure paths do not retain authorization
  Files: internal/plugin/lifecycle_failure_coverage_test.go
- `83f015f7` — 2026-11-27 15:09:21 +0330 — test: cover admin missing dependency and invalid project failures
  Files: internal/admin/lifecycle_failure_coverage_test.go
- `f7a83dac` — 2026-11-27 15:09:08 +0330 — Support native macOS development using Seatbelt worker isolation
  Files: .github/workflows/test.yml,README.md cmd/brain/main.go,docs/developers.md internal/plugin/macos_worker.go,scripts/run-macos-typescript-worker.sh scripts/start-local.sh,scripts/test-production-runtime.sh scripts/test-runtime-api.mjs,tests/plugin/macos_worker_test.go
- `15f2d1c7` — 2026-11-27 15:09:04 +0330 — style: fix CI-reported Go composite literal formatting
  Files: internal/plugin/durable_storage_coverage_test.go
- `11d8a88d` — 2026-11-27 15:08:48 +0330 — test: cover access policy fail-closed and worker admission paths
  Files: internal/plugin/access_failure_coverage_test.go
- `5efe9759` — 2026-11-27 15:08:31 +0330 — style: align guest response fields
  Files: internal/plugin/storage_failure_matrix_test.go
- `21f1e055` — 2026-11-27 15:08:14 +0330 — test: cover durable storage commit and entropy failures
  Files: internal/plugin/storage_failure_matrix_test.go
- `09b21be0` — 2026-11-27 15:07:56 +0330 — ci: print formatting diffs without relaxing the gate
  Files: .github/workflows/test.yml
- `de941b67` — 2026-11-27 14:58:55 +0330 — test: cover startup manifest rejection and sandbox prerequisites
  Files: cmd/brain/startup_coverage_test.go
- `4bb4c8ee` — 2026-11-27 14:58:18 +0330 — test: encode valid guest expiration before injected SQL failure
  Files: internal/plugin/durable_commit_coverage_test.go
- `84476972` — 2026-11-27 14:58:12 +0330 — test: cover durable capacity limits and database failure boundaries
  Files: internal/plugin/durable_commit_coverage_test.go
