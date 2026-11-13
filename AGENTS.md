# Agent instructions

## Project

This repository is building a Go multi-tenant mobile backend. The complete product context is in [plan.md](plan.md); the executable checklist is [roadmap.md](roadmap.md).

## Plugin-first architecture law

- Core host infrastructure is limited to transport, project isolation, configuration, lifecycle, permission brokering, storage primitives, worker supervision, and shared contracts.
- Every user-facing or product capability must be implemented as a separate TypeScript plugin under `plugins/<plugin-id>/`.
- Go is reserved for the host infrastructure; no plugin implementation may be written in Go.
- Do not add auth, files, tables, subscriptions, rooms, matchmaking, analytics, workflows, dashboard features, or client-facing product behavior directly to the core host.
- Every plugin must have a manifest, implementation, focused tests, permissions, configuration schema, developer notes, and administrator notes.
- Each plugin may contain multiple TypeScript files, but every file must be listed in `plugin.yaml` `source_files`; do not add plugin-local `package.json` or README files. `index.ts` is only the public export/bootstrap boundary.
- Each handler/function file must expose one product function and one responsibility. Classes with multiple unrelated responsibilities are prohibited; split them into focused modules. Shared logic must be imported from the SDK or an explicitly declared plugin-local shared file.
- Plugin implementations must stay clean and readable: audit all plugins for repeated request builders, validation helpers, handlers, and contracts; move genuinely shared behavior into `plugins/sdk-typescript/src/` and keep plugin `index.ts` focused on product behavior.
- TypeScript lint is mandatory: run `npm run lint:ts` with typecheck and tests; lint failures block a checkbox or phase gate.
- Apply SOLID and DRY across plugins: one responsibility per service/helper, explicit boundary interfaces, and no duplicated generic validation, request, or handler logic.
- In each TypeScript source file, keep imports, types, constants, and other file-level variables first; place classes next; place standalone functions after classes, with class fields before methods. Use descriptive domain names instead of ambiguous names.
- `plugin.yaml` is authoritative for plugin title, description, version, routes, function names, capabilities, extension points, and configuration; TypeScript must not duplicate manifest metadata or route allowlists.
- `plugin.yaml` is also the import boundary: declare every plugin source file in `source_files`, every runtime/module import in `imports`, and every plugin-to-plugin dependency in both `dependencies` and `plugin_imports`; the host must reject undeclared files, imports, or plugin dependencies before execution.
- Manifest lists must use block YAML sequences (`source_files:`/`imports:`/`plugin_imports:` followed by `-` entries), never inline flow lists; this keeps diffs readable and makes generated/automated edits safe.
- The host discovers YAML routes and denies undeclared operations or capability mismatches before invoking a plugin function. Documentation is generated from the YAML contract.
- TypeScript plugins communicate with the Go host through the versioned RPC contract and may not bypass the permission broker or project boundary.
- Plugins remain dynamically replaceable: version, enable, disable, promote, rollback, drain, and remove them through host lifecycle contracts.
- All tests for every layer—unit, contract, API, admin, plugin, integration, and end-to-end—must live under the top-level `tests/` directory. Never add test files beside implementation packages.
- Preserve test package boundaries under `tests/<area>/`, and run the full suite with `go test ./...` after moves.

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
- Project identity is host-authenticated execution context, never authorization data supplied by a plugin, script, or client. Request builders may serialize fixture input, but production handlers must receive and enforce the host context before invoking business functions.
- Never authorize from a request builder's project ID, payload, route parameter, or object key; rebind identity from the host context and reject mismatches.
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
- Changelog entries must be newest first.
- Changelog entries must be separated per commit and include the commit hash and subject.
- Changelog entries must include the commit date and time in ISO 8601 format (UTC).

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
