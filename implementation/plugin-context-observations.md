# Plugin project context and observation migration

Status: in progress. Do not mark an item complete until its focused and full CI gates pass.

## Runtime
- [ ] Verify shared worker injection for sdk.dispatch, sdk.request, sdk.context, sdk.response, handler and mapped exports; remove remaining legacy payload-only production exports.
- [ ] Resolve the existing Project model from trusted host state (including active status); expose a read-only safe projection, never caller-provided identity.
- [ ] Require context for every project-bound production function and propagate through functions, workflow, scheduled jobs and external workers. Pure helpers and request serializers are not authorization boundaries.
- [ ] Add host-validated manifest event and observation declarations, explicit subscriber grants and event payload contracts.
- [ ] Add SDK plugins.observe.fire(event, payload) with implicit publisher and project; stage bounded JSON events, dispatch only after successful host commit, never from failed invocations.
- [ ] Resolve subscribers from enabled installations in the same project, recheck readiness, permissions and lifecycle, invoke through shared registry; isolate failure, bound recursion and fanout, audit without sensitive payloads.
- [ ] Support YAML observers via existing manifest function dispatch, without importing publisher code. Define delivery and retry semantics.

## All plugin migration inventory
Each row requires context and relevant sanitized domain events, focused tests and manifest updates. No artificial events for pure reads; do not emit secrets or raw authentication payloads.
- [ ] analytics — capture context, subscribe to rooms.connected; avoid observation loops.
- [ ] arena — match started, finished; verify actual route exports.
- [ ] auth — account created, session established; exclude credentials, OTPs, tokens.
- [ ] dashboard — committed lifecycle changes; separate privileged system operations from project functions; verify actual route exports.
- [ ] files — uploaded, published; exclude file contents.
- [ ] functions — published, invoked; include scheduled and workflow paths, not source or arbitrary inputs.
- [ ] invites — created, accepted, cancelled; exclude invite codes; verify actual route exports.
- [ ] matchmaking — queued, assigned, cancelled.
- [ ] rooms — created, connected, reconnected, result committed; actual membership/connection semantics.
- [ ] subscriptions — opened, closed; exclude ticket secrets and private content; verify actual route exports.
- [ ] tables — record written; avoid raw record values unless explicitly authorized.
- [ ] template — context-first example and observer example.

## Exit gates
- [ ] Two-project isolation, forged identity, failed commit, disabled/revoked subscriber, recursive events, bounded payload and fanout, manifest validation, no observer dependency on publisher.
- [ ] Run npm run lint:ts, npm run typecheck, npm run test:ts, go test ./... and sandboxed worker tests.
- [ ] Update developer/operator documentation and changelog; no completion claim until verification.
