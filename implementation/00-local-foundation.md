# Phase 00 — Local foundation

[Phase checklist](../phases.md) · [Architecture plan](../plan.md)

**Outcome.** One repeatable local command runs the smallest Go host with PostgreSQL. The same checks run in GitHub Actions. Keep the host small enough to verify each boundary before adding plugins.

## 00.1 — Boot the host

- [ ] Add a failing test that boots the host with valid local configuration and receives a successful liveness response.
- [ ] Create the Go module and minimal HTTP entry point. Keep process setup in the command and HTTP behavior in a testable package. Pass the liveness test.
- [ ] Add a failing test for a missing or invalid required setting. Return a useful startup error with secret values omitted.

## 00.2 — Connect PostgreSQL

- [ ] Add an integration test that expects readiness to fail while PostgreSQL is unreachable and recover when it returns.
- [ ] Add bounded database startup and readiness checks. Use one connection owner and close it on shutdown.
- [ ] Add a fresh-database test for the first platform migration. Run migrations with a separate credential and verify the runtime role cannot perform DDL.
- [ ] Run migrations twice in the test stack and verify a repeated run leaves the schema at the same version.

## 00.3 — Run under Compose

- [ ] Write a smoke check for host liveness, database-backed readiness, and a clean stop in a fresh local stack.
- [ ] Add the smallest local Compose configuration to pass the smoke check. Keep the PostgreSQL port private and place test settings and data in an isolated project.
- [ ] Run the stack twice from scratch to verify the test setup is repeatable. Document setup, tests, and cleanup in one place.

## 00.4 — Run the checks in GitHub Actions

- [ ] Make formatting, Go tests, and the Compose smoke check fail independently when their expected behavior is broken.
- [ ] Wire the same local check commands into CI with read access. Verify a failed check blocks the publishing stage.
- [ ] Add the backend image build check for Linux amd64 and arm64. Add compiler and runner images when Phase 04 defines them, then assemble the final release in Phase 06.

## Exit gate

- [ ] The Go tests and Compose smoke check pass from a clean checkout locally and in CI.
- [ ] Readiness distinguishes a running process from a usable database, and migration credentials are separate from runtime credentials.
- [ ] Phase 01 can depend on the host's configuration, HTTP, and database entry points without introducing parallel bootstrap or test commands.
