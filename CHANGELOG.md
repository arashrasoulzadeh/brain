# Changelog

Automatically generated from commits on main.

## Unreleased
- `00ab96e7` — 2026-11-28 15:50:38 +0330 — Cover ten-Neuron rotation, operator pause and weighted load preference
  Files: tests/cortex/scheduler_test.go
- `996e5f05` — 2026-11-28 15:50:29 +0330 — Rotate equally loaded Neurons in Cortex placement
  Files: internal/cortex/scheduler.go
- `71019c57` — 2026-11-28 15:49:55 +0330 — Update Neuron refresh action regression tests
  Files: tests/admin-ui/neuron-console.test.mjs
- `23e03bf3` — 2026-11-28 15:49:46 +0330 — Surface Neuron performance report above inventory and rename refresh action
  Files: admin-nextjs/app/workers/view.tsx
- `59a9847e` — 2026-11-28 12:18:36 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `706e350f` — 2026-11-28 15:46:51 +0330 — Add Neuron performance reports and authenticated Prometheus metrics (#65)
  Files: admin-nextjs/app/lib/neuron-dashboard.ts,admin-nextjs/app/workers/view.tsx cmd/brain/main.go,docs/admin-pages/workers.md internal/app/api_response.go,internal/app/app.go internal/app/neuron_metrics.go,internal/app/neuron_status.go internal/cortex/performance.go,tests/admin-ui/neuron-console.test.mjs tests/app/neuron_metrics_test.go,tests/cortex/performance_test.go
- `c092877e` — 2026-11-28 12:09:42 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `0b7f3ec5` — 2026-11-28 15:39:31 +0330 — Fix managed Neuron drain calls to target the selected identity
  Files: cmd/brain/dev_neuron_management.go
- `fae344ae` — 2026-11-28 11:52:59 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `714c4281` — 2026-11-28 15:21:16 +0330 — Classify Core placement and reject host-only manifest operations before Neuron dispatch
  Files: docs/unified-execution-architecture.md,internal/plugin/execution_boundary.go internal/plugin/manifest.go,internal/plugin/scheduled_execution.go tests/plugin/execution_boundary_test.go
- `a4701f3d` — 2026-11-28 11:50:56 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `93bfcf4b` — 2026-11-28 15:20:08 +0330 — Persist operator Neuron placement pauses across Brain restarts
  Files: admin-nextjs/app/lib/neuron-dashboard.ts,admin-nextjs/app/workers/view.tsx cmd/brain/dev_neuron_management.go,cmd/brain/main.go docs/admins.md,docs/unified-execution-architecture.md internal/app/neuron_control.go,internal/cortex/drain_store.go internal/cortex/management.go,internal/cortex/scheduler.go internal/plugin/operator_drains.go,tests/cortex/drain_store_test.go
- `3ddacd34` — 2026-11-28 15:19:13 +0330 — Use trusted project resolvers in scheduled cancellation, credential and observation fixtures
  Files: tests/plugin/scheduled_cancellation_test.go,tests/plugin/scheduled_credential_test.go tests/plugin/scheduled_observation_test.go
- `5aedd02b` — 2026-11-28 15:17:53 +0330 — Merge pull request #64 from arashrasoulzadeh/architecture/durable-execution-history
  Files: 
- `748921e8` — 2026-11-28 15:17:46 +0330 — Integrate persistent history with nested execution and Neuron management
  Files: admin-nextjs/app/lib/neuron-dashboard.ts,admin-nextjs/app/workers/view.tsx cmd/brain/main.go,docs/unified-execution-architecture.md internal/cortex/scheduler.go,tests/admin-ui/neuron-console.test.mjs
- `57df4201` — 2026-11-28 11:47:29 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `15410b6f` — 2026-11-28 15:16:59 +0330 — Merge pull request #63 from arashrasoulzadeh/architecture/neuron-management
  Files: 
- `cad89382` — 2026-11-28 15:16:52 +0330 — Merge main into Neuron management and preserve nested execution
  Files: docs/unified-execution-architecture.md,internal/cortex/scheduler.go
- `470d264e` — 2026-11-28 11:46:37 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `368de351` — 2026-11-28 15:16:16 +0330 — Merge pull request #62 from arashrasoulzadeh/architecture/nested-draft-execution
  Files: 
- `13e24737` — 2026-11-28 11:44:17 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `00f9888e` — 2026-11-28 15:06:51 +0330 — Fix Go formatting of execution trace fields
  Files: internal/cortex/telemetry.go
- `e52a7401` — 2026-11-28 15:06:45 +0330 — Fix Go formatting of execution trace fields
  Files: internal/cortex/telemetry.go
- `152c4528` — 2026-11-28 15:06:38 +0330 — Fix Go formatting of execution trace fields
  Files: internal/cortex/telemetry.go
- `e48918b3` — 2026-11-28 15:06:31 +0330 — Fix Go formatting of execution trace fields
  Files: internal/cortex/telemetry.go
- `b6c972d0` — 2026-11-28 15:05:09 +0330 — Verify durable history diagnostics in the operator console
  Files: internal/cortex/scheduler.go,internal/cortex/telemetry.go tests/admin-ui/neuron-console.test.mjs
- `fa6237fb` — 2026-11-28 15:04:36 +0330 — Retain bounded redacted Cortex execution history across Brain restarts
  Files: admin-nextjs/app/lib/neuron-dashboard.ts,admin-nextjs/app/workers/view.tsx cmd/brain/main.go,docs/unified-execution-architecture.md internal/cortex/history.go,internal/cortex/scheduler.go internal/cortex/telemetry.go,internal/plugin/execution_history.go tests/cortex/history_test.go
- `f708a410` — 2026-11-28 15:03:10 +0330 — Test authenticated Neuron controls and managed process recovery
  Files: admin-nextjs/app/workers/view.tsx,tests/admin-ui/neuron-console.test.mjs tests/app/neuron_status_test.go,tests/cortex/management_test.go tests/development/test_dev_neurons.py
- `87772afb` — 2026-11-28 15:01:35 +0330 — Add operator Neuron drain/resume and managed local process controls
  Files: admin-nextjs/app/lib/neuron-dashboard.ts,admin-nextjs/app/workers/view.tsx cmd/brain/dev_neuron_management.go,cmd/brain/dev_neurons.go cmd/brain/main.go,docs/admins.md docs/http-api.md,docs/unified-execution-architecture.md internal/app/app.go,internal/app/neuron_control.go internal/app/neuron_status.go,internal/cortex/management.go internal/cortex/scheduler.go,tests/cortex/management_test.go
- `3bb5d66a` — 2026-11-28 14:58:14 +0330 — Fence nested call responses and add same-session transport and runtime regressions
  Files: docs/developers.md,docs/functions-runtime.md internal/cortex/scheduler.go,internal/synapse/grpc_client.go internal/synapse/grpc_host_client.go,internal/synapse/host_calls.go scripts/function-worker.mjs,scripts/test-runtime-api.mjs tests/admin/scheduled_function_test.go,tests/synapse/grpc_nested_test.go tests/synapse/host_calls_test.go
- `e1ef66a8` — 2026-11-28 11:27:27 +0000 — docs: update changelog [skip ci]
  Files: CHANGELOG.md
- `5a9f8440` — 2026-11-28 14:56:44 +0330 — Run operator drafts on Neurons with bounded read-only host bridge
  Files: docs/functions-runtime.md,docs/unified-execution-architecture.md internal/admin/function_execution.go,internal/admin/functions.go internal/cortex/scheduler.go,internal/neuron/function_runtime.go internal/plugin/typescript_worker.go,internal/plugin/worker_host_call.go internal/synapse/grpc_client.go,scripts/function-worker.mjs tests/admin/scheduled_function_test.go
- `21a41b2d` — 2026-11-28 14:55:32 +0330 — Enable lease-bound host calls and nonqueueing child execution admission
  Files: internal/admin/function_plugin_bridge.go,internal/cortex/scheduler.go internal/execution/host_call.go,internal/plugin/function_read.go internal/plugin/scheduled_execution.go,internal/synapse/grpc_client.go internal/synapse/grpc_host_client.go,internal/synapse/grpc_server.go internal/synapse/host_calls.go,internal/synapse/session.go tests/cortex/nested_test.go,tests/synapse/host_calls_test.go
- `40d72486` — 2026-11-28 14:45:50 +0330 — Ignore generated Next.js environment declarations during development watching
  Files: scripts/dev-stack.py,tests/development/test_dev_stack.py
- `033e839f` — 2026-11-28 14:45:24 +0330 — Watch source changes and restart the entire native development stack
  Files: Makefile,docs/developers.md scripts/dev-stack.py,tests/development/test_dev_stack.py
- `0fd05f02` — 2026-11-28 14:43:45 +0330 — Wait for Brain HTTP liveness before launching development admin
  Files: docs/developers.md,scripts/dev-stack.py tests/development/test_dev_stack.py
- `176a48d7` — 2026-11-28 14:41:38 +0330 — Distinguish admitted public function execution failures from concealed routes
  Files: docs/functions-runtime.md,docs/http-api.md internal/admin/function_execution.go,internal/admin/functions.go internal/app/admin_functions.go,tests/admin/scheduled_function_test.go tests/app/function_http_test.go
- `31e357c2` — 2026-11-28 14:37:51 +0330 — Allocate new immutable publication version when edited source conflicts
  Files: admin-nextjs/app/functions/[name]/view.tsx,docs/admin-pages/function-detail.md docs/functions-runtime.md,docs/http-api.md internal/admin/function_publication.go,internal/admin/functions.go tests/admin/function_publication_test.go,tests/app/function_http_test.go
- `43b69707` — 2026-11-28 14:34:36 +0330 — Merge pull request #61 from arashrasoulzadeh/architecture/cortex-execution-telemetry
  Files: 
- `2dfef6a1` — 2026-11-28 14:33:26 +0330 — Expose bounded Cortex execution telemetry and real launcher usage
  Files: admin-nextjs/app/lib/neuron-dashboard.ts,admin-nextjs/app/workers/view.tsx docs/unified-execution-architecture.md,internal/cortex/scheduler.go internal/cortex/telemetry.go,internal/neuron/function_runtime.go internal/neuron/typescript.go,internal/plugin/process_usage.go internal/plugin/process_usage_other.go,internal/plugin/process_usage_unix.go tests/admin-ui/neuron-console.test.mjs,tests/cortex/telemetry_test.go
- `85b5f2cb` — 2026-11-28 14:31:47 +0330 — Merge pull request #60 from arashrasoulzadeh/architecture/distributed-function-runtime
  Files: 
- `ae02ab4e` — 2026-11-28 14:31:22 +0330 — Exercise published function API in runtime integration tests
  Files: internal/admin/function_execution.go,internal/admin/functions.go internal/neuron/function_runtime.go,scripts/test-runtime-api.mjs
- `dc58257e` — 2026-11-28 14:30:02 +0330 — Propagate function cancellation and cover scheduled dispatch
  Files: cmd/brain/main.go,docs/functions-runtime.md docs/unified-execution-architecture.md,internal/admin/function_execution.go internal/admin/functions.go,internal/admin/lifecycle.go internal/app/admin_functions.go,tests/admin/scheduled_function_test.go tests/development/test_dev_neurons.py,tests/development/verify-neurons.mjs
- `199c72d8` — 2026-11-28 14:29:08 +0330 — Route published functions through isolated Neuron runtime
  Files: Dockerfile,cmd/brain/main.go cmd/brain/neuron.go,internal/admin/function_execution.go internal/admin/functions.go,internal/admin/lifecycle.go internal/execution/catalog.go,internal/execution/function.go internal/execution/runtime_identity.go,internal/neuron/function_runtime.go scripts/function-worker.mjs,scripts/run-macos-typescript-worker.sh scripts/run-sandboxed-typescript-worker.sh,tests/execution/runtime_identity_test.go tests/execution/transient_catalog_test.go
- `51f4d574` — 2026-11-28 14:26:34 +0330 — Pin published function source versions independently from editable drafts
  Files: docs/functions-runtime.md,docs/unified-execution-architecture.md internal/admin/function_publication.go,internal/admin/functions.go tests/admin/function_publication_test.go
- `0e64ea4e` — 2026-11-28 14:24:34 +0330 — Merge pull request #59 from arashrasoulzadeh/architecture/production-synapse
  Files: 
- `60b6c817` — 2026-11-28 14:24:08 +0330 — Wire explicitly enrolled Linux production Neurons to Cortex
  Files: cmd/brain/dev_neurons.go,cmd/brain/main.go cmd/brain/neuron.go,cmd/brain/production_neurons.go docs/admins.md,docs/developers.md docs/unified-execution-architecture.md,internal/execution/runtime_identity.go internal/synapse/enrollment.go,internal/synapse/grpc_server.go tests/execution/runtime_identity_test.go,tests/synapse/enrollment_test.go
- `0285a832` — 2026-11-28 14:22:20 +0330 — Merge pull request #58 from arashrasoulzadeh/architecture/cortex-default-execution
  Files: 
- `78da740b` — 2026-11-28 14:21:57 +0330 — Default managed development plugin execution to Cortex
  Files: Makefile,cmd/brain/dev_neurons.go cmd/brain/main.go,docs/admins.md docs/developers.md,docs/unified-execution-architecture.md scripts/start-local.sh,scripts/test-production-runtime.sh tests/development/test_dev_neurons.py,tests/development/test_makefile.py
- `e4342b27` — 2026-11-28 14:21:12 +0330 — Fix scheduled caller context name collision
  Files: internal/plugin/scheduled_execution.go
