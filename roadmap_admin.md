# Admin roadmap

Admin-panel and operator documentation work after backend release gates pass.

## Admin operations

- [x] The dashboard covers routine project, plugin, database, function, workflow, file, and live-session operations. Verified by the dashboard plugin contract tests and the project/function/workflow/file/table/live-session page contributions.
- [ ] Verify operator sessions remain separate from end-user sessions.
- [ ] Verify destructive actions require explicit confirmation and audit records.
- [ ] Verify plugin lifecycle controls expose readiness, grants, drain, rollback, and revocation state.
- [ ] Publish the administrator deployment runbook.
- [ ] Publish the administrator restore runbook.
- [ ] Document configuration, limitations, and recovery procedures.

## Admin release gate

- [ ] Verify admin operations against the packaged backend.
- [ ] Verify cross-project isolation in the admin panel.
- [ ] Verify permission revocation in the admin panel.
