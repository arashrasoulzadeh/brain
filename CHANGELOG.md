# Changelog

Automatically generated from commits on main.

## Unreleased
- `d8bdb9e` — 2026-11-13 17:04:32 +0330 — test: audit project uninstall
  Files: roadmap.md,tests/plugin/plugin_test.go
- `f39e3bc` — 2026-11-13 13:32:29 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `90953c3` — 2026-11-13 17:00:28 +0330 — feat: preserve durable plugin data on uninstall
  Files: internal/plugin/plugin.go,roadmap.md tests/plugin/plugin_test.go
- `19b1114` — 2026-11-13 13:28:25 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `66f3202` — 2026-11-13 16:58:02 +0330 — feat: revoke project grants during uninstall
  Files: internal/plugin/plugin.go,roadmap.md tests/plugin/plugin_test.go
- `b8e5533` — 2026-11-13 13:27:31 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `c423dd3` — 2026-11-13 16:56:42 +0330 — feat: wait for active work before uninstall
  Files: internal/plugin/plugin.go,roadmap.md tests/plugin/plugin_test.go
- `712fe8d` — 2026-11-13 13:24:25 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `42f12e5` — 2026-11-13 16:53:55 +0330 — feat: resume paused rooms on pinned versions
  Files: plugins/rooms/index.ts,plugins/rooms/plugin.yaml plugins/rooms/resume-tic-tac-toe-turn.ts,roadmap.md tests/rooms/rooms.test.ts
- `42f4c16` — 2026-11-13 13:23:14 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `e4e916b` — 2026-11-13 16:52:38 +0330 — feat: recheck lifecycle at invocation commit
  Files: plugins/functions/index.ts,plugins/functions/invocation-commit-state.ts plugins/functions/invocation-commit-writer.ts,roadmap.md tests/functions/functions.test.ts
- `0851194` — 2026-11-13 13:21:55 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `fd0a32d` — 2026-11-13 16:51:34 +0330 — feat: recheck matchmaking lifecycle at assignment
  Files: plugins/matchmaking/assignment-creator.ts,plugins/matchmaking/queue-runtime.ts plugins/matchmaking/queue.ts,roadmap.md tests/matchmaking/matchmaking.test.ts
- `d608ee7` — 2026-11-13 13:20:42 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `a19fe89` — 2026-11-13 16:50:19 +0330 — feat: add immediate plugin security stop
  Files: internal/plugin/plugin.go,roadmap.md tests/plugin/plugin_test.go
- `6551be1` — 2026-11-13 13:19:57 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `c2b688b` — 2026-11-13 16:49:36 +0330 — feat: interrupt expired action rooms
  Files: plugins/arena/index.ts,plugins/arena/interrupt-expired-action-room.ts plugins/arena/plugin.yaml,plugins/arena/types.ts roadmap.md,tests/arena/arena.test.ts
- `9bae6ae` — 2026-11-13 13:18:58 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `26403c5` — 2026-11-13 16:48:32 +0330 — feat: pause turn rooms at committed state
  Files: plugins/rooms/index.ts,plugins/rooms/pause-tic-tac-toe-turn.ts plugins/rooms/plugin.yaml,plugins/rooms/validate-tic-tac-toe-rule-result.ts roadmap.md,tests/rooms/rooms.test.ts
- `f78b1a4` — 2026-11-13 13:17:40 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
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
