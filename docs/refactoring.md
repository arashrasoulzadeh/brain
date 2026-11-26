# Refactoring boundaries and follow-up checks

The host remains the source of authority for project identity, lifecycle,
capabilities, and durable writes. Product behavior belongs in declared plugins.
Prefer explicit interfaces and focused modules over large generic abstractions.

## Compatibility decisions

- The SQLite storage binding still installs the legacy adapter for files,
  matchmaking, rooms, analytics, and tables. Its product logic has been isolated,
  not migrated out of Go. New product integrations should not extend this adapter.
- The invocation planner defers writes until success and reauthorization.
  In-process serialization prevents overlapping storage plans for the same
  project/plugin. This is not a distributed transaction or a rollback mechanism
  for side effects inside workers.
- Operator sessions are independent, revocable and expire after eight hours.
  Session storage remains process-local.
- The admin root coordinates existing state and callbacks; feature panels own
  rendering. Hidden legacy views are preserved to avoid unintended UI changes.

## Remaining verification and migrations

- Replace the legacy product adapter with declared plugin storage operations and
  migration tests that preserve existing persisted data.
- Define error handling and recovery for worker side effects and failed storage
  commits, including concurrent hosts; verify lifecycle persistence errors.
- Verify every production worker launch path against the sandbox policy, not
  only the sandbox fixture.
- Run native Flutter and Android integration checks on their respective
  toolchains. JavaScript and admin type checks do not substitute for native
  builds or browser end-to-end testing.
- Complete the release gates in the existing roadmaps; do not mark a gate
  complete based on a structural refactor alone.

CI results for a particular commit are the source of truth for verification.
