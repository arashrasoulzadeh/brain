# Changelog

Automatically generated from commits on main.

## Unreleased
- `9146582` — 2026-11-14 16:13:56 +0330 — refactor: split workflow policy helpers
  Files: plugins/functions/diagnose-source.ts,plugins/functions/evaluate-workflow-condition.ts plugins/functions/index.ts,plugins/functions/plugin.yaml plugins/functions/redact-run-status.ts,plugins/functions/run-idempotent-workflow.ts plugins/functions/validate-workflow-graph.ts
- `9a1018e` — 2026-11-14 12:43:09 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `b6ef221` — 2026-11-14 16:12:54 +0330 — refactor: isolate function publication stores
  Files: plugins/functions/create-approval-store.ts,plugins/functions/create-function-draft-store.ts plugins/functions/create-function-publication-store.ts,plugins/functions/create-shared-package-store.ts plugins/functions/index.ts,plugins/functions/plugin.yaml
- `2cbd143` — 2026-11-14 12:42:16 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `7d94ffb` — 2026-11-14 16:12:03 +0330 — refactor: isolate function lifecycle stores
  Files: plugins/functions/create-job-lease-store.ts,plugins/functions/create-record-change-trigger-store.ts plugins/functions/create-schedule-store.ts,plugins/functions/index.ts plugins/functions/plugin.yaml
- `415b273` — 2026-11-14 12:41:28 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `4af8671` — 2026-11-14 16:11:15 +0330 — refactor: isolate invocation stores
  Files: plugins/functions/create-invocation-commit-store.ts,plugins/functions/create-invocation-effects.ts plugins/functions/index.ts,plugins/functions/plugin.yaml
- `1dc3601` — 2026-11-14 12:40:34 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `1e381c5` — 2026-11-14 16:10:20 +0330 — refactor: isolate function retry policies
  Files: plugins/functions/commit-with-grant.ts,plugins/functions/index.ts plugins/functions/plugin.yaml,plugins/functions/run-capped-retries.ts
- `4324839` — 2026-11-14 12:39:45 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `7bed196` — 2026-11-14 16:09:31 +0330 — refactor: isolate sandbox boundary checks
  Files: plugins/functions/index.ts,plugins/functions/plugin.yaml plugins/functions/reject-sandbox-escape.ts,plugins/functions/run-with-timeout.ts plugins/functions/validate-execution-limits.ts,plugins/functions/validate-sandbox-capabilities.ts plugins/functions/validate-wasi-imports.ts
- `d8d665f` — 2026-11-14 12:38:51 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `c4c0f1b` — 2026-11-14 16:08:37 +0330 — refactor: split function boundary validators
  Files: plugins/functions/diagnostic-cap.ts,plugins/functions/function-input-validation.ts plugins/functions/index.ts,plugins/functions/plugin.yaml plugins/functions/write-intent-validation.ts
- `5f2241a` — 2026-11-14 12:37:59 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `eac7662` — 2026-11-14 16:07:44 +0330 — refactor: split invocation validation helpers
  Files: plugins/functions/execute-fixture.ts,plugins/functions/index.ts plugins/functions/invocation-types.ts,plugins/functions/invoke-direct-https.ts plugins/functions/parse-invocation-result.ts,plugins/functions/plugin.yaml plugins/functions/validate-invocation-result.ts
- `d225da5` — 2026-11-14 12:36:42 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `2a8c8c6` — 2026-11-14 16:06:25 +0330 — refactor: split function compiler validation
  Files: plugins/functions/index.ts,plugins/functions/plugin.yaml plugins/functions/validate-build-overrides.ts,plugins/functions/validate-function-limits.ts plugins/functions/validate-sdk-imports.ts,plugins/functions/validate-source-project.ts plugins/functions/wasm-build-settings.ts
- `716e06e` — 2026-11-14 12:35:43 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `ba2898c` — 2026-11-14 16:05:26 +0330 — refactor: isolate auth rate limiters
  Files: plugins/auth/account-rate-limiter.ts,plugins/auth/identity-linker.ts plugins/auth/index.ts,plugins/auth/otp-limiter.ts plugins/auth/plugin.yaml,plugins/auth/project-rate-limiter.ts
- `9b761c8` — 2026-11-14 12:34:32 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `51b8cdc` — 2026-11-14 16:04:19 +0330 — refactor: split auth utility functions
  Files: plugins/auth/create-pkce-challenge.ts,plugins/auth/index.ts plugins/auth/normalize-phone-number.ts,plugins/auth/oidc-types.ts plugins/auth/plugin.yaml,plugins/auth/public-auth-error.ts plugins/auth/require-recent-authentication.ts,plugins/auth/validate-delivery-config.ts plugins/auth/validate-oidc-token.ts,plugins/auth/verify-pkce.ts
- `61084c0` — 2026-11-14 12:33:27 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `53911f7` — 2026-11-14 16:03:02 +0330 — refactor: decompose tic-tac-toe room helpers
  Files: plugins/rooms/authorize-tic-tac-toe-command.ts,plugins/rooms/authorized-tic-tac-toe-state.ts plugins/rooms/command-deduplicator.ts,plugins/rooms/create-tic-tac-toe-coordinator.ts plugins/rooms/hidden-tic-tac-toe-view.ts,plugins/rooms/index.ts plugins/rooms/plugin.yaml,plugins/rooms/publish-tic-tac-toe-rules.ts plugins/rooms/reconnect-tic-tac-toe.ts,plugins/rooms/reconstruct-tic-tac-toe-history.ts plugins/rooms/serialize-tic-tac-toe-state.ts,plugins/rooms/validate-tic-tac-toe-command-payload.ts
