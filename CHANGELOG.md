# Changelog

Automatically generated from commits on main.

## Unreleased
- `a0445a8` — 2026-11-13 16:10:19 +0330 — verify direct scheduled and record invocations
  Files: roadmap.md,tests/functions/phase04_invocation_gate.test.ts
- `99fef0e` — 2026-11-13 11:35:49 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `5a50310` — 2026-11-13 15:05:28 +0330 — publish and roll back shared helper artifacts
  Files: plugins/functions/index.ts,plugins/functions/plugin.yaml plugins/functions/shared-helper-publisher.ts,roadmap.md tests/functions/functions.test.ts
- `3c69b38` — 2026-11-13 11:33:48 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `877c9cd` — 2026-11-13 15:03:25 +0330 — add CI compose plugin isolation profile
  Files: .github/workflows/test.yml,Dockerfile.test compose.test.yml,scripts/smoke-compose.sh
- `89c2164` — 2026-11-13 11:31:10 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `3936e85` — 2026-11-13 15:00:44 +0330 — add entry package fixture
  Files: plugins/functions/index.ts,plugins/functions/plugin.yaml plugins/tables/plugin.yaml,roadmap.md tests/functions/functions.test.ts
- `ec919b1` — 2026-11-13 11:27:43 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `ffba1cc` — 2026-11-13 14:57:19 +0330 — limit changelog to latest detailed commits
  Files: .github/workflows/changelog.yml
- `6d3eeda` — 2026-11-13 11:26:51 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `310de3d` — 2026-11-13 14:56:26 +0330 — split schedule operations
  Files: plugins/functions/function-version-enqueuer.ts,plugins/functions/index.ts plugins/functions/missed-schedule-enqueuer.ts,plugins/functions/plugin.yaml plugins/functions/schedule-adder.ts,plugins/functions/schedule-claimer.ts plugins/functions/schedule-occurrence-recorder.ts,plugins/functions/schedule-reader.ts plugins/functions/schedule-state.ts,tests/functions/functions.test.ts
- `c23aa85` — 2026-11-13 11:25:08 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `813fb10` — 2026-11-13 14:54:47 +0330 — split record trigger operations
  Files: plugins/functions/index.ts,plugins/functions/plugin.yaml plugins/functions/record-trigger-adder.ts,plugins/functions/record-trigger-dispatcher.ts plugins/functions/record-trigger-matcher.ts,plugins/functions/record-trigger-state.ts tests/functions/functions.test.ts
- `2616adf` — 2026-11-13 11:23:46 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `c53fac7` — 2026-11-13 14:53:26 +0330 — split invocation commit operations
  Files: plugins/functions/index.ts,plugins/functions/invocation-commit-reader.ts plugins/functions/invocation-commit-state.ts,plugins/functions/invocation-commit-writer.ts plugins/functions/invocation-existence-reader.ts,plugins/functions/invocation-result-reader.ts plugins/functions/plugin.yaml,tests/functions/functions.test.ts
- `e4ab489` — 2026-11-13 11:22:29 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `0864716` — 2026-11-13 14:52:07 +0330 — split function publication operations
  Files: plugins/functions/function-publication-reader.ts,plugins/functions/function-publication-rollback.ts plugins/functions/function-publication-state.ts,plugins/functions/function-publication-writer.ts plugins/functions/index.ts,plugins/functions/plugin.yaml tests/functions/functions.test.ts
- `dc905b6` — 2026-11-13 11:20:16 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `3517659` — 2026-11-13 14:49:53 +0330 — split invocation effect operations
  Files: plugins/functions/index.ts,plugins/functions/invocation-effects-reader.ts plugins/functions/invocation-effects-state.ts,plugins/functions/invocation-effects-writer.ts plugins/functions/plugin.yaml,tests/functions/functions.test.ts
- `ea14df2` — 2026-11-13 11:18:23 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `ca2db40` — 2026-11-13 14:48:00 +0330 — split function lease operations
  Files: plugins/functions/index.ts,plugins/functions/job-lease-acquirer.ts plugins/functions/job-lease-renewer.ts,plugins/functions/job-lease-state.ts plugins/functions/plugin.yaml,tests/functions/functions.test.ts
- `2e6216c` — 2026-11-13 11:15:45 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `08ae2f6` — 2026-11-13 14:45:25 +0330 — split function draft operations
  Files: plugins/functions/function-draft-reader.ts,plugins/functions/function-draft-state.ts plugins/functions/function-draft-writer.ts,plugins/functions/index.ts plugins/functions/plugin.yaml,tests/functions/functions.test.ts
- `7b7aa61` — 2026-11-13 11:14:28 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `bbe80a8` — 2026-11-13 14:44:07 +0330 — split shared package operations
  Files: plugins/functions/index.ts,plugins/functions/plugin.yaml plugins/functions/shared-package-reader.ts,plugins/functions/shared-package-state.ts plugins/functions/shared-package-writer.ts,tests/functions/functions.test.ts
- `b617972` — 2026-11-13 11:12:31 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `00deca8` — 2026-11-13 14:42:10 +0330 — split function approval operations
  Files: plugins/functions/approval-granter.ts,plugins/functions/approval-reader.ts plugins/functions/approval-requester.ts,plugins/functions/approval-state.ts plugins/functions/index.ts,plugins/functions/plugin.yaml tests/functions/functions.test.ts
