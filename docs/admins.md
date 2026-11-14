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

## Phase 02 completed

Administrators can now operate project-scoped realtime rooms with origin checks, single-use connection tickets, heartbeat expiry, bounded transport queues, member-only snapshots and commands, reconnects, replayable room events, FIFO matchmaking, and durable arena results. Eye exports are validated before delivery, and the eight-player profile has a repeatable gate fixture. Local verification uses `npm run typecheck && npm run test:ts` and `GOCACHE=/tmp/brain-go-cache go test ./...`; Docker remains opt-in for CI or release checks.

## Phase 03 completed

Administrators can review project-scoped identity and session controls, OIDC/PKCE and recovery configuration, table exposure and schema-preview audits, subscription tickets/origin grants, private bucket and quota settings, upload reconciliation, protected downloads, explicit public-asset publication, and dashboard record/table/file actions. API and direct-SQL isolation plus the guest co-op rerun are covered by gate fixtures. CI verifies Go formatting/tests, the full TypeScript suite, image builds, and Compose smoke checks; local development remains Docker-free. Review plugin grants, recent-auth requirements, origin allowlists, and audit records before enabling destructive or realtime operations.
## Plugin release gate

Plugin releases are blocked unless TypeScript lint, typecheck, and the complete test suite pass. Shared SDK changes should be reviewed as platform dependencies because they can affect multiple plugins.

Administrators can review `source_files`, `imports`, `plugin_imports`, and `dependencies` in each YAML manifest. Undeclared files, module imports, or plugin relationships are rejected before execution.

Plugin updates must preserve the one-callable-responsibility-per-file rule. The host review must reject undocumented legacy multi-responsibility modules and verify that all split files are declared in the plugin manifest.

Each handler file is independently versionable and replaceable, while `index.ts` remains the plugin export boundary. Review the YAML file list when approving a plugin update.

Project isolation is enforced from the host execution context. A plugin or client cannot select a project by changing a request payload or constructor argument at runtime.

Invocation credentials are single-use and project-bound. Tampered or replayed credentials are rejected at commit time.

Security review covers plugin network targets, authentication comparisons, project isolation, replay protection, and YAML execution boundaries before release. Authentication review also verifies project-scoped state, bounded OTP/recovery lifetimes, single-use credentials, and rejection of malformed or expired tokens.

Promotion is project-scoped and audited. Only an installed, enabled version can become active; pinned work remains on its pinned version during later promotion or rollback.

Before promotion, all manifest-declared configuration fields must be present and match their declared types.

The worker readiness state is checked explicitly before promotion; an unready version cannot become active.

## Phase 04 completed

Function compilation and execution are now covered by the hosted Compose isolation gate. Operators can rely on bounded compiler/runner execution, declared imports, denied filesystem/network/credential access, timeout and memory limits, safe retries, scheduled and record-change triggers, shared-helper rollback, grant rechecks, crash recovery, and deterministic Tic-tac-toe turn handling. CI run `43427610513` passed the isolation profile. Do not run Docker during ordinary local development; the normal local checks are TypeScript lint/typecheck/tests and the Go test suite.

The compiler and runner are separate trust boundaries. Review artifact digests, pinned toolchains, grants, retries, leases, and idempotency records before enabling production functions. A failed or revoked invocation must produce no unauthorized write.

## Phase 05 completed

Administrators can promote or roll back plugin versions, preview/apply dependency disables, drain/resume/stop plugin lifecycle, and uninstall plugins through the dashboard contract. Workflow actions require current project grants; deployment plugins require explicit `deployment.execute` authorization. Publish credentials and Docker/host sockets remain host-owned. Review the required test/build/preflight/deploy/post-deploy sequence and audit lifecycle operations before production use. Local checks are Docker-free; use CI for Compose checks.
