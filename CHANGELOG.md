# Changelog

Automatically generated from commits on main.

## Unreleased
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
- `f199861` — 2026-11-15 12:07:27 +0330 — reject empty compose services
  Files: internal/deploy/compose.go,tests/deploy/compose_test.go
- `3824c0a` — 2026-11-15 08:37:10 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `6cf468c` — 2026-11-15 12:06:53 +0330 — validate deployment project identity
  Files: internal/deploy/sequence.go,tests/deploy/sequence_test.go
- `d6660fa` — 2026-11-15 08:36:34 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `bcdb15d` — 2026-11-15 12:06:20 +0330 — require digest pinned release images
  Files: internal/deploy/manifest.go,tests/deploy/healthy_manifest_test.go tests/deploy/manifest_test.go,tests/deploy/rollback_test.go
- `5ec4f7d` — 2026-11-15 08:35:39 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `41710ac` — 2026-11-15 12:05:12 +0330 — bound measured capacity memory
  Files: internal/deploy/capacity_profile.go,tests/deploy/capacity_profile_test.go
- `efd8b75` — 2026-11-14 23:00:04 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `ef5fe97` — 2026-11-15 02:29:50 +0330 — validate caller project context
  Files: internal/project/transaction_context.go,tests/project/transaction_context_test.go
- `9b74e34` — 2026-11-14 22:59:23 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `132673c` — 2026-11-15 02:29:05 +0330 — reap terminated typescript workers
  Files: internal/plugin/typescript_worker.go,tests/plugin/typescript_worker_test.go
- `1efb3ad` — 2026-11-14 22:58:37 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `abe7af7` — 2026-11-15 02:28:26 +0330 — validate supervisor project teardown
  Files: internal/plugin/supervisor.go,tests/plugin/supervisor_test.go
- `cf64875` — 2026-11-14 22:58:07 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `958cd76` — 2026-11-15 02:27:51 +0330 — harden websocket project boundaries
  Files: internal/plugin/websocket_endpoint.go,internal/plugin/websocket_registry.go tests/plugin/websocket_endpoint_test.go,tests/plugin/websocket_registry_test.go
- `f2cbb2b` — 2026-11-14 22:57:28 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `9307c44` — 2026-11-15 02:27:11 +0330 — enforce host project identity format
  Files: internal/plugin/broker.go,internal/plugin/external.go internal/plugin/http_envelope.go,internal/plugin/project_identity.go internal/plugin/websocket_message.go,internal/plugin/worker.go tests/plugin/broker_test.go
- `89aebb8` — 2026-11-14 22:56:37 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `83e5f73` — 2026-11-15 02:26:24 +0330 — enforce canonical project identifiers
  Files: internal/project/project.go,tests/project/project_test.go
- `d36079c` — 2026-11-14 22:55:45 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `62ef061` — 2026-11-15 02:25:32 +0330 — bound analytics event values
  Files: plugins/analytics/validate-event.ts,tests/analytics/analytics.test.ts
- `6a378b8` — 2026-11-14 22:55:13 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `0c9c4a6` — 2026-11-15 02:24:58 +0330 — validate table authorization boundaries
  Files: plugins/tables/row-policy-checker.ts,plugins/tables/row-policy-granter.ts plugins/tables/scoped-transaction-change-reader.ts,plugins/tables/scoped-transaction-change-writer.ts plugins/tables/table-exposure-reader.ts,tests/tables/tables.test.ts
- `fd6ec24` — 2026-11-14 22:54:18 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `774a9c2` — 2026-11-15 02:24:05 +0330 — validate arena identity and result time
  Files: plugins/arena/arena-state.ts,plugins/arena/validate-result.ts tests/arena/arena.test.ts
- `d7c7446` — 2026-11-14 22:53:41 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `3897676` — 2026-11-15 02:23:27 +0330 — harden invite validation boundaries
  Files: plugins/invites/invite-accepter.ts,plugins/invites/invite-creator.ts tests/invites/invites.test.ts
- `abf0585` — 2026-11-14 22:53:00 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `e89084a` — 2026-11-15 02:22:47 +0330 — validate matchmaking scope boundaries
  Files: plugins/matchmaking/assignment-creator.ts,plugins/matchmaking/ticket-canceller.ts plugins/matchmaking/ticket-enqueuer.ts,plugins/matchmaking/ticket-matcher.ts plugins/matchmaking/ticket-reader.ts,plugins/matchmaking/ticket-requeuer.ts tests/matchmaking/matchmaking.test.ts
- `bcee2c9` — 2026-11-14 22:52:15 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
