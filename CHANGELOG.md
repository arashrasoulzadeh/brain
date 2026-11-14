# Changelog

Automatically generated from commits on main.

## Unreleased
- `e968db1` — 2026-11-14 17:32:22 +0330 — define PostgreSQL restore rehearsal
  Files: roadmap_backend.md,scripts/restore-postgres-backup.sh tests/deploy/postgres_restore_test.go
- `513d3ec` — 2026-11-14 14:01:58 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `6e90046` — 2026-11-14 17:31:44 +0330 — alert on stale verified backups
  Files: internal/deploy/backup_policy.go,roadmap_backend.md scripts/check-backup-age.sh,tests/deploy/backup_age_script_test.go tests/deploy/backup_policy_test.go
- `8243875` — 2026-11-14 14:01:17 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `d09d401` — 2026-11-14 17:31:01 +0330 — verify object version manifests
  Files: cmd/brain/main.go,roadmap_backend.md scripts/verify-object-manifest.sh,tests/deploy/object_manifest_script_test.go
- `c07d666` — 2026-11-14 13:59:35 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `7eb2eee` — 2026-11-14 17:29:19 +0330 — verify PostgreSQL backup artifacts
  Files: roadmap_backend.md,scripts/verify-postgres-backup.sh tests/deploy/postgres_backup_verify_test.go
- `a920cfe` — 2026-11-14 13:59:01 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `7443223` — 2026-11-14 17:28:46 +0330 — record referenced object versions
  Files: internal/deploy/object_manifest.go,roadmap_backend.md tests/deploy/object_manifest_test.go
- `1487dde` — 2026-11-14 13:58:13 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `a0f6f69` — 2026-11-14 17:28:00 +0330 — add consistent PostgreSQL snapshot
  Files: roadmap_backend.md,scripts/snapshot-postgres.sh tests/deploy/postgres_snapshot_test.go
- `c987deb` — 2026-11-14 13:57:36 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `2b05f3e` — 2026-11-14 17:27:21 +0330 — define verified backup policy
  Files: docs/admins.md,docs/developers.md internal/deploy/backup_policy.go,roadmap_backend.md tests/deploy/backup_policy_test.go
- `0236b8f` — 2026-11-14 13:56:49 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `820173f` — 2026-11-14 17:26:18 +0330 — document incompatible migration recovery
  Files: docs/admins.md,docs/developers.md docs/incompatible-migration-recovery.md,roadmap_backend.md tests/deploy/recovery_doc_test.go
- `98f4266` — 2026-11-14 13:55:49 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `4913227` — 2026-11-14 17:25:35 +0330 — guard rollback by schema compatibility
  Files: internal/deploy/rollback.go,roadmap_backend.md tests/deploy/rollback_test.go
- `62adf62` — 2026-11-14 13:55:13 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `660be6c` — 2026-11-14 17:24:58 +0330 — retain last healthy release manifest
  Files: internal/deploy/healthy_manifest.go,internal/deploy/manifest.go roadmap_backend.md,tests/deploy/healthy_manifest_test.go
- `e508729` — 2026-11-14 13:54:20 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `aff04b2` — 2026-11-14 17:24:03 +0330 — record deployed release version
  Files: internal/deploy/version.go,roadmap_backend.md tests/deploy/version_test.go
- `c755865` — 2026-11-14 13:53:33 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `b9a8b6f` — 2026-11-14 17:23:19 +0330 — check enabled worker health before release
  Files: internal/plugin/supervisor.go,roadmap_backend.md tests/plugin/supervisor_test.go
- `49e553f` — 2026-11-14 13:52:40 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `acfeebf` — 2026-11-14 17:22:24 +0330 — add PostgreSQL deployment access check
  Files: roadmap_backend.md,scripts/check-postgres-access.sh tests/deploy/postgres_script_test.go
- `d3be563` — 2026-11-14 13:51:59 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `104c589` — 2026-11-14 17:21:45 +0330 — add bounded API readiness check
  Files: roadmap_backend.md,scripts/check-api-readiness.sh tests/deploy/readiness_script_test.go
- `6c97588` — 2026-11-14 13:51:07 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
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
