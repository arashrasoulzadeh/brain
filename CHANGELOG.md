# Changelog

Automatically generated from commits on main.

## Unreleased
- `42af18e` — 2026-11-27 03:43:49 +0330 — Merge pull request #6 from arashrasoulzadeh/plugin-project-context-observations
  Files: 
- `4f00c7b` — 2026-11-27 03:43:35 +0330 — Reconcile host-owned storage snapshots, admin inventory and resource-scoped execution
  Files: internal/admin/lifecycle.go,internal/plugin/legacy_storage_adapter.go internal/plugin/manifest.go,internal/plugin/registry_workers.go plugins/sdk-typescript/src/index.ts,scripts/test-production-runtime.sh tests/plugin/project-context.test.ts
- `5a70d68` — 2026-11-27 03:43:04 +0330 — Stage concurrent admin inventory and resource-scoped authorization additions
  Files: CHANGELOG.md,admin-nextjs/app/components/admin-inventory.tsx admin-nextjs/app/components/admin-renderer.tsx,admin-nextjs/app/components/legacy-admin-page.tsx admin-nextjs/app/components/model-panel.tsx,admin-nextjs/app/lib/admin-registry.ts admin-nextjs/app/lib/admin-sdk.ts,admin-nextjs/app/lib/admin-types.ts admin-nextjs/app/plugins/[plugin]/page.tsx,docs/admin-pages/extensions.md docs/declarative-runtime.md,internal/app/admin.go internal/plugin/admin_inventory.go,internal/plugin/admin_inventory_test.go internal/plugin/admin_manifest.go,internal/plugin/admin_model_pages.go internal/plugin/admin_model_pages_test.go,internal/plugin/legacy_admin_inventory.go internal/plugin/legacy_storage_adapter_test.go,internal/plugin/route_resources_test.go plugins/files/plugin.yaml,plugins/matchmaking/plugin.yaml plugins/rooms/plugin.yaml,plugins/tables/plugin.yaml plugins/tables/routes.ts,roadmap_admin.md tests/admin/admin-registry.test.ts,tests/plugin/resource_scope_invoke_test.go
- `4fd3711` — 2026-11-27 03:42:23 +0330 — Unify project snapshot, declarative runtime, credential and access-policy authorization paths
  Files: internal/plugin/registry_workers.go,internal/plugin/worker.go
- `453fb72` — 2026-11-27 03:42:00 +0330 — Reconcile declarative worker, function VM helpers, resource-aware queue and manifest contracts
  Files: internal/admin/functions.go,internal/plugin/legacy_storage_adapter.go internal/plugin/manifest.go,internal/plugin/observation_queue.go internal/plugin/typescript_worker.go
- `3c3edd6` — 2026-11-27 03:41:28 +0330 — Preserve operator admission, project resolver, storage models and post-commit observation routing
  Files: cmd/brain/main.go,internal/admin/lifecycle.go internal/plugin/broker.go,internal/plugin/observations.go internal/plugin/plugin.go,internal/plugin/registry_catalog.go internal/plugin/sqlite_storage.go
- `c4406cd` — 2026-11-27 00:11:12 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `25f8f3e` — 2026-11-27 03:40:58 +0330 — Reconcile project context and observation manifests with declarative and resource-scoped routes
  Files: plugins/analytics/plugin.yaml,plugins/arena/plugin.yaml plugins/functions/plugin.yaml,plugins/invites/plugin.yaml plugins/sdk-typescript/src/index.ts,plugins/subscriptions/plugin.yaml plugins/template/plugin.yaml,scripts/typescript-worker.mjs tests/plugin/project-context.test.ts
- `2a909a3` — 2026-11-27 03:40:47 +0330 — Merge pull request #9 from arashrasoulzadeh/admin/resource-inventory
  Files: 
