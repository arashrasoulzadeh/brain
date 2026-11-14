# Changelog

Automatically generated from commits on main.

## Unreleased
- `5ca7c96` — 2026-11-14 22:36:57 +0330 — enforce sandbox resource limits
  Files: internal/plugin/typescript_worker.go,scripts/run-sandboxed-typescript-worker.sh tests/plugin/bubblewrap_worker_test.go,tests/plugin/sandbox_launcher_test.go
- `b370d18` — 2026-11-14 19:05:43 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `fdab3cb` — 2026-11-14 22:35:21 +0330 — wire sandbox launcher into plugin host
  Files: internal/plugin/typescript_worker.go,tests/plugin/bubblewrap_worker_test.go
- `867c433` — 2026-11-14 19:04:42 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `ec2f8d2` — 2026-11-14 22:34:27 +0330 — add fail-closed typescript sandbox launcher
  Files: scripts/run-sandboxed-typescript-worker.sh,tests/plugin/sandbox_launcher_test.go
- `fae6164` — 2026-11-14 14:11:57 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `2a0d887` — 2026-11-14 17:41:44 +0330 — bound typescript worker invocation time
  Files: internal/plugin/typescript_worker.go,tests/plugin/typescript_worker_test.go
- `a6b3282` — 2026-11-14 14:10:53 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `29104c7` — 2026-11-14 17:40:39 +0330 — connect typescript worker to host supervisor
  Files: internal/plugin/typescript_worker.go,tests/plugin/typescript_worker_test.go
- `c84ba77` — 2026-11-14 14:09:53 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `f904063` — 2026-11-14 17:39:40 +0330 — gate typescript worker execution context
  Files: scripts/typescript-worker.mjs,tests/plugin/typescript-worker.test.ts
- `a402693` — 2026-11-14 14:08:55 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `c4e20fb` — 2026-11-14 17:38:41 +0330 — add typescript worker process
  Files: scripts/typescript-worker.mjs,tests/plugin/typescript-worker.test.ts
- `6d4b51c` — 2026-11-14 14:06:56 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `37f14ab` — 2026-11-14 17:36:43 +0330 — run eight-player capacity fixture
  Files: roadmap_backend.md,tests/deploy/capacity-eight-player.test.ts
- `458cf7b` — 2026-11-14 14:06:16 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `0063cf9` — 2026-11-14 17:36:03 +0330 — define capacity reference profile
  Files: internal/deploy/capacity_profile.go,roadmap_backend.md tests/deploy/capacity_profile_test.go
- `42a69a2` — 2026-11-14 14:05:41 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `ff6c182` — 2026-11-14 17:35:27 +0330 — check backup data loss target
  Files: internal/deploy/backup_policy.go,roadmap_backend.md scripts/check-data-loss-target.sh,tests/deploy/backup_policy_test.go tests/deploy/data_loss_target_script_test.go
- `91019d6` — 2026-11-14 14:04:52 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `0e0ca7a` — 2026-11-14 17:34:40 +0330 — record recovery duration and retention
  Files: internal/deploy/recovery_record.go,roadmap_backend.md tests/deploy/recovery_record_test.go
- `e77288c` — 2026-11-14 14:04:09 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `d68bfe8` — 2026-11-14 17:33:56 +0330 — define fresh Compose restore rehearsal
  Files: roadmap_backend.md,scripts/restore-compose.sh tests/deploy/compose_restore_test.go
- `69efc6c` — 2026-11-14 14:03:30 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `303a8f4` — 2026-11-14 17:33:16 +0330 — detect missing object versions
  Files: internal/deploy/object_manifest.go,roadmap_backend.md tests/deploy/object_manifest_test.go
- `251c916` — 2026-11-14 14:02:44 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
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
