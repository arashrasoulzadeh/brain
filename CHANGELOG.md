# Changelog

Automatically generated from commits on main.

## Unreleased
- `86cf113` — 2026-11-27 02:23:56 +0330 — fix(ci): import dashboard project validation through public SDK
  Files: plugins/dashboard/custom-admin-view.ts,plugins/dashboard/plugin.yaml
- `f311d19` — 2026-11-26 22:52:28 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `39fa867` — 2026-11-27 02:21:46 +0330 — Merge pull request #2 from arashrasoulzadeh/refactor/plugin-sdk-yaml-helpers
  Files: 
- `ad1afc4` — 2026-11-27 02:20:32 +0330 — Merge current main while preserving SDK helpers and JSON worker contract
  Files: docs/developers.md,internal/plugin/manifest.go internal/plugin/typescript_worker.go,scripts/typescript-worker.mjs tests/plugin/manifest-helpers.test.ts
- `4707c7e` — 2026-11-27 02:19:14 +0330 — refactor: consolidate SDK routing, request and validation modules
  Files: plugins/sdk-typescript/src/create-function-handler.ts,plugins/sdk-typescript/src/create-host-handler.ts plugins/sdk-typescript/src/create-route-handler.ts,plugins/sdk-typescript/src/index.ts plugins/sdk-typescript/src/plugin-request.ts,plugins/sdk-typescript/src/project-store.ts plugins/sdk-typescript/src/request.ts,plugins/sdk-typescript/src/routes.ts plugins/sdk-typescript/src/validate-identifier.ts,plugins/sdk-typescript/src/validate-project-id.ts plugins/sdk-typescript/src/validation.ts,plugins/sdk-typescript/src/worker.ts tests/plugin/sdk.test.ts
- `8e42517` — 2026-11-27 02:19:13 +0330 — Merge pull request #1 from arashrasoulzadeh/refactor/admin-http-boundaries
  Files: 
- `6894d95` — 2026-11-27 02:18:25 +0330 — docs: record shared architecture boundaries and remaining release checks
  Files: README.md,docs/developers.md docs/refactoring.md
- `9ac98ff` — 2026-11-27 02:18:00 +0330 — fix(clients): align project identity validation with host
  Files: clients/android/src/main/kotlin/com/brain/client/BrainClient.kt,clients/flutter/lib/native_transport.dart clients/flutter/lib/project_identity.dart,clients/flutter/lib/web_transport.dart clients/javascript/src/client.ts,protocol/README.md tests/client/javascript-client.test.ts,tests/flutter/project_identity_test.dart
- `33410f2` — 2026-11-27 02:17:34 +0330 — refactor: support context-aware SDK request helpers in Auth
  Files: AGENTS.md,docs/developers.md internal/plugin/plugin.go,plugins/auth/auth-functions.ts plugins/auth/auth-guest-function.ts,plugins/auth/auth-session-function.ts plugins/auth/auth-signin-function.ts,plugins/auth/auth-signup-function.ts plugins/auth/index.ts,plugins/auth/plugin.yaml plugins/sdk-typescript/src/create-function-handler.ts,plugins/sdk-typescript/src/index.ts plugins/sdk-typescript/src/routes.ts,scripts/typescript-worker.mjs tests/auth/sdk-dispatch.test.ts,tests/plugin/manifest-helpers.test.ts tests/plugin/manifest_helpers_test.go
- `b45896c` — 2026-11-27 02:17:34 +0330 — refactor(admin): split project catalog draft and route panels
  Files: admin-nextjs/app/components/function-drafts-panel.tsx,admin-nextjs/app/components/plugin-catalog-panel.tsx admin-nextjs/app/components/projects-panel.tsx,admin-nextjs/app/components/routes-panel.tsx admin-nextjs/app/page.tsx
- `84fba6a` — 2026-11-27 02:16:42 +0330 — refactor: share transaction helpers across Tables and Functions
  Files: plugins/functions/create-invocation-effects.ts,plugins/sdk-typescript/src/execution.ts plugins/sdk-typescript/src/index.ts,plugins/tables/api-change-reader.ts plugins/tables/api-change-state.ts,plugins/tables/api-change-writer.ts plugins/tables/create-api-change-capture.ts,plugins/tables/create-direct-sql-change-capture.ts plugins/tables/create-schema-publication-audit.ts,plugins/tables/create-transactional-change-capture.ts plugins/tables/index.ts,plugins/tables/plugin.yaml plugins/tables/routes.ts,plugins/tables/schema-audit-state.ts plugins/tables/schema-publication-reader.ts,plugins/tables/schema-publication-recorder.ts plugins/tables/scoped-transaction-change-reader.ts,plugins/tables/scoped-transaction-change-writer.ts plugins/tables/sql-change-reader.ts,plugins/tables/sql-change-state.ts plugins/tables/sql-change-writer.ts,plugins/tables/tables-read.ts plugins/tables/tables-schema.ts,plugins/tables/tables-write.ts plugins/tables/transaction-change-reader.ts,plugins/tables/transaction-change-state.ts plugins/tables/transaction-change-writer.ts,plugins/tables/versioned-record-reader.ts plugins/tables/versioned-record-state.ts,plugins/tables/versioned-record-updater.ts plugins/tables/versioned-record.ts,tests/plugin/manifest-helpers.test.ts tests/plugin/transaction-helpers.test.ts