- `565b826` — 2026-11-27 00:10:43 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `58f49a9` — 2026-11-27 03:40:35 +0330 — Stage concurrent main changes and non-overlapping observation migration files for reconciliation
  Files: CHANGELOG.md,admin-nextjs/app/components/model-panel.tsx admin-nextjs/app/plugins/[plugin]/page.tsx,cmd/brain/main.go docs/declarative-runtime.md,examples/declarative-plugin/plugin.yaml implementation/dynamic-runtime.md,internal/admin/functions.go internal/admin/lifecycle.go,internal/plugin/access_management.go internal/plugin/broker.go,internal/plugin/contract_types.go internal/plugin/contract_types_test.go,internal/plugin/invocation_context.go internal/plugin/legacy_storage_adapter.go,internal/plugin/manifest.go internal/plugin/manifest_declarative_test.go,internal/plugin/manifest_model.go internal/plugin/manifest_pipeline.go,internal/plugin/manifest_pipeline_test.go internal/plugin/model_runtime.go,internal/plugin/observations.go internal/plugin/plugin.go,internal/plugin/plugin_call.go internal/plugin/registry_catalog.go,internal/plugin/registry_permissions.go internal/plugin/registry_workers.go,internal/plugin/sqlite_storage.go internal/plugin/typescript_worker.go,internal/plugin/worker.go plugins/README.md,plugins/access-management/README.md plugins/access-management/index.ts,plugins/access-management/plugin.yaml plugins/analytics/plugin.yaml,plugins/arena/plugin.yaml plugins/auth/plugin.yaml,plugins/auth/service.ts plugins/files/plugin.yaml,plugins/files/routes.ts plugins/functions/plugin.yaml,plugins/invites/plugin.yaml plugins/matchmaking/plugin.yaml,plugins/matchmaking/routes.ts plugins/rooms/plugin.yaml,plugins/rooms/routes.ts plugins/sdk-typescript/helpers.go,plugins/sdk-typescript/src/declarative.ts plugins/sdk-typescript/src/helpers.d.mts,plugins/sdk-typescript/src/helpers.mjs plugins/sdk-typescript/src/index.ts,plugins/sdk-typescript/src/request.ts plugins/subscriptions/plugin.yaml,plugins/tables/plugin.yaml plugins/tables/routes.ts,plugins/template/plugin.yaml scripts/run-sandboxed-typescript-worker.sh,scripts/typescript-worker.mjs tests/admin/function_helpers_test.go,tests/files/files.test.ts tests/matchmaking/matchmaking.test.ts,tests/plugin/access_management_test.go tests/plugin/declarative.test.ts,tests/plugin/model_runtime_test.go tests/plugin/pipeline-worker.test.ts,tests/plugin/plugin_call_test.go tests/plugin/project-context.test.ts,tests/plugin/resource-request.test.ts tests/rooms/rooms.test.ts
- `96089fb` — 2026-11-27 03:40:24 +0330 — Merge pull request #10 from arashrasoulzadeh/dynamic-runtime-storage-hardening
  Files: 
- `4e35a0c` — 2026-11-27 03:40:17 +0330 — Merge main and preserve central operator access policy in admin inventory
  Files: docs/admin-pages/extensions.md,internal/admin/lifecycle.go internal/plugin/admin_inventory.go,internal/plugin/admin_inventory_test.go plugins/files/plugin.yaml,plugins/matchmaking/plugin.yaml plugins/rooms/plugin.yaml,plugins/tables/plugin.yaml roadmap_admin.md
- `7bcad09` — 2026-11-27 03:40:14 +0330 — Reconcile storage hardening with access-management authorization
  Files: internal/plugin/manifest.go,internal/plugin/registry_workers.go internal/plugin/route_resources_test.go,plugins/sdk-typescript/src/index.ts
- `3342313` — 2026-11-27 00:09:34 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `904d0fe` — 2026-11-27 03:39:24 +0330 — Centralize project access management across plugins
  Files: internal/admin/functions.go,internal/admin/lifecycle.go internal/plugin/access_management.go,internal/plugin/broker.go internal/plugin/manifest.go,internal/plugin/model_runtime.go internal/plugin/observations.go,internal/plugin/plugin_call.go internal/plugin/registry_workers.go,internal/plugin/worker.go plugins/README.md,plugins/access-management/README.md plugins/access-management/index.ts,plugins/access-management/plugin.yaml plugins/analytics/plugin.yaml,plugins/arena/plugin.yaml plugins/files/plugin.yaml,plugins/files/routes.ts plugins/functions/plugin.yaml,plugins/invites/plugin.yaml plugins/matchmaking/plugin.yaml,plugins/matchmaking/routes.ts plugins/rooms/plugin.yaml,plugins/rooms/routes.ts plugins/sdk-typescript/src/index.ts,plugins/sdk-typescript/src/request.ts plugins/subscriptions/plugin.yaml,plugins/tables/plugin.yaml tests/files/files.test.ts,tests/matchmaking/matchmaking.test.ts tests/plugin/access_management_test.go,tests/plugin/resource-request.test.ts tests/rooms/rooms.test.ts
- `d6c635b` — 2026-11-26 23:54:59 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `904d101` — 2026-11-27 03:24:32 +0330 — Merge main formatting fix into storage hardening
  Files: 
