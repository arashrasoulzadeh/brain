# Changelog

Automatically generated from commits on main.

## Unreleased
- `0714be7` — 2026-11-13 17:18:44 +0330 — feat: add custom typescript admin view
  Files: plugins/dashboard/custom-admin-view.ts,plugins/dashboard/index.ts plugins/dashboard/plugin.yaml,roadmap.md tests/dashboard/dashboard.test.ts
- `5902a49` — 2026-11-13 13:48:28 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `1730e10` — 2026-11-13 17:17:56 +0330 — feat: audit dashboard view actions
  Files: plugins/dashboard/view-replacement-registry.ts,roadmap.md tests/dashboard/dashboard.test.ts
- `39b996c` — 2026-11-13 13:47:18 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `12a7072` — 2026-11-13 17:16:49 +0330 — feat: add dashboard csrf validation
  Files: plugins/dashboard/index.ts,plugins/dashboard/plugin.yaml plugins/dashboard/validate-csrf-token.ts,roadmap.md tests/dashboard/dashboard.test.ts
- `454a309` — 2026-11-13 13:46:21 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `543830d` — 2026-11-13 17:15:59 +0330 — feat: authorize dashboard view actions
  Files: plugins/dashboard/view-replacement-registry.ts,roadmap.md tests/dashboard/dashboard.test.ts
- `51f4aff` — 2026-11-13 13:45:40 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `bf450ed` — 2026-11-13 17:15:20 +0330 — feat: isolate dashboard view submissions
  Files: plugins/dashboard/view-replacement-registry.ts,roadmap.md tests/dashboard/dashboard.test.ts
- `412d004` — 2026-11-13 13:44:57 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `4abe251` — 2026-11-13 17:14:20 +0330 — test: verify duplicate view replacement rejection
  Files: roadmap.md
- `2cfe5ce` — 2026-11-13 13:43:48 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `7a1db4a` — 2026-11-13 17:13:25 +0330 — feat: activate scoped admin view replacements
  Files: plugins/dashboard/index.ts,plugins/dashboard/plugin.yaml plugins/dashboard/view-replacement-registry.ts,roadmap.md tests/dashboard/dashboard.test.ts
- `1665987` — 2026-11-13 13:42:56 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `f8d1805` — 2026-11-13 17:12:35 +0330 — feat: add declarative dashboard view contributions
  Files: plugins/dashboard/index.ts,plugins/dashboard/plugin-health-contribution.ts plugins/dashboard/plugin.yaml,plugins/dashboard/project-overview-contribution.ts roadmap.md,tests/dashboard/dashboard.test.ts
- `7e0c0b3` — 2026-11-13 13:42:07 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `2d18d15` — 2026-11-13 17:11:42 +0330 — feat: render host-owned admin fixture
  Files: plugins/dashboard/host-owned-admin-view.ts,plugins/dashboard/index.ts plugins/dashboard/plugin.yaml,roadmap.md tests/dashboard/dashboard.test.ts
- `a5a3e1f` — 2026-11-13 13:41:19 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `2fcb1ce` — 2026-11-13 17:10:55 +0330 — feat: discover declarative admin extension points
  Files: internal/plugin/plugin.go,roadmap.md tests/plugin/plugin_test.go
- `b059a1d` — 2026-11-13 13:40:34 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `c29beb0` — 2026-11-13 17:10:00 +0330 — feat: preserve operator plugin recovery
  Files: internal/plugin/plugin.go,roadmap.md tests/plugin/plugin_test.go
- `bfad1d1` — 2026-11-13 13:39:17 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `5e45d87` — 2026-11-13 17:08:41 +0330 — feat: separate durable data deletion
  Files: internal/plugin/plugin.go,roadmap.md tests/plugin/plugin_test.go
- `a5471cf` — 2026-11-13 13:38:16 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `b855075` — 2026-11-13 17:07:46 +0330 — test: verify referenced server cleanup
  Files: roadmap.md
- `3fd9d31` — 2026-11-13 13:37:29 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `e22425b` — 2026-11-13 17:05:55 +0330 — feat: protect referenced plugin versions from unregister
  Files: internal/plugin/plugin.go,roadmap.md tests/plugin/plugin_test.go
- `cc33511` — 2026-11-13 13:35:14 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
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
