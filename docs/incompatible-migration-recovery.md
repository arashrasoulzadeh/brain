# Incompatible migration recovery

This procedure is for a release whose database migration is not compatible with the last healthy image set. Do not force a binary rollback against an incompatible schema.

## Operator procedure

1. Stop new assignments and start the configured room drain.
2. Close remaining sockets after the drain deadline and record the failed release SHA.
3. Keep the failed deployment's logs and migration result for incident review; do not delete its volumes or backup artifacts.
4. Verify the latest known-good PostgreSQL backup and its referenced object-version manifest as a pair.
5. Restore both database and object data into a fresh isolated Compose environment.
6. Start the last healthy image set with `--no-build` and run API, PostgreSQL, and worker health checks.
7. If the restored environment is healthy, switch traffic to it and record the restored release SHA.
8. If restore is unavailable, keep the service stopped and prepare a forward-fix migration. Never run an unreviewed destructive down-migration in production.

## Recovery decision

- Compatible schema: select the retained healthy manifest and roll back.
- Incompatible schema with verified backup: restore the complete backup pair, then start the retained manifest.
- Incompatible schema without a verified backup: keep traffic stopped, preserve evidence, and require an approved forward migration or data-recovery decision.

All recovery actions require an operator audit record, and secrets must remain in the server environment rather than in manifests or logs.
