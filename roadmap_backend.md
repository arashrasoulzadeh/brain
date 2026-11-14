# Backend roadmap

Backend infrastructure, release, deployment, backup, capacity, and security work. This is the active roadmap.

## Backend laws

- [ ] Keep the core host limited to shared infrastructure and contracts.
- [ ] Enforce project isolation at every request, job, socket, worker, and commit boundary.
- [ ] Require host-authenticated execution context for every production plugin handler.
- [ ] Bind every production commit to a single-use host-issued invocation credential.
- [ ] Complete a deep security audit across every plugin and add regression tests for every finding.
- [ ] Keep operator access separate from end-user access.
- [ ] Treat destructive actions, migrations, releases, and permission changes as audited operations.
- [ ] Keep every test file under `tests/<area>/`.
- [ ] Run `go test ./...` after moving or adding tests.
- [ ] Do not run Docker or Compose during ordinary local development.
- [ ] Use SQLite for fast local tests when PostgreSQL-specific behavior is not required.
- [ ] Keep PostgreSQL, Compose, and multi-service checks explicit and opt-in.

## Backup and restore


## Final backend gates

- [ ] Project isolation holds across APIs, SQL, workers, jobs, sockets, and WASM.
- [ ] Permission revocation affects new work and in-flight commits.
- [ ] Recovery works after connection, worker, and deployment failures.
- [ ] Multi-architecture images are digest-pinned.
- [ ] Compose deployment and rollback are rehearsed.
- [ ] Production limits are based on measured capacity.
