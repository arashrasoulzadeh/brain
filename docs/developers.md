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

Product features belong under `plugins/<plugin-id>/` as TypeScript serverless functions. Each plugin contains only `index.ts` and `plugin.yaml`; the YAML is authoritative for metadata, routes, functions, permissions, and configuration. The host supplies runtime dependencies and generates documentation. Go is reserved for the core host, and tests live under top-level `tests/`.

Plugin communication follows [the plugin communication architecture](plugin-communication.md): JSON-RPC over Unix sockets locally, gRPC/mTLS externally, and the shared host authorization contract for both.

## Security rules

Never trust project IDs, plugin grants, room membership, or permissions supplied by clients. Recheck lifecycle, grants, versions, and idempotency at commit time. Keep secrets out of logs and tests.
