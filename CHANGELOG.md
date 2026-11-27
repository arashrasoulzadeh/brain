# Changelog

Automatically generated from commits on main.

## Unreleased
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
- `7c532baa` — 2026-11-27 14:06:38 +0330 — Merge pull request #15 from arashrasoulzadeh/refactor/rooms-games-pvp
  Files: 
- `113650e3` — 2026-11-27 14:06:30 +0330 — Integrate current main before merging game separation and coverage improvements
  Files: 
- `d509e68a` — 2026-11-27 10:29:05 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `9fbc5c9d` — 2026-11-27 13:57:50 +0330 — Clear generated model records and selections on permission revocation
  Files: admin-nextjs/app/components/model-panel.tsx,docs/admin-pages/extensions.md scripts/test-admin-browser.mjs
- `dc1cd6d9` — 2026-11-27 13:56:40 +0330 — Merge arena recovery fix, redact auth errors and verify credential storage
  Files: cmd/brain/main.go,docs/admin-pages/extensions.md internal/plugin/legacy_auth_storage_test.go,internal/plugin/registry_workers.go plugins/auth/plugin.yaml,plugins/auth/service.ts scripts/test-production-runtime.sh
- `4179edce` — 2026-11-27 13:55:33 +0330 — Verify session validation rejects malformed identities and ambiguous auth versions
  Files: tests/plugin/session_failure_test.go
- `afb75c3b` — 2026-11-27 13:55:31 +0330 — Merge browser CI fix and explicitly restart recovered arena workers
  Files: scripts/test-production-runtime.sh,tests/arena/arena-routes.test.ts
- `de04449a` — 2026-11-27 13:54:56 +0330 — Register broker fixture before installation
  Files: tests/plugin/broker_failure_test.go
- `7d63f098` — 2026-11-27 13:54:53 +0330 — Merge pull request #22 from arashrasoulzadeh/fix/admin-critical-ci
  Files: 
- `eb4b057a` — 2026-11-27 13:54:29 +0330 — Test broker and CSRF fail-closed behavior on entropy outage
  Files: tests/plugin/broker_failure_test.go,tests/security/csrf_failure_test.go
- `4d80fb39` — 2026-11-27 13:54:04 +0330 — Persist project authentication records through response-aware storage commits
  Files: docs/admin-pages/extensions.md,internal/plugin/legacy_auth_storage.go internal/plugin/legacy_storage_adapter.go,internal/plugin/registry_workers.go plugins/auth/service.ts,roadmap_admin.md scripts/test-production-runtime.sh,tests/plugin/user_session_test.go
- `36174a7e` — 2026-11-27 13:53:05 +0330 — Verify observation outbox transactions survive receipt, retry and deletion failures
  Files: tests/plugin/observation_failure_coverage_test.go
- `de1a9afe` — 2026-11-27 13:52:15 +0330 — Fix model SQL error propagation and exercise database and entropy failures
  Files: internal/plugin/model_runtime.go,tests/plugin/runtime_failure_coverage_test.go
- `6b870308` — 2026-11-27 13:51:01 +0330 — Persist arena snapshots and sequence counters with atomic state transitions
  Files: docs/admin-pages/extensions.md,internal/plugin/legacy_admin_inventory.go internal/plugin/legacy_arena_storage.go,internal/plugin/legacy_arena_storage_test.go internal/plugin/legacy_storage_adapter.go,internal/plugin/registry_workers.go plugins/arena/plugin.yaml,plugins/arena/state-machine.ts roadmap_admin.md,scripts/test-production-runtime.sh
- `fbee8579` — 2026-11-27 13:50:09 +0330 — Fix health-only shutdown and cover discovery and sandbox startup failures
  Files: cmd/brain/main.go,tests/deploy/cli_failure_coverage_test.go tests/plugin/worker_failure_coverage_test.go
- `322c17a3` — 2026-11-27 13:49:23 +0330 — Exclude generated frontend artifacts and local state from Docker build context
  Files: .dockerignore
- `da36eb34` — 2026-11-27 13:49:08 +0330 — Use runner Chrome for sandbox-compatible admin browser CI
  Files: .github/workflows/admin-browser.yml,docs/admin-pages/theme.md scripts/test-admin-browser.mjs,scripts/test-admin-live.mjs
- `a13ff8cc` — 2026-11-27 13:48:47 +0330 — Test lifecycle persistence rollback and cleanup failure matrix
  Files: tests/plugin/lifecycle_failure_matrix_test.go
- `f23ac892` — 2026-11-27 10:18:42 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `8f728cff` — 2026-11-27 13:47:49 +0330 — Adapt invitation coverage to room-scoped authorization from main
  Files: tests/invites/coverage-edges.test.ts
- `e5bfbec0` — 2026-11-27 13:47:06 +0330 — Test subprocess startup, RPC size and response validation failures
  Files: tests/plugin/worker_failure_coverage_test.go
- `0d513e57` — 2026-11-27 13:46:35 +0330 — Merge pull request #19 from arashrasoulzadeh/test/admin-live-backend
  Files: 
- `ec2d74da` — 2026-11-27 13:46:17 +0330 — Merge current main and align live Chromium runner with packaged runtime CI
  Files: .github/workflows/admin-browser.yml
- `c84d11a4` — 2026-11-27 13:46:01 +0330 — Merge current main and preserve operator SDK, game separation, and coverage improvements
  Files: .github/workflows/test.yml,plugins/sdk-typescript/src/index.ts plugins/sdk-typescript/src/validation.ts
- `c263bc03` — 2026-11-27 10:15:40 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `c350a896` — 2026-11-27 13:45:21 +0330 — Merge pull request #21 from arashrasoulzadeh/fix/admin-project-scope
  Files: 
- `f149ff75` — 2026-11-27 13:45:15 +0330 — Merge pull request #20 from arashrasoulzadeh/fix/admin-revocation-refresh
  Files: 
- `a42e093b` — 2026-11-27 13:15:05 +0330 — Prevent stale project catalog and function responses after switching projects
  Files: admin-nextjs/app/components/projects-panel.tsx,admin-nextjs/app/lib/project-requests.ts admin-nextjs/app/page.tsx,docs/admin-pages/projects.md tests/admin/project-requests.test.ts
- `be4d97d5` — 2026-11-27 13:13:39 +0330 — Keep live browser backend network-disabled and relay CI HTTP over docker exec
  Files: .github/workflows/admin-browser.yml,docs/admin-pages/theme.md scripts/admin-ci-backend-relay.mjs
- `e8846874` — 2026-11-27 13:12:32 +0330 — Test function result validation and unavailable runtime behavior
  Files: tests/admin/function_result_coverage_test.go
