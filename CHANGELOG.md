# Changelog

Automatically generated from commits on main.

## Unreleased
- `05ba5ce0` — 2026-11-27 19:26:14 +0330 — Merge pull request #38 from arashrasoulzadeh/feat/project-scoped-http-and-plugin-routes
  Files: 
- `173da987` — 2026-11-27 19:25:20 +0330 — Exercise authenticated Arena admission and invalid durable snapshots
  Files: tests/arena/arena-routes.test.ts
- `71fd24bd` — 2026-11-27 19:24:23 +0330 — Format WebSocket gateway table test
  Files: tests/plugin/websocket_gateway_test.go
- `0046b6b9` — 2026-11-27 19:24:18 +0330 — Cover project namespace configuration and persistence on memory and SQLite stores
  Files: tests/project/url_namespace_test.go
- `97834de2` — 2026-11-27 19:23:55 +0330 — Cover WebSocket gateway boundary and rejected socket messages
  Files: tests/plugin/websocket_gateway_test.go
- `d249e585` — 2026-11-27 19:23:28 +0330 — Cover WebSocket upgrade tokens and conflicting project identity
  Files: tests/plugin/websocket_endpoint_test.go
- `ca5405fe` — 2026-11-27 19:23:21 +0330 — Require an exact Connection upgrade token for WebSocket admission
  Files: internal/plugin/websocket_endpoint.go
- `5995ee43` — 2026-11-27 19:21:47 +0330 — Test project URL validation and reject invalid admin namespaces before transport
  Files: tests/client/admin-client.test.ts
- `9dd994b1` — 2026-11-27 19:18:37 +0330 — Exercise actual WebSocket upgrades, membership, replay and durable actor binding
  Files: tests/plugin/websocket_gateway_test.go
- `69a841a0` — 2026-11-27 19:18:04 +0330 — Vendor pinned coder WebSocket transport for offline reproducible builds
  Files: vendor/github.com/coder/websocket/LICENSE.txt,vendor/github.com/coder/websocket/README.md vendor/github.com/coder/websocket/accept.go,vendor/github.com/coder/websocket/close.go vendor/github.com/coder/websocket/compress.go,vendor/github.com/coder/websocket/conn.go vendor/github.com/coder/websocket/dial.go,vendor/github.com/coder/websocket/doc.go vendor/github.com/coder/websocket/errors.go,vendor/github.com/coder/websocket/frame.go vendor/github.com/coder/websocket/hijack.go,vendor/github.com/coder/websocket/internal/bpool/bpool.go vendor/github.com/coder/websocket/internal/errd/wrap.go,vendor/github.com/coder/websocket/internal/util/util.go vendor/github.com/coder/websocket/internal/wsjs/wsjs_js.go,vendor/github.com/coder/websocket/internal/xsync/go.go vendor/github.com/coder/websocket/mask.go,vendor/github.com/coder/websocket/mask_amd64.s vendor/github.com/coder/websocket/mask_arm64.s,vendor/github.com/coder/websocket/mask_asm.go vendor/github.com/coder/websocket/mask_go.go,vendor/github.com/coder/websocket/netconn.go vendor/github.com/coder/websocket/netconn_js.go,vendor/github.com/coder/websocket/netconn_notjs.go vendor/github.com/coder/websocket/read.go,vendor/github.com/coder/websocket/stringer.go vendor/github.com/coder/websocket/write.go,vendor/github.com/coder/websocket/ws_js.go vendor/modules.txt
- `1abba9c8` — 2026-11-27 19:16:11 +0330 — Implement authenticated arena membership admission and durable player inputs
  Files: docs/plugin-guides/arena.md,internal/plugin/legacy_arena_storage.go internal/plugin/websocket_gateway.go,plugins/arena/index.ts plugins/arena/plugin.yaml,plugins/arena/state-machine.ts plugins/sdk-typescript/src/manifest-operations.generated.d.ts,tests/arena/arena-routes.test.ts
- `558af223` — 2026-11-27 19:14:58 +0330 — Wire manifest-opted authenticated WebSocket gateway and bounded transport
  Files: docs/http-api.md,go.mod go.sum,internal/admin/lifecycle.go internal/app/api_response.go,internal/app/app.go internal/plugin/http_routes.go,internal/plugin/manifest.go internal/plugin/manifest_http.go,internal/plugin/websocket_gateway.go internal/plugin/websocket_ticket.go
