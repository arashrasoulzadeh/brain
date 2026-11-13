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
- [ ] Keep operator access separate from end-user access.
- [ ] Treat destructive actions, migrations, releases, and permission changes as audited operations.

## Plugin-first delivery law

- [ ] Keep the core host limited to shared infrastructure and contracts.
- [ ] Implement every product capability under `plugins/<plugin-id>/`.
- [ ] Give every plugin a manifest, grants, config schema, tests, developer notes, and administrator notes.
- [ ] Make every plugin dynamically versionable, enableable, disableable, promotable, rollbackable, drainable, and removable.
- [ ] Do not bypass the host permission broker or project boundary from plugin code.

## Test organization law

- [ ] Keep every test file for every layer under `tests/`.
- [ ] Group tests under `tests/<area>/` without placing test files beside implementation code.
- [ ] Run the complete `go test ./...` suite after moving or adding tests.

## Local development policy

- [ ] Do not run Docker or Docker Compose during ordinary local development.
- [ ] Use SQLite for fast local tests when PostgreSQL-specific behavior is not required.
- [ ] Keep PostgreSQL, Compose, and multi-service checks explicit and opt-in.
- [ ] Use PostgreSQL integration tests for roles, row-level security, migrations, locks, and database-specific transactions.

## Phase 00 — Local foundation

Goal: boot a small Go service with PostgreSQL, migrations, Compose, and CI.

### Host

- [x] Create the Go module.
- [x] Add the HTTP command entry point.
- [x] Add a testable application package.
- [x] Add a liveness endpoint.
- [x] Add a liveness test.
- [x] Add required configuration fields.
- [x] Reject missing configuration at startup.
- [x] Redact secrets from startup errors.

### PostgreSQL

- [x] Add a PostgreSQL connection owner.
- [x] Add bounded connection startup.
- [x] Add database readiness checks.
- [x] Fail readiness when PostgreSQL is unavailable.
- [x] Recover readiness when PostgreSQL returns.
- [x] Add the first platform migration.
- [x] Use a migration-only database credential.
- [x] Prevent the runtime role from running DDL.
- [x] Run migrations twice and verify idempotency.

### Compose and CI

- [x] Add the smallest local Compose file.
- [x] Keep the PostgreSQL port private.
- [x] Add isolated test settings and data.
- [x] Add a clean-start smoke test.
- [x] Add a clean-stop smoke test.
- [x] Repeat the Compose smoke test from scratch.
- [x] Add formatting checks.
- [x] Add Go tests to CI.
- [x] Add the Compose smoke test to CI.
- [x] Add the backend image build check for amd64 and arm64.

### Phase 00 gate

- [x] A clean checkout passes tests locally.
- [x] A clean checkout passes tests in GitHub Actions.
- [x] Readiness distinguishes process health from database health.
- [x] Migration and runtime credentials are separate.

## Phase 01 — Projects and plugin host

Goal: create isolated projects and run one plugin through both supported worker paths.

### Projects and operator access

- [x] Add a project record with ID and status.
- [x] Resolve the project at the request entry point.
- [x] Reject unknown projects.
- [x] Reject inactive projects.
- [x] Add operator bootstrap through the CLI.
- [x] Add operator sign-in.
- [x] Add MFA verification.
- [x] Reject invalid MFA.
- [x] Add operator recovery.
- [x] Add CSRF protection.
- [x] Reject an ordinary user attempting an operator action.
- [x] Audit operator sign-in and recovery.

### Plugin registration

- [x] Define the plugin manifest.
- [x] Add stable plugin ID and immutable version.
- [x] Add host API compatibility.
- [x] Add dependency declarations.
- [x] Add capability declarations.
- [x] Add extension-point declarations.
- [x] Add configuration schema validation.
- [x] Reject duplicate registrations.
- [x] Reject incompatible protocol versions.
- [x] Store approved artifact or endpoint information.

### Installation and grants

- [x] Install a plugin for one project.
- [x] Start every installation disabled.
- [x] Add requested capability display.
- [x] Grant a named capability.
- [x] Bind grants to project resources.
- [x] Reject enablement with missing grants.
- [x] Reject enablement when readiness fails.
- [x] Enable the same plugin independently in two projects.
- [x] Audit installation, grants, and enablement.

### Permission broker

