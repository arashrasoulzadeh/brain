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


## Cortex, Synapse and Neuron implementation progress (not phase exit)

| Commit | Verified increment |
| --- | --- |
| 74a5d05 | Cortex scheduler with compatible runtime selection, project and global queue bounds, capacity accounting, session fencing, expiration, draining and zero-Neuron rejection. |
| e0cf889 | Synapse v1 protobuf schema and transport-independent session engine: version negotiation, ordered envelopes, assignment leases, acceptance, correlated completion and bounded cancellation acknowledgements. |
| 4f89cc6 | Mutual TLS 1.3 helpers, canonical certificate-bound Neuron identity, loopback TLS verification and deterministic Cortex queue priority. |
| 31da1f0 | Lease-authorized artifact chunks and bounded contiguous assembly with complete digest verification. |
| 10f5fee | Neuron worker-local engine: exact runtime selection, verified bundles, independent input copies, execution slots, draining, cancellation and validated result envelopes. |
| 4127042 | In-process integration of the actual Cortex scheduler, Synapse session, artifact transfer and Neuron engine, including zero-capacity and duplicate-completion rejection. |

The combined new packages were formatted and tested in an isolated Go 1.23 module. Ten race-enabled repetitions of all five test packages passed with 100% combined statement coverage across the four new implementation packages; go vet passed. The Cortex concurrency test uses four in-process adapters, not four separate operating-system processes. The TLS tests use actual loopback TCP connections; the pipeline test uses the internal session API and a test runtime adapter, not gRPC or a production sandbox. Remote Git blob hashes were checked against the locally tested files.

### Transport dependency bootstrap

The protobuf schema compiles. The repository vendors Go dependencies, so gRPC generation and dependency synchronization are being tested separately on architecture/synapse-codegen. The first generation attempt with an older gRPC version failed the vendor consistency check. A patched gRPC v1.85.0 and Go 1.26.11 were selected rather than introducing an outdated network-facing dependency. The second bootstrap run generated bindings, synchronized vendor dependencies, built the generated package and passed the focused race/coverage tests. The full Go suite failed because crypto/rand.Read now terminates the process on injected entropy-reader failure. The separate bootstrap commit 68b12b2 replaces five existing error-handling calls with io.ReadFull(rand.Reader, ...) so the existing error paths can remain testable; its verification run is pending. The temporary workflow only publishes generated files to the bootstrap branch after all its checks pass. Neither the new dependency versions nor these compatibility edits have been merged into the architecture branch or main. The Docker and toolchain configurations still require coordinated upgrade and verification.

CI evidence: https://github.com/arashrasoulzadeh/brain/actions/runs/46073085872 (failed full suite); https://github.com/arashrasoulzadeh/brain/actions/runs/46073473685 (compatibility recheck). The pre-existing repository-wide coverage failures documented above remain unresolved.

### Outstanding production integration

The generated bindings, authenticated gRPC client/server stream adapters, actual Neuron process, concrete isolated runtime adapters, durable artifact catalog and host authorization/commit integration remain incomplete. The existing production routes and worker lifecycle remain unchanged. Core, plugin and user invocation migration, nested calls, remote process tests, CLI/admin, operational readiness and final cleanup still require implementation and full regression verification. No phase or final acceptance checklist is marked complete, and no existing security control or coverage threshold is relaxed.

## Phase 6: prepared plugin migration increment (verification pending)

The explicit loopback development mode `BRAIN_DEV_PLUGIN_EXECUTION=cortex make dev` binds registry installations to process-free lifecycle handles rather than dedicated TypeScript worker processes. Brain snapshots the validated manifest and its declared source files into the immutable shared catalog; each invocation still carries the admitted project, original actor, caller, operation, resource, configuration and available dependencies. Cortex assigns the exact artifact to a compatible Neuron, which reconstructs the prepared context and executes in a fresh Bubblewrap or local Seatbelt sandbox. The existing Brain storage planner and post-result lifecycle/grant checks remain the only durable commit path. Observation intents travel back through the same registry validation and delivery logic. No legacy fallback is used in this selected mode.

