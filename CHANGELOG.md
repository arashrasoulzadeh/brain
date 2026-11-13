# Changelog

Automatically generated from commits on main.

## Unreleased
- `e17db3a` — 2026-11-13 16:47:19 +0330 — feat: block room admissions during drain
  Files: plugins/rooms/index.ts,plugins/rooms/room-creator.ts plugins/rooms/room-runtime.ts,roadmap.md tests/rooms/rooms.test.ts
- `2551ac5` — 2026-11-13 13:16:28 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `bce1cf2` — 2026-11-13 16:46:03 +0330 — feat: allow participant reconnects during drain
  Files: internal/plugin/plugin.go,roadmap.md tests/plugin/plugin_test.go
- `34d4a22` — 2026-11-13 13:15:35 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `565f901` — 2026-11-13 16:44:52 +0330 — feat: default plugin drains to five minutes
  Files: internal/plugin/plugin.go,roadmap.md tests/plugin/plugin_test.go
- `356bd20` — 2026-11-13 13:14:20 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `b9a3a0f` — 2026-11-13 16:43:20 +0330 — feat: expose persisted plugin drain state
  Files: internal/plugin/plugin.go,roadmap.md tests/plugin/plugin_test.go
- `daf792b` — 2026-11-13 13:12:56 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `568fd29` — 2026-11-13 16:41:29 +0330 — feat: cancel queued matchmaking scope tickets
  Files: plugins/matchmaking/plugin.yaml,plugins/matchmaking/queue.ts plugins/matchmaking/scope-ticket-canceller.ts,roadmap.md tests/matchmaking/matchmaking.test.ts
- `7fe36ac` — 2026-11-13 13:10:53 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `94f6982` — 2026-11-13 16:39:22 +0330 — feat: block new work during plugin drain
  Files: internal/plugin/plugin.go,roadmap.md tests/plugin/plugin_test.go
- `f3216a4` — 2026-11-13 13:09:01 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `216353a` — 2026-11-13 16:37:17 +0330 — test: isolate matchmaking scope disable
  Files: roadmap.md,tests/matchmaking/matchmaking.test.ts
- `45c2e89` — 2026-11-13 13:06:51 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `c1ed766` — 2026-11-13 16:34:07 +0330 — feat: reject stale plugin disable previews
  Files: internal/plugin/plugin.go,roadmap.md tests/plugin/plugin_test.go
- `8fd0204` — 2026-11-13 13:03:45 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `9f3b4f8` — 2026-11-13 16:33:21 +0330 — feat: add plugin disable previews
  Files: internal/plugin/plugin.go,roadmap.md tests/plugin/plugin_test.go
- `37774a0` — 2026-11-13 13:02:44 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `8bbfbb1` — 2026-11-13 16:31:50 +0330 — feat: calculate transitive plugin dependents
  Files: internal/plugin/plugin.go,roadmap.md tests/plugin/plugin_test.go
- `2c1f5a5` — 2026-11-13 13:00:51 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `188fef8` — 2026-11-13 16:30:30 +0330 — feat: require approval for new plugin capabilities
  Files: internal/plugin/plugin.go,roadmap.md tests/plugin/plugin_test.go
- `6e034ee` — 2026-11-13 12:59:33 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `1ec481e` — 2026-11-13 16:29:06 +0330 — test: preserve plugin grants during rollback
  Files: roadmap.md,tests/plugin/plugin_test.go
- `d2bd281` — 2026-11-13 12:58:25 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `f7ecc1f` — 2026-11-13 16:26:33 +0330 — test: reject incompatible plugin rollbacks
  Files: roadmap.md,tests/plugin/plugin_test.go
- `3dea4f3` — 2026-11-13 12:55:10 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `fe9536b` — 2026-11-13 16:24:11 +0330 — test: enforce retained plugin rollback versions
  Files: roadmap.md,tests/plugin/plugin_test.go
- `159698b` — 2026-11-13 12:53:09 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `84c528b` — 2026-11-13 16:22:48 +0330 — feat: cancel incompatible matchmaking tickets
  Files: plugins/matchmaking/plugin.yaml,plugins/matchmaking/queue.ts plugins/matchmaking/ticket-canceller.ts,plugins/matchmaking/ticket-enqueuer.ts plugins/matchmaking/types.ts,roadmap.md tests/matchmaking/matchmaking.test.ts
- `83fce3b` — 2026-11-13 12:51:06 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `94c528a` — 2026-11-13 16:20:44 +0330 — test: verify pinned room versions survive promotion
  Files: roadmap.md
- `515dc00` — 2026-11-13 12:50:39 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `aa0a4c7` — 2026-11-13 16:20:17 +0330 — feat: validate plugin schema compatibility
  Files: internal/plugin/plugin.go,roadmap.md tests/plugin/plugin_test.go
- `8a29624` — 2026-11-13 12:49:42 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `5a8493e` — 2026-11-13 16:19:19 +0330 — feat: recheck plugin grants during promotion
  Files: internal/plugin/plugin.go,roadmap.md tests/plugin/plugin_test.go
- `ee51f53` — 2026-11-13 12:48:29 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `6ef4174` — 2026-11-13 16:18:07 +0330 — docs: close phase 04 isolation gate
  Files: docs/admins.md,docs/developers.md roadmap.md
- `404352c` — 2026-11-13 12:43:43 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `01ab594` — 2026-11-13 16:13:23 +0330 — exclude vendored files from Go formatting check
  Files: .github/workflows/test.yml
- `fdf4279` — 2026-11-13 12:42:43 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `7bff8e7` — 2026-11-13 16:12:23 +0330 — record function and turn metrics
  Files: roadmap.md,tests/functions/phase04_metrics_gate.test.ts
- `b9bc625` — 2026-11-13 12:42:08 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `37a6b4c` — 2026-11-13 16:11:46 +0330 — verify two client tic tac toe match
  Files: roadmap.md,tests/functions/phase04_tictactoe_gate.test.ts
- `056ed1b` — 2026-11-13 12:41:31 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `2df9268` — 2026-11-13 16:11:07 +0330 — verify function crash recovery and grant revocation
  Files: roadmap.md,tests/functions/phase04_recovery_gate.test.ts
- `e23347b` — 2026-11-13 12:40:49 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `a0445a8` — 2026-11-13 16:10:19 +0330 — verify direct scheduled and record invocations
  Files: roadmap.md,tests/functions/phase04_invocation_gate.test.ts
- `99fef0e` — 2026-11-13 11:35:49 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `5a50310` — 2026-11-13 15:05:28 +0330 — publish and roll back shared helper artifacts
  Files: plugins/functions/index.ts,plugins/functions/plugin.yaml plugins/functions/shared-helper-publisher.ts,roadmap.md tests/functions/functions.test.ts
- `3c69b38` — 2026-11-13 11:33:48 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