- `94cca6f` — 2026-11-27 03:24:21 +0330 — Fix gofmt alignment in declarative manifest test
  Files: internal/plugin/manifest_declarative_test.go
- `552c038` — 2026-11-27 03:23:50 +0330 — Unify declared model CRUD and state transitions with dynamic admin pages
  Files: admin-nextjs/app/components/admin-renderer.tsx,admin-nextjs/app/components/model-panel.tsx admin-nextjs/app/lib/admin-types.ts,admin-nextjs/app/plugins/[plugin]/page.tsx docs/admin-pages/extensions.md,roadmap_admin.md
- `0368f2b` — 2026-11-27 03:23:45 +0330 — Read table schema from durable storage and test direct route context
  Files: docs/declarative-runtime.md,internal/plugin/legacy_storage_adapter.go internal/plugin/legacy_storage_adapter_test.go,plugins/tables/routes.ts tests/plugin/project-context.test.ts
- `18d7a51` — 2026-11-27 03:23:27 +0330 — Generate operator admin pages automatically from declared model routes
  Files: internal/admin/lifecycle.go,internal/plugin/admin_manifest.go internal/plugin/admin_model_pages.go,internal/plugin/admin_model_pages_test.go
- `ecdeb92` — 2026-11-27 03:23:09 +0330 — Enforce manifest resource scope and forward context to direct exports
  Files: internal/plugin/manifest.go,internal/plugin/registry_workers.go internal/plugin/route_resources.go,internal/plugin/route_resources_test.go plugins/files/plugin.yaml,plugins/matchmaking/plugin.yaml plugins/rooms/plugin.yaml,plugins/sdk-typescript/src/index.ts plugins/tables/plugin.yaml,scripts/typescript-worker.mjs tests/plugin/resource_scope_invoke_test.go
- `43467eb` — 2026-11-26 23:52:36 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `98db7bb` — 2026-11-27 03:22:35 +0330 — Merge current main into storage hardening branch
  Files: 
- `7fbe5a8` — 2026-11-27 03:22:27 +0330 — Merge current main and connect inventories to new domain events
  Files: internal/plugin/admin_inventory.go,plugins/files/plugin.yaml plugins/matchmaking/plugin.yaml,plugins/rooms/plugin.yaml plugins/tables/plugin.yaml
- `61fae9b` — 2026-11-27 03:21:53 +0330 — Merge current main and preserve declarative admin pages with plugin observations
  Files: plugins/analytics/plugin.yaml,plugins/files/plugin.yaml plugins/functions/plugin.yaml,plugins/matchmaking/plugin.yaml plugins/rooms/plugin.yaml,plugins/tables/plugin.yaml
- `c993ddd` — 2026-11-27 03:21:46 +0330 — Merge pull request #7 from arashrasoulzadeh/dynamic-runtime-foundation
  Files: 
- `a32f4af` — 2026-11-27 03:21:42 +0330 — Reconcile admin observation core with durable delivery and normalize storage formatting
  Files: internal/admin/lifecycle.go,internal/plugin/legacy_storage_adapter.go internal/plugin/manifest.go,internal/plugin/observations.go internal/plugin/plugin.go,internal/plugin/registry_workers.go scripts/test-production-runtime.sh
- `8919b29` — 2026-11-27 03:21:32 +0330 — Render resource inventories and independently authorized plugin extension tabs
  Files: admin-nextjs/app/components/admin-inventory.tsx,admin-nextjs/app/components/admin-renderer.tsx admin-nextjs/app/components/legacy-admin-page.tsx,admin-nextjs/app/lib/admin-registry.ts admin-nextjs/app/lib/admin-sdk.ts,admin-nextjs/app/lib/admin-types.ts docs/admin-pages/extensions.md,internal/plugin/admin_manifest.go plugins/rooms/plugin.yaml,roadmap_admin.md tests/admin/admin-registry.test.ts
- `60022c4` — 2026-11-27 03:21:28 +0330 — Merge main and integrate declarative runtime with manifest-driven admin
  Files: admin-nextjs/app/plugins/[plugin]/page.tsx,internal/plugin/manifest.go internal/plugin/plugin.go,plugins/files/plugin.yaml plugins/matchmaking/plugin.yaml,plugins/rooms/plugin.yaml plugins/tables/plugin.yaml
