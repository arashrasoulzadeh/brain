# Phase 05 — Plugin lifecycle, admin UI, and workflow extensions

[Phase checklist](../phases.md) · [Architecture plan](../plan.md)

Complete the lifecycle around the plugin host, live rooms, and workflows from the earlier phases. Reuse the scoped permission broker from [phase 01](01-project-and-plugin-host.md) for every operation in this phase.

## 05.1 — Promote one version per project

- [ ] **Red.** Test promotion with two projects and one room already running. Verify that new work in the promoted project uses the selected version while the active room and the second project keep their versions.
- [ ] **Green.** Add an admin promotion action that validates config, readiness, requested grants, and schema compatibility. Commit the selected version and revision atomically. Preserve the versions pinned by active work. Cancel incompatible unmatched tickets with a requeue reason.
- [ ] **Refactor.** Keep version selection separate from invocation. Use the existing broker policy on both pinned and newly assigned work.

## 05.2 — Roll back without restoring old permissions

- [ ] **Red.** Test that rollback routes new work to a retained compatible version, rejects a missing or incompatible version, and respects a grant revoked since the old version last ran.
- [ ] **Green.** Add the rollback action and preserve the current grant set. Require explicit approval before a version uses new or expanded capabilities.
- [ ] **Refactor.** Reuse promotion validation and the version change audit path.

## 05.3 — Preview and apply a dependency disable

- [ ] **Red.** Test transitive dependents, a stale preview, and disabling at one matchmaking scope while a second scope continues to admit work. Cover an optional plugin whose removal still allows its active room to play.
- [ ] **Green.** Preview affected plugins and scopes, then apply the approved cascade atomically against the previewed revision. Block new calls, jobs, and subscriptions, cancel affected queued and provisioning matches, and recheck permissions before outstanding effects commit.
- [ ] **Refactor.** Use one dependency traversal for preview and apply. Keep lifecycle decisions in the host so workers cannot override them.

## 05.4 — Drain or stop affected rooms

- [ ] **Red.** With a controlled clock, test that draining blocks new rooms and admits reconnects only for current participants. Verify that expiry pauses a turn-based room at its last committed move and interrupts an action room. Race an assignment and a result commit with **Stop now**.
- [ ] **Green.** Persist the transition, scope, and configurable deadline with a five-minute default. Honor the deadline on reconnect and commit. Use immediate stop for security revocation and record interrupted work. Resume a paused turn-based room only when its pinned rules and runtime are available and permitted.
- [ ] **Refactor.** Reuse the lifecycle revision checks for queue admission, socket activity, and room commits. Keep the clock injectable in lifecycle tests.

## 05.5 — Uninstall safely

- [ ] **Red.** Test that uninstall waits for active work to resolve, removes the project's grants, and preserves durable plugin data. Verify that server unregister fails while another project still references the version.
- [ ] **Green.** Add project uninstall and server unregister with separate scopes. Keep operator recovery available from the CLI and audit both actions.
- [ ] **Refactor.** Share reference checks with version retirement. Keep data deletion in its own explicit action.

## 05.6 — Add and replace one declarative admin view

- [ ] **Red.** Render a host-owned fixture view with two contributions. Test that a grant can activate one replacement for a named view and scope, and reject a second replacement or a cross-project submission.
- [ ] **Green.** Add one named extension point with Go-rendered forms and typed actions. Validate view data and route submissions through host-owned authorization, CSRF checks, and audit.
- [ ] **Refactor.** Keep view rendering, extension selection, and actions separate. Apply the same replacement rule to further named pages, navigation, record panels, and the project dashboard shell.

## 05.7 — Embed one custom admin view

- [ ] **Red.** Test a TypeScript view in a sandboxed frame. Verify that a valid typed message reaches an approved action and that unbound channels, malformed requests, cross-project calls, and ungranted actions fail. Check that operator tokens stay in the trusted UI.
- [ ] **Green.** Serve the custom bundle from a separate origin with restricted frame capabilities and network access. Dispatch typed actions through the same authorization path as declarative views. Keep grant dialogs and sensitive confirmations in the trusted control surface.
- [ ] **Refactor.** Reuse action schemas and view selection rules. Give the bridge only the operations the view needs.

## 05.8 — Extend a project workflow

- [ ] **Red.** Run one workflow with a registered plugin action, then replace a named step with a permitted plugin. Cover revoked grants, invalid graphs, and one retried step whose effect commits once.
- [ ] **Green.** Connect named triggers, conditions, actions, approval steps, and workflow providers to versioned project definitions. Use scoped project identities and reuse the queue, leases, retries, and idempotent commits from phase 04.
- [ ] **Refactor.** Keep workflow execution independent of the admin renderer and plug-in transport. Reuse the host permission broker for every step.

## 05.9 — Extend a release workflow

- [ ] **Red.** Test that a project-installed plugin cannot register a release step, an approved deployment plugin can run a named step, and its process has no access to publish credentials or the Docker socket.
- [ ] **Green.** Expose named test, build, preflight, and post-deploy steps to explicitly granted deployment plugins. Pass validated inputs and outputs across each stage. Keep privileged publishing and rollout operations in the host release path through GitHub Actions, public GHCR, and manual Compose deployment.
- [ ] **Refactor.** Reuse manifest and audit conventions across project and deployment extensions while keeping their grants separate.

## Exit gate

- [ ] From the dashboard, promote and roll back a plugin, preview and apply a dependency disable, and uninstall a project plugin with its durable records preserved.
- [ ] From the live-session view, demonstrate drain, resume, and immediate stop with the expected outcomes for both room types.
- [ ] Replace the same named admin view with a declarative view and a custom frame in separate test projects.
- [ ] Run a project workflow extension and a granted server workflow extension. Verify that permissions and credentials stay scoped to their respective stages.
