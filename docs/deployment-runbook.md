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

## Rollback

Use the last healthy, schema-compatible image set. Re-run readiness, database, worker-health, and version checks after rollback. Never perform an automatic destructive down-migration. For an incompatible schema, follow [`incompatible-migration-recovery.md`](incompatible-migration-recovery.md).

## Verification and audit

Record the operator, change ID, release SHA, image digests, migration result, pause/drain timestamps, readiness result, worker health, and rollback decision. Preserve failed deployment logs and backup/object-manifest references.

## Local development boundary

Do not run Docker or Compose during ordinary local development. Use the in-process and SQLite-compatible checks locally; run Compose, PostgreSQL, image, and packaged-release checks through explicit CI or release workflows.
