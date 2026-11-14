# Changelog

Automatically generated from commits on main.

## Unreleased
- `d928172` — 2026-11-15 00:16:08 +0330 — add representative app request capacity workload
  Files: internal/deploy/capacity_profile.go,tests/deploy/capacity_profile_test.go
- `265f342` — 2026-11-14 20:45:36 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `a453c74` — 2026-11-15 00:15:16 +0330 — require assignment pause before room drain
  Files: internal/deploy/sequence.go,tests/deploy/sequence_test.go
- `f94118b` — 2026-11-14 19:45:31 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `4aa27f7` — 2026-11-14 23:15:16 +0330 — stop workers when disable applies
  Files: internal/plugin/plugin.go,tests/plugin/plugin_test.go
- `e4477c9` — 2026-11-14 19:44:53 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `9a0cbaf` — 2026-11-14 23:14:39 +0330 — stop project workers when grants revoke
  Files: internal/plugin/plugin.go,internal/plugin/supervisor.go tests/plugin/plugin_test.go
- `9032485` — 2026-11-14 19:44:06 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `c3a1d91` — 2026-11-14 23:13:52 +0330 — stop revoked plugin worker versions
  Files: internal/plugin/supervisor.go,tests/plugin/supervisor_test.go
- `c664b18` — 2026-11-14 19:43:13 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `340c867` — 2026-11-14 23:12:54 +0330 — enforce project context in typescript workers
  Files: internal/plugin/typescript_worker.go,tests/plugin/typescript_worker_test.go
- `953ae02` — 2026-11-14 19:42:00 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `6633594` — 2026-11-14 23:11:46 +0330 — bind worker invocation to host project
  Files: internal/plugin/worker.go,tests/plugin/worker_test.go
- `c1e0000` — 2026-11-14 19:40:58 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `8fdadc2` — 2026-11-14 23:10:43 +0330 — bind typescript workers to plugin manifests
  Files: internal/plugin/typescript_worker.go,tests/plugin/bubblewrap_worker_test.go
- `9649c33` — 2026-11-14 19:38:03 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `083d722` — 2026-11-14 23:07:49 +0330 — avoid unavailable sandbox test dependency
  Files: scripts/test-sandboxed-typescript-worker.sh
- `418f2b8` — 2026-11-14 19:36:36 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `0dd6679` — 2026-11-14 23:06:18 +0330 — mount node runtime modules in sandbox
  Files: scripts/run-sandboxed-typescript-worker.sh
- `5fdbef1` — 2026-11-14 19:34:37 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `da26143` — 2026-11-14 23:04:22 +0330 — allow node snapshot address space in sandbox
  Files: scripts/test-sandboxed-typescript-worker.sh
- `63a82e5` — 2026-11-14 19:33:05 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `bb3122e` — 2026-11-14 23:02:47 +0330 — allow node startup memory in sandbox fixture
  Files: scripts/test-sandboxed-typescript-worker.sh
- `414b866` — 2026-11-14 19:31:14 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `b020f41` — 2026-11-14 23:00:57 +0330 — bound sandbox ci duration
  Files: .github/workflows/test.yml
- `c2d091c` — 2026-11-14 19:29:39 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `0cf1f85` — 2026-11-14 22:59:26 +0330 — bound sandbox network probe
  Files: scripts/test-sandboxed-typescript-worker.sh
- `d91d110` — 2026-11-14 19:28:03 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `6cbde4f` — 2026-11-14 22:57:50 +0330 — mount linux loader for sandbox node
  Files: scripts/run-sandboxed-typescript-worker.sh
- `b295124` — 2026-11-14 19:26:43 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `7b73658` — 2026-11-14 22:56:30 +0330 — launch nodejs path without symlink resolution
  Files: scripts/run-sandboxed-typescript-worker.sh
- `8958429` — 2026-11-14 19:25:28 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `ee7a868` — 2026-11-14 22:55:14 +0330 — use installed nodejs binary in sandbox ci
  Files: .github/workflows/test.yml,scripts/run-sandboxed-typescript-worker.sh
- `a09c001` — 2026-11-14 19:24:06 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
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
