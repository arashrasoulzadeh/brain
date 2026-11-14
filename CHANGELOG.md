# Changelog

Automatically generated from commits on main.

## Unreleased
- `6a82173` — 2026-11-15 02:15:23 +0330 — validate file object identities at access boundaries
  Files: plugins/files/authorized-upload-writer.ts,plugins/files/file-access-deleter.ts plugins/files/file-access-reader.ts,plugins/files/pending-upload-publisher.ts plugins/files/pending-upload-reader.ts,tests/files/files.test.ts
- `992991f` — 2026-11-14 22:44:07 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `ee5a470` — 2026-11-15 02:13:48 +0330 — validate project identity in auth token consumers
  Files: plugins/auth/email-otp-verifier.ts,plugins/auth/recovery-token-consumer.ts plugins/auth/sms-otp-verifier.ts,tests/auth/auth.test.ts
- `f290136` — 2026-11-14 22:43:06 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `1e4acf0` — 2026-11-15 02:12:46 +0330 — audit every plugin manifest boundary
  Files: tests/plugin/all_manifests_test.go
- `143c435` — 2026-11-14 22:42:13 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `33a5162` — 2026-11-15 02:11:52 +0330 — validate dashboard project ids before html rendering
  Files: docs/admins.md,docs/developers.md plugins/dashboard/custom-admin-view.ts,plugins/dashboard/plugin.yaml tests/dashboard/dashboard.test.ts
- `d26605e` — 2026-11-14 22:37:26 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `6ffa6ed` — 2026-11-15 02:07:11 +0330 — verify single use commit credentials
  Files: roadmap_backend.md
- `063afca` — 2026-11-14 22:37:09 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `f479cbe` — 2026-11-15 02:06:53 +0330 — verify host authenticated plugin execution
  Files: roadmap_backend.md
- `def42f1` — 2026-11-14 22:36:48 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `2d3f209` — 2026-11-15 02:06:31 +0330 — verify project isolation boundaries
  Files: roadmap_backend.md
- `00dd5b2` — 2026-11-14 22:36:26 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `b68fd4f` — 2026-11-15 02:06:07 +0330 — verify plugin-first host boundary
  Files: docs/admins.md,docs/developers.md roadmap_backend.md
- `3956b5c` — 2026-11-14 22:35:11 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `3dd11a5` — 2026-11-15 02:04:52 +0330 — refresh developer and administrator documentation
  Files: docs/admins.html,docs/admins.md docs/developers.html,docs/developers.md docs/index.html,docs/index.md docs/site.css
- `b0d6cb2` — 2026-11-14 22:30:16 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `036e53f` — 2026-11-15 01:59:57 +0330 — validate manifests at registry registration
  Files: internal/plugin/plugin.go,tests/plugin/plugin_test.go
- `d1ddf96` — 2026-11-14 22:28:24 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `2ff2af8` — 2026-11-15 01:58:08 +0330 — enforce project identity in disable previews
  Files: internal/plugin/plugin.go,tests/plugin/plugin_test.go
- `ab341ac` — 2026-11-14 22:27:39 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `52d292f` — 2026-11-15 01:57:23 +0330 — reject duplicate plugin installations
  Files: internal/plugin/plugin.go,tests/plugin/plugin_test.go
- `cc88191` — 2026-11-14 22:26:54 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `e7472cb` — 2026-11-15 01:56:29 +0330 — lock permission transitions during registry changes
  Files: internal/plugin/plugin.go,tests/plugin/plugin_test.go
- `e83fb51` — 2026-11-14 22:24:38 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `163a7d2` — 2026-11-15 01:54:24 +0330 — order plugin approval locks
  Files: internal/plugin/plugin.go,tests/plugin/plugin_test.go
- `a39cc10` — 2026-11-14 22:24:01 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `0bb6e66` — 2026-11-15 01:53:47 +0330 — require plugin resource identifiers
  Files: internal/plugin/plugin.go,tests/plugin/plugin_test.go
- `2f99655` — 2026-11-14 22:23:28 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `9b009d1` — 2026-11-15 01:53:14 +0330 — validate plugin capability approvals
  Files: internal/plugin/plugin.go,tests/plugin/plugin_test.go
- `18bf5bf` — 2026-11-14 22:22:53 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `47c24bc` — 2026-11-15 01:52:42 +0330 — reject ungranted plugin revocations
  Files: internal/plugin/plugin.go,tests/plugin/plugin_test.go
- `3e10e2d` — 2026-11-14 22:22:21 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `d31f7c9` — 2026-11-15 01:52:08 +0330 — validate plugin revocation capability
  Files: internal/plugin/plugin.go,tests/plugin/plugin_test.go
- `8ec5743` — 2026-11-14 22:21:33 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `ba89d08` — 2026-11-15 01:51:21 +0330 — stop runtime on immediate stop
  Files: internal/plugin/plugin.go,tests/plugin/plugin_test.go
- `ef6cec1` — 2026-11-14 22:20:56 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `e51070f` — 2026-11-15 01:50:43 +0330 — stop runtime on plugin removal
  Files: internal/plugin/plugin.go,tests/plugin/plugin_test.go
- `5fe8a5b` — 2026-11-14 22:20:18 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `74c953b` — 2026-11-15 01:50:06 +0330 — surface failed project recovery
  Files: internal/plugin/plugin.go,tests/plugin/plugin_test.go
- `170c90e` — 2026-11-14 22:19:22 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `9a3d085` — 2026-11-15 01:49:09 +0330 — close sockets during project recovery
  Files: internal/plugin/plugin.go,tests/plugin/plugin_test.go
- `038867f` — 2026-11-14 22:18:47 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `7d91db3` — 2026-11-15 01:48:35 +0330 — recover project plugin workers
  Files: internal/plugin/plugin.go,tests/plugin/plugin_test.go
- `fd0b797` — 2026-11-14 22:17:54 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `2492765` — 2026-11-15 01:47:37 +0330 — stop project workers on uninstall
  Files: internal/plugin/plugin.go,tests/plugin/plugin_test.go
- `7fa32a1` — 2026-11-14 22:17:11 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `5a921bc` — 2026-11-15 01:46:57 +0330 — invalidate plugin lifecycle previews
  Files: internal/plugin/plugin.go,tests/plugin/plugin_test.go
- `605a9c9` — 2026-11-14 22:16:26 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
