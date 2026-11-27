# Unified distributed execution — architecture audit and migration plan

Status: Phase 1 audit draft; not an implementation or release sign-off.
Baseline: `main` at `5015ce8aa51373ad03fd4abafd4c68603b7fda33`.

## Objective and governing constraints

Brain is the trusted control plane; Cortex schedules individual function invocations; Neurons are independent, generic execution resources; Synapse is their versioned gRPC bidirectional protocol. Impulse is an invocation, Circuit a workflow, Pulse a heartbeat/capacity report, Memory durable state. Prefer ordinary execution terminology in internal interfaces.

Follow AGENTS.md: Go is host infrastructure only, product capabilities remain manifest-declared TypeScript plugins, tests belong under tests/, single-use host credentials and trusted project identity must survive the migration. The existing plugin-first law is not repealed by unified execution. Core infrastructure primitives remain local; any eligible executable core operation must be inventoried and must not introduce new Go product functionality. No Kubernetes dependency and no hidden execution fallback in the final architecture.

## Source audit

| Current boundary | Evidence and impact |
| --- | --- |
| Startup and readiness | cmd/brain/main.go:76–159,219–299 binds platform-specific workers, runs a sandbox probe, restores enabled installation workers, starts observation delivery, and checks installation readiness. Move probe/runtime dependencies into Neuron; split control-plane health from execution readiness. |
| Execution contract | internal/plugin/worker.go:12–113 couples requests to plugin/version and Handle to Worker lifecycle; project validation, credential binding and 1 MiB request limit are already present. Extract a context-aware execution-only interface while keeping old callers through an adapter. |
| Placement | internal/plugin/registry_workers.go:44–229,315–380 uses project:plugin@version worker keys and selects a ready dedicated worker; internal/plugin/supervisor.go supervises keyed instances. Replace placement only, not registry authorization. |
| Sandbox runtime | internal/plugin/typescript_worker.go:84–151,160–217,255 onward builds manifest/project-specific process descriptors and serializes Handle under a mutex. scripts/typescript-worker.mjs and platform launchers are the runtime boundary. Preserve declared source/import validation, no-network restrictions, output limits, and platform-specific isolation; a Neuron agent may be long lived but invocation sandboxes must not leak mutable state. |
| Lifecycle | internal/plugin/registry_lifecycle.go and registry_enable_dependencies.go:75–159 start workers during enable, rollback worker starts on persistence failure, and compare the exact approved dependency plan before enable. Split artifact availability and enabled installation state from cluster capacity; preserve atomic rollback, confirmation, promotion, drain, and audit. |
| Host authorization and effects | internal/plugin/registry_workers.go:88–229 resolves trusted projects, takes a storage planning lock, authorizes, prepares reads/writes, creates invocation context, executes, rechecks project and grants, then commits. internal/plugin/broker.go and invocation_context.go carry credential and caller boundaries. Keep planner, database, broker and commit closures in Brain. Never serialize host closures or credentials to a sandbox. |
| Other plugin execution paths | internal/plugin/plugin_call.go:21–98 has fixed manifest-declared pure delegation and explicit source/target checks; http_routes.go, model_runtime.go, external.go, observation_queue.go and WebSocket paths require call-site inventory before cutover. Declarative/model execution currently can run in host; schedule application computation without moving authorization or trusted storage primitives. |
| User functions | internal/admin/functions.go:338–622 has separate publish/run logic and Node VM subprocess; function_plugin_bridge.go:32–169 restricts plugin reads to operator tests, uses bounded framed bridge, and enforces saved function grants. docs/functions-runtime.md explicitly says VM is not a production OS sandbox, and public invocations cannot use the operator bridge. Migrate draft and published execution without loosening these rules. |
| Durable state | internal/plugin/sqlite_storage.go, internal/project/sqlite.go and legacy storage adapters retain host-owned state; productionServer refuses implicit SQLite fallback for configured PostgreSQL DSNs. Preserve database semantics; do not claim new PostgreSQL wiring. |
| Operator and public surface | internal/app/app.go, operator_auth.go, project_http.go; admin-nextjs Functions/Workers pages and docs/admin-pages/workers.md. Preserve authenticated recovery and existing HTTP error contracts; add real Neuron and execution metrics, never payloads or secrets. |
| Development and CI | Makefile, scripts/start-local.sh, .github/workflows/test.yml, docs/runtime-operations.md. Existing Go, macOS Seatbelt, Linux Bubblewrap, TypeScript lint/typecheck/test, admin/browser and release checks remain gates. |

