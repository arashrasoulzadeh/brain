# Restore runbook

Use this procedure when restoring a Brain deployment or rehearsing recovery. Restore PostgreSQL data and referenced object data as one recovery unit.

## Preconditions

1. Stop writes and preserve the failed release logs, deployment record, and recovery decision.
2. Identify the last verified PostgreSQL backup, object manifest, referenced object versions, and release SHA.
3. Confirm the restore target is a fresh isolated environment; never overwrite the only surviving copy.
4. Record the recovery start time and operator/change ID.

## Restore sequence

1. Provision the fresh PostgreSQL and object-storage targets.
2. Restore the verified PostgreSQL snapshot.
3. Restore every object version listed in the matching object manifest.
4. Verify database checksums/row counts and compare the restored object manifest.
5. Validate schema compatibility with the intended backend, compiler, runner, and plugin versions.
6. Start the environment with traffic disabled.
7. Run API readiness, PostgreSQL access, worker health, project-isolation, and plugin-revocation checks.
8. Record restore completion time, recovered release SHA, backup age, object verification result, and any missing data.
9. Re-enable traffic only after the recovery checks and an operator approval.

## Failure handling

If an object version is missing, the database/object pair is not complete. Keep traffic stopped, preserve evidence, and locate a matching backup or prepare an approved forward recovery. Do not silently substitute an unrelated object version.

For an incompatible migration, follow [`incompatible-migration-recovery.md`](incompatible-migration-recovery.md). Never run an unreviewed destructive down-migration.

## Recovery targets and audit

Compare the recorded recovery duration and recovered data-loss window with the 24-hour data-loss target and the documented recovery-time objective. Store the backup identifier, object-manifest hash, release SHA, verification output, and approval with the recovery record.

## Local development boundary

Do not run Docker or Compose during ordinary local development. Restore and PostgreSQL/object-storage rehearsals are explicit CI or release checks; local unit tests may use SQLite only where PostgreSQL-specific behavior is not under test.
