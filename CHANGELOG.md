# Changelog

Automatically generated from commits on main.

## Unreleased
- `029e8746` — 2026-11-27 13:07:42 +0330 — Merge pull request #18 from arashrasoulzadeh/test/admin-browser-smoke
  Files: 
- `7f2ecef9` — 2026-11-27 13:05:09 +0330 — Resolve browser test paths at step execution time
  Files: .github/workflows/admin-browser.yml
- `fbbb0e69` — 2026-11-27 13:04:32 +0330 — Add production-build Chromium admin smoke tests and screenshots
  Files: .github/workflows/admin-browser.yml,docs/admin-pages/theme.md scripts/test-admin-browser.mjs
- `6fceee46` — 2026-11-27 09:33:06 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `53f3a75b` — 2026-11-27 13:02:53 +0330 — Merge pull request #16 from arashrasoulzadeh/admin/compact-control-plane
  Files: 
- `d4f6b960` — 2026-11-27 12:51:23 +0330 — Unify catalog, inventory and generated form layouts around compact panels
  Files: admin-nextjs/app/components/admin-config-form.tsx,admin-nextjs/app/components/admin-inventory.tsx admin-nextjs/app/components/admin-renderer.tsx,admin-nextjs/app/components/admin-results.tsx admin-nextjs/app/components/model-panel.tsx,admin-nextjs/app/components/plugin-catalog-panel.tsx admin-nextjs/app/theme.css,docs/admin-pages/theme.md
- `c97e77a4` — 2026-11-27 12:50:16 +0330 — Test operator-only admin inspection, redaction and resource isolation
  Files: tests/admin/admin-plugin-operations.test.ts
- `963471ac` — 2026-11-27 12:49:13 +0330 — Add safe manifest admin pages for arena, invites, subscriptions and authentication
  Files: docs/admin-pages/README.md,docs/admin-pages/extensions.md plugins/arena/plugin.yaml,plugins/auth/plugin.yaml plugins/auth/service.ts,plugins/invites/index.ts plugins/invites/plugin.yaml,plugins/sdk-typescript/src/index.ts plugins/sdk-typescript/src/validation.ts,plugins/subscriptions/index.ts plugins/subscriptions/plugin.yaml,plugins/subscriptions/routes.ts roadmap_admin.md,tests/invites/invite-routes.test.ts
- `dc528e8b` — 2026-11-27 12:47:31 +0330 — Filter admin inventories using exact operator resource access rules
  Files: docs/admin-pages/extensions.md,internal/admin/lifecycle.go internal/plugin/access_management.go,internal/plugin/admin_inventory.go internal/plugin/legacy_admin_inventory.go,roadmap_admin.md tests/plugin/access_management_test.go
- `7e9b28e4` — 2026-11-27 12:46:26 +0330 — Polish responsive menu behavior and accessible content navigation
  Files: admin-nextjs/app/components/admin-chrome.tsx,admin-nextjs/app/components/admin-shell.tsx admin-nextjs/app/lib/admin-menu.ts,admin-nextjs/app/page.tsx admin-nextjs/app/theme.css,roadmap_admin.md
- `b4b2f4e8` — 2026-11-27 12:44:35 +0330 — Refresh admin with compact professional theme and dynamic navigation
  Files: admin-nextjs/app/components/admin-chrome.tsx,admin-nextjs/app/components/admin-shell.tsx admin-nextjs/app/components/projects-panel.tsx,admin-nextjs/app/layout.tsx admin-nextjs/app/lib/admin-api.ts,admin-nextjs/app/lib/admin-menu.ts admin-nextjs/app/styles.css,admin-nextjs/app/theme.css docs/admin-pages/theme.md,tests/admin/admin-menu.test.ts tests/admin/admin-shell.test.ts
- `6140113d` — 2026-11-27 00:28:59 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `174113c7` — 2026-11-27 03:57:22 +0330 — Merge pull request #12 from arashrasoulzadeh/feature/plugin-auth-context
  Files: 
- `26ef6069` — 2026-11-27 03:56:57 +0330 — Merge current main and preserve coverage and actor assertions
  Files: tests/plugin/access_management_test.go,tests/plugin/observation_queue_test.go
- `e5b4bfc9` — 2026-11-27 00:26:00 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `dfd53636` — 2026-11-27 03:55:46 +0330 — Merge pull request #13 from arashrasoulzadeh/docs/coverage-and-pages
  Files: 
- `668700d3` — 2026-11-27 03:54:43 +0330 — Fix published function authorization test fixture
  Files: tests/plugin/user_session_test.go