- `cc2620c` — 2026-11-27 03:20:51 +0330 — Add scoped paginated operator inventories backed by trusted storage adapters
  Files: internal/admin/lifecycle.go,internal/app/admin.go internal/plugin/admin_inventory.go,internal/plugin/admin_inventory_test.go internal/plugin/admin_manifest.go,internal/plugin/legacy_admin_inventory.go plugins/files/plugin.yaml,plugins/matchmaking/plugin.yaml plugins/rooms/plugin.yaml,plugins/tables/plugin.yaml
- `8bd03b2` — 2026-11-26 23:50:47 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `783461f` — 2026-11-27 03:20:15 +0330 — Prevent forged storage snapshots in legacy plugin routes
  Files: internal/plugin/legacy_storage_adapter.go,internal/plugin/legacy_storage_adapter_test.go
- `7f9181d` — 2026-11-27 03:19:43 +0330 — Merge pull request #8 from arashrasoulzadeh/admin/manifest-driven-ui
  Files: 
- `52a7b86` — 2026-11-27 03:13:05 +0330 — Document project observation contracts and exercise packaged cross-plugin delivery
  Files: docs/admins.md,docs/developers.md docs/plugin-communication.md,implementation/plugin-context-observations.md internal/plugin/manifest.go,scripts/test-production-runtime.sh tests/plugin/all_manifests_test.go
- `ddea07e` — 2026-11-27 03:13:04 +0330 — Restore executable mode of sandbox launcher
  Files: scripts/run-sandboxed-typescript-worker.sh
- `4ae66bc` — 2026-11-27 03:12:31 +0330 — Generate project configuration UI and document declarative runtime
  Files: admin-nextjs/app/plugins/[plugin]/page.tsx,docs/declarative-runtime.md
- `d57b5fd` — 2026-11-27 03:12:21 +0330 — Add declarative result component registry and update admin migration roadmap
  Files: admin-nextjs/app/components/admin-renderer.tsx,admin-nextjs/app/components/admin-results.tsx admin-nextjs/app/lib/admin-result-data.ts,admin-nextjs/app/lib/admin-types.ts docs/admin-pages/extensions.md,internal/plugin/admin_manifest.go internal/plugin/admin_results.go,internal/plugin/admin_results_test.go plugins/files/plugin.yaml,plugins/rooms/plugin.yaml plugins/tables/plugin.yaml,roadmap_admin.md tests/admin/admin-result-data.test.ts
- `47ba08b` — 2026-11-27 03:11:57 +0330 — Bind workflow and scheduling helpers to project context; complete observer template
  Files: plugins/analytics/index.ts,plugins/analytics/plugin.yaml plugins/functions/index.ts,plugins/functions/scheduling.ts plugins/functions/workflow.ts,plugins/template/index.ts plugins/template/plugin.yaml,tests/plugin/project-bound-functions.test.ts
- `9fd583b` — 2026-11-27 03:11:47 +0330 — Add capability-checked declarative cross-plugin calls
  Files: internal/plugin/manifest.go,internal/plugin/plugin_call.go internal/plugin/registry_workers.go,scripts/typescript-worker.mjs tests/plugin/plugin_call_test.go
- `cf607ad` — 2026-11-27 03:11:35 +0330 — Reject forged host-owned state before observation-producing invocations
  Files: internal/plugin/legacy_storage_adapter.go,internal/plugin/registry_workers.go tests/plugin/host_state_observation_test.go
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
- `b25b326` — 2026-11-27 03:09:40 +0330 — Unify credential-bound worker execution with registry commit and observations
  Files: implementation/plugin-context-observations.md,internal/plugin/registry_workers.go internal/plugin/worker.go,tests/plugin/worker_test.go
- `c2b89b7` — 2026-11-27 03:08:57 +0330 — Commit SQLite plugin effects and observation outbox atomically
  Files: implementation/plugin-context-observations.md,internal/plugin/legacy_storage_adapter.go internal/plugin/observation_storage.go,internal/plugin/observations.go internal/plugin/registry_workers.go,internal/plugin/sqlite_storage.go tests/plugin/observation_queue_test.go
- `5436ac2` — 2026-11-27 03:08:39 +0330 — Share pure helper registry across YAML, SDK and user function VM
  Files: implementation/dynamic-runtime.md,internal/admin/functions.go internal/plugin/manifest_pipeline.go,plugins/sdk-typescript/helpers.go plugins/sdk-typescript/src/helpers.d.mts,plugins/sdk-typescript/src/helpers.mjs plugins/sdk-typescript/src/index.ts,scripts/run-sandboxed-typescript-worker.sh scripts/typescript-worker.mjs,tests/admin/function_helpers_test.go tests/plugin/pipeline-worker.test.ts,tests/plugin/project-context.test.ts
