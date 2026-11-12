# mBaaS implementation phases

Use [plan.md](plan.md) for product and architecture decisions. Work through the implementation guides in order and check off a slice after its tests and exit condition pass. Start locally. The co-op arena is the first end-to-end milestone.

## Working method

- [ ] **Red.** Name one observable behavior and add a focused failing test. Run it and confirm the failure matches the missing behavior.
- [ ] **Green.** Add the smallest change that passes through the normal entry point. Run the focused test and the affected package or contract tests.
- [ ] **Refactor.** Keep responsibilities small, name code for the behavior it owns, and remove duplication once the common behavior is clear. Keep dependency direction from feature code toward small host contracts. Use SOLID principles where they make boundaries easier to test or extend.
- [ ] **Verify.** Use unit tests for decisions and state transitions, PostgreSQL or Compose integration tests for real boundaries, and a short end-to-end test for each critical operator or player flow. Use controlled clocks and fixtures for concurrency, expiry, and retries.
- [ ] **Record.** Keep acceptance commands and measured results in the phase notes when the implementation is verified. Close each phase with a passing exit gate.

## Phase checklist

- [ ] **00 — Local foundation.** Boot a Go host and PostgreSQL under Compose. Prove migrations, health, a repeatable test command, and GitHub Actions checks. [Implementation guide](implementation/00-local-foundation.md)
- [ ] **01 — Projects and plugin host.** Create tenants in the dashboard, install and grant scoped plugins, and verify both isolated Go worker paths with the same host contract. Deliver guest identity and the event contract for the arena. [Implementation guide](implementation/01-project-and-plugin-host.md)
- [ ] **02 — First end-to-end co-op arena.** Play a complete match in two browsers, recover a connection, verify a denied plugin action and one durable result, and export an Eye event. Exercise private rooms and FIFO matchmaking, then measure an eight-player room. [Implementation guide](implementation/02-co-op-arena.md)
- [ ] **03 — App data and identity.** Complete dashboard-managed auth, native PostgreSQL tables and API exposure, files, and app subscriptions. Exercise email and SMS through the local test inbox. [Implementation guide](implementation/03-app-data-and-auth.md)
- [ ] **04 — Go functions and turn games.** Write, compile, test, and publish Go from the dashboard. Run it in the Wasm sandbox, add triggers and project workflows, and complete a durable Tic-tac-toe match. [Implementation guide](implementation/04-go-functions-and-turns.md)
- [ ] **05 — Plugin UI and workflows.** Complete per-project version promotion, dependency changes, drain and stop, admin UI replacement, and server workflow extensions. [Implementation guide](implementation/05-plugin-ui-and-workflows.md)
- [ ] **06 — Clients and release.** Verify browser, Flutter, and native Android clients, publish Zarvan protocol fixtures, and test release images, manual deployment, restore, and mixed-load capacity. [Implementation guide](implementation/06-clients-and-release.md)

## Release gates

- [ ] The dashboard covers routine tenant, plugin, database, function, workflow, and live-session operations.
- [ ] Isolation holds across projects, matchmaking configs, both plugin execution paths, Go builds, and Wasm execution.
- [ ] Play the arena and Tic-tac-toe through supported client contracts. Verify recovery and permission revocation with real connections.
- [ ] Pull the complete multi-architecture release from public GHCR in a clean test environment. Verify a digest-pinned Compose rollout and recovery.
- [ ] Restore PostgreSQL and matching object versions from an off-host backup. Measure the room cap on the 4 vCPU, 8 GB profile before choosing production limits.