- `a933e973` — 2026-11-27 03:53:53 +0330 — Verify published functions reject forged and unauthorized user sessions
  Files: tests/plugin/user_session_test.go
- `17f072e4` — 2026-11-27 03:53:37 +0330 — Authenticate project sessions on published function invocations
  Files: docs/plugin-auth-context.md,internal/admin/functions.go internal/app/admin_functions.go,internal/plugin/access_management.go internal/plugin/user_session.go
- `982787e5` — 2026-11-27 03:53:17 +0330 — test: cover plugin observation, worker and admin client boundaries
  Files: .github/workflows/coverage.yml,tests/coverage/plugin-boundaries.test.ts
- `c62e17c7` — 2026-11-27 03:52:36 +0330 — Budget shared UID threads for the complete packaged plugin inventory
  Files: cmd/brain/main.go,docs/plugin-auth-context.md scripts/test-production-runtime.sh
- `8aadd2b9` — 2026-11-27 03:51:20 +0330 — test: align policy and queue regressions with host contracts
  Files: .github/workflows/coverage.yml,tests/android/BrainClientTest.kt tests/plugin/access_management_test.go,tests/plugin/observation_queue_test.go
- `aafec670` — 2026-11-27 03:51:15 +0330 — Diagnose packaged arena startup failure with synthetic CI probe
  Files: scripts/test-production-runtime.sh
- `2ba1c2c8` — 2026-11-27 03:50:13 +0330 — test: cover Flutter authentication and cookie rejection paths
  Files: tests/flutter/auth_client_test.dart,tests/flutter/web_cookie_test.dart
- `e15491d8` — 2026-11-27 03:49:44 +0330 — fix: validate admin inventory project and normalize V8 coverage inputs
  Files: .github/workflows/coverage.yml,internal/plugin/admin_inventory.go scripts/prepare-v8-coverage.py,tests/documentation/test_prepare_v8_coverage.py tests/plugin/admin_inventory_project_test.go
- `95cbdeb3` — 2026-11-27 03:48:52 +0330 — Isolate durable outbox fixture and accommodate cold sandbox startup
  Files: internal/admin/functions.go,tests/plugin/observation_queue_test.go
- `13780f2c` — 2026-11-27 03:48:48 +0330 — test: add mobile client coverage gates and Kotlin contract tests
  Files: .github/workflows/coverage.yml,.gitignore clients/android/build.gradle.kts,clients/android/settings.gradle.kts docs/quality.md,scripts/check-lcov.py tests/android/BrainClientTest.kt,tests/documentation/test_check_lcov.py
- `318962ba` — 2026-11-27 03:48:21 +0330 — Test verified user session identity, project isolation and revocation
  Files: tests/auth/sdk-dispatch.test.ts,tests/plugin/user_session_test.go
- `43160fbb` — 2026-11-27 03:48:00 +0330 — Resolve user context from authenticated project sessions and reconcile inventory resolver
  Files: docs/plugin-auth-context.md,internal/plugin/admin_inventory.go internal/plugin/observation_queue.go,internal/plugin/observations.go internal/plugin/plugin_call.go,internal/plugin/registry_workers.go internal/plugin/user_session.go,internal/plugin/worker.go plugins/auth/accounts.ts,plugins/auth/plugin.yaml plugins/auth/service.ts,plugins/auth/sessions.ts
- `2549b7f4` — 2026-11-27 03:47:55 +0330 — docs: build searchable Pages reference and enforce measured coverage
  Files: .github/workflows/coverage.yml,.github/workflows/pages.yml .gitignore,docs/assets/brain.css docs/index.md,docs/quality.md docs/requirements.txt,mkdocs.yml scripts/build-docs.py,tests/README.md tests/documentation/test_build_docs.py
- `93e8e47f` — 2026-11-27 03:46:07 +0330 — Pass verified actor to function VM and host observations
  Files: docs/plugin-auth-context.md,internal/admin/functions.go internal/admin/lifecycle.go,tests/admin/functions_test.go
- `8990792a` — 2026-11-27 03:45:40 +0330 — Reconcile authentication context with durable observation runtime
  Files: internal/plugin/broker.go,internal/plugin/observation_queue.go internal/plugin/observations.go,internal/plugin/worker.go plugins/sdk-typescript/src/index.ts,scripts/typescript-worker.mjs tests/plugin/access_management_test.go,tests/plugin/observation_queue_test.go tests/plugin/project-context.test.ts,tests/plugin/worker_test.go
- `f65be743` — 2026-11-27 00:14:03 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `42af18e9` — 2026-11-27 03:43:49 +0330 — Merge pull request #6 from arashrasoulzadeh/plugin-project-context-observations
  Files: 
