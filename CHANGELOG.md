# Changelog

Automatically generated from commits on main.

## Unreleased
- `904d0fe` — 2026-11-27 03:39:24 +0330 — Centralize project access management across plugins
  Files: internal/admin/functions.go,internal/admin/lifecycle.go internal/plugin/access_management.go,internal/plugin/broker.go internal/plugin/manifest.go,internal/plugin/model_runtime.go internal/plugin/observations.go,internal/plugin/plugin_call.go internal/plugin/registry_workers.go,internal/plugin/worker.go plugins/README.md,plugins/access-management/README.md plugins/access-management/index.ts,plugins/access-management/plugin.yaml plugins/analytics/plugin.yaml,plugins/arena/plugin.yaml plugins/files/plugin.yaml,plugins/files/routes.ts plugins/functions/plugin.yaml,plugins/invites/plugin.yaml plugins/matchmaking/plugin.yaml,plugins/matchmaking/routes.ts plugins/rooms/plugin.yaml,plugins/rooms/routes.ts plugins/sdk-typescript/src/index.ts,plugins/sdk-typescript/src/request.ts plugins/subscriptions/plugin.yaml,plugins/tables/plugin.yaml tests/files/files.test.ts,tests/matchmaking/matchmaking.test.ts tests/plugin/access_management_test.go,tests/plugin/resource-request.test.ts tests/rooms/rooms.test.ts
- `d6c635b` — 2026-11-26 23:54:59 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `94cca6f` — 2026-11-27 03:24:21 +0330 — Fix gofmt alignment in declarative manifest test
  Files: internal/plugin/manifest_declarative_test.go
- `43467eb` — 2026-11-26 23:52:36 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `c993ddd` — 2026-11-27 03:21:46 +0330 — Merge pull request #7 from arashrasoulzadeh/dynamic-runtime-foundation
  Files: 
- `60022c4` — 2026-11-27 03:21:28 +0330 — Merge main and integrate declarative runtime with manifest-driven admin
  Files: admin-nextjs/app/plugins/[plugin]/page.tsx,internal/plugin/manifest.go internal/plugin/plugin.go,plugins/files/plugin.yaml plugins/matchmaking/plugin.yaml,plugins/rooms/plugin.yaml plugins/tables/plugin.yaml
- `8bd03b2` — 2026-11-26 23:50:47 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `7f9181d` — 2026-11-27 03:19:43 +0330 — Merge pull request #8 from arashrasoulzadeh/admin/manifest-driven-ui
  Files: 
- `ddea07e` — 2026-11-27 03:13:04 +0330 — Restore executable mode of sandbox launcher
  Files: scripts/run-sandboxed-typescript-worker.sh
- `4ae66bc` — 2026-11-27 03:12:31 +0330 — Generate project configuration UI and document declarative runtime
  Files: admin-nextjs/app/plugins/[plugin]/page.tsx,docs/declarative-runtime.md
- `d57b5fd` — 2026-11-27 03:12:21 +0330 — Add declarative result component registry and update admin migration roadmap
  Files: admin-nextjs/app/components/admin-renderer.tsx,admin-nextjs/app/components/admin-results.tsx admin-nextjs/app/lib/admin-result-data.ts,admin-nextjs/app/lib/admin-types.ts docs/admin-pages/extensions.md,internal/plugin/admin_manifest.go internal/plugin/admin_results.go,internal/plugin/admin_results_test.go plugins/files/plugin.yaml,plugins/rooms/plugin.yaml plugins/tables/plugin.yaml,roadmap_admin.md tests/admin/admin-result-data.test.ts
- `9fd583b` — 2026-11-27 03:11:47 +0330 — Add capability-checked declarative cross-plugin calls
  Files: internal/plugin/manifest.go,internal/plugin/plugin_call.go internal/plugin/registry_workers.go,scripts/typescript-worker.mjs tests/plugin/plugin_call_test.go
- `e944040` — 2026-11-27 03:11:33 +0330 — Add reusable schema configuration editor and scoped admin integration tests
  Files: admin-nextjs/app/components/admin-config-form.tsx,admin-nextjs/app/functions/[name]/page.tsx admin-nextjs/app/functions/new/page.tsx,admin-nextjs/app/functions/page.tsx admin-nextjs/app/lib/admin-api.ts,admin-nextjs/app/lib/admin-config.ts admin-nextjs/app/page.tsx,admin-nextjs/app/plugins/[plugin]/page.tsx admin-nextjs/app/workers/page.tsx,internal/app/admin.go tests/admin/admin-config.test.ts,tests/admin/admin-registry.test.ts tests/admin/manifest_admin_test.go
