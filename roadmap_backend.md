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

## Release images and CI

- [ ] Build the compiler image for amd64 and arm64.
- [ ] Build the runner image for amd64 and arm64.
- [ ] Publish images to public GHCR.
- [ ] Publish the full commit SHA as the release version.
- [ ] Publish all image digests in one manifest.
- [ ] Scope publish credentials to the release job.
- [ ] Pull and smoke-test both architectures.

## Manual Compose deployment

- [x] Enforce pulling all pinned images before draining.
- [ ] Pause new match assignments.
- [ ] Drain active rooms for the configured window.
- [x] Close remaining sockets explicitly.
- [x] Run the one-shot migration when required.
- [x] Start with Compose `--no-build`.
- [ ] Check API readiness.
- [ ] Check PostgreSQL access.
- [ ] Check enabled worker health.
- [ ] Record the deployed version.
- [ ] Keep the last healthy image set.
- [ ] Test rollback with a compatible schema.
- [ ] Document incompatible-migration recovery.

## Backup and restore

- [ ] Schedule encrypted off-host backups every 12 hours.
- [ ] Snapshot PostgreSQL consistently.
- [ ] Record referenced object versions.
- [ ] Verify the database backup.
- [ ] Verify the object manifest.
- [ ] Alert when the latest verified backup is over 18 hours old.
- [ ] Test a fresh restore.
- [ ] Test a missing object version.
- [ ] Restore into a fresh Compose instance.
- [ ] Record recovery time and retention.
- [ ] Check the 24-hour data-loss target.

## Capacity and backend release gate

- [ ] Record the reference test profile.
- [ ] Start with one eight-player room.
- [ ] Add representative app requests.
- [ ] Add uploads and background jobs.
- [ ] Add Eye export activity.
- [ ] Simulate latency and packet loss.
- [ ] Measure p95 input-to-snapshot latency.
- [ ] Measure tick delay.
- [ ] Measure CPU, RAM, queue depth, disconnects, and traffic.
- [ ] Set a conservative room cap.
- [ ] Run cross-project isolation tests on the packaged release.
- [ ] Run plugin revocation tests on the packaged release.
- [ ] Publish the deployment runbook.
- [ ] Publish the restore runbook.
- [ ] Verify deploy, rollback, and restore in a clean environment.

## Final backend gates

- [ ] Project isolation holds across APIs, SQL, workers, jobs, sockets, and WASM.
- [ ] Permission revocation affects new work and in-flight commits.
- [ ] Recovery works after connection, worker, and deployment failures.
- [ ] Multi-architecture images are digest-pinned.
- [ ] Compose deployment and rollback are rehearsed.
- [ ] PostgreSQL and object data restore together.
- [ ] Production limits are based on measured capacity.