This is a targeted static audit, not an exhaustive call graph or executed security review. Before each phase enumerate the actual callers and fixtures, including direct model/response helpers, external RPC, public routes, observations, WebSocket events and admin operations.

## Branch reconciliation

At the audited baseline the only open PR is draft #49 (feat/functions-studio-redesign), 5 commits ahead and 17 behind main. Its 14 modified/added files are Functions Studio UI, supporting frontend libraries, styles, documentation and tests, not Go runtime. Rebase/update and validate it separately; coordinate its function inspection, executions UI and contract fixtures with Phase 7/11. Do not merge a divergent draft simply to clear the branch list. No source conflict is established for this documentation-only increment.

The following nine branches have zero commits ahead of main: admin/confirmed-dependency-enable (6 behind), admin/function-development-studio (12), admin/functions-catalog-rework (16), feat/admin-operations-completion (19), feat/confirmed-plugin-install (96), feat/files-megabyte-limit (91), fix/admin-page-consistency (85), functions/host-plugin-read-bridge (11), quality/coverage-completion (53). No re-merges or deletions are needed. Recheck before implementation. roadmap_backend.md still contains an unchecked PR #38 migration gate; verify current CI and PR state before reconciling that historical checklist; do not silently check it off.

## Target boundaries

1. Brain resolves a manifest-declared operation or registered user artifact from trusted project and caller identity; it pins immutable version and digest (including dependency/source manifest), authorizes resource access and prepares bounded input.
2. Cortex creates an execution record with execution/trace/parent IDs, attempt ID, deadline, limits, authorization scope, idempotency key and runtime/security profile. Queue and slot accounting are host-owned. Runtime adapters implement Execute(context.Context, ExecutionRequest) (ExecutionResult, error); worker lifecycle is a distinct interface.
3. Synapse uses protobuf v1 and authenticated bidirectional gRPC. Require negotiated version, message size limits, bounded queues, session identity, correlation and assignment lease/attempt IDs. No worker-selected project or ownership. Remote transport uses mutually authenticated TLS or equivalently authenticated secure transport; local transport also authenticates the Neuron.
4. Generic Neuron advertises supported runtime/security profiles and slots; Brain assigns an authorized immutable artifact. Neuron verifies digest, runs a fresh isolated sandbox, and returns bounded structured results/metrics/intents. No direct database access, administrative token or unrestricted host filesystem/network. Trusted native execution must have a distinct admission/security profile; defer WASM/native adapters until independently implemented and tested.
5. Brain accepts only current-session, current-attempt results, validates output/effects, rechecks authorization/lifecycle/idempotency and commits atomically where the current host contract requires. Acknowledgment loss or stale result must not commit twice.

## Proposed decisions requiring implementation validation

