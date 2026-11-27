# Changelog

Automatically generated from commits on main.

## Unreleased
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
- `0d45e69e` — 2026-11-27 14:56:45 +0330 — style: fix Go formatting
  Files: internal/plugin/durable_storage_coverage_test.go
- `53188fea` — 2026-11-27 14:56:43 +0330 — style: fix Go formatting
  Files: internal/plugin/wasm_runner_coverage_test.go
- `b9275059` — 2026-11-27 14:56:37 +0330 — style: normalize durable authentication tests
  Files: internal/plugin/durable_auth_coverage_test.go
- `16f3970d` — 2026-11-27 14:56:10 +0330 — style: normalize durable storage regression test formatting
  Files: internal/plugin/durable_storage_coverage_test.go
- `470dd5d6` — 2026-11-27 14:54:44 +0330 — test: exercise valid WASM execution within project context
  Files: internal/plugin/wasm_runner_coverage_test.go
- `c4691a0f` — 2026-11-27 14:54:36 +0330 — test: cover durable authentication denial and revocation races
  Files: internal/plugin/durable_auth_coverage_test.go
- `f6dee120` — 2026-11-27 14:54:19 +0330 — test: cover durable storage rejection and commit boundaries
  Files: internal/plugin/durable_storage_coverage_test.go
- `56cd958a` — 2026-11-27 11:23:15 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `3e70d30e` — 2026-11-27 14:52:55 +0330 — Merge pull request #29 from arashrasoulzadeh/fix/reconcile-restore-verification
  Files: 
- `544c722f` — 2026-11-27 14:52:36 +0330 — test: reconcile isolated restore verification
  Files: tests/deploy/clean_environment_script_test.go
- `eb3caacf` — 2026-11-27 11:22:11 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `9ac473ff` — 2026-11-27 14:51:55 +0330 — Merge pull request #28 from arashrasoulzadeh/test/complete-go-coverage
  Files: internal/plugin/registry_workers.go
- `ccd23f08` — 2026-11-27 11:20:51 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
