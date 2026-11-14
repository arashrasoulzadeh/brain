# Changelog

Automatically generated from commits on main.

## Unreleased
- `702f508` — 2026-11-14 22:53:53 +0330 — resolve sandbox node symlink
  Files: scripts/run-sandboxed-typescript-worker.sh
- `6ceafce` — 2026-11-14 19:22:48 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `00f88bd` — 2026-11-14 22:52:34 +0330 — use explicit debian node in sandbox ci
  Files: .github/workflows/test.yml,scripts/run-sandboxed-typescript-worker.sh
- `96a5b3c` — 2026-11-14 19:21:30 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `d9cad6a` — 2026-11-14 22:51:16 +0330 — bind sandbox node executable explicitly
  Files: scripts/run-sandboxed-typescript-worker.sh
- `636007d` — 2026-11-14 19:19:55 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `b29c103` — 2026-11-14 22:49:38 +0330 — bind node runtime trees without overlap
  Files: scripts/run-sandboxed-typescript-worker.sh
- `26f21a9` — 2026-11-14 19:18:22 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `9fe666d` — 2026-11-14 22:48:09 +0330 — bind node executable after runtime directory
  Files: scripts/run-sandboxed-typescript-worker.sh
- `a3ba68c` — 2026-11-14 19:16:59 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `b14ff0e` — 2026-11-14 22:46:42 +0330 — preserve node runtime mount in sandbox
  Files: scripts/run-sandboxed-typescript-worker.sh
- `72a7447` — 2026-11-14 19:15:41 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `2aa22bb` — 2026-11-14 22:45:28 +0330 — preserve sandbox mounts under isolated tmp
  Files: scripts/run-sandboxed-typescript-worker.sh
- `51fc3a4` — 2026-11-14 19:14:20 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `ba30916` — 2026-11-14 22:44:06 +0330 — stabilize sandbox worker host test
  Files: tests/plugin/bubblewrap_worker_test.go
- `dbcf3e7` — 2026-11-14 19:13:19 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `56fb77e` — 2026-11-14 22:43:05 +0330 — run sandbox verification in privileged ci container
  Files: .github/workflows/test.yml
- `287c439` — 2026-11-14 19:12:08 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `40cbedd` — 2026-11-14 22:41:50 +0330 — pin sandbox ci to namespace capable runner
  Files: .github/workflows/test.yml
- `efd818b` — 2026-11-14 19:10:15 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `3e7650e` — 2026-11-14 22:40:00 +0330 — run sandbox launcher with bash limits
  Files: .github/workflows/test.yml,scripts/run-sandboxed-typescript-worker.sh scripts/test-sandboxed-typescript-worker.sh
- `bca43e5` — 2026-11-14 19:08:41 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `c553fda` — 2026-11-14 22:38:27 +0330 — verify typescript sandbox in linux ci
  Files: .github/workflows/test.yml,scripts/run-sandboxed-typescript-worker.sh scripts/test-sandboxed-typescript-worker.sh
- `828efc2` — 2026-11-14 19:07:13 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
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
