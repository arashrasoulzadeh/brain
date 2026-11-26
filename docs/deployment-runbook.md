# Deployment runbook

This runbook is for an operator deploying a validated release with Compose. Do not run these steps against an unreviewed image or an unverified database backup.

## Preconditions

1. Confirm the release uses a full 40-character commit SHA.
2. Confirm backend, compiler, and runner images are digest-pinned.
3. Confirm the release manifest records packaged cross-project isolation and plugin-revocation verification.
4. Confirm the latest verified PostgreSQL backup and object manifest are within the recovery policy.
5. Confirm the deployment lock and an approved change record.

## Deployment sequence

1. Acquire the deployment lock.
2. Validate release, image digests, Compose configuration, and schema compatibility.
3. Pull all pinned images before changing live traffic.
4. Pause new match assignments.
5. Drain active rooms for the configured window.
6. Close remaining sockets after the drain deadline.
7. Run the one-shot migration when required, using migration credentials only.
8. Start Compose with `--no-build`.
9. Check API readiness, PostgreSQL access, and every enabled worker.
10. Record the deployed release SHA and retain the last healthy image set.

For a complete clean-environment rehearsal with verified recovery artifacts, run `scripts/clean-environment-rehearsal.sh` with `COMPOSE_FILE`, `BACKUP_INPUT`, `RESTORE_DSN`, `OBJECT_MANIFEST`, and `BRAIN_BIN` configured.

## Rollback

Use the last healthy, schema-compatible image set. Re-run readiness, database, worker-health, and version checks after rollback. Never perform an automatic destructive down-migration. For an incompatible schema, follow [`incompatible-migration-recovery.md`](incompatible-migration-recovery.md).

## Verification and audit

Record the operator, change ID, release SHA, image digests, migration result, pause/drain timestamps, readiness result, worker health, and rollback decision. Preserve failed deployment logs and backup/object-manifest references.

## Local development boundary

Do not run Docker or Compose during ordinary local development. Use the in-process and SQLite-compatible checks locally; run Compose, PostgreSQL, image, and packaged-release checks through explicit CI or release workflows.

## Application startup safety

The default backend process requires operator credentials, an explicit durable
`PROJECTS_FILE` SQLite path, valid plugin manifests and sources, and a working
Bubblewrap/Node/TypeScript runtime. Mount a persistent writable data directory
owned by the image's non-root user. The host currently uses SQLite for project
and plugin state: configuring PostgreSQL DSNs now fails startup rather than
silently ignoring them. PostgreSQL application storage wiring remains separate
release work.

The runtime must support unprivileged user, PID, IPC, UTS and network namespaces
for Bubblewrap. Do not grant the production container privileged mode or mount
host directories to bypass this requirement. The startup probe verifies a real
sandboxed TypeScript invocation; if it fails, do not route traffic to the host.
The runner image requires the launcher inputs documented in its entrypoint.

Use `/livez` for liveness and `/readyz` for traffic readiness. Readiness checks
both SQLite connections and enabled non-draining workers. The explicit
`health-only` command used by the legacy Compose fixture exposes liveness but
never readiness or admin routes; it is not an application deployment. CI has a
separate packaged runtime smoke test with disposable, network-isolated test
containers. A production namespace configuration must still be verified on the
target deployment host.
