# Changelog

Automatically generated from commits on main.

## Unreleased
- `57724d8` — 2026-11-15 12:29:26 +0330 — record verified recovery gate
  Files: roadmap_backend.md
- `9ced1dd` — 2026-11-15 08:59:12 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `c84040f` — 2026-11-15 12:28:59 +0330 — harden malformed authentication digests
  Files: plugins/auth/email-otp-verifier.ts,plugins/auth/recovery-token-consumer.ts plugins/auth/sms-otp-verifier.ts,tests/auth/auth.test.ts
- `06167ad` — 2026-11-15 08:58:11 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `1dae3fd` — 2026-11-15 12:27:58 +0330 — record verified backend release gates
  Files: roadmap_backend.md
- `ccba3f7` — 2026-11-15 08:57:42 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `dbb1983` — 2026-11-15 12:27:30 +0330 — complete project boundary audit pass
  Files: plugins/functions/record-trigger-matcher.ts,plugins/functions/shared-helper-publisher.ts plugins/matchmaking/matchmaking-config-reader.ts,plugins/matchmaking/matchmaking-config-status.ts plugins/rooms/authorized-room-snapshot.ts,plugins/rooms/authorized-tic-tac-toe-state.ts plugins/rooms/reconnect-tic-tac-toe.ts,tests/matchmaking/matchmaking.test.ts tests/rooms/rooms.test.ts
- `2b97bcf` — 2026-11-15 08:56:14 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `873210b` — 2026-11-15 12:26:00 +0330 — harden function state project boundaries
  Files: plugins/functions/authorize-project-workflow-action.ts,plugins/functions/function-draft-reader.ts plugins/functions/function-publication-reader.ts,plugins/functions/shared-package-reader.ts tests/functions/functions.test.ts,tests/functions/workflow-components.test.ts
- `f9d36eb` — 2026-11-15 08:55:13 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `b49cb7d` — 2026-11-15 12:24:56 +0330 — mark operator session isolation complete
  Files: roadmap_backend.md
- `df33381` — 2026-11-15 08:54:56 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `a51ac81` — 2026-11-15 12:24:41 +0330 — verify operator session isolation
  Files: tests/auth/auth.test.ts
- `8084f95` — 2026-11-15 08:54:19 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `4680c1f` — 2026-11-15 12:24:06 +0330 — harden data and deployment project boundaries
  Files: plugins/functions/authorize-deployment-plugin.ts,plugins/tables/api-change-writer.ts plugins/tables/plan-schema-change.ts,plugins/tables/sql-change-writer.ts plugins/tables/transaction-change-writer.ts,tests/functions/workflow-components.test.ts tests/tables/tables.test.ts
- `38c9c5d` — 2026-11-15 08:53:21 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `dd932b3` — 2026-11-15 12:23:09 +0330 — harden dashboard project boundaries
  Files: plugins/dashboard/auth-settings-page.ts,plugins/dashboard/browser-quickstart.ts plugins/dashboard/form-audit-recorder.ts,plugins/dashboard/host-owned-admin-view.ts plugins/dashboard/plugin-health-contribution.ts,plugins/dashboard/project-overview-contribution.ts plugins/dashboard/record-browser-page.ts,plugins/dashboard/record-edit-page.ts plugins/dashboard/signup-settings-page.ts,plugins/dashboard/table-management-page.ts tests/dashboard/dashboard.test.ts
- `39688df` — 2026-11-15 08:52:04 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `4014f79` — 2026-11-15 12:21:51 +0330 — validate auth state project boundaries
  Files: plugins/auth/account-project.ts,plugins/auth/index.ts tests/auth/auth.test.ts
- `3ea2358` — 2026-11-15 08:51:07 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `f521c5a` — 2026-11-15 12:20:52 +0330 — validate plugin project boundaries
  Files: plugins/arena/arena-start.ts,plugins/dashboard/file-management-page.ts tests/arena/arena.test.ts,tests/dashboard/dashboard.test.ts
- `6561360` — 2026-11-15 08:49:30 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `df4bb08` — 2026-11-15 12:19:16 +0330 — enforce canonical plugin project identities
  Files: internal/plugin/plugin.go,tests/plugin/plugin_test.go
- `5b07349` — 2026-11-15 08:47:37 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `02710a4` — 2026-11-15 12:17:22 +0330 — mount audited admin server boundary
  Files: internal/app/app.go,tests/app/admin_test.go
- `41894f7` — 2026-11-15 08:47:04 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `9137037` — 2026-11-15 12:16:47 +0330 — wire audited admin HTTP boundary
  Files: internal/app/admin.go,internal/operator/context.go tests/app/admin_test.go
- `fd7ca9c` — 2026-11-15 08:46:12 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `487b4f9` — 2026-11-15 12:15:57 +0330 — audit plugin permission actions
  Files: internal/admin/lifecycle.go,tests/admin/lifecycle_test.go
- `8e5eb31` — 2026-11-15 08:45:14 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `b03ada6` — 2026-11-15 12:15:00 +0330 — audit migration and rollback actions
  Files: internal/admin/lifecycle.go,tests/admin/lifecycle_test.go
- `33f0aba` — 2026-11-15 08:44:31 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `355ec53` — 2026-11-15 12:14:17 +0330 — audit plugin recovery actions
  Files: internal/admin/lifecycle.go,tests/admin/lifecycle_test.go
- `b0150a4` — 2026-11-15 08:43:56 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `8ef5720` — 2026-11-15 12:13:42 +0330 — audit durable data deletion
  Files: internal/admin/lifecycle.go,tests/admin/lifecycle_test.go
- `c6ee526` — 2026-11-15 08:43:25 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `a5a65c7` — 2026-11-15 12:13:11 +0330 — add audited admin lifecycle service
  Files: internal/admin/lifecycle.go,tests/admin/lifecycle_test.go
- `8b1e024` — 2026-11-15 08:42:23 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `4658128` — 2026-11-15 12:12:09 +0330 — add audited operator action boundary
  Files: internal/operator/operator.go,tests/operator/operator_test.go
- `16ec3fc` — 2026-11-15 08:41:24 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `199a380` — 2026-11-15 12:11:11 +0330 — separate sensitive operator setup
  Files: internal/operator/mfa.go,internal/operator/operator.go tests/operator/operator_test.go
- `34836ca` — 2026-11-15 08:40:21 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `57ee8e1` — 2026-11-15 12:10:07 +0330 — reject compose host isolation bypasses
  Files: internal/deploy/compose.go,tests/deploy/compose_test.go
- `7054dcd` — 2026-11-15 08:39:45 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `427e5fa` — 2026-11-15 12:09:31 +0330 — validate backup age timestamps
  Files: scripts/check-backup-age.sh,tests/deploy/postgres_backup_verify_test.go
- `ae7a3c0` — 2026-11-15 08:38:53 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `d246728` — 2026-11-15 12:08:40 +0330 — validate rehearsal compose projects
  Files: scripts/clean-environment-rehearsal.sh,tests/deploy/clean_environment_script_test.go
- `a4e3ebc` — 2026-11-15 08:38:23 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `d58d684` — 2026-11-15 12:08:06 +0330 — validate compose project names
  Files: scripts/rollback-compose.sh,scripts/start-compose-no-build.sh tests/deploy/compose_script_test.go,tests/deploy/rollback_compose_test.go
- `b227c0b` — 2026-11-15 08:37:40 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