- `6e930c9` — 2026-11-27 02:16:01 +0330 — refactor(plugin): inject storage planners and normalize registry lock ordering
  Files: internal/plugin/legacy_storage_adapter.go,internal/plugin/plugin.go internal/plugin/registry_catalog.go,internal/plugin/registry_lifecycle.go internal/plugin/registry_permissions.go,internal/plugin/registry_storage.go internal/plugin/registry_workers.go,tests/plugin/storage_planner_test.go
- `d3e37d3` — 2026-11-27 02:15:33 +0330 — refactor: migrate Files and Rooms to shared operation routes
  Files: docs/developers.md,plugins/files/files-download.ts plugins/files/files-metadata.ts,plugins/files/files-publish.ts plugins/files/files-upload.ts,plugins/files/index.ts plugins/files/plugin.yaml,plugins/files/routes.ts plugins/functions/routes.ts,plugins/matchmaking/routes.ts plugins/rooms/index.ts,plugins/rooms/plugin.yaml plugins/rooms/rooms-create-route.ts,plugins/rooms/rooms-join-route.ts plugins/rooms/rooms-reconnect-route.ts,plugins/rooms/rooms-result-route.ts plugins/rooms/routes.ts,plugins/sdk-typescript/src/index.ts plugins/sdk-typescript/src/routes.ts,tests/plugin/manifest-helpers.test.ts tests/plugin/sdk.test.ts
- `2b9eb61` — 2026-11-27 02:15:16 +0330 — fix(operator): enforce absolute session expiry with a controlled clock
  Files: docs/admins.md,internal/operator/operator.go internal/operator/session.go,tests/operator/session_expiry_test.go
- `ba62d07` — 2026-11-27 02:14:51 +0330 — refactor: consolidate Functions stores and share execution helpers
  Files: .github/workflows/test.yml,plugins/functions/approval-granter.ts plugins/functions/approval-reader.ts,plugins/functions/approval-requester.ts plugins/functions/approval-state.ts,plugins/functions/commit-with-grant.ts plugins/functions/create-approval-store.ts,plugins/functions/create-function-draft-store.ts plugins/functions/create-function-publication-store.ts,plugins/functions/create-invocation-commit-store.ts plugins/functions/create-invocation-effects.ts,plugins/functions/create-job-lease-store.ts plugins/functions/create-record-change-trigger-store.ts,plugins/functions/create-schedule-store.ts plugins/functions/create-shared-package-store.ts,plugins/functions/function-draft-reader.ts plugins/functions/function-draft-state.ts,plugins/functions/function-draft-writer.ts plugins/functions/function-publication-reader.ts,plugins/functions/function-publication-rollback.ts plugins/functions/function-publication-state.ts,plugins/functions/function-publication-writer.ts plugins/functions/function-version-enqueuer.ts,plugins/functions/index.ts plugins/functions/invocation-commit-reader.ts,plugins/functions/invocation-commit-state.ts plugins/functions/invocation-commit-writer.ts,plugins/functions/invocation-effects-reader.ts plugins/functions/invocation-effects-state.ts,plugins/functions/invocation-effects-writer.ts plugins/functions/invocation-existence-reader.ts,plugins/functions/invocation-result-reader.ts plugins/functions/job-lease-acquirer.ts,plugins/functions/job-lease-renewer.ts plugins/functions/job-lease-state.ts,plugins/functions/missed-schedule-enqueuer.ts plugins/functions/plugin.yaml,plugins/functions/record-trigger-adder.ts plugins/functions/record-trigger-dispatcher.ts,plugins/functions/record-trigger-matcher.ts plugins/functions/record-trigger-state.ts,plugins/functions/run-capped-retries.ts plugins/functions/run-idempotent-workflow.ts,plugins/functions/run-with-timeout.ts plugins/functions/schedule-adder.ts,plugins/functions/schedule-claimer.ts plugins/functions/schedule-occurrence-recorder.ts,plugins/functions/schedule-reader.ts plugins/functions/schedule-state.ts,plugins/functions/shared-package-reader.ts plugins/functions/shared-package-state.ts,plugins/functions/shared-package-writer.ts plugins/sdk-typescript/src/execution.ts,plugins/sdk-typescript/src/index.ts plugins/sdk-typescript/src/project-store.ts,tests/functions/shared-sdk-stores.test.ts
- `a96932d` — 2026-11-27 02:14:45 +0330 — refactor(admin): share navigation policy and verify frontend types in CI
  Files: .github/workflows/test.yml,admin-nextjs/app/components/admin-chrome.tsx admin-nextjs/app/lib/admin-navigation.ts,admin-nextjs/app/page.tsx admin-nextjs/package.json,tests/admin/admin-navigation.test.ts