- [x] Define typed host calls.
- [x] Check actor permission.
- [x] Check project installation grant.
- [x] Check matchmaking scope when present.
- [x] Check resource scope.
- [x] Issue short-lived invocation credentials.
- [x] Expire invocation credentials.
- [x] Recheck grants before commit.
- [x] Return bounded errors when a worker is unavailable.

### Workers

- [x] Add a bundled worker fixture.
- [x] Add the private worker protocol.
- [x] Add worker startup supervision.
- [x] Restrict worker filesystem access.
- [x] Restrict worker network access.
- [x] Restrict worker CPU and memory.
- [x] Run the contract fixture through a bundled worker.
- [x] Add an external worker endpoint.
- [x] Authenticate the external worker handshake.
- [x] Bind external identity to the project installation.
- [x] Run the same fixture through the external worker.

### Dashboard and first event

- [x] Bootstrap the dashboard through the CLI.
- [x] Add project creation form.
- [x] Add plugin installation form.
- [x] Add grant approval form.
- [x] Add enablement form.
- [x] Show worker readiness separately from enablement.
- [x] Add guest sign-in through a plugin.
- [x] Define the analytics event contract.
- [x] Reject unapproved event fields.
- [x] Capture one event in a local test sink.

### Phase 01 gate

- [x] Create two projects through the dashboard.
- [x] Configure the same plugin independently in both.
- [x] Pass the same contract tests through both worker paths.
- [x] Prove cross-project isolation.
- [x] Prove revoked grants reject an in-flight effect.
- [x] Issue one project-scoped guest session.

## Phase 02 — First co-op arena

Goal: complete a private multiplayer match with matchmaking, realtime updates, and durable results.

### Protocol and room model

- [x] Define versioned HTTP request envelopes.
- [x] Define versioned WebSocket messages.
- [x] Define request IDs and error codes.
- [x] Define retry and conflict behavior.
- [x] Define project-scoped room IDs.
- [x] Define player slots and membership.
- [x] Define room lifecycle states.
- [x] Define durable match results.

### Private rooms

- [x] Add room creation.
- [x] Require project membership for room access.
- [x] Add invite creation.
- [x] Make invites single-use.
- [x] Reject an invite from another project.
- [x] Reject a non-member socket.
- [x] Add player join and leave events.
- [x] Persist room membership.

### FIFO matchmaking

- [x] Add a matchmaking configuration.
- [x] Add a project-scoped queue.
- [x] Add a FIFO ticket.
- [x] Prevent duplicate active tickets.
- [x] Match compatible tickets.
- [x] Reject tickets for disabled scopes.
- [x] Add assignment expiry.
- [x] Requeue failed assignments.
- [x] Record assignment decisions.

### Realtime gameplay

- [x] Add a WebSocket connection endpoint.
- [x] Issue single-use connection tickets.
- [x] Validate origin and project.
- [x] Add heartbeat handling.
- [x] Add bounded frame sizes.
- [x] Add bounded send queues.
- [x] Broadcast authorized snapshots.
- [x] Reject unauthorized commands.
- [x] Reconnect an existing player.
- [x] Replay missed durable events.

### Arena and Eye analytics

- [x] Implement a minimal arena state machine.
- [x] Add a start command.
- [x] Add a player action.
- [x] Add a result command.
- [x] Persist one completed result.
- [x] Export one validated Eye event.
- [x] Add an eight-player fixture.
- [x] Measure room latency and memory.

### Phase 02 gate

- [x] Start a match in two browsers.
- [x] Recover one dropped connection.
- [x] Reject one denied plugin action.
- [x] Persist one durable result.
- [x] Exercise private rooms.
- [x] Exercise FIFO matchmaking.
- [x] Record the eight-player test result.

## Phase 03 — App data and authentication

Goal: provide project-scoped identity, PostgreSQL APIs, files, subscriptions, and dashboard controls.

### Identity and sessions

- [x] Add project-scoped user IDs.
- [x] Add username uniqueness per project.
- [x] Add Argon2id password hashing.
- [x] Add password registration without email delivery.
- [x] Add password sign-in.
- [x] Add per-account rate limits.
- [x] Add per-project rate limits.
- [x] Add session token digests.
- [x] Add session expiry.
- [x] Add session rotation.
- [x] Revoke refresh tokens after reuse.
- [x] Add logout.
- [x] Add operator session revocation.
- [x] Reject a session in another project.

### Invitations and identity providers

