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

## Security rules

Never trust project IDs, plugin grants, room membership, or permissions supplied by clients. Recheck lifecycle, grants, versions, and idempotency at commit time. Keep secrets out of logs and tests.
