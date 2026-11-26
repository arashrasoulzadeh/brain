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
