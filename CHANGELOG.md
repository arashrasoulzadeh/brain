# Changelog

Automatically generated from commits on main.

## Unreleased
- `6c4a8d3` — 2026-11-14 17:20:47 +0330 — start deployment compose without rebuilding
  Files: roadmap_backend.md,scripts/start-compose-no-build.sh tests/deploy/compose_script_test.go
- `94a5102` — 2026-11-14 13:50:23 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `eed37f9` — 2026-11-14 17:20:02 +0330 — add one-shot deployment migration runner
  Files: internal/migrate/oneshot.go,roadmap_backend.md tests/migrate/migrate_test.go
- `cd03169` — 2026-11-14 13:49:28 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `0b92e4c` — 2026-11-14 17:19:14 +0330 — close tracked sockets during deploy
  Files: internal/plugin/websocket_registry.go,roadmap_backend.md tests/plugin/websocket_registry_test.go
- `ef4c55b` — 2026-11-14 13:48:33 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `e25b42a` — 2026-11-14 17:18:18 +0330 — drain room admissions during deploy
  Files: plugins/rooms/create-room-directory.ts,plugins/rooms/is-room-draining.ts plugins/rooms/plugin.yaml,plugins/rooms/room-creator.ts plugins/rooms/room-runtime.ts,plugins/rooms/start-room-drain.ts plugins/rooms/stop-room-drain.ts,tests/rooms/rooms.test.ts
- `6385163` — 2026-11-14 13:47:00 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `6df7801` — 2026-11-14 17:16:44 +0330 — pause matchmaking assignments during deploy
  Files: plugins/matchmaking/assignment-creator.ts,plugins/matchmaking/is-project-assignments-paused.ts plugins/matchmaking/pause-project-assignments.ts,plugins/matchmaking/plugin.yaml plugins/matchmaking/queue-runtime.ts,plugins/matchmaking/queue.ts plugins/matchmaking/resume-project-assignments.ts,plugins/matchmaking/ticket-enqueuer.ts tests/matchmaking/matchmaking.test.ts
- `beb53f0` — 2026-11-14 13:45:29 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `15805d7` — 2026-11-14 17:15:13 +0330 — enforce image pull before deployment drain
  Files: internal/deploy/sequence.go,roadmap_backend.md tests/deploy/sequence_test.go
- `452556c` — 2026-11-14 13:44:17 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `ed03dc6` — 2026-11-14 17:13:57 +0330 — make backend roadmap the active checklist
  Files: AGENTS.md
- `b4be8d3` — 2026-11-14 13:43:01 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `c619d75` — 2026-11-14 17:12:44 +0330 — split roadmap by backend admin and client
  Files: roadmap.md,roadmap_admin.md roadmap_backend.md,roadmap_client.md
- `89a45f0` — 2026-11-14 13:40:51 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `a6ff0ed` — 2026-11-14 17:10:36 +0330 — trim completed roadmap phases
  Files: roadmap.md
- `04539b0` — 2026-11-14 13:37:34 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `7f77480` — 2026-11-14 17:07:19 +0330 — validate deployment compose configuration
  Files: internal/deploy/compose.go,roadmap.md tests/deploy/compose_test.go
- `2b73bf2` — 2026-11-14 13:36:05 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `2e8f689` — 2026-11-14 17:05:46 +0330 — validate release image manifest
  Files: internal/deploy/manifest.go,roadmap.md tests/deploy/manifest_test.go
- `945c957` — 2026-11-14 13:35:31 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `c5eac14` — 2026-11-14 17:05:17 +0330 — validate deployment release sha
  Files: internal/deploy/release_sha.go,roadmap.md tests/deploy/release_sha_test.go
- `2599e51` — 2026-11-14 13:35:09 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `c5ae197` — 2026-11-14 17:04:52 +0330 — add atomic deployment lock
  Files: internal/deploy/lock.go,roadmap.md tests/deploy/lock_test.go
- `8c29446` — 2026-11-14 13:34:25 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `74ca814` — 2026-11-14 17:04:10 +0330 — bind plugin envelopes to host project
  Files: internal/plugin/http_envelope.go,internal/plugin/worker.go tests/plugin/http_envelope_test.go,tests/plugin/worker_test.go
- `6ea8068` — 2026-11-14 13:32:46 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `a877763` — 2026-11-14 17:02:31 +0330 — add protected backend release workflow
  Files: .github/workflows/release-backend.yml,roadmap.md
- `37e1143` — 2026-11-14 13:31:56 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `3689372` — 2026-11-14 17:01:44 +0330 — align plugin enable test with grant policy
  Files: tests/plugin/plugin_test.go
- `61f53c6` — 2026-11-14 13:31:03 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `17fd047` — 2026-11-14 17:00:50 +0330 — scaffold kotlin android client
  Files: clients/android/README.md,clients/android/src/main/kotlin/com/brain/client/AuthCallback.kt clients/android/src/main/kotlin/com/brain/client/BrainClient.kt
- `9791301` — 2026-11-14 13:30:29 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `07c316b` — 2026-11-14 17:00:15 +0330 — document protocol compatibility rules
  Files: protocol/README.md,roadmap.md
- `5620009` — 2026-11-14 13:30:15 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `ea2d311` — 2026-11-14 16:59:59 +0330 — publish versioned protocol fixtures
  Files: protocol/README.md,roadmap.md
- `8202ec5` — 2026-11-14 13:29:55 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `1419fc3` — 2026-11-14 16:59:42 +0330 — validate protocol fixtures with dart
  Files: clients/flutter/lib/protocol_fixtures.dart,roadmap.md tests/flutter/protocol_fixture_consumer_test.dart
- `534efb2` — 2026-11-14 13:29:16 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `44a2c30` — 2026-11-14 16:59:01 +0330 — define reconnect protocol fixture
  Files: protocol/fixtures/reconnect.json,roadmap.md tests/protocol/fixtures_test.go
- `8fe29c6` — 2026-11-14 13:28:50 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `6ebad02` — 2026-11-14 16:58:38 +0330 — define authorized snapshot fixture
  Files: protocol/fixtures/authorized-snapshot.json,roadmap.md tests/protocol/fixtures_test.go
- `b056235` — 2026-11-14 13:28:27 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `3980d5f` — 2026-11-14 16:58:14 +0330 — define player input fixture
  Files: protocol/fixtures/player-input.json,roadmap.md tests/protocol/fixtures_test.go
- `f42f7ab` — 2026-11-14 13:28:03 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `4a48adf` — 2026-11-14 16:57:50 +0330 — define assignment fixture
  Files: protocol/fixtures/assignment.json,roadmap.md tests/protocol/fixtures_test.go
- `0618c1d` — 2026-11-14 13:27:41 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `2f3117a` — 2026-11-14 16:57:27 +0330 — define room admission fixture
  Files: protocol/fixtures/room-admission.json,roadmap.md tests/protocol/fixtures_test.go
- `bec31f5` — 2026-11-14 13:27:16 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
