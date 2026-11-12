# Phase 01 — Project and plugin host

[Phase checklist](../phases.md) · [Architecture plan](../plan.md)

Build the smallest project and permission model the co-op arena needs. Use the local Go, Compose, and PostgreSQL foundation from phase 00. Finish the full lifecycle and UI extension surface in [phase 05](05-plugin-ui-and-workflows.md).

## 01.1 — Create and resolve a project

- [ ] **Red.** Test that a request can resolve an existing project and that unknown, inactive, and unrelated projects cannot reach a protected operation.
- [ ] **Green.** Persist the project ID and status in PostgreSQL. Resolve the project at the entry point and carry its identity through protected operations.
- [ ] **Refactor.** Put project resolution in one place and use project-qualified keys in the first integration fixture.

## 01.2 — Bootstrap operator access

- [ ] **Red.** Test operator bootstrap, sign-in, and recovery. Cover rejected MFA, CSRF, and an ordinary session attempting an operator action.
- [ ] **Green.** Add CLI bootstrap and recovery and an MFA-protected operator session. Store session digests and audit operator access.
- [ ] **Refactor.** Keep operator authentication separate from project account authentication. Share session storage primitives where both paths use them.

## 01.3 — Register one reviewed plugin version

- [ ] **Red.** Test that one valid immutable manifest registers and that incompatible protocol versions, duplicate registrations, and invalid config schemas are rejected.
- [ ] **Green.** Record the stable plugin ID, version, approved artifact or endpoint, compatible protocol, required dependencies, requested grants, and hook declarations. Separate deployment registration from project installation.
- [ ] **Refactor.** Put the first fixture in a shared contract test package that both worker transports can run.

## 01.4 — Install, grant, and enable a plugin

- [ ] **Red.** Test that an install starts disabled, enabling fails with missing grants or failed readiness, and two projects can enable the same registered version independently.
- [ ] **Green.** Add scoped installation, grant approval, and enablement as operator use cases. Persist revisions, audit mutations, and store worker readiness separately from the saved enablement choice.
- [ ] **Refactor.** Use one activation rule for all callers. Keep the permission decision independent of HTML rendering and worker transport.

## 01.5 — Gate a broker call at project and matchmaking scope

- [ ] **Red.** Test one permitted operation and denied cases for a wrong project, missing actor permission, missing plugin grant, disabled matchmaking config, wrong resource, and expired invocation credential.
- [ ] **Green.** Derive project and match scope from trusted server context. Use short-lived invocation identities and check the intersection of actor, installation, and config permissions. Recheck the current grant and lifecycle revision before a host-side effect commits.
- [ ] **Refactor.** Share the broker policy across invocation paths. Use typed requests and results for the first operation rather than a generic command interface.

## 01.6 — Run the contract fixture in a bundled worker

- [ ] **Red.** Run the fixture against a worker process on the local Compose stack. Cover successful dispatch, crash, timeout, excessive output, and attempts to access host secrets or forbidden filesystem and network destinations.
- [ ] **Green.** Start a bundled Go worker under a restricted identity. Use bounded private IPC, cap resources, and stop the worker on deadline. Return a bounded error when the worker is unavailable.
- [ ] **Refactor.** Separate the worker supervisor from the invocation contract. Preserve a way to observe worker readiness and failures without logging request secrets.

## 01.7 — Run the same fixture through an external worker

- [ ] **Red.** Reuse the contract suite with a private service endpoint. Add identity mismatch, incompatible handshake, and callback replay cases.
- [ ] **Green.** Add authenticated registration and invocation over the private versioned protocol. Bind service identity to the project installation and use the same broker, deadlines, and effect validation as the bundled worker.
- [ ] **Refactor.** Keep transport-specific code at the edges. Run the same positive and denial fixtures through both adapters.

## 01.8 — Expose the operator use cases in the UI

- [ ] **Red.** Test that an operator can create a project, approve a grant, and enable a plugin using the bundled dashboard. Verify that rejected operations display their reasons.
- [ ] **Green.** Bootstrap the dashboard plugin from the operator CLI. Add Go-rendered forms for projects, plugin installation, grants, and enablement. Show worker readiness separately from saved enablement.
- [ ] **Refactor.** Reuse the existing operator use cases and request authorization. Keep presentation decisions in the dashboard adapter.

## 01.9 — Add guest auth and the first analytics event

- [ ] **Red.** Test that an enabled guest plugin issues a restricted session for its project, that a disabled guest plugin issues none, and that event payloads reject unapproved fields.
- [ ] **Green.** Add guest sign-in through a project plugin and a versioned analytics event contract carrying the project, plugin, event ID, timestamp, and optional matchmaking config. Use a local test sink for this phase.
- [ ] **Refactor.** Keep session issuance in the host. Reuse the event contract when the arena adds the real Eye exporter.

## Exit gate

- [ ] Create two projects through the admin UI and independently configure the same plugin version in each.
- [ ] Run the same contract tests through bundled and external workers. Verify real isolation and bounded failure in Compose.
- [ ] Demonstrate that revoked grants or a disabled scope reject an in-flight effect at commit.
- [ ] Issue a project-scoped guest session and capture one validated event for the arena fixture.
