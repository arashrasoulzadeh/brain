# Administrator guide

## Current capabilities

The current implementation provides the foundations for project and plugin administration:

- Project status resolution.
- Plugin registration with host API compatibility.
- Per-project plugin installation.
- Capability grants.
- Enablement only after all requested grants are present.
- Project-scoped authorization.
- Runtime versus migration credential separation.

## Current limitations

The dashboard, operator MFA, recovery, worker supervision, external workers, and audit storage are not implemented yet. Do not treat the current HTTP host as production-ready.

## Operational policy

- Keep development and production project IDs separate.
- Use migration credentials only for schema changes.
- Keep runtime credentials unable to run DDL.
- Use Docker/Compose only for explicit integration or release checks.
- Record the commit and test result for every completed roadmap item.

At the end of every phase, this guide must be updated with new operator capabilities, configuration changes, migration notes, security considerations, limitations, and rollback guidance.
