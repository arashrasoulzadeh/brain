# Changelog

Automatically generated from commits on main.

## Unreleased
- `9a3d085` — 2026-11-15 01:49:09 +0330 — close sockets during project recovery
  Files: internal/plugin/plugin.go,tests/plugin/plugin_test.go
- `038867f` — 2026-11-14 22:18:47 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `7d91db3` — 2026-11-15 01:48:35 +0330 — recover project plugin workers
  Files: internal/plugin/plugin.go,tests/plugin/plugin_test.go
- `fd0b797` — 2026-11-14 22:17:54 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `2492765` — 2026-11-15 01:47:37 +0330 — stop project workers on uninstall
  Files: internal/plugin/plugin.go,tests/plugin/plugin_test.go
- `7fa32a1` — 2026-11-14 22:17:11 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `5a921bc` — 2026-11-15 01:46:57 +0330 — invalidate plugin lifecycle previews
  Files: internal/plugin/plugin.go,tests/plugin/plugin_test.go
- `605a9c9` — 2026-11-14 22:16:26 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `c383e15` — 2026-11-15 01:46:13 +0330 — finish socket cleanup after stop errors
  Files: internal/plugin/plugin.go,tests/plugin/plugin_test.go
- `9235e5d` — 2026-11-14 22:15:29 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `a926504` — 2026-11-15 01:45:15 +0330 — close all project sockets after errors
  Files: internal/plugin/websocket_registry.go,tests/plugin/websocket_registry_test.go
- `f9ea4df` — 2026-11-14 22:14:54 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `3ad9f9b` — 2026-11-15 01:44:41 +0330 — close project sockets on disable
  Files: internal/plugin/plugin.go,tests/plugin/plugin_test.go
- `b0cefc8` — 2026-11-14 22:14:19 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `b40c59d` — 2026-11-15 01:44:07 +0330 — close project sockets on revocation
  Files: internal/plugin/plugin.go,tests/plugin/plugin_test.go
- `c375b09` — 2026-11-14 22:13:29 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
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
