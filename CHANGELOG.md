# Changelog

Automatically generated from commits on main.

## Unreleased
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
