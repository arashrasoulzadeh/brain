# Backend roadmap

Backend infrastructure, release, deployment, backup, capacity, and security work. This is the active roadmap.

## Backend laws

- [x] Keep the core host limited to shared infrastructure and contracts. Verified with `go test ./...`, `go vet ./...`, and the plugin-first source layout: product capabilities remain under `plugins/`; Go host packages provide transport, isolation, lifecycle, permissions, storage primitives, worker supervision, and contracts.
- [x] Enforce project isolation at every request, job, socket, worker, and commit boundary. Verified by Go project middleware/SQL/plugin/worker/WebSocket/WASM tests, packaged isolation tests, and TypeScript isolation gates across dashboard, files, tables, auth, functions, rooms, matchmaking, subscriptions, and arena.
- [x] Require host-authenticated execution context for every production plugin handler. Verified by project middleware trust checks, broker authorization, worker invocation binding, credential tamper/replay tests, and publish-credential boundary tests.
- [x] Bind every production commit to a single-use host-issued invocation credential. Verified by credential project/operation binding, tamper rejection, replay rejection, concurrent single-use, and revoked-grant recheck tests; raw host-call commits remain test-only.
- [ ] Complete a deep security audit across every plugin and add regression tests for every finding.
- [x] Keep operator access separate from end-user access. Verified by isolated operator/end-user token stores, cross-use rejection tests, and operator revocation coverage in `tests/auth/auth.test.ts`.
- [x] Treat destructive actions, migrations, releases, and permission changes as audited operations. Verified by the audited `LifecycleService` wrapper, denied/failure/success audit tests, and admin HTTP boundary tests in `tests/admin/` and `tests/app/`.

## Backup and restore


## Final backend gates

- [x] Project isolation holds across APIs, SQL, workers, jobs, sockets, and WASM. Verified by the full Go suite, race suite, TypeScript isolation gates, and the plugin project-boundary audit commits.
- [x] Recovery works after connection, worker, and deployment failures. Verified by connection/reconnect tests, bounded worker restart tests, idempotent invocation recovery, release rollback compatibility tests, and deployment recovery-record/runbook checks.
- [x] Multi-architecture images are digest-pinned. Verified by release-manifest validation requiring immutable SHA-256 references for backend, compiler, and runner images.
- [x] Compose deployment and rollback are rehearsed in CI. Verified by the `test.yml` clean-environment and full deploy/rollback/restore rehearsal jobs, with local development explicitly remaining Docker-free.
- [x] Production limits are based on measured capacity. Verified by bounded memory/capacity validation and the eight-player capacity gate.
