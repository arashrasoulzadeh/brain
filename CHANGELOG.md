# Changelog

Automatically generated from commits on main.

## Unreleased
- `0084160` — 2026-11-15 01:04:36 +0330 — reject unsafe restore project names
  Files: scripts/restore-compose.sh,tests/deploy/compose_restore_test.go
- `f2e48d1` — 2026-11-14 21:33:59 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `6b97bb5` — 2026-11-15 01:03:45 +0330 — verify release images by digest
  Files: .github/workflows/release-backend.yml,tests/deploy/release_image_verification_test.go tests/deploy/release_workflow_test.go
- `0da2616` — 2026-11-14 21:32:53 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `5d49433` — 2026-11-15 01:02:40 +0330 — reject symlinked plugin source directories
  Files: internal/plugin/plugin.go,tests/plugin/plugin_test.go
- `439b8e7` — 2026-11-14 21:32:20 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `d299283` — 2026-11-15 01:01:59 +0330 — reject symlinked plugin sources
  Files: internal/plugin/plugin.go,tests/plugin/plugin_test.go
- `26693dd` — 2026-11-14 21:31:15 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `66345a9` — 2026-11-15 01:01:00 +0330 — require declared route capabilities
  Files: internal/plugin/plugin.go,plugins/auth/plugin.yaml tests/plugin/discovery_test.go,tests/plugin/plugin_test.go
- `5c92d2d` — 2026-11-14 21:29:35 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `9368677` — 2026-11-15 00:59:23 +0330 — require trusted caller for project middleware
  Files: internal/project/http.go,tests/project/http_test.go
- `25b0285` — 2026-11-14 21:28:46 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `362ae3b` — 2026-11-15 00:58:32 +0330 — align dashboard bootstrap with plugin manifest
  Files: internal/app/dashboard.go,tests/app/dashboard_bootstrap_test.go
- `d168679` — 2026-11-14 21:28:16 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `bec63b1` — 2026-11-15 00:58:03 +0330 — remove completed revocation gate
  Files: roadmap_backend.md
- `4df3d1f` — 2026-11-14 21:27:55 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `3dbfa92` — 2026-11-15 00:57:41 +0330 — remove completed backend process laws
  Files: roadmap_backend.md
- `c7d8d55` — 2026-11-14 21:27:23 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `784d958` — 2026-11-15 00:57:11 +0330 — reject ambiguous plugin routes
  Files: internal/plugin/plugin.go,tests/plugin/plugin_test.go
- `a811db5` — 2026-11-14 21:26:49 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `b41dfeb` — 2026-11-15 00:55:19 +0330 — bind websocket endpoint to host project
  Files: internal/plugin/websocket_endpoint.go,tests/plugin/websocket_endpoint_test.go
- `7044e28` — 2026-11-14 21:24:39 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `2388a95` — 2026-11-15 00:54:26 +0330 — reject duplicate manifest sources
  Files: internal/plugin/plugin.go,tests/plugin/plugin_test.go
- `ca2a89f` — 2026-11-14 21:24:15 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `0c1f923` — 2026-11-15 00:53:53 +0330 — reject manifest source traversal at load time
  Files: internal/plugin/plugin.go,tests/plugin/plugin_test.go
- `82f2747` — 2026-11-14 21:20:31 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `13f042c` — 2026-11-15 00:50:17 +0330 — wire full recovery rehearsal in CI
  Files: .github/workflows/test.yml,tests/deploy/clean_environment_script_test.go
- `450efb4` — 2026-11-14 21:19:50 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `5063d62` — 2026-11-15 00:49:36 +0330 — rehearse explicit compose rollback
  Files: scripts/clean-environment-rehearsal.sh,tests/deploy/clean_environment_script_test.go
- `002176b` — 2026-11-14 21:19:19 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `2f37359` — 2026-11-15 00:49:04 +0330 — add explicit compose rollback script
  Files: scripts/rollback-compose.sh,tests/deploy/rollback_compose_test.go
- `13b1c4c` — 2026-11-14 21:18:41 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `af4c449` — 2026-11-15 00:48:26 +0330 — make worker supervisor concurrency safe
  Files: internal/plugin/supervisor.go,tests/plugin/supervisor_test.go
- `48ea91a` — 2026-11-14 21:18:01 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `7fff3e1` — 2026-11-15 00:47:48 +0330 — cover revoked in-flight credentials
  Files: tests/plugin/broker_test.go
- `0ef8193` — 2026-11-14 21:16:59 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `0bb0413` — 2026-11-15 00:46:42 +0330 — bind external worker handshakes to project
  Files: internal/plugin/external.go,internal/plugin/external_http.go tests/plugin/external_http_test.go,tests/plugin/external_test.go
- `f33e44a` — 2026-11-14 21:15:52 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `af2e8c4` — 2026-11-15 00:45:37 +0330 — remove unscoped worker invocation
  Files: internal/plugin/worker.go,tests/plugin/bubblewrap_worker_test.go tests/plugin/external_fixture_test.go,tests/plugin/typescript_worker_test.go tests/plugin/worker_fixture_test.go
- `28b1301` — 2026-11-14 21:15:10 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `b60dba0` — 2026-11-15 00:44:57 +0330 — bind wasm execution to host project
  Files: internal/plugin/wasm_runner.go,tests/plugin/wasm_runner_test.go
- `450c51e` — 2026-11-14 21:14:27 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `0ea8b29` — 2026-11-15 00:44:13 +0330 — scope function job leases by project
  Files: plugins/functions/job-lease-acquirer.ts,plugins/functions/job-lease-renewer.ts tests/functions/functions.test.ts
- `3924c1a` — 2026-11-14 21:13:14 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `6d8b8c9` — 2026-11-15 00:43:02 +0330 — require project context for runtime SQL
  Files: internal/db/db.go,tests/db/db_test.go
- `ff8cabd` — 2026-11-14 21:12:23 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `45d622f` — 2026-11-15 00:42:09 +0330 — bind websocket messages to host project
  Files: internal/plugin/websocket_message.go,tests/plugin/websocket_message_test.go
- `b5d22b6` — 2026-11-14 21:11:45 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `8046ba0` — 2026-11-15 00:41:32 +0330 — reject traversal in websocket project paths
  Files: internal/plugin/websocket_endpoint.go,tests/plugin/websocket_endpoint_test.go
- `3b13d37` — 2026-11-14 21:10:58 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
