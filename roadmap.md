# Go multi-tenant mBaaS roadmap

This is the execution roadmap for the project. Work from top to bottom. Each checkbox should be small enough to implement and verify independently.

## Required delivery loop

For every checkbox:

- [ ] Write the focused failing test.
- [ ] Implement the smallest passing change.
- [ ] Run the focused and affected tests.
- [ ] Check off the item only after the tests pass.
- [ ] Commit the completed item in a focused commit.
- [ ] Push the commit to the configured remote.
- [ ] Confirm the remote contains the commit before starting the next item.

At the end of every phase, run its exit gate, commit any gate changes, push them, and record the verification result.

Every phase exit must also update:

- [ ] `CHANGELOG.md` with the completed capabilities and commit reference.
- [ ] `docs/developers.md` with implementation and testing notes.
- [ ] `docs/admins.md` with available operations, configuration, and limitations.

## Working rules

- [ ] Start each feature with one failing test.
- [ ] Make the smallest implementation pass the test.
- [ ] Refactor only after the focused test passes.
- [ ] Run unit tests for decisions and integration tests for real boundaries.
- [ ] Record the command and result when a step is complete.
- [ ] Keep development and production project IDs separate.
- [ ] Enforce project isolation at every request, job, socket, worker, and commit boundary.
- [ ] Require host-authenticated execution context for every production plugin handler; reject project IDs supplied only by scripts, payloads, or request constructors.
- [ ] Bind every production commit to a single-use host-issued invocation credential; reject tampered or replayed credentials.
- [ ] Run a deep security audit across every plugin for project binding, authorization, replay/idempotency, input/path validation, network targets, secret handling, and YAML boundaries; fix findings and add regression tests.
- [ ] Refactor every legacy plugin source file to the one-callable-responsibility-per-file law; update manifests, imports, tests, developer docs, and admin docs.
- [ ] Keep operator access separate from end-user access.
- [ ] Treat destructive actions, migrations, releases, and permission changes as audited operations.

## Plugin-first delivery law

- [ ] Keep the core host limited to shared infrastructure and contracts.
- [ ] Implement every product capability under `plugins/<plugin-id>/`.
- [ ] Give every plugin a manifest, grants, config schema, tests, developer notes, and administrator notes.
- [ ] Make every plugin dynamically versionable, enableable, disableable, promotable, rollbackable, drainable, and removable.
- [ ] Do not bypass the host permission broker or project boundary from plugin code.
- [ ] Audit every plugin for duplicate helpers and move shared behavior into the TypeScript SDK; keep plugin entrypoints small, readable, and product-focused.
- [ ] Enforce TypeScript lint with `npm run lint:ts`; lint, typecheck, and tests are required verification for every plugin change.
- [ ] Keep plugin code SOLID and DRY: shared generic helpers belong in the SDK, plugin code has one clear responsibility, and duplicate handlers/request/validation code is prohibited.
- [ ] Keep source declarations ordered: imports, types, constants, and file-level variables first; classes next; standalone functions after classes; class fields before methods. Use descriptive names.
- [ ] Split plugins into one focused product function per TypeScript handler file; keep `index.ts` as exports/bootstrap only, declare every file and import in YAML, and split multi-responsibility classes.
- [ ] **Permanent plugin decomposition law:** no new or modified plugin may introduce a multi-responsibility file, handler, class, helper, or facade; decomposition is required before feature implementation and must be verified by lint/tests.
- [ ] Gate every plugin source file, runtime import, and plugin-to-plugin import through `plugin.yaml`; the host must reject undeclared files, imports, and dependencies before execution.
- [ ] Use block YAML lists for every plugin source/import/dependency declaration; inline flow lists are prohibited.

## Test organization law

- [ ] Keep every test file for every layer under `tests/`.
- [ ] Group tests under `tests/<area>/` without placing test files beside implementation code.
- [ ] Run the complete `go test ./...` suite after moving or adding tests.

## Local development policy

- [ ] Do not run Docker or Docker Compose during ordinary local development.
- [ ] Use SQLite for fast local tests when PostgreSQL-specific behavior is not required.
- [ ] Keep PostgreSQL, Compose, and multi-service checks explicit and opt-in.
- [ ] Use PostgreSQL integration tests for roles, row-level security, migrations, locks, and database-specific transactions.

## Phase 06 — Clients and release

Goal: finish the supported clients, publish the protocol, and rehearse production deployment and recovery.

### Flutter client

- [ ] Test secure token storage.

### Android client

- [ ] Test the verified App Link callback.
- [ ] Test reconnect after backgrounding.
- [ ] Test expired access-token recovery.
- [ ] Finish the Kotlin client.
- [ ] Add coroutine-friendly requests.
- [ ] Store refresh credentials with Android Keystore.
- [ ] Run shared fixtures on an emulator or device.
- [ ] Play one match.
- [ ] Publish the Kotlin quickstart.

### Release images and CI

- [ ] Build the compiler image for amd64 and arm64.
- [ ] Build the runner image for amd64 and arm64.
- [ ] Publish images to public GHCR.
- [ ] Publish the full commit SHA as the release version.
- [ ] Publish all image digests in one manifest.
- [ ] Scope publish credentials to the release job.
- [ ] Pull and smoke-test both architectures.

### Manual Compose deployment

- [ ] Pull all pinned images before draining.
- [ ] Pause new match assignments.
- [ ] Drain active rooms for the configured window.
- [ ] Close remaining sockets explicitly.
- [ ] Run the one-shot migration when required.
- [ ] Start with Compose `--no-build`.
- [ ] Check API readiness.
- [ ] Check PostgreSQL access.
- [ ] Check enabled worker health.
- [ ] Record the deployed version.
- [ ] Keep the last healthy image set.
- [ ] Test rollback with a compatible schema.
- [ ] Document incompatible-migration recovery.

### Backup and restore

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

### Capacity and final gate

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
- [ ] Publish browser, Flutter, Android, and Zarvan quickstarts.
- [ ] Verify all three clients complete a match.
- [ ] Verify deploy, rollback, and restore in a clean environment.

## Final release gates

- [ ] The dashboard covers routine project, plugin, database, function, workflow, file, and live-session operations.
- [ ] Project isolation holds across APIs, SQL, workers, jobs, sockets, and WASM.
- [ ] Permission revocation affects new work and in-flight commits.
- [ ] The arena and Tic-tac-toe work through supported clients.
- [ ] Recovery works after connection, worker, and deployment failures.
- [ ] Multi-architecture images are digest-pinned.
- [ ] Compose deployment and rollback are rehearsed.
- [ ] PostgreSQL and object data restore together.
- [ ] Production limits are based on measured capacity.