- `8d82d580` — 2026-11-27 19:00:28 +0330 — Add bounded single-use WebSocket transport ticket binding
  Files: docs/http-api.md,internal/plugin/websocket_ticket.go tests/plugin/websocket_ticket_test.go
- `5efdf699` — 2026-11-27 18:55:17 +0330 — Update admin transport fixture and format routing regression test
  Files: tests/coverage/plugin-boundaries.test.ts,tests/plugin/http_version_routing_test.go
- `86e71f47` — 2026-11-27 18:54:29 +0330 — Revalidate manifest HTTP route selection at invocation and commit
  Files: CHANGELOG.md,internal/admin/lifecycle.go internal/plugin/http_routes.go,internal/plugin/model_runtime.go internal/plugin/registry_workers.go,internal/plugin/worker.go tests/plugin/http_version_routing_test.go
- `79a83e02` — 2026-11-27 18:52:06 +0330 — Enforce project command namespace and reserve global lifecycle operations
  Files: admin-nextjs/app/page.tsx,admin-nextjs/app/plugins/[plugin]/page.tsx clients/javascript/src/admin.ts,docs/http-api.md internal/app/admin.go,internal/app/app.go scripts/test-admin-browser.mjs,scripts/test-admin-live.mjs scripts/test-runtime-api.mjs,tests/app/admin_namespace_test.go tests/app/admin_test.go,tests/client/admin-client.test.ts
- `f15998af` — 2026-11-27 15:19:36 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `6e6115eb` — 2026-11-27 18:49:07 +0330 — Integrate merged manifest SDK improvements into project routing
  Files: docs/sdk-reference.md
- `c6f51829` — 2026-11-27 18:48:32 +0330 — Merge pull request #39 from arashrasoulzadeh/refactor/manifest-driven-runtime
  Files: 
- `07f5ee22` — 2026-11-27 18:44:19 +0330 — Use atomic CLI coverage mode consistently
  Files: cmd/manifest-sdk/main_test.go
- `8dea3425` — 2026-11-27 18:43:57 +0330 — Exercise manifest generator command success and failure with CLI coverage
  Files: cmd/manifest-sdk/main_test.go
- `7cce76cd` — 2026-11-27 18:41:17 +0330 — Test manifest generator failure paths and emit type-only declaration file
  Files: cmd/manifest-sdk/main.go,cmd/manifest-sdk/main_test.go plugins/sdk-typescript/src/manifest-operations.generated.d.ts,tests/plugin/manifest_sdk_test.go
- `5a1e26f6` — 2026-11-27 18:40:19 +0330 — Migrate remaining admin views to project routes and retain macOS startup fix
  Files: .github/workflows/test.yml,admin-nextjs/app/components/legacy-admin-page.tsx admin-nextjs/app/functions/page.tsx
- `c9633148` — 2026-11-27 18:38:20 +0330 — Reconcile manifest runtime refactor with current secure host and SDK
  Files: docs/sdk-reference.md,plugins/functions/scheduling.ts plugins/matchmaking/queue.ts,plugins/sdk-typescript/src/index.ts plugins/sdk-typescript/src/manifest-operations.generated.ts,plugins/sdk-typescript/src/request.ts tests/plugin/manifest-request.test.ts
- `c1b6c8e8` — 2026-11-27 15:05:18 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `8bf6cb5c` — 2026-11-27 18:34:14 +0330 — Merge pull request #40 from arashrasoulzadeh/fix/macos-homebrew-openssl
  Files: 
- `03f401b1` — 2026-11-27 18:33:23 +0330 — Correct Homebrew isolation CI step formatting
  Files: .github/workflows/test.yml
- `c80d4a0e` — 2026-11-27 18:33:02 +0330 — Fix native macOS worker startup with Homebrew OpenSSL
  Files: .github/workflows/test.yml,internal/plugin/macos_profile_coverage_test.go internal/plugin/macos_worker.go
- `421f3c29` — 2026-11-27 18:25:52 +0330 — Migrate transport coverage to project WebSocket route and reject legacy URL
  Files: tests/plugin/transport_coverage_edges_test.go
- `8ec51d35` — 2026-11-27 18:24:17 +0330 — Fix HTTP router build and document project WebSocket boundary
  Files: docs/http-api.md,internal/app/app.go
- `d9e9c74a` — 2026-11-27 18:22:52 +0330 — Merge current main into project routing; preserve macOS fixes and migrate CI probes
  Files: .github/workflows/test.yml,CHANGELOG.md