- Keep artifact activation separate from worker readiness. Plugin enable can succeed without capacity once artifact integrity, dependencies and grants are validated; application readiness remains false when no eligible Neuron exists. Keep lifecycle recovery APIs available.
- Start with least-loaded compatible Neuron placement and bounded per-project/global queues; default fail fast with structured service unavailable when no eligible capacity, configurable bounded wait later. Never execute in Brain to hide missing capacity after cutover.
- Track queued, assigned, accepted/running, succeeded, failed, cancelled and lost outcomes with an attempt fencing token. Heartbeat expiry/disconnect removes schedulable capacity; retry only explicitly idempotent work with safe commit deduplication. Unknown outcome of non-idempotent work is not automatically retried.
- Nested execute is host mediated, multiplexed and permission-narrowing. Resolve declared dependency/active version policy, retain original caller, and apply depth/fan-out/cycle/deadline budgets. Waiting parent must release runnable capacity or use a bounded, separately accounted continuation mechanism; bound suspended memory and reject when continuation capacity is exhausted. A parent holding a storage planning lock must not wait on a child needing that lock. Do not ship nested execution until transaction/lock design is proven by one-slot and saturated-pool tests.
- Initially nested operations are restricted to existing authorized read/pure semantics. Independent child commits are not a distributed transaction; introduce explicit write semantics separately, with commit-time checks and failure outcomes. Do not expand public user function permissions via the compatibility adapter.
- Preserve the original HTTP envelopes, exact dependency consent, grants, audit, worker sandbox restrictions and operator/user separation. Avoid dual execution and automatic fallback during migration; legacy adapter is temporary, explicitly selected and removed after parity gates.

## Incremental delivery checklist (all unchecked)

| Phase | Increment | Exit evidence |
| --- | --- | --- |
| 1 | Review this audit; reconcile branch #49 and historical roadmap gate; complete call-site/test inventory; capture baseline CI. | Reviewed migration plan; verified baseline and branch status. |
| 2 | Add internal/execution contract, typed errors, artifact reference/digest, context and limits. Adapt existing plugin worker invocation without changing lifecycle or wire format. | Tests under tests/execution for validation, trusted identity, exact version, cancellation semantics; existing plugin regression and race tests. |
| 3 | Add proto/synapse/v1, generated bindings, authenticated stream, registration, heartbeat, assignment, accept/result/cancel with fencing. | Protocol/version, auth, replay, oversized frame and disconnect tests. |
| 4 | Add cmd/neuron and runtime adapter, managed local process supervisor and configurable slots. Relocate platform sandbox probes. | Four real local Neurons register, execute isolated workloads from distinct projects, drain and cancel; sandbox checks. |
| 5 | Cortex compatible placement, bounded queue, quotas, state, timeouts, backpressure and disconnect handling. Separate health and readiness. | Zero/four Neurons, slot saturation, stale result, quota, cancellation, safe retry and capacity removal integration tests. |
| 6 | Move plugin execution and lifecycle to registry/scheduler; include model/declarative routes, delegation, external RPC, observations and WS dispatch inventory. | No dedicated project/plugin worker; lifecycle rollback, dependency consent and version pinning, resource grants, HTTP/storage and observation parity. |
| 7 | Route draft and published user functions through Cortex with OS isolation; preserve public and operator authorization distinctions. | Cross-project denial, VM escape boundary, limits, existing API/Studio contracts and no privilege expansion. |
| 8 | Classify remaining Core operations. Schedule eligible application computation; keep transport, recovery auth, cluster admin, storage and policy primitives in host. | Zero-capacity recovery remains usable; plugin-first contract tests; no circular scheduling dependency. |
| 9 | Add context.execute, multiplexed host calls, execution trees, bounded continuation strategy and permission-narrowed compatibility adapter. | One-slot nested progress or bounded rejection, saturated pool, parallel placement, cancellation, cycles, lock safety, budgets and grant checks. |
| 10 | Remote Neurons with secure enrollment, artifact distribution/verification and reconnect semantics. | Separate-machine integration, identity/session fencing, tampered artifact, network partition, certificate/auth checks. |
| 11 | CLI/dev and admin Neurons/Cortex monitoring, traces, actual usage metrics and managed local drain/restart. | Real data, authorization and redaction, admin/browser and local dev regression. |
| 12 | Delete obsolete paths only after parity; update runbooks, developer/admin documentation, release/rollback guidance. | Full regression, race, lint, typecheck, platform sandbox, deployment and security gates; all acceptance cases below. |