This increment keeps host.admin, host.model and host.call mediation in Brain. Host model semantics and delegated pure calls need separate migration analysis. The compatibility profile is local-only; the trusted SDK and runtime package scope are resolved from the installed Brain runtime, not transferred from another machine. Durable artifact version/digest pins, remote runtime identity, cancellation propagation from the HTTP caller, nested continuations, observation and WebSocket parity, and a complete security audit are still exit criteria. Production and normal development retain the existing explicit legacy path until the new mode is verified. Zero managed Neurons must leave operator recovery available and application readiness unavailable; enabling an installation only activates its catalog-backed handle. Capacity loss does not trigger a retry or local fallback.

The CI changes exercise the real plugin-compatible SDK probe on both platforms and the existing packaged HTTP/storage/lifecycle test suite against the scheduled mode in addition to the legacy production run. Results must be checked before declaring this phase complete. Coverage gates remain unchanged.


### Durable plugin artifact pins (verification pending)

Cortex-mode startup now binds the artifact catalog to the existing Brain SQLite database before registering manifest-selected bundles or restoring installations. The host persists the exact scope, artifact ID, version, digest, runtime profile and canonical entrypoint list in an immutable insert-or-verify table. A conflicting source, runtime or entrypoint change under an existing version fails startup; an unavailable pin store fails registration closed. Repeated identical registration and independent project/version scopes remain valid. The first Cortex startup establishes pins from the then-current validated source; historical versions from before this change cannot be reconstructed. Back up the database together with the trusted source distribution; pins do not persist the bundle bytes or attest to the external SDK/runtime files. This is not remote runtime enrollment or production cutover. Regression tests exercise restart, concurrent conflicting registration, unavailable storage and late binding. CI verification is pending; coverage gates remain unchanged.


### Scheduled plugin caller cancellation (verification pending)

Manifest HTTP handlers now preserve the trusted request context through the registry and any fixed pure delegation. WebSocket admission carries its HTTP request context and established sockets carry their existing bounded session context. Scheduled execution uses the earlier caller or invocation deadline and is cancelled on either caller cancellation or installation stop; the registry rejects cancellation before dispatch and under the final commit locks. Cancellation cannot undo an already-started durable transaction. Legacy synchronous workers are not interrupted, observation jobs remain independently durable, and custom function execution and operator command adapters still require their own migration. The regression test exercises actual Cortex placement with a controlled late-result executor, deadline propagation, pre-cancelled admission, no commit on cancellation and slot reuse. It is not a new end-to-end remote process test. CI verification is pending.


### Durable observer capacity deferral (verification pending)

Before dispatching a queued observation, Brain revalidates the publisher, target, active version, original principal and grants. If the scheduled installation is active but no compatible Neuron is registered, it defers the existing durable row by five seconds without consuming an execution attempt. Queue capacity limits remain in force; a prolonged outage can fill the queue and cause existing admission backpressure. Busy compatible Neurons and ambiguous post-dispatch failures retain the existing bounded retry policy; this change does not add exactly-once observer commits, multi-host leases or automatic retries of arbitrary plugin invocations. Legacy-worker retry behavior is unchanged. The regression test exercises durable outbox delivery across more than three capacity deferrals, identity preservation, recovery through Cortex and receipt acknowledgement. CI verification is pending.


### Credential-bound scheduled dispatch (verification pending)

Cortex plugin startup now explicitly binds the registry to a scheduled-only worker factory before restoring installations. The mode cannot be switched to a legacy factory, and manual worker bindings must match the exact scheduled installation. Credential-bound calls use the normal registry dispatch in this mode rather than executing an external worker supplied by a trusted legacy adapter, including projects without Access Management. The original single-use credential is consumed, resource binding is checked, and the broker is rechecked before ordinary scheduled results can commit or pure delegated results can return. No compatible Neuron means unavailable, never fallback. Host model operations retain their existing host-owned execution and authorization path; this is not remote worker enrollment or conversion of the legacy handshake to Synapse. The legacy registry mode remains supported. The regression test covers capacity loss, replay, resource substitution, caller identity, wrong-project binding and broker revocation before commit. CI verification is pending.


