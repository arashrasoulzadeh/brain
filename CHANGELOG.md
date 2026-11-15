# Changelog

Automatically generated from commits on main.

## Unreleased
- `91b82bc` — 2026-11-15 12:51:05 +0330 — move admin panel to nextjs
  Files: admin-nextjs/app/layout.tsx,admin-nextjs/app/page.tsx admin-nextjs/app/styles.css,admin-nextjs/next-env.d.ts admin-nextjs/next.config.ts,admin-nextjs/package.json admin-nextjs/tsconfig.json,cmd/brain/main.go internal/admin/lifecycle.go,internal/app/admin.go internal/app/app.go,internal/project/project.go plugins/dashboard/admin-panel.html,plugins/dashboard/plugin.yaml tests/app/admin_test.go
- `74597af` — 2026-11-15 09:18:45 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `ede74b7` — 2026-11-15 12:48:31 +0330 — add browser admin login panel
  Files: cmd/brain/main.go,internal/app/app.go plugins/dashboard/admin-panel.html,plugins/dashboard/plugin.yaml tests/app/admin_test.go
- `54a2560` — 2026-11-15 09:13:22 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `7984810` — 2026-11-15 12:43:09 +0330 — add dashboard admin panel and login client
  Files: clients/javascript/src/admin.ts,clients/javascript/src/index.ts plugins/dashboard/admin-panel.ts,plugins/dashboard/index.ts plugins/dashboard/plugin.yaml,tests/client/admin-client.test.ts tests/dashboard/dashboard.test.ts
- `60b0376` — 2026-11-15 09:09:45 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `88f7b87` — 2026-11-15 12:39:23 +0330 — mark packaged admin verification complete
  Files: roadmap_admin.md
- `64162e9` — 2026-11-15 09:09:26 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `166f0f6` — 2026-11-15 12:39:10 +0330 — verify packaged admin composition
  Files: cmd/brain/main_test.go
- `d8fbe27` — 2026-11-15 09:08:58 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `ec00c8d` — 2026-11-15 12:38:45 +0330 — compose production operator admin server
  Files: cmd/brain/main.go,internal/app/admin.go internal/app/app.go,tests/app/admin_test.go
- `222ccf4` — 2026-11-15 09:07:47 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `729e411` — 2026-11-15 12:37:32 +0330 — document authenticated admin sessions
  Files: docs/admins.html,docs/admins.md
- `e99ed11` — 2026-11-15 09:07:13 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `baa2d17` — 2026-11-15 12:37:00 +0330 — wire operator sessions to admin boundary
  Files: internal/app/admin.go,internal/operator/operator.go internal/operator/session.go,tests/app/admin_test.go tests/operator/operator_test.go
- `b051ca7` — 2026-11-15 09:05:17 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `2bd1147` — 2026-11-15 12:35:04 +0330 — record admin isolation release gates
  Files: roadmap_admin.md
- `e39e00b` — 2026-11-15 09:04:54 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `640915a` — 2026-11-15 12:34:40 +0330 — record administrator runbooks complete
  Files: roadmap_admin.md
- `52d00a6` — 2026-11-15 09:04:31 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `46c1cba` — 2026-11-15 12:34:11 +0330 — record verified admin security controls
  Files: roadmap_admin.md
- `d7168e6` — 2026-11-15 09:04:14 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `5e2e969` — 2026-11-15 12:33:56 +0330 — require admin confirmation for destructive actions
  Files: internal/app/admin.go,tests/app/admin_test.go
- `8d9cae8` — 2026-11-15 09:03:18 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `7519519` — 2026-11-15 12:33:03 +0330 — mark dashboard routine operations complete
  Files: roadmap_admin.md
- `00e528d` — 2026-11-15 09:03:01 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `1e442c2` — 2026-11-15 12:32:48 +0330 — add routine dashboard operation pages
  Files: plugins/dashboard/function-management-page.ts,plugins/dashboard/index.ts plugins/dashboard/live-session-page.ts,plugins/dashboard/plugin.yaml plugins/dashboard/types.ts,plugins/dashboard/workflow-management-page.ts tests/dashboard/dashboard.test.ts
- `734a219` — 2026-11-15 09:01:57 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `227a014` — 2026-11-15 12:31:42 +0330 — expose audited admin lifecycle operations
  Files: internal/app/admin.go,tests/app/admin_test.go
- `0f37015` — 2026-11-15 09:01:07 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `b7571e2` — 2026-11-15 12:30:54 +0330 — complete plugin security audit gate
  Files: roadmap_backend.md
- `c451bc2` — 2026-11-15 09:00:26 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `f49fd18` — 2026-11-15 12:30:13 +0330 — record compose rehearsal gate
  Files: roadmap_backend.md
- `11b7521` — 2026-11-15 09:00:02 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `0d96aae` — 2026-11-15 12:29:47 +0330 — record audited action gate
  Files: roadmap_backend.md
- `21e848c` — 2026-11-15 08:59:38 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `57724d8` — 2026-11-15 12:29:26 +0330 — record verified recovery gate
  Files: roadmap_backend.md
- `9ced1dd` — 2026-11-15 08:59:12 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `c84040f` — 2026-11-15 12:28:59 +0330 — harden malformed authentication digests
  Files: plugins/auth/email-otp-verifier.ts,plugins/auth/recovery-token-consumer.ts plugins/auth/sms-otp-verifier.ts,tests/auth/auth.test.ts
- `06167ad` — 2026-11-15 08:58:11 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `1dae3fd` — 2026-11-15 12:27:58 +0330 — record verified backend release gates
  Files: roadmap_backend.md
- `ccba3f7` — 2026-11-15 08:57:42 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `dbb1983` — 2026-11-15 12:27:30 +0330 — complete project boundary audit pass
  Files: plugins/functions/record-trigger-matcher.ts,plugins/functions/shared-helper-publisher.ts plugins/matchmaking/matchmaking-config-reader.ts,plugins/matchmaking/matchmaking-config-status.ts plugins/rooms/authorized-room-snapshot.ts,plugins/rooms/authorized-tic-tac-toe-state.ts plugins/rooms/reconnect-tic-tac-toe.ts,tests/matchmaking/matchmaking.test.ts tests/rooms/rooms.test.ts
- `2b97bcf` — 2026-11-15 08:56:14 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `873210b` — 2026-11-15 12:26:00 +0330 — harden function state project boundaries
  Files: plugins/functions/authorize-project-workflow-action.ts,plugins/functions/function-draft-reader.ts plugins/functions/function-publication-reader.ts,plugins/functions/shared-package-reader.ts tests/functions/functions.test.ts,tests/functions/workflow-components.test.ts
- `f9d36eb` — 2026-11-15 08:55:13 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `b49cb7d` — 2026-11-15 12:24:56 +0330 — mark operator session isolation complete
  Files: roadmap_backend.md
- `df33381` — 2026-11-15 08:54:56 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `a51ac81` — 2026-11-15 12:24:41 +0330 — verify operator session isolation
  Files: tests/auth/auth.test.ts
- `8084f95` — 2026-11-15 08:54:19 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