- [x] Add one-time invitation links.
- [x] Add invitation codes.
- [x] Reject invitation replay.
- [x] Reject expired invitations.
- [x] Add guest accounts.
- [x] Add guest upgrade without changing user ID.
- [x] Detect identity-link conflicts.
- [x] Add Google OIDC configuration.
- [x] Validate issuer, audience, expiry, signature, and nonce.
- [x] Require recent authentication for linking.
- [x] Add native authorization-code exchange.
- [x] Add PKCE verification.
- [x] Reject callback state mismatch.

### Email, SMS, and recovery

- [x] Add a local test delivery inbox.
- [x] Add email OTP.
- [x] Add SMS OTP.
- [x] Normalize phone numbers to E.164.
- [x] Bind challenges to project, destination, and purpose.
- [x] Store keyed code digests.
- [x] Consume codes atomically.
- [x] Expire codes quickly.
- [x] Limit sends and guesses.
- [x] Add email recovery.
- [x] Revoke sessions after password reset.
- [x] Hide account enumeration signals.
- [x] Reject fake delivery configuration in production.

### PostgreSQL tables and API exposure

- [x] Create one schema per project.
- [x] Create scoped database roles.
- [x] Test migration and runtime roles separately.
- [x] Bind caller context transaction-locally.
- [x] Add default-deny row policies.
- [x] Reject cross-project reads and writes.
- [x] Reject pooled caller-context reuse.
- [x] Add table exposure records.
- [x] Require a supported primary key.
- [x] Reject unsupported types.
- [x] Hide unapproved columns.
- [x] Allow only approved filters.
- [x] Add cursor pagination.
- [x] Validate values and identifiers.
- [x] Add record versions.
- [x] Reject stale writes.
- [x] Make retried creates idempotent.
- [x] Add guided schema changes.
- [x] Add SQL migration preview.
- [x] Require recent authentication for destructive schema actions.
- [x] Audit schema publication.

### Change capture and subscriptions

- [x] Add transactional row capture.
- [x] Capture approved old and new authorization scopes.
- [x] Capture API changes.
- [x] Capture direct SQL changes.
- [x] Do not capture rolled-back changes.
- [x] Add subscription tickets.
- [x] Make tickets single-use.
- [x] Validate browser origins.
- [x] Authorize each subscription.
- [x] Handle owner-transfer invalidations.
- [x] Handle writes during initial fetch.
- [x] Handle slow consumers.
- [x] Reconnect with missed events.

### Files

- [x] Configure a private S3-compatible bucket.
- [x] Namespace object keys by project.
- [x] Add pending upload metadata.
- [x] Reserve file quota.
- [x] Stream authorized uploads.
- [x] Publish metadata after object success.
- [x] Reject unauthorized reads and deletes.
- [x] Reject cross-project object IDs.
- [x] Add upload size limits.
- [x] Reconcile interrupted uploads.
- [x] Reconcile orphaned objects.
- [x] Serve protected downloads as attachments.
- [x] Require explicit publication for public assets.

### Dashboard and gate

- [x] Add auth settings pages.
- [x] Add signup settings.
- [x] Add record browsing.
- [x] Add record editing.
- [x] Add table management.
- [x] Add file management.
- [x] Add audited form actions.
- [x] Add the browser quickstart.
- [x] Run the full suite against local Compose.
- [x] Prove API and direct SQL isolation.
- [x] Re-run the Phase 02 guest co-op flow.

## Phase 04 — Go functions and turn-based rooms

Goal: compile dashboard-authored Go to WASM, execute it safely, run durable triggers, and complete Tic-tac-toe.

### Offline compiler

- [x] Define the approved SDK imports.
- [ ] Add an entry-package fixture.
- [x] Add a project-local helper fixture.
- [x] Compile the fixture to `wasip1/wasm`.
- [x] Pin the Go toolchain.
- [x] Generate platform module settings.
- [x] Vendor approved dependencies.
- [x] Reject missing imports.
- [x] Reject unsupported imports.
- [x] Reject cross-project source access.
- [x] Reject workspace and toolchain overrides.
- [x] Enforce source and artifact limits.
- [x] Use a disposable compiler workspace.
- [x] Run without credentials or network access.
- [x] Bound compiler CPU, memory, processes, and time.

### WASM runner

