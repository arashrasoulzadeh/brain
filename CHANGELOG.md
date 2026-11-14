# Changelog

Automatically generated from commits on main.

## Unreleased
- `6ca3489` — 2026-11-15 01:43:14 +0330 — add project scoped websocket cleanup
  Files: internal/plugin/websocket_registry.go,tests/plugin/websocket_registry_test.go
- `6f4abd2` — 2026-11-14 22:11:57 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `778f6dd` — 2026-11-15 01:41:41 +0330 — bound plugin transport payloads
  Files: internal/plugin/http_envelope.go,internal/plugin/websocket_message.go tests/plugin/http_envelope_test.go,tests/plugin/websocket_message_test.go
- `e0206bf` — 2026-11-14 22:10:58 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `a48d2c9` — 2026-11-15 01:40:46 +0330 — require plugin worker keys
  Files: internal/plugin/supervisor.go,tests/plugin/supervisor_test.go
- `c57ac11` — 2026-11-14 22:10:25 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `8e4019b` — 2026-11-15 01:40:10 +0330 — fail closed for nil caller context
  Files: internal/project/transaction_context.go,tests/project/transaction_context_test.go
- `351cb47` — 2026-11-14 22:09:32 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `554707a` — 2026-11-15 01:39:19 +0330 — reject nil plugin workers
  Files: internal/plugin/supervisor.go,tests/plugin/supervisor_test.go
- `0ea39c0` — 2026-11-14 22:09:03 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `e046094` — 2026-11-15 01:38:50 +0330 — fail closed without plugin worker
  Files: internal/plugin/worker.go,tests/plugin/worker_test.go
- `4cd0809` — 2026-11-14 22:08:22 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `a2bc42c` — 2026-11-15 01:38:09 +0330 — reject trailing handshake data
  Files: internal/plugin/external_http.go,tests/plugin/external_http_test.go
- `1fddf06` — 2026-11-14 22:07:44 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `2f1188f` — 2026-11-15 01:37:31 +0330 — bind external handshake authentication
  Files: internal/plugin/external.go,tests/plugin/external_test.go
- `2efbc61` — 2026-11-14 22:07:10 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `9e81357` — 2026-11-15 01:36:57 +0330 — reject projectless worker stops
  Files: internal/plugin/supervisor.go,tests/plugin/supervisor_test.go
- `01a4011` — 2026-11-14 22:06:23 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `c3faacf` — 2026-11-15 01:36:10 +0330 — fail closed without project store
  Files: internal/project/http.go,tests/project/http_test.go
- `b06e300` — 2026-11-14 22:05:32 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `01638f9` — 2026-11-15 01:35:18 +0330 — fail closed when plugin registry unavailable
  Files: internal/plugin/broker.go,tests/plugin/broker_test.go
- `7d494af` — 2026-11-14 22:04:08 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `a795ca9` — 2026-11-15 01:33:49 +0330 — protect broker manifest authorization
  Files: internal/plugin/broker.go,tests/plugin/plugin_test.go
- `c972917` — 2026-11-14 22:03:15 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `fd94528` — 2026-11-15 01:33:01 +0330 — protect plugin disable previews
  Files: internal/plugin/plugin.go,tests/plugin/plugin_test.go
- `8d27d6f` — 2026-11-14 22:02:37 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `4741525` — 2026-11-15 01:32:24 +0330 — enforce plugin project boundaries
  Files: internal/plugin/plugin.go,tests/plugin/plugin_test.go
- `87c5a5b` — 2026-11-14 22:01:40 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `360fb3b` — 2026-11-15 01:31:27 +0330 — serialize plugin unregistration state
  Files: internal/plugin/plugin.go,tests/plugin/plugin_test.go
- `ef39714` — 2026-11-14 22:00:59 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `a8f8efe` — 2026-11-15 01:30:45 +0330 — protect plugin dependency metadata
  Files: internal/plugin/plugin.go,tests/plugin/plugin_test.go
- `2e2105f` — 2026-11-14 22:00:06 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `446e983` — 2026-11-15 01:29:53 +0330 — serialize plugin grant state
  Files: internal/plugin/plugin.go,tests/plugin/plugin_test.go
- `bf216da` — 2026-11-14 21:59:29 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `e4b4562` — 2026-11-15 01:29:16 +0330 — serialize plugin lifecycle state
  Files: internal/plugin/plugin.go,tests/plugin/plugin_test.go
- `521c5f0` — 2026-11-14 21:58:34 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `bdfe0d5` — 2026-11-15 01:27:28 +0330 — serialize plugin install unregistration
  Files: internal/plugin/plugin.go,tests/plugin/plugin_test.go
- `c2ca14e` — 2026-11-14 21:57:05 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `5a21fc6` — 2026-11-15 01:26:47 +0330 — serialize plugin artifact cleanup
  Files: internal/plugin/plugin.go,tests/plugin/plugin_test.go
- `0a89897` — 2026-11-14 21:56:27 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
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
