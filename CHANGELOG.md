# Changelog

Automatically generated from commits on main.

## Unreleased
- `0d513e57` — 2026-11-27 13:46:35 +0330 — Merge pull request #19 from arashrasoulzadeh/test/admin-live-backend
  Files: 
- `ec2d74da` — 2026-11-27 13:46:17 +0330 — Merge current main and align live Chromium runner with packaged runtime CI
  Files: .github/workflows/admin-browser.yml
- `c263bc03` — 2026-11-27 10:15:40 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `c350a896` — 2026-11-27 13:45:21 +0330 — Merge pull request #21 from arashrasoulzadeh/fix/admin-project-scope
  Files: 
- `f149ff75` — 2026-11-27 13:45:15 +0330 — Merge pull request #20 from arashrasoulzadeh/fix/admin-revocation-refresh
  Files: 
- `a42e093b` — 2026-11-27 13:15:05 +0330 — Prevent stale project catalog and function responses after switching projects
  Files: admin-nextjs/app/components/projects-panel.tsx,admin-nextjs/app/lib/project-requests.ts admin-nextjs/app/page.tsx,docs/admin-pages/projects.md tests/admin/project-requests.test.ts
- `be4d97d5` — 2026-11-27 13:13:39 +0330 — Keep live browser backend network-disabled and relay CI HTTP over docker exec
  Files: .github/workflows/admin-browser.yml,docs/admin-pages/theme.md scripts/admin-ci-backend-relay.mjs
- `02c6c423` — 2026-11-27 13:12:24 +0330 — Cover delegated admin read checks and revoked target grants
  Files: tests/plugin/plugin_call_test.go
- `999e95b0` — 2026-11-27 09:41:41 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `9824af94` — 2026-11-27 13:11:02 +0330 — Merge pull request #17 from arashrasoulzadeh/fix/compose-node-runtime
  Files: 
- `aa7506d1` — 2026-11-27 13:10:38 +0330 — Revalidate displayed admin results and inventories after silent policy revocation
  Files: admin-nextjs/app/components/admin-inventory.tsx,admin-nextjs/app/components/admin-renderer.tsx admin-nextjs/app/components/model-panel.tsx,admin-nextjs/app/lib/admin-sdk.ts docs/admin-pages/extensions.md,internal/admin/lifecycle.go internal/app/admin.go,internal/plugin/access_management.go roadmap_admin.md,scripts/test-admin-browser.mjs tests/admin/manifest_admin_test.go,tests/plugin/access_management_test.go
- `96a2a32b` — 2026-11-27 13:08:54 +0330 — Exercise real packaged admin backend with Chromium policy and persistence tests
  Files: .github/workflows/admin-browser.yml,docs/admin-pages/theme.md scripts/test-admin-live.mjs
- `45b52cd2` — 2026-11-27 09:38:17 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `029e8746` — 2026-11-27 13:07:42 +0330 — Merge pull request #18 from arashrasoulzadeh/test/admin-browser-smoke
  Files: 
- `7f2ecef9` — 2026-11-27 13:05:09 +0330 — Resolve browser test paths at step execution time
  Files: .github/workflows/admin-browser.yml
- `fbbb0e69` — 2026-11-27 13:04:32 +0330 — Add production-build Chromium admin smoke tests and screenshots
  Files: .github/workflows/admin-browser.yml,docs/admin-pages/theme.md scripts/test-admin-browser.mjs
- `b52ce1a5` — 2026-11-27 13:03:34 +0330 — Install production-matched Node runtime in Go integration test image
  Files: .github/workflows/test.yml,Dockerfile.test
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