- [x] Define the typed invocation request.
- [x] Define the typed invocation result.
- [x] Reject malformed output.
- [x] Reject extra results.
- [x] Cap diagnostics.
- [x] Reject unexpected WASI imports.
- [x] Run each invocation in a fresh wazero instance.
- [x] Allow only approved WASI interfaces.
- [x] Deny filesystem access.
- [x] Deny network access.
- [x] Deny environment credentials.
- [x] Test infinite-loop timeouts.
- [x] Test linear-memory limits.
- [x] Test startup timeouts.
- [x] Test sandbox escape attempts.
- [x] Confirm failed invocations produce no writes.

### Publication and invocation

- [x] Add draft saving.
- [x] Add shared-package compilation.
- [x] Add source diagnostics.
- [x] Add fixture execution.
- [x] Add publication.
- [x] Add rollback.
- [x] Keep the active version after a failed draft.
- [x] Pin artifact digest and toolchain.
- [x] Add direct HTTPS invocation.
- [x] Load only declared inputs.
- [x] Validate write intents.
- [x] Validate stale input versions.
- [x] Commit effects and idempotency records together.
- [x] Retry safely after a lost response.

### Triggers and workflows

- [x] Add scheduled functions.
- [x] Claim schedules safely under concurrency.
- [x] Handle duplicate UTC occurrences.
- [x] Enqueue the latest missed occurrence.
- [x] Pin the function version at enqueue.
- [x] Add record-change triggers.
- [x] Trigger only committed writes.
- [x] Suppress recursive worker writes.
- [x] Add short job leases.
- [x] Recheck grants at job commit.
- [x] Cap retries.
- [x] Add workflow graph validation.
- [x] Add conditions.
- [x] Add approvals.
- [x] Add workflow retries and idempotency.
- [x] Hide secrets from run status.

### Turn-based rooms

- [x] Add Tic-tac-toe initialization.
- [x] Add legal move validation.
- [x] Reject out-of-turn moves.
- [x] Detect wins.
- [x] Detect draws.
- [x] Publish the rules through the compiler.
- [x] Give rules only authorized state and commands.
- [x] Add duplicate command protection.
- [x] Reject command payload mismatches.
- [x] Handle concurrent moves.
- [ ] Reject malformed rule results.
- [ ] Reconnect after restart.
- [ ] Add hidden-state player views.
- [ ] Add timed-out moves.
- [ ] Reconstruct history from committed outcomes.

### Phase 04 gate

- [ ] Pass compiler and runner isolation tests in Compose.
- [ ] Publish and roll back a shared-helper artifact.
- [ ] Run direct, scheduled, and record-change invocations.
- [ ] Verify crash recovery and grant revocation.
- [ ] Play a complete two-client Tic-tac-toe match.
- [ ] Record compiler time, turn latency, and memory use.

## Phase 05 — Plugin lifecycle, admin UI, and workflows

Goal: safely promote, roll back, disable, drain, uninstall, and extend plugin-controlled UI and workflows.

### Version lifecycle

- [ ] Promote one plugin version for one project.
- [ ] Validate config before promotion.
- [ ] Validate readiness before promotion.
- [ ] Validate requested grants.
- [ ] Validate schema compatibility.
- [ ] Preserve versions pinned by active rooms.
- [ ] Cancel incompatible unmatched tickets.
- [ ] Roll back to a retained version.
- [ ] Reject missing or incompatible rollback versions.
- [ ] Keep current grants during rollback.
- [ ] Require approval for new capabilities.

### Disable and drain

- [ ] Calculate transitive dependents.
- [ ] Add a disable preview revision.
- [ ] Reject stale disable previews.
- [ ] Disable one matchmaking scope independently.
- [ ] Block new calls, jobs, subscriptions, and assignments.
- [ ] Cancel affected queued matches.
- [ ] Persist drain scope and deadline.
- [ ] Use a five-minute default drain window.
- [ ] Allow current participants to reconnect.
- [ ] Reject new room admissions during drain.
- [ ] Pause turn rooms at the last committed move.
- [ ] Interrupt action rooms at expiry.
- [ ] Add immediate stop for security revocation.
- [ ] Recheck lifecycle revision at assignment.
- [ ] Recheck lifecycle revision at result commit.
- [ ] Resume paused rooms only with permitted pinned versions.

### Uninstall and registry cleanup

- [ ] Wait for active work before project uninstall.
- [ ] Revoke project grants.
- [ ] Preserve durable plugin data.
- [ ] Audit project uninstall.
- [ ] Reject server unregister while references remain.
- [ ] Remove a server version only after all references are gone.
- [ ] Keep data deletion as a separate action.
- [ ] Keep CLI recovery available.

