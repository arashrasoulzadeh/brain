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

All feature capabilities are managed as separate TypeScript plugins. Review the host-generated YAML documentation, requested grants, configuration, readiness, lifecycle, rollback, and removal behavior before enabling a plugin. The catalog is in [`plugins/README.md`](../plugins/README.md).

Phase 01 operator notes: dashboard lifecycle actions are host-authorized; worker readiness is distinct from enablement; guest sessions are project-scoped; and analytics rejects fields not declared by the plugin contract. Local verification does not require Docker.

Plugins communicate through host-controlled local sockets or authenticated external services. They do not receive unrestricted database, operator-token, filesystem, or network access.

## Current limitations

The dashboard, operator MFA, recovery, worker supervision, external workers, and audit storage are not implemented yet. Do not treat the current HTTP host as production-ready.

## Operational policy

- Keep development and production project IDs separate.
- Use migration credentials only for schema changes.
- Keep runtime credentials unable to run DDL.
- Use Docker/Compose only for explicit integration or release checks.
- Record the commit and test result for every completed roadmap item.

At the end of every phase, this guide must be updated with new operator capabilities, configuration changes, migration notes, security considerations, limitations, and rollback guidance.

The changelog is newest first and separated into one auditable entry per commit.

## Phase 00 completed

The foundation is complete and verified. Administrators now have documented separation between runtime and migration credentials, runtime DDL protection, a platform migration contract, isolated local test settings, and an opt-in Compose smoke test. CI verifies formatting, Go tests, and multi-architecture image builds. Project/plugin administration remains the next implementation phase.
