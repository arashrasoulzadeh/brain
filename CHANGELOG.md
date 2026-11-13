# Changelog

Automatically generated from commits on main.

## Unreleased
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