### Admin UI extensions

- [ ] Add one declarative admin extension point.
- [ ] Render a host-owned fixture view.
- [ ] Add two view contributions.
- [ ] Activate one replacement per named view and scope.
- [ ] Reject a second replacement.
- [ ] Reject cross-project submissions.
- [ ] Route actions through host authorization.
- [ ] Add CSRF checks.
- [ ] Audit view actions.
- [ ] Add one custom TypeScript view.
- [ ] Serve it from a separate origin.
- [ ] Embed it in a restricted frame.
- [ ] Validate typed bridge messages.
- [ ] Reject malformed or unbound messages.
- [ ] Keep operator tokens in the trusted UI.
- [ ] Keep sensitive confirmations in the trusted UI.

### Workflow extensions

- [ ] Add one registered project workflow action.
- [ ] Replace one named workflow step.
- [ ] Reject revoked workflow grants.
- [ ] Reject invalid workflow graphs.
- [ ] Ensure retried effects commit once.
- [ ] Add named triggers, conditions, actions, and approvals.
- [ ] Prevent project plugins from registering release steps.
- [ ] Add explicitly granted deployment plugins.
- [ ] Add test, build, preflight, and post-deploy hooks.
- [ ] Keep publish credentials outside plugin processes.
- [ ] Keep the Docker socket outside plugin processes.
- [ ] Audit deployment workflow steps.

### Phase 05 gate

- [ ] Promote and roll back a plugin from the dashboard.
- [ ] Preview and apply a dependency disable.
- [ ] Demonstrate drain, resume, and immediate stop.
- [ ] Uninstall a plugin while preserving durable records.
- [ ] Replace one admin view declaratively.
- [ ] Replace one admin view with a custom frame.
- [ ] Run one project workflow extension.
- [ ] Run one granted server workflow extension.
- [ ] Verify permission and credential separation.

## Phase 06 — Clients and release

Goal: finish the supported clients, publish the protocol, and rehearse production deployment and recovery.

### Browser client

- [ ] Add shared sign-in fixtures.
- [ ] Add shared data fixtures.
- [ ] Add shared file fixtures.
- [ ] Add shared function fixtures.
- [ ] Add shared subscription fixtures.
- [ ] Add shared matchmaking fixtures.
- [ ] Add shared gameplay fixtures.
- [ ] Add shared reconnect fixtures.
- [ ] Test first-party cookie sessions.
- [ ] Test CSRF protection.
- [ ] Test invalidation during a fetch.
- [ ] Test safe retry after reconnect.
- [ ] Finish the typed JavaScript client.
- [ ] Publish the browser quickstart.
- [ ] Publish the playable arena.

### Flutter client

- [ ] Run shared fixtures in Flutter.
- [ ] Test the native system-browser callback.
- [ ] Test secure token storage.
- [ ] Test Flutter Web cookies.
- [ ] Add native transport adapter.
- [ ] Add web transport adapter.
- [ ] Hide platform storage behind adapters.
- [ ] Complete sign-in on a native target.
- [ ] Complete sign-in on Flutter Web.
- [ ] Complete one app flow.
- [ ] Complete one match flow.
- [ ] Document both setup paths.

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

### Zarvan protocol

- [ ] Define transport-neutral login fixtures.
- [ ] Define room-admission fixtures.
- [ ] Define assignment fixtures.
- [ ] Define player-input fixtures.
- [ ] Define authorized-snapshot fixtures.
- [ ] Define reconnect fixtures.
- [ ] Validate fixtures with a Dart consumer.
- [ ] Publish fixtures with the protocol version.
- [ ] Document compatibility and error behavior.

### Release images and CI

- [ ] Gate releases on protected `main`.
- [ ] Reject incomplete builds.
- [ ] Build the backend image for amd64 and arm64.
- [ ] Build the compiler image for amd64 and arm64.
- [ ] Build the runner image for amd64 and arm64.
- [ ] Publish images to public GHCR.
- [ ] Publish the full commit SHA as the release version.
- [ ] Publish all image digests in one manifest.
- [ ] Scope publish credentials to the release job.
- [ ] Pull and smoke-test both architectures.

### Manual Compose deployment

- [ ] Add a deployment lock.
- [ ] Validate the requested release SHA.
- [ ] Validate the release manifest.
- [ ] Validate the Compose configuration.
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
