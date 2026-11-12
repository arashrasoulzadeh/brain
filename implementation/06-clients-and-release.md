# Phase 06 — Clients and release

[Phase checklist](../phases.md) · [Architecture plan](../plan.md)

Use the app features, arena, Go function editor, and plugin contracts delivered in earlier phases. Start each slice with the failing test or rehearsal, make it pass, then simplify with the check still in place.

## 1. Finish the browser client

- [ ] Add shared protocol fixtures for sign-in, data, files, functions, subscriptions, matchmaking, gameplay, and reconnect.
- [ ] Write a browser smoke test for first-party cookie sessions, CSRF, an invalidation during a fetch, and safe retry after reconnect.
- [ ] Finish the typed JavaScript and TypeScript client against the versioned HTTP and socket contracts.
- [ ] Publish the browser quickstart with one app flow and the playable arena.

## 2. Finish the Flutter client

- [ ] Run the shared protocol fixtures in Flutter. Add tests for the native system-browser callback, secure token storage, and Flutter Web cookies.
- [ ] Finish one Dart client with native and web transport adapters. Keep session storage and platform callbacks behind those adapters.
- [ ] Verify the same sign-in, app, and match flows on a supported native target and Flutter Web. Document both setup paths.

## 3. Finish the native Android client

- [ ] Test the verified App Link callback and a reconnect after backgrounding with an expired access token.
- [ ] Finish the Kotlin client and its coroutine-friendly requests. Protect refresh credentials with Android Keystore-backed storage.
- [ ] Run the shared fixtures and play one match on an emulator or device. Publish the Kotlin quickstart.

## 4. Freeze the Zarvan integration contract

- [ ] Add transport-neutral fixtures for login, room admission, assignment, player input, authorized snapshots, and reconnect.
- [ ] Validate the fixtures against the backend with a Dart consumer and publish them with the versioned protocol.
- [ ] Document the compatibility and error behavior the future SDK can use in the Zarvan repository.

## 5. Publish the release images

- [ ] Test that the GitHub Actions release job is gated by passing tests on protected `main`, and that incomplete builds produce no release manifest.
- [ ] Build and publish the backend, runner, and compiler images to public GitHub Container Registry for `linux/amd64` and `linux/arm64`.
- [ ] Publish a manifest keyed by the full commit SHA with all three image digests. Keep publishing credentials scoped to the release job.
- [ ] Pull and smoke-test the published image set on both architectures.

## 6. Rehearse the manual Compose deployment

- [ ] Test an invalid manifest, concurrent deployment, migration failure, health failure, and successful rollback on a disposable server stack.
- [ ] Add the server-side deploy script. Acquire the lock, validate the requested SHA and Compose config, and pull all pinned images before starting the drain.
- [ ] Drain affected matches, stop the old release, run the migration when needed, and start the selected release with Compose using `--no-build`.
- [ ] Check backend readiness, PostgreSQL access, and the health of enabled workers. Record the deployed version after the checks pass.
- [ ] Verify rollback to the last healthy image set with a compatible schema. Document the recovery procedure for an incompatible migration.

## 7. Restore a production-shaped backup

- [ ] Test a fresh restore with a valid backup pair, a missing object version, and a stale backup alert.
- [ ] Schedule encrypted off-host backups every 12 hours. Pair a consistent PostgreSQL snapshot with a manifest of referenced object versions.
- [ ] Verify both parts before marking a backup complete. Alert when the latest verified backup is more than 18 hours old.
- [ ] Restore to a fresh Compose instance. Record recovery time, retention, and whether the 24-hour data-loss target was met.

## 8. Verify the release workload

- [ ] Record the test profile and start with one eight-player room on a 4-vCPU, 8-GB production-shaped host.
- [ ] Increase room count with representative app requests, uploads, jobs, and Eye export running. Include simulated latency and loss, then check RTT from intended player networks when available.
- [ ] Record p95 input-to-snapshot latency and tick delay, CPU, RAM, queue depth, disconnects, and outbound traffic. Set a conservative room cap from the measured headroom.
- [ ] Run the cross-project and plugin revocation tests against the packaged release. Publish the deployment and restore runbook with the client quickstarts.

## Exit gate

- [ ] The three clients pass the shared protocol suite and complete a match. Zarvan fixtures validate against the released protocol.
- [ ] A clean Compose rehearsal pulls the pinned image set and verifies deploy, rollback, and restore. Record the tested capacity and the operator's manual VPS launch steps.
