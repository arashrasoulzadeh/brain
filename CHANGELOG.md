# Changelog

Automatically generated from commits on main.

## Unreleased
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
- `e466e09` — 2026-11-15 02:22:03 +0330 — validate room connection ticket context
  Files: plugins/rooms/connection-ticket-consumer.ts,tests/rooms/rooms.test.ts
- `2dd5de8` — 2026-11-14 22:51:48 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `cb290d6` — 2026-11-15 02:21:37 +0330 — harden subscription authorization boundaries
  Files: plugins/subscriptions/subscription-authorization.ts,plugins/subscriptions/subscription-owner-invalidator.ts plugins/subscriptions/subscription-ticket-validator.ts,tests/subscriptions/subscriptions.test.ts
- `f2196af` — 2026-11-14 22:49:52 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `697474f` — 2026-11-15 02:19:37 +0330 — validate function publication rollback inputs
  Files: plugins/functions/function-publication-rollback.ts,tests/functions/functions.test.ts
- `c76cd21` — 2026-11-14 22:49:18 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `781d419` — 2026-11-15 02:19:00 +0330 — close numeric direct https SSRF bypasses
  Files: plugins/functions/invoke-direct-https.ts,tests/functions/functions.test.ts
- `864ee7f` — 2026-11-14 22:48:19 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `ab8f75f` — 2026-11-15 02:18:03 +0330 — serialize concurrent idempotent workflows
  Files: plugins/functions/run-idempotent-workflow.ts,tests/functions/functions.test.ts
- `913c0b0` — 2026-11-14 22:47:19 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `fcca6c6` — 2026-11-15 02:17:02 +0330 — harden sandbox path traversal validation
  Files: plugins/functions/reject-sandbox-escape.ts,tests/functions/functions.test.ts
- `c9a7ffb` — 2026-11-14 22:46:35 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `d8763d5` — 2026-11-15 02:16:17 +0330 — validate project identity for function job leases
  Files: plugins/functions/job-lease-acquirer.ts,plugins/functions/job-lease-renewer.ts tests/functions/functions.test.ts
- `1b20cc1` — 2026-11-14 22:45:39 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `6a82173` — 2026-11-15 02:15:23 +0330 — validate file object identities at access boundaries
  Files: plugins/files/authorized-upload-writer.ts,plugins/files/file-access-deleter.ts plugins/files/file-access-reader.ts,plugins/files/pending-upload-publisher.ts plugins/files/pending-upload-reader.ts,tests/files/files.test.ts
- `992991f` — 2026-11-14 22:44:07 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `ee5a470` — 2026-11-15 02:13:48 +0330 — validate project identity in auth token consumers
  Files: plugins/auth/email-otp-verifier.ts,plugins/auth/recovery-token-consumer.ts plugins/auth/sms-otp-verifier.ts,tests/auth/auth.test.ts
- `f290136` — 2026-11-14 22:43:06 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `1e4acf0` — 2026-11-15 02:12:46 +0330 — audit every plugin manifest boundary
  Files: tests/plugin/all_manifests_test.go
- `143c435` — 2026-11-14 22:42:13 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `33a5162` — 2026-11-15 02:11:52 +0330 — validate dashboard project ids before html rendering
  Files: docs/admins.md,docs/developers.md plugins/dashboard/custom-admin-view.ts,plugins/dashboard/plugin.yaml tests/dashboard/dashboard.test.ts
- `d26605e` — 2026-11-14 22:37:26 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `6ffa6ed` — 2026-11-15 02:07:11 +0330 — verify single use commit credentials
  Files: roadmap_backend.md
- `063afca` — 2026-11-14 22:37:09 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
