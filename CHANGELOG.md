# Changelog

Automatically generated from commits on main.

## Unreleased
- `6733ca7f` — 2026-11-27 14:12:19 +0330 — Merge pull request #25 from arashrasoulzadeh/fix/admin-model-revocation
  Files: 
- `4217b7ef` — 2026-11-27 14:12:11 +0330 — Reconcile verified model permission refresh with latest main
  Files: 
- `7a76ca89` — 2026-11-27 10:40:46 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `7c532baa` — 2026-11-27 14:06:38 +0330 — Merge pull request #15 from arashrasoulzadeh/refactor/rooms-games-pvp
  Files: 
- `113650e3` — 2026-11-27 14:06:30 +0330 — Integrate current main before merging game separation and coverage improvements
  Files: 
- `d509e68a` — 2026-11-27 10:29:05 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `9fbc5c9d` — 2026-11-27 13:57:50 +0330 — Clear generated model records and selections on permission revocation
  Files: admin-nextjs/app/components/model-panel.tsx,docs/admin-pages/extensions.md scripts/test-admin-browser.mjs
- `4179edce` — 2026-11-27 13:55:33 +0330 — Verify session validation rejects malformed identities and ambiguous auth versions
  Files: tests/plugin/session_failure_test.go
- `de04449a` — 2026-11-27 13:54:56 +0330 — Register broker fixture before installation
  Files: tests/plugin/broker_failure_test.go
- `7d63f098` — 2026-11-27 13:54:53 +0330 — Merge pull request #22 from arashrasoulzadeh/fix/admin-critical-ci
  Files: 
- `eb4b057a` — 2026-11-27 13:54:29 +0330 — Test broker and CSRF fail-closed behavior on entropy outage
  Files: tests/plugin/broker_failure_test.go,tests/security/csrf_failure_test.go
- `36174a7e` — 2026-11-27 13:53:05 +0330 — Verify observation outbox transactions survive receipt, retry and deletion failures
  Files: tests/plugin/observation_failure_coverage_test.go
- `de1a9afe` — 2026-11-27 13:52:15 +0330 — Fix model SQL error propagation and exercise database and entropy failures
  Files: internal/plugin/model_runtime.go,tests/plugin/runtime_failure_coverage_test.go
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
- `02c6c423` — 2026-11-27 13:12:24 +0330 — Cover delegated admin read checks and revoked target grants
  Files: tests/plugin/plugin_call_test.go
- `f034f0ec` — 2026-11-27 13:11:43 +0330 — Verify dependency context and delegation commit-time permission checks
  Files: tests/plugin/dependency_coverage_test.go
- `999e95b0` — 2026-11-27 09:41:41 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `9824af94` — 2026-11-27 13:11:02 +0330 — Merge pull request #17 from arashrasoulzadeh/fix/compose-node-runtime
  Files: 
- `aa7506d1` — 2026-11-27 13:10:38 +0330 — Revalidate displayed admin results and inventories after silent policy revocation
  Files: admin-nextjs/app/components/admin-inventory.tsx,admin-nextjs/app/components/admin-renderer.tsx admin-nextjs/app/components/model-panel.tsx,admin-nextjs/app/lib/admin-sdk.ts docs/admin-pages/extensions.md,internal/admin/lifecycle.go internal/app/admin.go,internal/plugin/access_management.go roadmap_admin.md,scripts/test-admin-browser.mjs tests/admin/manifest_admin_test.go,tests/plugin/access_management_test.go
- `33878463` — 2026-11-27 13:10:05 +0330 — Test observation retry exhaustion and corrupt queue failure handling
  Files: tests/plugin/observation_failure_coverage_test.go
- `96a2a32b` — 2026-11-27 13:08:54 +0330 — Exercise real packaged admin backend with Chromium policy and persistence tests
  Files: .github/workflows/admin-browser.yml,docs/admin-pages/theme.md scripts/test-admin-live.mjs
- `a0ccbdd6` — 2026-11-27 13:08:48 +0330 — Keep failed legacy project writes from granting in-memory access
  Files: internal/project/project.go,tests/project/persistence_failure_test.go
- `45b52cd2` — 2026-11-27 09:38:17 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `ea5e91d5` — 2026-11-27 13:07:56 +0330 — Test worker restart authorization and lifecycle rejection paths
  Files: tests/plugin/registry_lifecycle_coverage_test.go
- `029e8746` — 2026-11-27 13:07:42 +0330 — Merge pull request #18 from arashrasoulzadeh/test/admin-browser-smoke
  Files: 
- `f8419ce4` — 2026-11-27 13:06:50 +0330 — Test admin metadata validation and typed request authorization
  Files: tests/plugin/admin_contract_coverage_test.go
- `1cf04d71` — 2026-11-27 13:05:42 +0330 — Cover SQLite storage failures and legacy plugin persistence lifecycle
  Files: tests/plugin/storage_failure_coverage_test.go
- `7f2ecef9` — 2026-11-27 13:05:09 +0330 — Resolve browser test paths at step execution time
  Files: .github/workflows/admin-browser.yml
- `c50efdfd` — 2026-11-27 13:04:45 +0330 — Test model CRUD corruption and worker supervision failure boundaries
  Files: tests/plugin/runtime_failure_coverage_test.go
- `fbbb0e69` — 2026-11-27 13:04:32 +0330 — Add production-build Chromium admin smoke tests and screenshots
  Files: .github/workflows/admin-browser.yml,docs/admin-pages/theme.md scripts/test-admin-browser.mjs
- `b52ce1a5` — 2026-11-27 13:03:34 +0330 — Install production-matched Node runtime in Go integration test image
  Files: .github/workflows/test.yml,Dockerfile.test
- `6fceee46` — 2026-11-27 09:33:06 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `53f3a75b` — 2026-11-27 13:02:53 +0330 — Merge pull request #16 from arashrasoulzadeh/admin/compact-control-plane
  Files: 
- `9a1c1b8d` — 2026-11-27 12:58:24 +0330 — Format manifest coverage tests and remove temporary CI workspace export
  Files: .github/workflows/coverage.yml,tests/plugin/manifest_coverage_edges_test.go
- `69e37aa3` — 2026-11-27 12:54:59 +0330 — Temporarily export locked Go test workspace for coverage debugging
  Files: .github/workflows/coverage.yml
