# Dynamic runtime migration

Status: incremental; not production-ready until full CI passes.

- [ ] Finish mandatory trusted Project context for all plugin, workflow, scheduled and external invocations.
- [ ] Unify invocation contracts across routes, observers, workflows and jobs.
- [ ] Introduce a shared capability-scoped helper registry. First increment: pure worker pipeline helpers.
- [ ] Extend bounded declarative execution, with host authorization for effects and no arbitrary evaluation.
- [ ] Finish observation migration and test post-commit isolation, grants, recursion and fanout.
- [ ] Introduce declared models, explicit versioned migrations and authorized project CRUD.
- [ ] Introduce transactional state machines with concurrency guards.
- [ ] Generate dashboard navigation, forms and pages from validated manifests.
- [ ] Resolve optional dependencies and validate project-specific configuration.
- [ ] Generate SDK types, documentation and contract tests; migrate all plugins.
- [ ] Run full Go, TypeScript, lint and sandbox gates.

First increment: sdk.pipeline supports up to 32 ordered pure steps, explicit inputs and saved outputs, and trusted project/plugin identity references. The example plugin demonstrates it. No storage, network, cross-plugin calls or arbitrary code are exposed by this increment.

Second increment: fixes worker initialization order; adds typed pipeline SDK declarations, integer/boolean helpers, staged observation helper, pure model validation and state transition primitives. Model persistence, YAML schema integration, transactional state updates and generated dashboard remain open.

Third increment: host-only model routes with explicit schema versions, project/plugin/model isolation, bounded list pages, create/read/update/delete and declarative state transitions backed by SQLite compare-and-swap revisions. Model versions fail closed on mismatched stored records; migration tooling and PostgreSQL support remain open. Template example and registry integration test included.

Fourth increment: production registry active-project resolver and generated admin model forms using declared schema and granted routes. The model dashboard is available on plugin settings pages for enabled installations. Go formatting corrections included; full CI still required.

Fifth increment: per-invocation host snapshot carries verified active project, invocation ID, installation config and optional dependency availability; optional dependencies do not block installation. JSON integer config works through persistence. Dependency availability is a lifecycle snapshot, not authorization for a subsequent call.

Sixth increment: existing auth, rooms, files, matchmaking and tables route handlers stage sanitized domain observations. Missing arena, invites, subscription and dashboard execution paths and user-function integration still require migration. Observation delivery remains best effort after the host commit boundary.