- `5ab8807` — 2026-11-27 02:14:21 +0330 — refactor(admin): extract reusable plugin consoles and focused dashboard panels
  Files: admin-nextjs/app/components/audit-panel.tsx,admin-nextjs/app/components/monitoring-panel.tsx admin-nextjs/app/components/plugin-console.tsx,admin-nextjs/app/components/system-panel.tsx admin-nextjs/app/lib/admin-copy.ts,admin-nextjs/app/lib/admin-types.ts admin-nextjs/app/page.tsx
- `beb3cac` — 2026-11-27 02:13:42 +0330 — refactor: introduce manifest-selected plugin SDK dispatch
  Files: AGENTS.md,internal/plugin/plugin.go internal/plugin/typescript_worker.go,plugins/functions/call-route.ts plugins/functions/index.ts,plugins/functions/invoke-route.ts plugins/functions/plugin.yaml,plugins/functions/publish-route.ts plugins/functions/routes.ts,plugins/matchmaking/index.ts plugins/matchmaking/matchmaking-assign-route.ts,plugins/matchmaking/matchmaking-cancel-route.ts plugins/matchmaking/matchmaking-queue-route.ts,plugins/matchmaking/matchmaking-status-route.ts plugins/matchmaking/plugin.yaml,plugins/matchmaking/routes.ts plugins/sdk-typescript/src/index.ts,plugins/sdk-typescript/src/routes.ts scripts/typescript-worker.mjs,tests/functions/public-call.test.ts tests/plugin/manifest-helpers.test.ts,tests/plugin/manifest_helpers_test.go tests/plugin/sdk.test.ts
- `078d996` — 2026-11-27 02:13:18 +0330 — refactor(admin): centralize session storage and authenticated request headers
  Files: admin-nextjs/app/components/admin-chrome.tsx,admin-nextjs/app/components/admin-shell.tsx admin-nextjs/app/functions/[name]/page.tsx,admin-nextjs/app/functions/new/page.tsx admin-nextjs/app/functions/page.tsx,admin-nextjs/app/lib/admin-api.ts admin-nextjs/app/page.tsx,admin-nextjs/app/plugins/[plugin]/page.tsx admin-nextjs/app/workers/page.tsx,admin-nextjs/app/workflows/page.tsx tests/admin/admin-api.test.ts
- `b9e97c1` — 2026-11-27 02:12:14 +0330 — refactor(sdk): share manifest route and host context validation
  Files: plugins/sdk-typescript/src/create-function-handler.ts,plugins/sdk-typescript/src/create-host-handler.ts plugins/sdk-typescript/src/create-route-handler.ts,plugins/sdk-typescript/src/require-host-context.ts plugins/sdk-typescript/src/require-route.ts,tests/plugin/sdk-handler-contract.test.ts
- `ca0ae0a` — 2026-11-27 02:12:04 +0330 — ci: enforce mandatory TypeScript lint and format registry fields
  Files: .github/workflows/test.yml,internal/plugin/plugin.go
- `ba6fb7c` — 2026-11-27 02:11:31 +0330 — fix(plugin): defer durable writes until successful authorized execution
  Files: internal/plugin/legacy_storage_adapter.go,internal/plugin/plugin.go internal/plugin/registry_storage.go,internal/plugin/registry_workers.go tests/plugin/registry_storage_commit_test.go
- `974d1f2` — 2026-11-27 02:10:50 +0330 — fix(plugin): align Go and TypeScript worker JSON contracts
  Files: internal/plugin/typescript_worker.go,scripts/test-sandboxed-typescript-worker.sh scripts/typescript-worker.mjs,tests/plugin/typescript-worker.test.ts tests/plugin/typescript_worker_json_test.go
- `f213b2c` — 2026-11-27 02:09:28 +0330 — fix(plugin): authorize invocations before preparing durable storage
  Files: internal/plugin/registry_workers.go,tests/plugin/registry_storage_authorization_test.go
- `d37647d` — 2026-11-27 02:08:07 +0330 — refactor(plugin): separate registry and manifest responsibilities
  Files: internal/plugin/manifest.go,internal/plugin/manifest_sources.go internal/plugin/plugin.go,internal/plugin/registry_audit.go internal/plugin/registry_catalog.go,internal/plugin/registry_lifecycle.go internal/plugin/registry_permissions.go,internal/plugin/registry_storage.go internal/plugin/registry_versions.go,internal/plugin/registry_workers.go
- `5cafb71` — 2026-11-27 02:05:54 +0330 — refactor(operator): isolate HTTP sessions and synchronize store access
  Files: .github/workflows/test.yml,internal/app/operator_auth.go internal/operator/mfa.go,internal/operator/operator.go internal/operator/session.go,tests/app/operator_session_isolation_test.go tests/operator/session_isolation_test.go
- `fe2be80` — 2026-11-27 02:03:42 +0330 — refactor(app): centralize operator authorization and split function routes
  Files: internal/app/admin.go,internal/app/admin_functions.go internal/app/operator_auth.go,tests/app/operator_auth_test.go
- `4475c14` — 2026-11-26 22:31:50 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
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