- `4f00c7ba` — 2026-11-27 03:43:35 +0330 — Reconcile host-owned storage snapshots, admin inventory and resource-scoped execution
  Files: internal/admin/lifecycle.go,internal/plugin/legacy_storage_adapter.go internal/plugin/manifest.go,internal/plugin/registry_workers.go plugins/sdk-typescript/src/index.ts,scripts/test-production-runtime.sh tests/plugin/project-context.test.ts
- `ba191356` — 2026-11-27 03:43:26 +0330 — Expose host-verified actor and scoped permissions in plugin context
  Files: docs/plugin-auth-context.md,internal/plugin/broker.go internal/plugin/invocation_context.go,internal/plugin/manifest_pipeline.go internal/plugin/manifest_response.go,internal/plugin/observations.go internal/plugin/plugin_call.go,internal/plugin/worker.go plugins/sdk-typescript/src/index.ts,scripts/typescript-worker.mjs tests/plugin/project-context.test.ts,tests/plugin/worker_test.go
- `5a70d688` — 2026-11-27 03:43:04 +0330 — Stage concurrent admin inventory and resource-scoped authorization additions
  Files: CHANGELOG.md,admin-nextjs/app/components/admin-inventory.tsx admin-nextjs/app/components/admin-renderer.tsx,admin-nextjs/app/components/legacy-admin-page.tsx admin-nextjs/app/components/model-panel.tsx,admin-nextjs/app/lib/admin-registry.ts admin-nextjs/app/lib/admin-sdk.ts,admin-nextjs/app/lib/admin-types.ts admin-nextjs/app/plugins/[plugin]/page.tsx,docs/admin-pages/extensions.md docs/declarative-runtime.md,internal/app/admin.go internal/plugin/admin_inventory.go,internal/plugin/admin_inventory_test.go internal/plugin/admin_manifest.go,internal/plugin/admin_model_pages.go internal/plugin/admin_model_pages_test.go,internal/plugin/legacy_admin_inventory.go internal/plugin/legacy_storage_adapter_test.go,internal/plugin/route_resources_test.go plugins/files/plugin.yaml,plugins/matchmaking/plugin.yaml plugins/rooms/plugin.yaml,plugins/tables/plugin.yaml plugins/tables/routes.ts,roadmap_admin.md tests/admin/admin-registry.test.ts,tests/plugin/resource_scope_invoke_test.go
- `4fd3711c` — 2026-11-27 03:42:23 +0330 — Unify project snapshot, declarative runtime, credential and access-policy authorization paths
  Files: internal/plugin/registry_workers.go,internal/plugin/worker.go
- `453fb721` — 2026-11-27 03:42:00 +0330 — Reconcile declarative worker, function VM helpers, resource-aware queue and manifest contracts
  Files: internal/admin/functions.go,internal/plugin/legacy_storage_adapter.go internal/plugin/manifest.go,internal/plugin/observation_queue.go internal/plugin/typescript_worker.go
- `3c3edd6a` — 2026-11-27 03:41:28 +0330 — Preserve operator admission, project resolver, storage models and post-commit observation routing
  Files: cmd/brain/main.go,internal/admin/lifecycle.go internal/plugin/broker.go,internal/plugin/observations.go internal/plugin/plugin.go,internal/plugin/registry_catalog.go internal/plugin/sqlite_storage.go
- `c4406cdf` — 2026-11-27 00:11:12 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `25f8f3ef` — 2026-11-27 03:40:58 +0330 — Reconcile project context and observation manifests with declarative and resource-scoped routes
  Files: plugins/analytics/plugin.yaml,plugins/arena/plugin.yaml plugins/functions/plugin.yaml,plugins/invites/plugin.yaml plugins/sdk-typescript/src/index.ts,plugins/subscriptions/plugin.yaml plugins/template/plugin.yaml,scripts/typescript-worker.mjs tests/plugin/project-context.test.ts
- `2a909a3a` — 2026-11-27 03:40:47 +0330 — Merge pull request #9 from arashrasoulzadeh/admin/resource-inventory
  Files: 