Each phase gets a focused failing test first, reviewable commits and a separate PR; update roadmap only after verification, and update CHANGELOG.md, docs/developers.md and docs/admins.md per phase exit instructions. No coverage reduction or skipped tests to enable merge.

## Final acceptance

- Brain starts with zero Neurons and reports control-plane health; application readiness is unavailable and every application execution path fails closed or enters an explicitly bounded queue.
- Four generic Neurons execute eligible plugin, user and Core workloads across projects with no permanent assignment; stopping one removes capacity and stopping all preserves recovery only.
- Parallel calls distribute across Neurons; nested calls never indefinitely deadlock even with one slot or saturation.
- Cross-project/plugin/operation denial, original caller identity, exact version and digest, authenticated remote transport, replay/session fencing, resource limits and isolation are tested.
- Worker failure, late/duplicate results, disabled/revoked grants and non-idempotent unknown outcomes cannot corrupt durable state.
- Existing lifecycle, dependency confirmation, HTTP routes, observation behavior and authorization regressions pass; UI shows actual state and redacts secrets.
- All existing and new Go, race, TypeScript, browser, sandbox, deployment and documentation gates pass.

## Verification and limitations of this draft

Read-only GitHub source/tree and branch comparisons were performed against the pinned baseline. No runtime code was edited; no build, test, sandbox, race or deployment command was executed in a repository checkout. Direct git access from the execution environment failed DNS resolution; GitHub API access is available. The branch/PR carrying this document is for review, not a phase-complete or security sign-off. Before Phase 2, obtain a working checkout and run the actual repository gates; do not mark unchecked items complete merely because the plan exists.


## Phase 2 implementation progress (not phase exit)

The following reviewable implementation commits are present on the architecture branch:

| Commit | Verified increment |
| --- | --- |
| d249425 | Host-owned execution contract and structural request/result validation, exact version syntax, SHA-256 verification, identity binding, deadlines and limits. |
| 87060df | Canonical immutable file bundles; path, case, ancestor, format, depth, count and size validation. |
| 2014342 | Bounded in-memory artifact catalog with explicit project/shared scope, immutable version/runtime/entrypoint registration and defensive copies. |
| 6d56319 | Catalog-bound assignment validation, effective deadline propagation, fail-closed missing capacity and rejected late/cancelled or mismatched results. |

The initial artifact format is canonical JSON with base64 file bytes, maximum encoded size 64 MiB (configurable downward), 4096 files, 32 path components, ASCII canonical paths, and no direct filesystem extraction. The registry must select all declared files and pin runtime/dependency identities. The catalog is deliberately in-memory; durable pins and host authorization are still required before live use. ExecuteAssigned is a boundary helper, not a scheduler; it relies on the selected runtime adapter to honor cancellation and enforce actual resource limits.

Verification: in an isolated Go 1.23 module with module name brain, the new tests under tests/execution passed with the race detector, ten repetitions, and 100% statement coverage of internal/execution. go vet ./... passed for that isolated module. Remote Git blob hashes were checked against the exact locally formatted and tested files before pushing each code commit.

The full repository checkout remains unavailable in this environment (GitHub DNS resolution fails). GitHub Actions was triggered for each commit. The first implementation commit's repository-wide coverage gate reported 7248/7594 covered statements and 346 missing; the latest CI-tested baseline main commit d781058 reported 7219/7565 and the same 346 missing. The only change between d781058 and the pinned main baseline is CHANGELOG.md. Both the baseline and the first implementation commit also failed the TypeScript 100% coverage gate. Do not lower thresholds or treat these pre-existing failures as a passing phase. The latest implementation CI is not yet verified.

No existing production invocation path has been switched. The next increment is a genuine legacy worker adapter with immutable source/runtime binding, preserving the existing broker, storage planner, manifest authorization, commit rechecks, and legacy HTTP behavior. It requires a working full-repository build and sandbox regression environment. No phase checkbox or final acceptance criterion has been marked complete.
