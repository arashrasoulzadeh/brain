# Developer guide

## Workflow

1. Read `AGENTS.md` and the relevant `.agents/phase-XX.md`.
2. Select the first unchecked item in `roadmap.md`.
3. Add one focused failing test.
4. Implement the smallest passing change.
5. Run `test -z "$(gofmt -l .)"` and `GOCACHE=/tmp/brain-go-cache go test ./...`.
6. Update only the verified roadmap item.
7. Commit and push before starting the next item.

At the end of every phase, update this guide with the implementation notes, test commands, compatibility changes, and the next phase before committing the phase result.

Keep changelog entries newest first, with one entry per commit containing the short commit hash and subject.

## Tests

All tests live under the top-level `tests/` directory. Keep tests grouped by area, preserve external package boundaries, and run `GOCACHE=/tmp/brain-go-cache go test ./...` after every test move or implementation change.

## Phase 00 completed

The foundation phase now has a Go host, validated configuration, an injectable database owner, bounded readiness retry, migration orchestration, runtime DDL protection, isolated local settings, an opt-in Compose smoke test, and CI checks for formatting, tests, and amd64/arm64 image builds. The local verification command is `GOCACHE=/tmp/brain-go-cache go test ./...`. Phase 01 begins with project and plugin-host operations.

## Local development

Do not run Docker during normal development. Use in-process tests and isolated local settings. Use real PostgreSQL integration tests when behavior depends on roles, row-level security, migrations, locks, or PostgreSQL transactions.

## Package boundaries

- `cmd/brain`: process startup and wiring.
- `internal/app`: HTTP routing.
- `internal/config`: validated runtime configuration.
- `internal/db`: database ownership, readiness, and role policy.
- `internal/migrate`: migration orchestration.
- `internal/project`: project resolution.
- `internal/plugin`: plugin registration and scoped authorization.

## Plugin-first feature development

Product features belong under `plugins/<plugin-id>/` as TypeScript serverless functions. Each handler has one focused product function; `index.ts` is only the public export/bootstrap boundary. Every source file and import is declared in YAML, which is authoritative for metadata, routes, functions, permissions, and configuration. The host supplies runtime dependencies and generates documentation. Go is reserved for the core host, and tests live under top-level `tests/`.

This split is intentional: a handler can be replaced, scaled, tested, and audited independently. Analytics is the first migrated plugin; follow its `types.ts`, one-handler-file, and export-only `index.ts` pattern for the remaining plugins.

Phase 01 verification: the dashboard is bootstrapped through the CLI; YAML routes are host-discovered and unauthorized operations are rejected; readiness is reported separately from enablement; auth supports project-scoped guest sessions; analytics events are validated against allowed fields and captured by an in-process sink. Run `go test ./...`, `npm run typecheck`, and `npm run test:ts` locally.

## Phase 02 completed

Realtime gameplay is covered by project-scoped WebSocket endpoint validation, exact origin allowlists, single-use connection tickets, heartbeat expiry, bounded frames, bounded FIFO send queues, authorized snapshots and commands, reconnect checks, and sequence-based event replay. The arena plugin now supports a validated start/action/result flow, idempotent completed-result storage, Eye event export, an eight-player fixture, and latency/memory measurement. The Phase 02 gate is in `tests/arena/phase02_gate.test.ts`. Run `npm run typecheck && npm run test:ts` and `GOCACHE=/tmp/brain-go-cache go test ./...`; do not run Docker locally.

## Phase 03 completed

Phase 03 now covers project-scoped identity and sessions, invitations, OIDC/PKCE validation, OTP delivery and recovery, Argon2id passwords, PostgreSQL exposure contracts, versioned and idempotent records, guided schema previews, transactional/API/direct-SQL change capture, subscription tickets and replay, private project-namespaced files, quota and upload reconciliation, protected downloads, explicit public publication, and dashboard settings, record, table, file, audit, and browser-quickstart models. The gate fixtures prove API/direct-SQL isolation and rerun guest co-op. CI runs Go formatting/tests, the full TypeScript suite, image build, and Compose smoke verification. Locally run `GOCACHE=/tmp/brain-go-cache go test ./...`, `npm run typecheck`, and `npm run test:ts`; do not run Docker locally.

Plugin communication follows [the plugin communication architecture](plugin-communication.md): JSON-RPC over Unix sockets locally, gRPC/mTLS externally, and the shared host authorization contract for both.

## Security rules

Never trust project IDs, plugin grants, room membership, or permissions supplied by clients. Recheck lifecycle, grants, versions, and idempotency at commit time. Keep secrets out of logs and tests.
## Plugin quality gate

Plugin changes must pass `npm run lint:ts`, `npm run typecheck`, and `npm run test:ts`. Shared generic request, identity, and validation helpers belong in `plugins/sdk-typescript/src/`; plugin entrypoints should remain product-focused.

`plugin.yaml` is the execution boundary. Declare all plugin files in `source_files`, all module/runtime imports in `imports`, and plugin-to-plugin calls in `plugin_imports` plus `dependencies`. The host validates these declarations before loading a plugin.

The host also scans each declared source file and rejects missing files or import specifiers absent from the YAML allowlist via `ValidatePluginSources`.

Turn-based rooms expose deterministic timeout handling through `expireTicTacToeTurn(state, now, nextDeadlineAt)`; callers supply persisted clock values and must commit the resulting revision atomically.

Committed Tic-tac-toe outcomes can be replayed with `reconstructTicTacToeHistory`; revisions must be contiguous from zero and every move is revalidated during replay.

The host registry promotes only installed, enabled plugin versions and keeps active-work pins on their original version during promotion or rollback.

Promotion also validates every configured field against the manifest schema and rejects missing or incorrectly typed values before changing the active version.