- `565b8260` — 2026-11-27 00:10:43 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `58f49a91` — 2026-11-27 03:40:35 +0330 — Stage concurrent main changes and non-overlapping observation migration files for reconciliation
  Files: CHANGELOG.md,admin-nextjs/app/components/model-panel.tsx admin-nextjs/app/plugins/[plugin]/page.tsx,cmd/brain/main.go docs/declarative-runtime.md,examples/declarative-plugin/plugin.yaml implementation/dynamic-runtime.md,internal/admin/functions.go internal/admin/lifecycle.go,internal/plugin/access_management.go internal/plugin/broker.go,internal/plugin/contract_types.go internal/plugin/contract_types_test.go,internal/plugin/invocation_context.go internal/plugin/legacy_storage_adapter.go,internal/plugin/manifest.go internal/plugin/manifest_declarative_test.go,internal/plugin/manifest_model.go internal/plugin/manifest_pipeline.go,internal/plugin/manifest_pipeline_test.go internal/plugin/model_runtime.go,internal/plugin/observations.go internal/plugin/plugin.go,internal/plugin/plugin_call.go internal/plugin/registry_catalog.go,internal/plugin/registry_permissions.go internal/plugin/registry_workers.go,internal/plugin/sqlite_storage.go internal/plugin/typescript_worker.go,internal/plugin/worker.go plugins/README.md,plugins/access-management/README.md plugins/access-management/index.ts,plugins/access-management/plugin.yaml plugins/analytics/plugin.yaml,plugins/arena/plugin.yaml plugins/auth/plugin.yaml,plugins/auth/service.ts plugins/files/plugin.yaml,plugins/files/routes.ts plugins/functions/plugin.yaml,plugins/invites/plugin.yaml plugins/matchmaking/plugin.yaml,plugins/matchmaking/routes.ts plugins/rooms/plugin.yaml,plugins/rooms/routes.ts plugins/sdk-typescript/helpers.go,plugins/sdk-typescript/src/declarative.ts plugins/sdk-typescript/src/helpers.d.mts,plugins/sdk-typescript/src/helpers.mjs plugins/sdk-typescript/src/index.ts,plugins/sdk-typescript/src/request.ts plugins/subscriptions/plugin.yaml,plugins/tables/plugin.yaml plugins/tables/routes.ts,plugins/template/plugin.yaml scripts/run-sandboxed-typescript-worker.sh,scripts/typescript-worker.mjs tests/admin/function_helpers_test.go,tests/files/files.test.ts tests/matchmaking/matchmaking.test.ts,tests/plugin/access_management_test.go tests/plugin/declarative.test.ts,tests/plugin/model_runtime_test.go tests/plugin/pipeline-worker.test.ts,tests/plugin/plugin_call_test.go tests/plugin/project-context.test.ts,tests/plugin/resource-request.test.ts tests/rooms/rooms.test.ts
- `96089fb9` — 2026-11-27 03:40:24 +0330 — Merge pull request #10 from arashrasoulzadeh/dynamic-runtime-storage-hardening
  Files: 
- `4e35a0c7` — 2026-11-27 03:40:17 +0330 — Merge main and preserve central operator access policy in admin inventory
  Files: docs/admin-pages/extensions.md,internal/admin/lifecycle.go internal/plugin/admin_inventory.go,internal/plugin/admin_inventory_test.go plugins/files/plugin.yaml,plugins/matchmaking/plugin.yaml plugins/rooms/plugin.yaml,plugins/tables/plugin.yaml roadmap_admin.md
- `7bcad09c` — 2026-11-27 03:40:14 +0330 — Reconcile storage hardening with access-management authorization
  Files: internal/plugin/manifest.go,internal/plugin/registry_workers.go internal/plugin/route_resources_test.go,plugins/sdk-typescript/src/index.ts
- `33423138` — 2026-11-27 00:09:34 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `904d0fef` — 2026-11-27 03:39:24 +0330 — Centralize project access management across plugins
  Files: internal/admin/functions.go,internal/admin/lifecycle.go internal/plugin/access_management.go,internal/plugin/broker.go internal/plugin/manifest.go,internal/plugin/model_runtime.go internal/plugin/observations.go,internal/plugin/plugin_call.go internal/plugin/registry_workers.go,internal/plugin/worker.go plugins/README.md,plugins/access-management/README.md plugins/access-management/index.ts,plugins/access-management/plugin.yaml plugins/analytics/plugin.yaml,plugins/arena/plugin.yaml plugins/files/plugin.yaml,plugins/files/routes.ts plugins/functions/plugin.yaml,plugins/invites/plugin.yaml plugins/matchmaking/plugin.yaml,plugins/matchmaking/routes.ts plugins/rooms/plugin.yaml,plugins/rooms/routes.ts plugins/sdk-typescript/src/index.ts,plugins/sdk-typescript/src/request.ts plugins/subscriptions/plugin.yaml,plugins/tables/plugin.yaml tests/files/files.test.ts,tests/matchmaking/matchmaking.test.ts tests/plugin/access_management_test.go,tests/plugin/resource-request.test.ts tests/rooms/rooms.test.ts
