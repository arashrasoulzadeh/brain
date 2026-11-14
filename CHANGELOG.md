# Changelog

Automatically generated from commits on main.

## Unreleased
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
