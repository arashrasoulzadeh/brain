# Changelog

Automatically generated from commits on main.

## Unreleased
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
- `b3369bec` — 2026-11-27 14:50:40 +0330 — Merge pull request #27 from arashrasoulzadeh/fix/durable-subscriptions
  Files: 
- `c249e66a` — 2026-11-27 10:53:10 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `3743518e` — 2026-11-27 14:14:16 +0330 — Integrate all durable runtime fixes with current main and document migration limitations
  Files: docs/admin-pages/extensions.md
- `ff186c39` — 2026-11-27 14:14:14 +0330 — Verify registry catalog ordering and revoked version approvals
  Files: tests/plugin/catalog_failure_test.go
- `61c35258` — 2026-11-27 14:13:00 +0330 — Test subscription restart, deduplication and in-flight credential revocation
  Files: internal/plugin/legacy_subscription_storage.go,internal/plugin/legacy_subscription_storage_test.go
- `e2b61cf8` — 2026-11-27 14:12:42 +0330 — Reject nil worker factory results and test startup rollback
  Files: internal/plugin/registry_lifecycle.go,internal/plugin/registry_workers.go tests/plugin/worker_factory_failure_test.go
- `6733ca7f` — 2026-11-27 14:12:19 +0330 — Merge pull request #25 from arashrasoulzadeh/fix/admin-model-revocation
  Files: 
- `4217b7ef` — 2026-11-27 14:12:11 +0330 — Reconcile verified model permission refresh with latest main
  Files: 
- `13112d96` — 2026-11-27 14:11:32 +0330 — Validate baseline observation manifests before checking rejected mutations
  Files: tests/plugin/observation_limits_test.go
- `4ee9b2f8` — 2026-11-27 14:11:23 +0330 — Persist subscription tickets and bounded observation streams across worker recovery
  Files: docs/admin-pages/extensions.md,internal/plugin/legacy_storage_adapter.go internal/plugin/legacy_subscription_storage.go,internal/plugin/registry_workers.go plugins/subscriptions/routes.ts,roadmap_admin.md scripts/test-production-runtime.sh
- `a477829d` — 2026-11-27 14:11:12 +0330 — Verify observation fanout limits and reject stale durable deliveries
  Files: tests/plugin/observation_limits_test.go
- `7a76ca89` — 2026-11-27 10:40:46 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `5de27cb8` — 2026-11-27 14:10:38 +0330 — Reconcile invitations with current main and test durable single-use semantics
  Files: internal/plugin/legacy_invite_storage_test.go,scripts/test-production-runtime.sh
- `365c4cd8` — 2026-11-27 14:10:17 +0330 — Reconcile auth persistence with current main and validate exact worker identity results
  Files: cmd/brain/main.go,internal/plugin/legacy_auth_storage.go scripts/test-production-runtime.sh,tests/plugin/session_failure_test.go
- `0ad77473` — 2026-11-27 14:09:55 +0330 — Verify worker admission and commit reject resource substitution and context loss
  Files: tests/plugin/invocation_failure_matrix_test.go
- `ee4ad3a8` — 2026-11-27 14:09:41 +0330 — Reconcile arena recovery with game separation and instrumented runtime smoke
  Files: scripts/test-production-runtime.sh
- `065c73fe` — 2026-11-27 14:08:31 +0330 — Keep failed function draft writes from exposing uncommitted publications
  Files: internal/admin/functions.go,tests/admin/function_persistence_failure_test.go
- `a46c5c38` — 2026-11-27 14:08:29 +0330 — Persist project-scoped invitations with digest-only codes and conditional consumption
  Files: docs/admin-pages/extensions.md,internal/plugin/legacy_invite_storage.go internal/plugin/legacy_storage_adapter.go,internal/plugin/registry_workers.go plugins/invites/index.ts,plugins/invites/plugin.yaml roadmap_admin.md,scripts/test-production-runtime.sh