- `d268662` — 2026-11-27 03:11:07 +0330 — Support source-free declarative plugin manifests and worker startup
  Files: examples/declarative-plugin/plugin.yaml,internal/plugin/contract_types.go internal/plugin/manifest.go,internal/plugin/manifest_declarative_test.go internal/plugin/registry_catalog.go,internal/plugin/typescript_worker.go scripts/run-sandboxed-typescript-worker.sh,scripts/typescript-worker.mjs
- `2e40592` — 2026-11-27 03:10:44 +0330 — Generate manifest SDK contracts and persist room observations
  Files: cmd/brain/main.go,internal/admin/functions.go internal/plugin/contract_types.go,internal/plugin/contract_types_test.go internal/plugin/legacy_storage_adapter.go,internal/plugin/manifest.go
- `2641aaa` — 2026-11-27 03:10:40 +0330 — Migrate four legacy admin consoles to manifest renderer and preserve project selection
  Files: admin-nextjs/app/[legacy]/page.tsx,admin-nextjs/app/components/admin-chrome.tsx admin-nextjs/app/components/admin-renderer.tsx,admin-nextjs/app/components/legacy-admin-page.tsx admin-nextjs/app/database/page.tsx,admin-nextjs/app/extensions/[project]/[plugin]/[version]/[page]/page.tsx admin-nextjs/app/files/page.tsx,admin-nextjs/app/lib/admin-api.ts admin-nextjs/app/lib/admin-navigation.ts,admin-nextjs/app/live-sessions/page.tsx admin-nextjs/app/page.tsx,admin-nextjs/app/workflows/page.tsx internal/plugin/admin_manifest.go,tests/admin/admin-navigation.test.ts tests/admin/admin-shell.test.ts
- `b6c5e63` — 2026-11-27 03:10:01 +0330 — Declare complete legacy console operations in plugin manifests
  Files: admin-nextjs/app/lib/admin-registry.ts,admin-nextjs/app/lib/admin-types.ts internal/admin/lifecycle.go,internal/plugin/admin_manifest.go plugins/files/plugin.yaml,plugins/functions/plugin.yaml plugins/matchmaking/plugin.yaml,plugins/rooms/plugin.yaml plugins/tables/plugin.yaml
- `5436ac2` — 2026-11-27 03:08:39 +0330 — Share pure helper registry across YAML, SDK and user function VM
  Files: implementation/dynamic-runtime.md,internal/admin/functions.go internal/plugin/manifest_pipeline.go,plugins/sdk-typescript/helpers.go plugins/sdk-typescript/src/helpers.d.mts,plugins/sdk-typescript/src/helpers.mjs plugins/sdk-typescript/src/index.ts,scripts/run-sandboxed-typescript-worker.sh scripts/typescript-worker.mjs,tests/admin/function_helpers_test.go tests/plugin/pipeline-worker.test.ts,tests/plugin/project-context.test.ts
- `9558c45` — 2026-11-27 03:08:16 +0330 — Connect declarative admin queries to project observation invalidation
  Files: admin-nextjs/app/extensions/[project]/[plugin]/[version]/[page]/page.tsx,admin-nextjs/app/lib/admin-sdk.ts admin-nextjs/app/lib/admin-types.ts,docs/admin-pages/extensions.md plugins/files/plugin.yaml,plugins/matchmaking/plugin.yaml plugins/rooms/plugin.yaml,plugins/tables/plugin.yaml roadmap_admin.md
- `fe3e065` — 2026-11-27 03:07:52 +0330 — Add bounded project observation revisions and read-only admin query boundary
  Files: internal/admin/lifecycle.go,internal/app/admin.go internal/plugin/admin_manifest.go,internal/plugin/admin_manifest_test.go internal/plugin/admin_observations.go,internal/plugin/admin_observations_test.go internal/plugin/observations.go,internal/plugin/plugin.go
- `2246897` — 2026-11-27 03:07:39 +0330 — Migrate existing plugin routes to sanitized domain observations
  Files: implementation/dynamic-runtime.md,plugins/auth/plugin.yaml plugins/auth/service.ts,plugins/files/plugin.yaml plugins/files/routes.ts,plugins/matchmaking/plugin.yaml plugins/matchmaking/routes.ts,plugins/rooms/plugin.yaml plugins/rooms/routes.ts,plugins/tables/plugin.yaml plugins/tables/routes.ts
- `051fdaf` — 2026-11-27 03:07:16 +0330 — Inject project configuration and optional dependency context
  Files: implementation/dynamic-runtime.md,internal/plugin/invocation_context.go internal/plugin/manifest.go,internal/plugin/registry_permissions.go internal/plugin/registry_workers.go,internal/plugin/typescript_worker.go internal/plugin/worker.go,plugins/arena/plugin.yaml plugins/matchmaking/plugin.yaml,plugins/sdk-typescript/src/index.ts scripts/typescript-worker.mjs
