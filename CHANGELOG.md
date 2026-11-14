# Changelog

Automatically generated from commits on main.

## Unreleased
- `a16ca7c` — 2026-11-15 01:26:07 +0330 — protect plugin installation cleanup
  Files: internal/plugin/plugin.go,tests/plugin/plugin_test.go
- `e061bf3` — 2026-11-14 21:55:45 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `aa5c8d4` — 2026-11-15 01:25:29 +0330 — serialize plugin registration lifecycle
  Files: internal/plugin/plugin.go,tests/plugin/plugin_test.go
- `d176f97` — 2026-11-14 21:55:08 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `ddca07e` — 2026-11-15 01:24:51 +0330 — clear project plugin approvals on unregister
  Files: internal/plugin/plugin.go,tests/plugin/plugin_test.go
- `d15f461` — 2026-11-14 21:53:46 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `430b413` — 2026-11-15 01:23:29 +0330 — serialize plugin uninstall lifecycle
  Files: internal/plugin/plugin.go,tests/plugin/plugin_test.go
- `fd08c9d` — 2026-11-14 21:53:09 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `497ba7b` — 2026-11-15 01:22:50 +0330 — serialize plugin removal and pinning
  Files: internal/plugin/plugin.go,tests/plugin/plugin_test.go
- `52f7dfc` — 2026-11-14 21:52:14 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `e9f2ba3` — 2026-11-15 01:21:51 +0330 — protect plugin approval and artifact state
  Files: internal/plugin/plugin.go,tests/plugin/plugin_test.go
- `60dde13` — 2026-11-14 21:51:19 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `a95a7fa` — 2026-11-15 01:21:02 +0330 — protect plugin pin state
  Files: internal/plugin/plugin.go,tests/plugin/plugin_test.go
- `7c9b68f` — 2026-11-14 21:50:42 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `301d035` — 2026-11-15 01:20:23 +0330 — protect active plugin version state
  Files: internal/plugin/plugin.go,tests/plugin/plugin_test.go
- `8a207dd` — 2026-11-14 21:50:02 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `0a90c9c` — 2026-11-15 01:19:45 +0330 — protect plugin installation state
  Files: internal/plugin/plugin.go,tests/plugin/plugin_test.go
- `1bd54af` — 2026-11-14 21:49:22 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `4a17205` — 2026-11-15 01:19:04 +0330 — protect plugin authorization state
  Files: internal/plugin/plugin.go,tests/plugin/plugin_test.go
- `7ae4b79` — 2026-11-14 21:48:25 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `405c065` — 2026-11-15 01:18:03 +0330 — protect plugin audit history
  Files: internal/plugin/plugin.go,tests/plugin/plugin_test.go
- `1955f5a` — 2026-11-14 21:46:07 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `46b2e24` — 2026-11-15 01:15:51 +0330 — reject projectless plugin lifecycle checks
  Files: internal/plugin/plugin.go,tests/plugin/plugin_test.go
- `6259576` — 2026-11-14 21:45:16 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `f86e2a0` — 2026-11-15 01:14:59 +0330 — reject projectless plugin removal
  Files: internal/plugin/plugin.go,tests/plugin/plugin_test.go
- `7c67478` — 2026-11-14 21:44:37 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `74a7306` — 2026-11-15 01:14:15 +0330 — serialize plugin readiness unregistration
  Files: internal/plugin/plugin.go,tests/plugin/plugin_test.go
- `d44cff9` — 2026-11-14 21:43:51 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `1979883` — 2026-11-15 01:13:33 +0330 — serialize plugin lifecycle previews
  Files: internal/plugin/plugin.go,tests/plugin/plugin_test.go
- `1f182c7` — 2026-11-14 21:43:07 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `a065f6f` — 2026-11-15 01:12:46 +0330 — reject projectless lifecycle mutations
  Files: internal/plugin/plugin.go,tests/plugin/plugin_test.go
- `a447cde` — 2026-11-14 21:42:27 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `2c53022` — 2026-11-15 01:12:09 +0330 — reject projectless plugin installation
  Files: internal/plugin/plugin.go,tests/plugin/plugin_test.go
- `182a7ae` — 2026-11-14 21:41:43 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `55aa276` — 2026-11-15 01:11:26 +0330 — make invocation credentials concurrency safe
  Files: internal/plugin/broker.go,tests/plugin/broker_test.go
- `03e8724` — 2026-11-14 21:40:05 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `5011e16` — 2026-11-15 01:09:46 +0330 — make plugin readiness concurrency safe
  Files: internal/plugin/plugin.go,tests/plugin/plugin_test.go
- `c74b214` — 2026-11-14 21:38:22 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `17803ab` — 2026-11-15 01:08:06 +0330 — make plugin durable data concurrency safe
  Files: internal/plugin/plugin.go,tests/plugin/plugin_test.go
- `9981393` — 2026-11-14 21:37:30 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `ebf0e13` — 2026-11-15 01:07:18 +0330 — make websocket send queue concurrency safe
  Files: internal/plugin/websocket_queue.go,tests/plugin/websocket_queue_test.go
- `8b77ff4` — 2026-11-14 21:36:39 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `543ea63` — 2026-11-15 01:06:25 +0330 — make caller leases single use under concurrency
  Files: internal/project/transaction_context.go,tests/project/transaction_context_test.go
- `ed20fc6` — 2026-11-14 21:35:50 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `ab714e3` — 2026-11-15 01:05:36 +0330 — make project store concurrency safe
  Files: internal/project/project.go,tests/project/project_test.go
- `5631278` — 2026-11-14 21:34:48 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `0084160` — 2026-11-15 01:04:36 +0330 — reject unsafe restore project names
  Files: scripts/restore-compose.sh,tests/deploy/compose_restore_test.go
- `f2e48d1` — 2026-11-14 21:33:59 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `6b97bb5` — 2026-11-15 01:03:45 +0330 — verify release images by digest
  Files: .github/workflows/release-backend.yml,tests/deploy/release_image_verification_test.go tests/deploy/release_workflow_test.go
- `0da2616` — 2026-11-14 21:32:53 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
