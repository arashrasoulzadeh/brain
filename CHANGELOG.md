# Changelog

Automatically generated from commits on main.

## Unreleased

- Breaking: all application HTTP routes require a project slug or immutable ID; server infrastructure uses `/_system`.
- Active project plugin manifests declare host-mounted HTTP routes and authentication; public functions use `/{project}/functions/{name}`.
- Project names, URL slugs and settings are mutable without changing immutable IDs.
- Clients, operator UI and deployment integrations must use the new URLs.

- `c6f51829` — 2026-11-27 18:48:32 +0330 — Merge pull request #39 from arashrasoulzadeh/refactor/manifest-driven-runtime
  Files: 
- `07f5ee22` — 2026-11-27 18:44:19 +0330 — Use atomic CLI coverage mode consistently
  Files: cmd/manifest-sdk/main_test.go
- `8dea3425` — 2026-11-27 18:43:57 +0330 — Exercise manifest generator command success and failure with CLI coverage
  Files: cmd/manifest-sdk/main_test.go
- `7cce76cd` — 2026-11-27 18:41:17 +0330 — Test manifest generator failure paths and emit type-only declaration file
  Files: cmd/manifest-sdk/main.go,cmd/manifest-sdk/main_test.go plugins/sdk-typescript/src/manifest-operations.generated.d.ts,tests/plugin/manifest_sdk_test.go
- `c9633148` — 2026-11-27 18:38:20 +0330 — Reconcile manifest runtime refactor with current secure host and SDK
  Files: docs/sdk-reference.md,plugins/functions/scheduling.ts plugins/matchmaking/queue.ts,plugins/sdk-typescript/src/index.ts plugins/sdk-typescript/src/manifest-operations.generated.ts,plugins/sdk-typescript/src/request.ts tests/plugin/manifest-request.test.ts
- `c1b6c8e8` — 2026-11-27 15:05:18 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `8bf6cb5c` — 2026-11-27 18:34:14 +0330 — Merge pull request #40 from arashrasoulzadeh/fix/macos-homebrew-openssl
  Files: 
- `03f401b1` — 2026-11-27 18:33:23 +0330 — Correct Homebrew isolation CI step formatting
  Files: .github/workflows/test.yml
- `c80d4a0e` — 2026-11-27 18:33:02 +0330 — Fix native macOS worker startup with Homebrew OpenSSL
  Files: .github/workflows/test.yml,internal/plugin/macos_profile_coverage_test.go internal/plugin/macos_worker.go
- `abd9aacf` — 2026-11-27 14:02:28 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `5611abbb` — 2026-11-27 17:31:56 +0330 — Merge pull request #35 from arashrasoulzadeh/fix/macos-seatbelt-startup
  Files: 
- `b6110f87` — 2026-11-27 17:13:55 +0330 — Emit actual newline between Seatbelt directory rules
  Files: internal/plugin/macos_worker.go
- `1ac264e9` — 2026-11-27 17:13:50 +0330 — Fix macOS worker getcwd with literal ancestor directory permissions
  Files: internal/plugin/macos_worker.go
- `a3cfcb65` — 2026-11-27 13:07:55 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `2432526a` — 2026-11-27 16:28:23 +0330 — Merge pull request #37 from arashrasoulzadeh/docs/current-complete-reference
  Files: 
- `fd4ed616` — 2026-11-27 16:27:17 +0330 — Merge main into macOS startup fix preserving platform coverage helper
  Files: internal/plugin/macos_worker.go
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
- `b3c2af8c` — 2026-11-27 16:09:35 +0330 — Run the actual native launcher with the confirmed dyld root lookup fix
  Files: .github/workflows/test.yml,tests/plugin/macos_worker_test.go
- `334bcded` — 2026-11-27 16:09:20 +0330 — Merge pull request #34 from arashrasoulzadeh/test/remaining-host-coverage
  Files: 
- `a0ad040c` — 2026-11-27 16:09:16 +0330 — Merge pull request #33 from arashrasoulzadeh/fix/sqlite-coverage-stability
  Files: 
- `d6eaa8fe` — 2026-11-27 16:06:29 +0330 — Permit literal root directory lookup and narrowly scoped dyld runtime files
  Files: internal/plugin/macos_worker.go
- `008a8f49` — 2026-11-27 16:04:37 +0330 — Allow macOS system runtime paths and diagnose Seatbelt startup
  Files: .github/workflows/test.yml,internal/plugin/macos_worker.go tests/plugin/macos_worker_test.go
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