- `8b334d5` — 2026-11-14 12:24:21 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `0e86dbc` — 2026-11-14 15:54:01 +0330 — refactor: split room authorization boundaries
  Files: plugins/rooms/authorize-room-command.ts,plugins/rooms/authorized-room-snapshot.ts plugins/rooms/create-room-directory.ts,plugins/rooms/index.ts plugins/rooms/plugin.yaml
- `d53136a` — 2026-11-14 12:23:11 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `9152603` — 2026-11-14 15:52:50 +0330 — refactor: split tables index responsibilities
  Files: plugins/tables/create-api-change-capture.ts,plugins/tables/create-direct-sql-change-capture.ts plugins/tables/create-schema-publication-audit.ts,plugins/tables/create-transactional-change-capture.ts plugins/tables/index.ts,plugins/tables/plan-schema-change.ts plugins/tables/plugin.yaml,plugins/tables/preview-schema-sql.ts plugins/tables/require-recent-schema-auth.ts,plugins/tables/schema-types.ts
- `fb6503f` — 2026-11-14 12:21:06 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `5242d81` — 2026-11-14 15:51:02 +0330 — refactor: split auth state responsibilities
  Files: plugins/auth/account-project.ts,plugins/auth/auth-state.ts plugins/auth/create-auth-state.ts,plugins/auth/index.ts plugins/auth/new-session.ts,plugins/auth/password-hash.ts plugins/auth/plugin.yaml,plugins/auth/project-user-id.ts plugins/auth/validate-account.ts
- `e2c3198` — 2026-11-14 15:50:53 +0330 — feat: embed custom views in restricted frames
  Files: plugins/dashboard/index.ts,plugins/dashboard/plugin.yaml plugins/dashboard/restricted-frame.ts,roadmap.md tests/dashboard/dashboard.test.ts
- `b927fb1` — 2026-11-13 13:49:57 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `017c178` — 2026-11-13 17:19:34 +0330 — feat: isolate custom admin view origin
  Files: plugins/dashboard/custom-view-origin.ts,plugins/dashboard/index.ts plugins/dashboard/plugin.yaml,roadmap.md tests/dashboard/dashboard.test.ts
- `093dac3` — 2026-11-13 13:49:07 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `0714be7` — 2026-11-13 17:18:44 +0330 — feat: add custom typescript admin view
  Files: plugins/dashboard/custom-admin-view.ts,plugins/dashboard/index.ts plugins/dashboard/plugin.yaml,roadmap.md tests/dashboard/dashboard.test.ts
- `5902a49` — 2026-11-13 13:48:28 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `1730e10` — 2026-11-13 17:17:56 +0330 — feat: audit dashboard view actions
  Files: plugins/dashboard/view-replacement-registry.ts,roadmap.md tests/dashboard/dashboard.test.ts
- `39b996c` — 2026-11-13 13:47:18 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `12a7072` — 2026-11-13 17:16:49 +0330 — feat: add dashboard csrf validation
  Files: plugins/dashboard/index.ts,plugins/dashboard/plugin.yaml plugins/dashboard/validate-csrf-token.ts,roadmap.md tests/dashboard/dashboard.test.ts
- `454a309` — 2026-11-13 13:46:21 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `543830d` — 2026-11-13 17:15:59 +0330 — feat: authorize dashboard view actions
  Files: plugins/dashboard/view-replacement-registry.ts,roadmap.md tests/dashboard/dashboard.test.ts
- `51f4aff` — 2026-11-13 13:45:40 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `bf450ed` — 2026-11-13 17:15:20 +0330 — feat: isolate dashboard view submissions
  Files: plugins/dashboard/view-replacement-registry.ts,roadmap.md tests/dashboard/dashboard.test.ts
- `412d004` — 2026-11-13 13:44:57 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `4abe251` — 2026-11-13 17:14:20 +0330 — test: verify duplicate view replacement rejection
  Files: roadmap.md
- `2cfe5ce` — 2026-11-13 13:43:48 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `7a1db4a` — 2026-11-13 17:13:25 +0330 — feat: activate scoped admin view replacements
  Files: plugins/dashboard/index.ts,plugins/dashboard/plugin.yaml plugins/dashboard/view-replacement-registry.ts,roadmap.md tests/dashboard/dashboard.test.ts
- `1665987` — 2026-11-13 13:42:56 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `f8d1805` — 2026-11-13 17:12:35 +0330 — feat: add declarative dashboard view contributions
  Files: plugins/dashboard/index.ts,plugins/dashboard/plugin-health-contribution.ts plugins/dashboard/plugin.yaml,plugins/dashboard/project-overview-contribution.ts roadmap.md,tests/dashboard/dashboard.test.ts
- `7e0c0b3` — 2026-11-13 13:42:07 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `2d18d15` — 2026-11-13 17:11:42 +0330 — feat: render host-owned admin fixture
  Files: plugins/dashboard/host-owned-admin-view.ts,plugins/dashboard/index.ts plugins/dashboard/plugin.yaml,roadmap.md tests/dashboard/dashboard.test.ts
