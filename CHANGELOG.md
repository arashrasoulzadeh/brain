# Changelog

Automatically generated from commits on main.

## Unreleased
- `8046ba0` — 2026-11-15 00:41:32 +0330 — reject traversal in websocket project paths
  Files: internal/plugin/websocket_endpoint.go,tests/plugin/websocket_endpoint_test.go
- `3b13d37` — 2026-11-14 21:10:58 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `895d229` — 2026-11-15 00:40:45 +0330 — make assignment gate concurrency safe
  Files: internal/deploy/assignments.go,tests/deploy/assignments_test.go
- `e42fa25` — 2026-11-14 21:10:19 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `c389b41` — 2026-11-15 00:40:05 +0330 — bind database and object restores to release
  Files: internal/deploy/restore_pair.go,roadmap_backend.md tests/deploy/restore_pair_test.go
- `5b79130` — 2026-11-14 21:09:27 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `647fa51` — 2026-11-15 00:39:13 +0330 — remove completed backend roadmap items
  Files: roadmap_backend.md
- `d38d3f6` — 2026-11-14 21:08:32 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `42b85d3` — 2026-11-15 00:38:18 +0330 — apply worker test formatting
  Files: tests/plugin/worker_test.go
- `7e69e83` — 2026-11-15 00:37:57 +0330 — harden worker credentials and deployment controls
  Files: go.mod,internal/deploy/assignments.go internal/deploy/room_drain.go,internal/plugin/worker.go tests/plugin/worker_test.go
- `76926b0` — 2026-11-14 21:07:30 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `0b9cd72` — 2026-11-15 00:36:59 +0330 — verify multi-architecture release images
  Files: .github/workflows/release-backend.yml,tests/deploy/assignments_test.go tests/deploy/release_image_verification_test.go,tests/deploy/release_workflow_test.go tests/deploy/room_drain_test.go,tests/plugin/worker_test.go
- `e85471a` — 2026-11-14 21:07:07 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `e03cbf4` — 2026-11-15 00:36:56 +0330 — add manual assignment pause contract
  Files: internal/deploy/sequence.go,internal/plugin/plugin.go tests/deploy/sequence_test.go,tests/plugin/plugin_test.go
- `6a8b810` — 2026-11-14 21:04:38 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `4799692` — 2026-11-15 00:34:24 +0330 — remove completed backend roadmap items
  Files: roadmap_backend.md
- `319ddd9` — 2026-11-14 21:04:16 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `b36f5a7` — 2026-11-15 00:34:02 +0330 — use dedicated compiler and runner images
  Files: .github/workflows/release-backend.yml,Dockerfile.compiler Dockerfile.runner,tests/deploy/release_workflow_test.go
- `74f2af8` — 2026-11-14 21:03:08 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `1d644fd` — 2026-11-15 00:32:14 +0330 — publish multi-architecture backend images
  Files: .github/workflows/release-backend.yml,tests/deploy/release_workflow_test.go
- `7f77e43` — 2026-11-14 21:02:52 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `9e88566` — 2026-11-15 00:32:05 +0330 — gate packaged cross-project isolation
  Files: tests/deploy/packaged_isolation_test.go,tests/deploy/testdata/packaged-isolation-probe/main.go
- `1da1c17` — 2026-11-14 21:02:36 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `ffd22c3` — 2026-11-15 00:31:51 +0330 — add bounded worker recovery
  Files: internal/plugin/supervisor.go,tests/plugin/supervisor_test.go
- `7029390` — 2026-11-14 21:02:17 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `324637a` — 2026-11-15 00:31:39 +0330 — bind operator actions to authenticated sessions
  Files: internal/operator/operator.go,tests/operator/operator_test.go
- `e2065c0` — 2026-11-14 20:58:07 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `816bb46` — 2026-11-15 00:27:50 +0330 — add complete clean environment rehearsal
  Files: docs/deployment-runbook.md,scripts/clean-environment-rehearsal.sh tests/deploy/clean_environment_script_test.go
- `33b66a5` — 2026-11-14 20:56:58 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `f800672` — 2026-11-15 00:26:45 +0330 — rehearse clean deploy rollback in ci
  Files: .github/workflows/test.yml
- `3cee528` — 2026-11-14 20:56:25 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `5ca2610` — 2026-11-15 00:26:12 +0330 — guard clean environment runbook gate
  Files: tests/deploy/runbook_test.go
- `b1b0951` — 2026-11-14 20:55:46 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `478118b` — 2026-11-15 00:25:33 +0330 — publish restore runbook
  Files: docs/admins.md,docs/developers.md docs/restore-runbook.md
- `7345ece` — 2026-11-14 20:55:17 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `fce1623` — 2026-11-15 00:25:03 +0330 — publish deployment runbook
  Files: docs/admins.md,docs/deployment-runbook.md docs/developers.md
- `033edc9` — 2026-11-14 20:54:43 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `7020298` — 2026-11-15 00:24:29 +0330 — gate releases on packaged revocation
  Files: internal/deploy/manifest.go,tests/deploy/healthy_manifest_test.go tests/deploy/manifest_test.go,tests/deploy/rollback_test.go
- `a346e1c` — 2026-11-14 20:54:02 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `63933fa` — 2026-11-15 00:23:50 +0330 — gate releases on packaged isolation
  Files: internal/deploy/manifest.go,tests/deploy/healthy_manifest_test.go tests/deploy/manifest_test.go,tests/deploy/rollback_test.go
- `ab313ca` — 2026-11-14 20:53:19 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `810fe8a` — 2026-11-15 00:23:00 +0330 — set conservative capacity room cap
  Files: internal/deploy/capacity_profile.go,tests/deploy/capacity_profile_test.go
- `6e3b6fb` — 2026-11-14 20:51:58 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `fdaeb7e` — 2026-11-15 00:21:45 +0330 — require capacity resource and traffic measurements
  Files: internal/deploy/capacity_profile.go,tests/deploy/capacity_profile_test.go
- `cb0bda6` — 2026-11-14 20:50:31 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `4d35e91` — 2026-11-15 00:20:19 +0330 — require tick delay capacity measurement
  Files: internal/deploy/capacity_profile.go,tests/deploy/capacity_profile_test.go
- `5381577` — 2026-11-14 20:48:52 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `fb2ee42` — 2026-11-15 00:18:40 +0330 — require p95 snapshot latency measurement
  Files: internal/deploy/capacity_profile.go,tests/deploy/capacity_profile_test.go
- `917b06f` — 2026-11-14 20:48:04 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `ce3d564` — 2026-11-15 00:17:51 +0330 — bound packet loss capacity simulation
  Files: internal/deploy/capacity_profile.go,tests/deploy/capacity_profile_test.go
