# Admin roadmap

Admin-panel and operator documentation work after backend release gates pass.

## Admin operations

- [x] The dashboard covers routine project, plugin, database, function, workflow, file, and live-session operations. Verified by the dashboard plugin contract tests and the project/function/workflow/file/table/live-session page contributions.
- [x] Verify operator sessions remain separate from end-user sessions. Covered by cross-use rejection and revocation tests in `tests/auth/auth.test.ts`.
- [x] Verify destructive actions require explicit confirmation and audit records. The admin HTTP boundary now fails closed without `confirmed: true`, and lifecycle tests verify audit records.
- [x] Verify plugin lifecycle controls expose readiness, grants, drain, rollback, and revocation state. Covered by dashboard manifest routes, lifecycle HTTP routes, and lifecycle service tests.
- [x] Publish the administrator deployment runbook. Available at `docs/deployment-runbook.md`.
- [x] Publish the administrator restore runbook. Available at `docs/restore-runbook.md`.
- [x] Document configuration, limitations, and recovery procedures in `docs/admins.md`, `docs/deployment-runbook.md`, `docs/restore-runbook.md`, and `docs/incompatible-migration-recovery.md`.

## Admin release gate

- [x] Verify admin operations against the packaged backend. The production command composition test boots the configured server, logs in through `/admin/session`, and exercises bearer-authenticated `/admin/actions`; Compose CI remains the deployment-level check.
- [x] Verify cross-project isolation in the admin panel. Covered by dashboard host-handler, view, auth, table-policy, and file-access isolation tests.
- [x] Verify permission revocation in the admin panel. Covered by broker credential revocation tests, plugin lifecycle revocation tests, and audited admin lifecycle tests.

## Manifest-driven admin migration

- [x] Versioned admin metadata and generic action discovery.
- [x] Typed forms and explicit server-side action confirmation and payload validation.
- [x] Project/version/grant-aware page contributions and cross-plugin extension links.
- [x] Declarative result components with filtering and client-side pagination.
- [x] Authenticated observation-driven invalidation and read-only refresh (bounded polling).
- [x] Schema-generated configuration forms.
- [x] Migrate database, files, workflows and live-session consoles away from hard-coded plugin dispatch.
- [ ] Dedicated resource list APIs and richer embedded extension slots.
- [ ] Dedicated admin operations for plugins without implemented runtime handlers.
- [ ] End-to-end and deployment validation of the new paths.
