# Changelog

Automatically generated from commits on main.

## Unreleased
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
- `f62fe0e` — 2026-11-14 20:47:32 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `e2fbc55` — 2026-11-15 00:17:19 +0330 — add eye export capacity workload
  Files: internal/deploy/capacity_profile.go,tests/deploy/capacity_profile_test.go
- `2db493b` — 2026-11-14 20:47:00 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `858302f` — 2026-11-15 00:16:43 +0330 — add upload and job capacity workload
  Files: internal/deploy/capacity_profile.go,tests/deploy/capacity_profile_test.go
- `b15892d` — 2026-11-14 20:46:22 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
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