- `ab00a24` — 2026-11-27 03:06:42 +0330 — Correct manifest pattern and format observation response structs
  Files: internal/plugin/admin_manifest.go,internal/plugin/typescript_worker.go internal/plugin/worker.go
- `888c3d3` — 2026-11-27 03:06:12 +0330 — Declare typed admin fields in six plugin manifests
  Files: plugins/analytics/plugin.yaml,plugins/files/plugin.yaml plugins/functions/plugin.yaml,plugins/matchmaking/plugin.yaml plugins/rooms/plugin.yaml,plugins/tables/plugin.yaml
- `a186724` — 2026-11-27 03:06:02 +0330 — Add typed admin form validation and cross-plugin extension links
  Files: admin-nextjs/app/extensions/[project]/[plugin]/[version]/[page]/page.tsx,admin-nextjs/app/lib/admin-registry.ts admin-nextjs/app/lib/admin-types.ts,docs/admin-pages/extensions.md internal/admin/lifecycle.go,internal/plugin/admin_manifest.go internal/plugin/admin_manifest_test.go,roadmap_admin.md tests/admin/admin-registry.test.ts
- `7047a7f` — 2026-11-27 03:05:39 +0330 — Tighten project and grant checks and remove unimplemented actions
  Files: admin-nextjs/app/lib/admin-registry.ts,docs/admin-pages/extensions.md internal/admin/lifecycle.go,plugins/arena/plugin.yaml plugins/invites/plugin.yaml,plugins/subscriptions/plugin.yaml roadmap_admin.md,tests/app/admin_test.go
- `9d26b49` — 2026-11-27 03:05:08 +0330 — Generate admin model forms and finish Go formatting corrections
  Files: admin-nextjs/app/components/model-panel.tsx,admin-nextjs/app/plugins/[plugin]/page.tsx implementation/dynamic-runtime.md,internal/plugin/manifest_model.go internal/plugin/model_runtime.go,plugins/sdk-typescript/src/index.ts tests/plugin/model_runtime_test.go
- `5273321` — 2026-11-27 03:04:08 +0330 — Add versioned manifest-driven admin action pages and permission-aware registry
  Files: admin-nextjs/app/components/admin-chrome.tsx,admin-nextjs/app/extensions/[project]/[plugin]/[version]/[page]/page.tsx admin-nextjs/app/lib/admin-registry.ts,admin-nextjs/app/lib/admin-sdk.ts admin-nextjs/app/lib/admin-types.ts,docs/admin-pages/extensions.md internal/admin/lifecycle.go,internal/app/admin.go internal/plugin/admin_manifest.go,internal/plugin/admin_manifest_test.go internal/plugin/manifest.go,plugins/analytics/plugin.yaml plugins/arena/plugin.yaml,plugins/files/plugin.yaml plugins/functions/plugin.yaml,plugins/invites/plugin.yaml plugins/matchmaking/plugin.yaml,plugins/rooms/plugin.yaml plugins/subscriptions/plugin.yaml,plugins/tables/plugin.yaml tests/admin/admin-registry.test.ts
- `1b8c981` — 2026-11-27 03:03:50 +0330 — Bind active project resolution and normalize manifest formatting
  Files: cmd/brain/main.go,internal/plugin/manifest.go internal/plugin/manifest_model.go,internal/plugin/manifest_pipeline.go internal/plugin/manifest_pipeline_test.go,internal/plugin/plugin.go internal/plugin/registry_catalog.go,internal/plugin/registry_workers.go internal/plugin/typescript_worker.go,internal/plugin/worker.go
- `1565186` — 2026-11-27 03:02:48 +0330 — Implement host-managed declarative model CRUD and state transitions
  Files: implementation/dynamic-runtime.md,internal/plugin/manifest.go internal/plugin/manifest_model.go,internal/plugin/manifest_pipeline.go internal/plugin/model_runtime.go,internal/plugin/registry_workers.go internal/plugin/sqlite_storage.go,plugins/template/plugin.yaml scripts/typescript-worker.mjs,tests/plugin/model_runtime_test.go
- `ce839ab` — 2026-11-27 03:01:26 +0330 — Extend declarative helpers and add model and transition primitives
  Files: implementation/dynamic-runtime.md,internal/plugin/manifest_pipeline.go plugins/sdk-typescript/src/declarative.ts,plugins/sdk-typescript/src/index.ts scripts/typescript-worker.mjs,tests/plugin/declarative.test.ts
