# Agent instructions

## Project

This repository is building a Go multi-tenant mobile backend. The complete product context is in [plan.md](plan.md); the executable checklist is [roadmap.md](roadmap.md).

## How to work

1. Read this file and the relevant `.agents/phase-XX.md` file.
2. Select the first unchecked roadmap item in the active phase unless the user names another item.
3. Inspect existing code before editing. Preserve unrelated user changes.
4. Write one focused failing test first.
5. Implement the smallest change that passes it.
6. Run the focused test, then the affected package tests.
7. Update the matching checkbox only after verification.
8. Commit the completed checkbox as a small, focused commit.
9. Push the commit to the configured remote and confirm the push succeeded.
10. Report the commit, push result, changed files, tests run, and next smallest step.

## Non-negotiable boundaries

- Every request, job, socket, worker, file, function, and database operation carries a project identity.
- Never trust project IDs, plugin grants, room membership, or caller permissions from client input.
- Recheck lifecycle, grants, versions, and idempotency at commit time.
- Keep operator sessions separate from end-user sessions.
- Keep secrets out of logs, images, client bundles, fixtures, and diagnostics.
- Use parameterized SQL and validated identifiers.
- Do not add broad dependencies or infrastructure before a focused test requires them.
- Do not mark a task complete without a reproducible verification command.
- Do not mark a checkbox complete without committing and pushing its verified change.
- Keep one checkbox's implementation in one commit when practical; do not mix unrelated work.
- After the final checkbox in a phase, run the phase exit gate, commit its result, and push again if the gate creates changes.
- After every phase, update `CHANGELOG.md`, `docs/developers.md`, and `docs/admins.md` with the verified work, operator impact, limitations, and next phase.
- Treat the phase documentation update as part of the phase exit gate; do not declare a phase complete without it.

## Coding defaults

- Prefer small packages with explicit interfaces at boundaries.
- Keep transport adapters thin and business decisions in testable services.
- Use controlled clocks for expiry, retries, leases, drain windows, and schedules.
- Make retries idempotent and bounded.
- Use real PostgreSQL or Compose tests when behavior depends on transactions, roles, row policies, or queues.
- Keep worker, compiler, runner, and gameplay work outside unrelated request paths.

## Local development

- Do not run Docker or Docker Compose during ordinary local development.
- Use an in-process test server and SQLite for fast local unit and integration tests where PostgreSQL-specific behavior is not under test.
- Keep PostgreSQL and Compose checks as explicit CI, release, or opt-in integration checks.
- Do not replace PostgreSQL-specific tests involving roles, row-level security, migrations, locks, or transactions with SQLite; mark those tests as PostgreSQL integration tests.
- Document any command that requires Docker before running it.

## Phase routing

Read only the relevant short guide before coding:

- [Phase 00](.agents/phase-00.md) foundation
- [Phase 01](.agents/phase-01.md) projects and plugins
- [Phase 02](.agents/phase-02.md) arena
- [Phase 03](.agents/phase-03.md) app data and auth
- [Phase 04](.agents/phase-04.md) functions and turns
- [Phase 05](.agents/phase-05.md) plugin lifecycle and UI
- [Phase 06](.agents/phase-06.md) clients and release

## Completion format

End each coding task with:

- Outcome
- Files changed
- Tests and commands run
- Security or migration considerations
- Next unchecked roadmap step