- `df8123b` — 2026-11-13 11:10:35 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `d0689ee` — 2026-11-13 14:40:10 +0330 — split transactional change capture operations
  Files: plugins/tables/index.ts,plugins/tables/plugin.yaml plugins/tables/scoped-transaction-change-reader.ts,plugins/tables/scoped-transaction-change-writer.ts plugins/tables/transaction-change-reader.ts,plugins/tables/transaction-change-state.ts plugins/tables/transaction-change-writer.ts,tests/tables/tables.test.ts
- `08a44e8` — 2026-11-13 11:08:44 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `967f65c` — 2026-11-13 14:38:22 +0330 — split SQL change capture operations
  Files: plugins/tables/index.ts,plugins/tables/plugin.yaml plugins/tables/sql-change-reader.ts,plugins/tables/sql-change-state.ts plugins/tables/sql-change-writer.ts,tests/tables/tables.test.ts
- `0134c9b` — 2026-11-13 11:07:22 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `bbf272b` — 2026-11-13 14:37:01 +0330 — split API change capture operations
  Files: plugins/tables/api-change-reader.ts,plugins/tables/api-change-state.ts plugins/tables/api-change-writer.ts,plugins/tables/index.ts plugins/tables/plugin.yaml,tests/tables/tables.test.ts
- `a453106` — 2026-11-13 11:06:13 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `9908a14` — 2026-11-13 14:35:52 +0330 — split schema audit operations
  Files: plugins/tables/index.ts,plugins/tables/plugin.yaml plugins/tables/schema-audit-state.ts,plugins/tables/schema-publication-reader.ts plugins/tables/schema-publication-recorder.ts,tests/tables/tables.test.ts
- `7c248fc` — 2026-11-13 11:02:53 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `740d4a8` — 2026-11-13 14:32:35 +0330 — split versioned record operations
  Files: plugins/tables/index.ts,plugins/tables/plugin.yaml plugins/tables/versioned-record-reader.ts,plugins/tables/versioned-record-state.ts plugins/tables/versioned-record-updater.ts,plugins/tables/versioned-record.ts tests/tables/tables.test.ts
- `842e8b9` — 2026-11-13 11:01:53 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `03902e9` — 2026-11-13 14:31:30 +0330 — make plugin decomposition a permanent law
  Files: AGENTS.md,docs/admins.html docs/developers.md,roadmap.md
- `fcacca7` — 2026-11-13 11:00:25 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `566016b` — 2026-11-13 14:30:05 +0330 — split matchmaking queue operations
  Files: plugins/matchmaking/assignment-creator.ts,plugins/matchmaking/index.ts plugins/matchmaking/plugin.yaml,plugins/matchmaking/queue-runtime.ts plugins/matchmaking/queue.ts,plugins/matchmaking/ticket-enqueuer.ts plugins/matchmaking/ticket-expirer.ts,plugins/matchmaking/ticket-matcher.ts plugins/matchmaking/ticket-reader.ts,plugins/matchmaking/ticket-requeuer.ts tests/arena/phase02_gate.test.ts,tests/matchmaking/matchmaking.test.ts
- `10fe316` — 2026-11-13 10:57:52 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `cd51721` — 2026-11-13 14:27:26 +0330 — split tic tac toe coordinator
  Files: plugins/rooms/index.ts,plugins/rooms/plugin.yaml plugins/rooms/tic-tac-toe-mover.ts,plugins/rooms/tic-tac-toe-revision-reader.ts plugins/rooms/tic-tac-toe-runtime.ts,tests/rooms/rooms.test.ts
- `971d6d1` — 2026-11-13 10:51:23 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `0d37bba` — 2026-11-13 14:21:03 +0330 — split room directory operations
  Files: plugins/rooms/index.ts,plugins/rooms/plugin.yaml plugins/rooms/room-creator.ts,plugins/rooms/room-event-reader.ts plugins/rooms/room-joiner.ts,plugins/rooms/room-leaver.ts plugins/rooms/room-membership-reader.ts,plugins/rooms/room-reader.ts plugins/rooms/room-runtime.ts,tests/arena/phase02_gate.test.ts tests/rooms/rooms.test.ts
- `f10bb6b` — 2026-11-13 10:49:27 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `ba63107` — 2026-11-13 14:19:03 +0330 — split arena state operations
  Files: plugins/arena/arena-action-recorder.ts,plugins/arena/arena-finisher.ts plugins/arena/arena-measurer.ts,plugins/arena/arena-start.ts plugins/arena/arena-starter.ts,plugins/arena/arena-state-reader.ts plugins/arena/arena-state.ts,plugins/arena/index.ts plugins/arena/plugin.yaml,plugins/arena/state-machine.ts tests/arena/arena.test.ts,tests/arena/phase02_gate.test.ts tests/dashboard/phase03_guest_coop_gate.test.ts
- `47269da` — 2026-11-13 10:45:34 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `d4baf82` — 2026-11-13 14:15:02 +0330 — split auth account operations
  Files: plugins/auth/account-password-resetter.ts,plugins/auth/account-session-reader.ts plugins/auth/account-session-revoker.ts,plugins/auth/account-session-rotator.ts plugins/auth/account-signin.ts,plugins/auth/account-signup.ts plugins/auth/auth-state.ts,plugins/auth/guest-account-upgrader.ts plugins/auth/guest-session-issuer.ts,plugins/auth/index.ts plugins/auth/plugin.yaml,tests/auth/auth.test.ts tests/dashboard/phase03_guest_coop_gate.test.ts,tests/dashboard/phase03_isolation_gate.test.ts
- `d07843c` — 2026-11-13 10:40:28 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