### Default managed development execution (verification pending)

Both `make dev` and `make dev server` now start four managed Neurons by default and select the scheduled-only registry before restoring installations. The explicit local comparison mode is `BRAIN_DEV_PLUGIN_EXECUTION=legacy`; it is never an automatic fallback. Zero workers remain a supported recovery configuration. The Linux and macOS managed-pool API parity tests now exercise the default without a Cortex opt-in variable. Production launch configuration, external certificate enrollment and remote runtime identity are not changed by this increment. Full CI is pending.


### Explicit production Synapse enrollment (verification pending)

Linux Brain accepts an explicit `BRAIN_PLUGIN_EXECUTION=cortex` configuration with a private Synapse address, server certificate/key, dedicated Neuron CA and a JSON map of Neuron IDs to one or two SHA-256 leaf-certificate fingerprints. CA trust and canonical SPIFFE URI checks still apply; a second certificate for the same ID is rejected unless explicitly enrolled. The server requires the runtime-source digest in the authenticated stream and fences sessions at certificate expiry. The digest covers the trusted SDK tree, runtime scripts and locked package metadata; it is a compatibility check, not remote attestation or proof of installed dependency bytes. Deploy verified runtime images and manage certificates externally. An enrolled Linux Neuron uses `brain neuron connect --enrolled-plugin-runtime` with its existing connection and TLS flags. Brain binds the scheduled registry even with zero Neurons; no legacy fallback is possible after selection. Partial production settings or mixed development/production settings fail closed. Legacy unconfigured production launch remains supported until deployment rollout; live certificate reload, automated issuance and separate-machine validation remain outstanding.


### Function publication identity (verification pending)

Before migrating user function computation, Brain now records exact published source bytes, method, version and digest independently from editable drafts. The active publication is selected from bounded immutable history; reuse of a published version with different bytes is denied. Legacy public records adopt their first known source, and public execution reads durable state rather than a stale draft cache. A changed publication or permission is rechecked before success observation and result exposure. This increment does not yet run Functions in Cortex or add nested delegation; the existing VM remains unchanged. Regression tests cover draft edits, version conflicts, method pinning, visibility and corrupt storage. CI verification is pending.

### Published functions on Neurons (verification pending)

Cortex-mode Brain routes published public and authenticated-user functions through a separate user-function runtime. Each admitted invocation registers a bounded project-scoped transient artifact containing exactly the selected immutable publication, and releases it after dispatch. The Neuron uses a fresh OS-sandboxed process containing a restricted JavaScript VM, pure helpers, trusted context and no host bridge. Both development and enrolled workers probe this capability before registration. HTTP cancellation reaches the scheduler; results are checked again against the current publication and caller policy. No compatible worker means failure, not local fallback. Operator draft tests still use the previous host bridge, and nested continuations, distributed draft storage, complete production cutover and verification remain outstanding.

### Cortex execution telemetry (verification pending)

Cortex records assigned outcomes after fencing, with process-lifetime counters and a newest-first bounded history of 128 executions. The operator-only Neurons endpoint and UI expose runtime, Neuron, trace/parent identity, host-measured wait/elapsed duration and reported process usage, not request payloads, caller/resource metadata, raw errors or results. Linux and macOS runtime adapters capture launcher CPU and peak RSS after process reaping; zero/unavailable usage is rendered as unavailable, not fabricated. These measurements do not prove resource enforcement or successful durable commits. History is not durable and managed drain/restart remains outstanding. Regression and UI tests are added; CI verification is pending.

### Neuron placement and managed-process controls (verification pending)

Authenticated, explicitly confirmed system operator commands can pause and resume Cortex placement for a Neuron. Host pauses are bounded, survive reconnects within the current Brain process, and do not clear a worker's autonomous drain. Local process controls stop/restart only identities launched by the existing development supervisor, wait for all active assignments (including fenced old sessions), preserve the host-selected launch configuration and wait for authenticated reconnection before removing the pause. Remote process control remains external. The operator console shows real managed process liveness independently of connections, with confirmation and refresh. Durable history, durable operator pauses across Brain restart, and platform verification are still pending.