- `a05f58c` — 2026-11-27 03:00:34 +0330 — Add bounded declarative plugin pipelines and dynamic runtime roadmap
  Files: implementation/dynamic-runtime.md,internal/plugin/manifest.go internal/plugin/manifest_pipeline.go,internal/plugin/manifest_pipeline_test.go internal/plugin/typescript_worker.go,plugins/template/plugin.yaml scripts/typescript-worker.mjs
- `e857abd` — 2026-11-26 23:27:50 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `8252383` — 2026-11-27 02:57:40 +0330 — Merge pull request #5 from arashrasoulzadeh/plugin-project-context-observations
  Files: 
- `d4cff4a` — 2026-11-27 02:55:58 +0330 — Stage project-scoped observations and deliver through host after commit
  Files: internal/plugin/manifest.go,internal/plugin/observations.go internal/plugin/registry_workers.go,internal/plugin/typescript_worker.go internal/plugin/worker.go,plugins/analytics/index.ts plugins/analytics/plugin.yaml,plugins/rooms/plugin.yaml plugins/rooms/routes.ts,plugins/sdk-typescript/src/index.ts scripts/typescript-worker.mjs,tests/plugin/project-context.test.ts
- `fa2cf70` — 2026-11-27 02:54:22 +0330 — Introduce shared plugin project context and observation migration checklist
  Files: implementation/plugin-context-observations.md,internal/plugin/manifest.go internal/plugin/manifest_response.go,plugins/analytics/index.ts plugins/analytics/plugin.yaml,plugins/sdk-typescript/src/index.ts plugins/sdk-typescript/src/routes.ts,scripts/typescript-worker.mjs tests/plugin/project-context.test.ts
- `7f6ef27` — 2026-11-26 23:19:38 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `2b9a52e` — 2026-11-27 02:49:22 +0330 — fix(sandbox): apply bounded Wasm mode to JavaScript workers too
  Files: scripts/run-sandboxed-typescript-worker.sh
- `34e57b3` — 2026-11-26 23:18:18 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `4a550df` — 2026-11-27 02:48:06 +0330 — fix(sandbox): allow bounded WebAssembly allocation without raising memory cap
  Files: .github/workflows/test.yml,cmd/brain/main.go docs/deployment-runbook.md,scripts/run-sandboxed-typescript-worker.sh scripts/test-sandboxed-typescript-worker.sh
- `c3c3006` — 2026-11-26 23:13:00 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `43b7dc0` — 2026-11-27 02:42:47 +0330 — Merge pull request #3 from arashrasoulzadeh/fix/production-safety
  Files: 
- `0ef81fb` — 2026-11-27 02:40:37 +0330 — Merge main and preserve declarative manifest responses in hardened worker
  Files: internal/plugin/typescript_worker.go,scripts/typescript-worker.mjs
- `18f22d6` — 2026-11-27 02:39:48 +0330 — fix(plugin): persist restrictive transitions and stop bound workers
  Files: .github/workflows/test.yml,docs/deployment-runbook.md internal/plugin/registry_lifecycle.go,internal/plugin/registry_permissions.go internal/plugin/registry_workers.go,tests/plugin/lifecycle_runtime_persistence_test.go
- `3ec3726` — 2026-11-26 23:08:54 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `d78a0fb` — 2026-11-27 02:38:49 +0330 — fix(plugin): fail closed on lifecycle persistence errors and persist enable after startup
  Files: cmd/brain/main.go,internal/plugin/plugin.go internal/plugin/registry_catalog.go,internal/plugin/registry_lifecycle.go internal/plugin/registry_permissions.go,internal/plugin/registry_storage.go internal/plugin/registry_workers.go,tests/plugin/lifecycle_storage_failure_test.go
- `0f995c3` — 2026-11-27 02:38:43 +0330 — Merge pull request #4 from arashrasoulzadeh/refactor/compact-plugin-domains
  Files: 
- `8c5fcbd` — 2026-11-27 02:37:45 +0330 — fix(runtime): expose trusted sandbox probe diagnostics and correct smoke response contract
  Files: cmd/brain/main.go,internal/plugin/typescript_worker.go scripts/run-sandboxed-typescript-worker.sh,scripts/test-production-runtime.sh
- `028dfc5` — 2026-11-27 02:34:57 +0330 — Merge main and preserve concurrent dashboard SDK import fix and changelog
  Files: 
- `4b56685` — 2026-11-27 02:34:42 +0330 — test(plugins): cover declarative descriptor and shared resource contracts
  Files: docs/developers.md,internal/plugin/manifest_response.go scripts/typescript-worker.mjs,tests/plugin/manifest_response_test.go tests/plugin/resource-request.test.ts
