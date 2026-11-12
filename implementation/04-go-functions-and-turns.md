# Phase 04 — Go functions and turn-based rooms

[Phase checklist](../phases.md) · [Architecture plan](../plan.md)

Build on the project plugin grants, native table exposure, and change capture from the earlier phases. Deliver dashboard-authored Go functions and turn rules, an offline compiler, a sandboxed runtime, durable triggers, and project workflow execution.

For each slice, reproduce the behavior in a failing check, make it pass through the compiler, host, or room boundary that owns it, then refactor. Keep compilation, execution, and data commits separate. Reuse the host's permission broker, project checks, and transaction logic for every invocation path.

## Offline Go compiler

- [ ] Add a failing fixture that compiles an entry package and a project-local helper to `wasip1/wasm` with the approved SDK imports.
- [ ] Build the separate compiler image with a pinned standard Go toolchain, platform-generated module settings, and approved offline dependencies. Add the compiler and runner to the local Compose profile and release manifest fixture.
- [ ] Add failing tests for an unsupported or missing import, cross-project source, workspace or toolchain overrides, and source or artifact limits.
- [ ] Validate source bundles and import policies before compiling. Use a disposable workspace, a clean environment, a local toolchain, fixed build flags, and vendored dependencies.
- [ ] Add sandbox tests for forbidden filesystem and network access, CPU or memory exhaustion, and oversized compiler diagnostics.
- [ ] Run the compiler in a credential-free, network-isolated workspace with bounded scratch storage and enforced CPU, memory, process, and wall-clock limits. Start with one active build at a time on the 4 vCPU, 8 GB VPS profile.

## WASM runner and publication

- [ ] Add failing invocation fixtures for a valid typed request, malformed output, an extra result, excessive diagnostic output, and an unexpected WASI import.
- [ ] Run the pinned compiler artifacts in fresh wazero instances. Use bounded JSON over stdin and stdout, capped stderr, and only approved WASI interfaces. Keep filesystem mounts, environment credentials, network access, and direct database calls outside the sandbox.
- [ ] Add failing tests for an infinite loop, linear-memory growth, a timeout during startup, and a sandbox escape attempt. Confirm that failed invocations leave no writes.
- [ ] Enforce wazero execution deadlines and linear-memory limits, plus outer worker limits. Bound compiled-artifact cache size, concurrency, input and result sizes, and diagnostic retention.
- [ ] Add failing editor tests for saving a draft, compiling a shared package, reporting source diagnostics, testing with fixture data, publishing, and rolling back. Cover a failed draft that leaves the active version available.
- [ ] Add dashboard source and helper editing, fixture runs, and audited publication. Pin the validated artifact digest, toolchain, SDK contract, source bundle, and grants. Use retained artifacts for rollback and rooms that reference earlier versions.

## Direct invocation and durable triggers

- [ ] Add failing HTTPS invocation tests for unauthorized reads, unapproved write intents, stale input versions, and retries after a committed response is lost.
- [ ] Load only declared, authorized inputs. Validate output and commit the approved intents, result, and idempotency record together after rechecking actor, plugin grants, and input versions.
- [ ] Add failing scheduler tests for concurrent claims, duplicate UTC occurrences, and restart after missed runs.
- [ ] Enqueue the latest missed occurrence per schedule and atomically advance the next due time. Pin the published function version at enqueue.
- [ ] Add failing record-trigger tests for committed versus rolled-back writes, direct SQL and interactive function writes, and async writes that still invalidate subscriptions.
- [ ] Enqueue matching jobs through the table capture transaction. Identify asynchronous writes by their dedicated PostgreSQL login to suppress recursive jobs and retain realtime events.
- [ ] Add failing worker tests for an expired lease, revoked grant, duplicate attempt, terminal error, and a crash before or after a commit.
- [ ] Claim jobs with short leases and a scoped project identity. Recheck the active trigger, lifecycle, grants, and lease at commit. Commit the validated effects with job completion, and cap retries.
- [ ] Add dashboard controls for schedules and record triggers, pause and resume, sanitized run status, and audited retry of terminal failures.

## Project workflow execution

- [ ] Add failing tests for a workflow that starts on a record event, applies a condition, waits for approval, and invokes a published function. Reject an invalid graph and a step whose project grant is missing or revoked.
- [ ] Add versioned workflow definitions with scoped inputs and explicit step capabilities. Validate the graph before publication and expose typed step contracts for Phase 05 extension points.
- [ ] Add failing tests for repeated workflow dispatch, approval replay, and retries across a worker restart.
- [ ] Dispatch workflow steps through the durable queue and apply the existing lease, idempotency, and commit checks. Show run status without exposing input secrets or credentials.

## Turn-based rooms

- [ ] Add failing Tic-tac-toe fixtures for initialization, legal moves, out-of-turn moves, a win, and a draw.
- [ ] Publish the example Go rules through the same editor and compiler. Give each turn only the state, verified player slot, command, and explicit host-provided randomness or time when needed. Use a pure runtime profile and reserve capacity for turns.
- [ ] Add failing room tests for duplicate command IDs, payload mismatch, concurrent moves, malformed rule results, and reconnect after a service restart.
- [ ] Run turns outside database locks. Recheck membership, phase, and version before committing state, player views, and the accepted command together. Keep the pinned rules artifact available for durable rooms.
- [ ] Add a hidden-state fixture that proves players receive only their authorized views. Cover reconnect, a timed-out move, and replay checks for the reviewed rules. Reconstruct history from committed outcomes.

## Exit gate

- [ ] Run compiler and runner isolation tests in the local Compose release profile. Verify that the editor can publish and roll back an artifact with a shared helper.
- [ ] Run direct, scheduled, and record-change invocations end to end. Verify crash recovery, grant revocation, and job retries against real PostgreSQL transactions.
- [ ] Play a complete Tic-tac-toe match and the hidden-state fixture with two clients. Repeat with background function work active and record turn latency, compiler queue time, and memory use on the reference profile.