- `eab67f5c` — 2026-11-27 18:05:34 +0330 — fix: distinguish root routes from parameters and authorize before command parsing
  Files: internal/app/app.go,internal/app/project_http.go internal/plugin/manifest_http.go
- `b31d765a` — 2026-11-27 18:04:32 +0330 — feat: mount WebSocket ticket boundary only at host-resolved /{project}/ws
  Files: internal/app/app.go,internal/plugin/websocket_endpoint.go tests/plugin/websocket_endpoint_test.go
- `84aeda75` — 2026-11-27 18:03:38 +0330 — fix: grant macOS worker narrow system directory lookup needed by Node startup
  Files: internal/plugin/macos_worker.go
- `40c32407` — 2026-11-27 18:02:24 +0330 — style: run gofmt on manifest route registry and HTTP lifecycle tests
  Files: internal/plugin/http_routes.go,tests/app/project_plugin_http_test.go tests/plugin/http_version_routing_test.go
- `413f8617` — 2026-11-27 18:01:07 +0330 — feat: centralized project URL validation for JavaScript SDK and operator actions
  Files: clients/javascript/src/admin.ts,clients/javascript/src/client.ts clients/javascript/src/index.ts,tests/client/admin-client.test.ts
- `73382bc9` — 2026-11-27 18:00:33 +0330 — docs: plugin manifest HTTP schema, admin route discovery and Go vet CI gate
  Files: .github/workflows/test.yml,docs/admin-pages/projects.md docs/admin-pages/routes.md,docs/developers.md docs/plugin-manifest.md,docs/sdk-reference.md roadmap_backend.md
- `09a74b1f` — 2026-11-27 17:59:42 +0330 — fix: export project HTTP route model as valid TypeScript
  Files: admin-nextjs/app/lib/admin-types.ts
- `ef4572e3` — 2026-11-27 17:58:50 +0330 — fix: retain existing project API casing for new metadata
  Files: internal/project/project.go
- `3439d92c` — 2026-11-27 17:57:17 +0330 — test: project HTTP dispatch authorization, concurrent disable and version rollback
  Files: tests/app/project_plugin_http_test.go,tests/plugin/http_version_routing_test.go
- `5849613a` — 2026-11-27 17:56:37 +0330 — feat: project metadata editor, live route discovery and scoped operator actions
  Files: admin-nextjs/app/components/projects-panel.tsx,admin-nextjs/app/components/routes-panel.tsx admin-nextjs/app/lib/admin-types.ts,admin-nextjs/app/page.tsx
- `c9f21f50` — 2026-11-27 17:55:32 +0330 — style: apply CI gofmt to strict project HTTP dispatcher
  Files: internal/app/project_http.go
- `6bdb06ef` — 2026-11-27 17:55:14 +0330 — test: open namespaced operator panel in live browser smoke
  Files: scripts/test-admin-live.mjs
- `61441a06` — 2026-11-27 17:54:56 +0330 — docs: correct project.configure inventory row separator
  Files: docs/http-api.md
- `ec4717ea` — 2026-11-27 17:53:01 +0330 — feat: expose route bindings and capabilities, and reject ambiguous manifest paths
  Files: internal/plugin/http_routes.go,internal/plugin/manifest_http.go
- `188b7013` — 2026-11-27 17:51:46 +0330 — security: reject duplicated project identities and published function body conflicts
  Files: internal/app/admin_functions.go,internal/project/url_middleware.go
- `ad9c7833` — 2026-11-27 17:51:24 +0330 — fix: bind raw JSON project identities without precision loss and reject duplicate fields
  Files: internal/app/project_http.go,internal/plugin/manifest_http.go tests/app/operator_auth_test.go
- `51edf1a5` — 2026-11-27 17:49:56 +0330 — feat: support project user authentication via existing host session validator
  Files: internal/admin/lifecycle.go,internal/app/project_http.go internal/plugin/manifest_http.go
- `bfd8fd9f` — 2026-11-27 17:48:27 +0330 — fix: exclude reserved system UI and static assets from dynamic plugin proxy rewrite
  Files: admin-nextjs/next.config.ts
- `e97266e1` — 2026-11-27 17:48:00 +0330 — fix: preserve JSON response envelope and migrate CI health URLs
  Files: .github/workflows/test.yml,docs/http-api.md internal/app/api_response.go
